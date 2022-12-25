package ohbemgo

import (
	"errors"
	"math"
	"sort"
)

func calculateCpMultiplier(level float64) float64 {
	var intLevel = int(level * 10)
	if level <= maxLevel {
		return cpMultipliers[intLevel]
	}
	var baseLevel = math.Floor(level)
	var baseCpm = 0.5903 + baseLevel*0.005
	if baseLevel == level {
		return baseCpm
	}
	var nextCpm = 0.5903 + (baseLevel+1)*0.005
	return math.Sqrt((baseCpm*baseCpm + nextCpm*nextCpm) / 2)
}

func calculateHp(stats PokemonStats, stamina int, level float64) int {
	var staminaSum = float64(stats.Stamina + stamina)
	if staminaSum <= 10 {
		staminaSum = 10
	}
	return int(math.Floor(staminaSum * calculateCpMultiplier(level)))
}

func calculateStatProduct(stats PokemonStats, attack int, defense int, stamina int, level float64) float64 {
	var multiplier = calculateCpMultiplier(level)
	var hp = math.Floor(float64(stamina+stats.Stamina) * multiplier)
	if hp < 10 {
		hp = 10.0
	}
	return float64(attack+stats.Attack) * multiplier * float64(defense+stats.Defense) * multiplier * hp
}

func calculateCp(stats PokemonStats, attack int, defense int, stamina int, level float64) int {
	var multiplier = calculateCpMultiplier(level)

	if multiplier == 0 {
		return 0
	}

	var a = float64(stats.Attack + attack)
	var d = float64(stats.Defense + defense)
	var s = float64(stats.Stamina + stamina)

	var cp = int(math.Floor(multiplier * multiplier * a * math.Sqrt(d*s) / 10))
	if cp < 10 {
		return 10
	}
	return cp
}

func calculatePvPStat(stats PokemonStats, attack int, defense int, stamina int, cap int, lvCap float64, minLevel float64) (Ranking, error) {
	var mid float64
	var cp int

	var bestCP = calculateCp(stats, attack, defense, stamina, minLevel)
	if bestCP > cap {
		return Ranking{}, errors.New("bestCP > cap")
	}
	var lowest, highest = minLevel, lvCap
	for mid = math.Ceil(lowest+highest) / 2; lowest < highest; mid = math.Ceil(lowest+highest) / 2 {
		cp = calculateCp(stats, attack, defense, stamina, mid)
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

func calculateRanks(stats PokemonStats, cpCap int, lvCap float64) ([16][16][16]Ranking, [4096]Ranking) {
	var combinations [16][16][16]Ranking
	var sortedRanks [4096]Ranking
	var c uint16

	for a := 0; a <= 15; a++ {
		for d := 0; d <= 15; d++ {
			for s := 0; s <= 15; s++ {
				var currentStat, err = calculatePvPStat(stats, a, d, s, cpCap, lvCap, 1)
				if err != nil {
					continue
				}
				combinations[a][d][s] = currentStat
				sortedRanks[c] = currentStat
				c++
			}
		}
	}

	sort.Slice(sortedRanks[:], func(i, j int) bool {
		return sortedRanks[i].Value > sortedRanks[j].Value
	})

	var best = sortedRanks[0].Value
	var i, j int16
	for i, j = 0, 0; i < int16(len(sortedRanks)); i++ {
		entry := &sortedRanks[i]
		combinationsEntry := &combinations[entry.Attack][entry.Defense][entry.Stamina]

		var percentage = roundFloat(entry.Value/best, 5)

		entry.Percentage = percentage
		combinationsEntry.Percentage = percentage

		if entry.Value < sortedRanks[j].Value {
			j = i
		}
		var rank = j + 1

		entry.Rank = rank
		combinationsEntry.Rank = rank
	}
	return combinations, sortedRanks
}

func calculateRanksCompact(stats PokemonStats, cpCap int, lvCap float64, ivFloor int) ([4096]int16, [4096]Ranking) {
	var combinations [4096]int16
	var sortedRanks [4096]Ranking

	for a := ivFloor; a <= 15; a++ {
		for d := ivFloor; d <= 15; d++ {
			for s := ivFloor; s <= 15; s++ {
				var entry, err = calculatePvPStat(stats, a, d, s, cpCap, lvCap, 1)
				if err != nil {
					continue
				}
				entry.Index = (a*16+d)*16 + s
				sortedRanks[entry.Index] = entry
			}
		}
	}

	sort.Slice(sortedRanks[:], func(i, j int) bool {
		return sortedRanks[i].Value > sortedRanks[j].Value || (sortedRanks[i].Value == sortedRanks[j].Value && sortedRanks[i].Index < sortedRanks[j].Index)
	})

	for i, j := 0, 0; i < len(sortedRanks); i++ {
		entry := &sortedRanks[i]
		// skip empty entries
		if entry.Value == 0 {
			continue
		}
		if entry.Value < sortedRanks[j].Value {
			j = i
		}
		combinations[entry.Index] = int16(j + 1)
	}
	return combinations, sortedRanks
}
