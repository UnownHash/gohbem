package gohbem

import (
	"math"
	"sort"
)

// calculateCpMultiplier is used to calculate CP multiplier for provided level. It's using precalculated values from cpm.go file.
func calculateCpMultiplier(level float64) float64 {
	intLevel := int(level * 10)
	if level <= MaxLevel {
		return cpMultipliers[intLevel]
	}
	baseLevel := math.Floor(level)
	baseCpm := 0.5903 + baseLevel*0.005
	if baseLevel == level {
		return baseCpm
	}
	nextCpm := 0.5903 + (baseLevel+1)*0.005
	return math.Sqrt((baseCpm*baseCpm + nextCpm*nextCpm) / 2)
}

// calculateHp is used to calculate Pokemon HP.
func calculateHp(stats PokemonStats, stamina int, level float64) int {
	staminaSum := float64(stats.Stamina + stamina)
	if staminaSum <= 10 {
		staminaSum = 10
	}
	return int(math.Floor(staminaSum * calculateCpMultiplier(level)))
}

// calculateStatProduct is used to calculate Pokemon stat product.
func calculateStatProduct(stats PokemonStats, attack int, defense int, stamina int, level float64) float64 {
	multiplier := calculateCpMultiplier(level)
	hp := math.Floor(float64(stamina+stats.Stamina) * multiplier)
	if hp < 10 {
		hp = 10.0
	}
	return float64(attack+stats.Attack) * multiplier * float64(defense+stats.Defense) * multiplier * hp
}

// calculateStatProduct is used to calculate CP for provided Pokemon data.
func calculateCp(stats PokemonStats, attack int, defense int, stamina int, level float64) int {
	multiplier := calculateCpMultiplier(level)

	if multiplier == 0 {
		return 0
	}

	a := float64(stats.Attack + attack)
	d := float64(stats.Defense + defense)
	s := float64(stats.Stamina + stamina)

	cp := int(math.Floor(multiplier * multiplier * a * math.Sqrt(d*s) / 10))
	if cp < 10 {
		return 10
	}
	return cp
}

// calculatePvPStat is core method used to calculate PvP stats for provided Pokemon data.
func calculatePvPStat(stats PokemonStats, attack int, defense int, stamina int, cap int, lvCap float64, minLevel float64) (Ranking, error) {
	bestCP := calculateCp(stats, attack, defense, stamina, minLevel)

	if bestCP > cap {
		return Ranking{}, ErrPvpStatBestCp
	}
	lowest, highest := minLevel, lvCap
	for mid := math.Ceil(lowest+highest) / 2; lowest < highest; mid = math.Ceil(lowest+highest) / 2 {
		cp := calculateCp(stats, attack, defense, stamina, mid)
		if cp <= cap {
			lowest = mid
			bestCP = cp
		} else {
			highest = mid - 0.5
		}
	}

	return Ranking{
		Value:   calculateStatProduct(stats, attack, defense, stamina, lowest),
		Attack:  attack,
		Defense: defense,
		Stamina: stamina,
		Level:   lowest,
		Cp:      bestCP,
	}, nil
}

// calculateRanks is core method used to calculate PvP ranks for provided Pokemon data.
func calculateRanks(stats PokemonStats, cpCap int, lvCap float64) ([16][16][16]Ranking, [4096]Ranking) {
	var combinations [16][16][16]Ranking
	var sortedRanks [4096]Ranking
	var c uint16

	for a := 0; a <= 15; a++ {
		for d := 0; d <= 15; d++ {
			for s := 0; s <= 15; s++ {
				currentStat, err := calculatePvPStat(stats, a, d, s, cpCap, lvCap, 1)
				if err != nil {
					continue
				}
				combinations[a][d][s] = currentStat
				sortedRanks[c] = currentStat
				c++
			}
		}
	}

	sort.Sort(rankingSortable(sortedRanks[:]))

	best := sortedRanks[0].Value
	var i, j int16
	for i, j = 0, 0; i < int16(len(sortedRanks)); i++ {
		entry := &sortedRanks[i]
		combinationsEntry := &combinations[entry.Attack][entry.Defense][entry.Stamina]

		percentage := roundFloat(entry.Value/best, 5)

		entry.Percentage = percentage
		combinationsEntry.Percentage = percentage

		if entry.Value < sortedRanks[j].Value {
			j = i
		}
		rank := j + 1

		entry.Rank = rank
		combinationsEntry.Rank = rank
	}
	return combinations, sortedRanks
}

// calculateRanksCompact is optimized (for cache) core method used to calculate PvP ranks for provided Pokemon data.
func calculateRanksCompact(stats PokemonStats, cpCap int, lvCap float64, ivFloor int) ([4096]int16, [4096]Ranking) {
	var combinations [4096]int16
	var sortedRanks [4096]Ranking

	for a := ivFloor; a <= 15; a++ {
		for d := ivFloor; d <= 15; d++ {
			for s := ivFloor; s <= 15; s++ {
				entry, _ := calculatePvPStat(stats, a, d, s, cpCap, lvCap, 1)
				entry.Index = (a*16+d)*16 + s
				sortedRanks[entry.Index] = entry
			}
		}
	}

	sort.Sort(rankingSortableIndexed(sortedRanks[:]))

	for i, j := 0, 0; i < len(sortedRanks); i++ {
		entry := &sortedRanks[i]
		if entry.Value < sortedRanks[j].Value {
			j = i
		}
		combinations[entry.Index] = int16(j + 1)
	}
	return combinations, sortedRanks
}
