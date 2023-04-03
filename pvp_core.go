package gohbem

import (
	"math"
	"sort"
)

// calculateCpMultiplier is used to calculate CP multiplier for provided level. It's using precalculated values from cpm.go file.
func calculateCpMultiplier(level float64) float64 {
	intLevel := int(level * 2)
	if intLevel <= 55*2 {
		return cpMultipliers[intLevel-2]
	}
	baseLevel := intLevel / 2
	baseCpm := float64(float32(0.5903 + float64(baseLevel)*0.005))
	if baseLevel+baseLevel == intLevel {
		return baseCpm
	}
	nextCpm := float64(float32(0.5903 + float64(baseLevel+1)*0.005))
	return math.Sqrt((baseCpm*baseCpm + nextCpm*nextCpm) / 2)
}

// calculateHp is used to calculate Pokemon HP.
func calculateHp(stats *PokemonStats, stamina int, level float64) int {
	hp := int(float64(stats.Stamina+stamina) * calculateCpMultiplier(level))
	if hp <= 10 {
		return 10
	}
	return hp
}

// calculateCp is used to calculate CP for provided Pokemon data.
func calculateCp(stats *PokemonStats, attack, defense, stamina int, level float64) int {
	multiplier := calculateCpMultiplier(level)

	cp := int(multiplier * multiplier * float64(stats.Attack+attack) *
		math.Sqrt(float64((stats.Defense+defense)*(stats.Stamina+stamina))) / 10)
	if cp < 10 {
		return 10
	}
	return cp
}

// calculatePvPStat is core method used to calculate PvP stats for provided Pokemon data.
func calculatePvPStat(out *PvPRankingStats, stats *PokemonStats, attack, defense, stamina, cap int, lvCap, minLevel float64) error {
	bestCP := calculateCp(stats, attack, defense, stamina, minLevel)

	if bestCP > cap {
		return ErrPvpStatBestCp
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

	multiplier := calculateCpMultiplier(lowest)
	out.Attack = float64(attack+stats.Attack) * multiplier
	hp := math.Floor(float64(stamina+stats.Stamina) * multiplier)
	if hp < 10 {
		hp = 10.0
	}
	out.Value = out.Attack * float64(defense+stats.Defense) * multiplier * hp
	out.Level = lowest
	out.Cp = bestCP
	return nil
}

// calculateRanks is core method used to calculate PvP ranks for provided Pokemon data.
/*
func calculateRanks(stats *PokemonStats, cpCap int, lvCap float64, comparator RankingComparator) (*[16][16][16]Ranking, *[4096]*Ranking) {
	combinations := new([16][16][16]Ranking)
	sortedRanks := new([4096]*Ranking)
	var c uint16

	for a := 0; a <= 15; a++ {
		for d := 0; d <= 15; d++ {
			for s := 0; s <= 15; s++ {
				currentStat, err := calculatePvPStat(stats, a, d, s, cpCap, lvCap, 1)
				if err != nil {
					continue
				}
				combinations[a][d][s] = currentStat
				sortedRanks[c] = &currentStat
				c++
			}
		}
	}

	sort.SliceStable(sortedRanks, func(i, j int) bool {
		return comparator(sortedRanks[i], sortedRanks[j]) < 0
	})

	best := sortedRanks[0].Value
	var i, j int16
	for i, j = 0, 0; i < int16(len(sortedRanks)); i++ {
		entry := sortedRanks[i]
		percentage := roundFloat(entry.Value/best, 5)
		entry.Percentage = percentage
		if entry.Value < sortedRanks[j].Value {
			j = i
		}
		rank := j + 1
		entry.Rank = rank
	}
	return combinations, sortedRanks
}
*/

func RankingComparator_Default(a, b *PvPRankingStats) int {
	d := b.Value - a.Value
	if d > 0 {
		return 1
	}
	if d < 0 {
		return -1
	}
	d = b.Attack - a.Attack
	if d > 0 {
		return 1
	}
	if d < 0 {
		return -1
	}
	return 0
}
func RankingComparator_PreferHigherCp(a, b *PvPRankingStats) int {
	d := RankingComparator_Default(a, b)
	if d > 0 {
		return 1
	}
	if d < 0 {
		return -1
	}
	d = b.Cp - a.Cp
	if d > 0 {
		return 1
	}
	if d < 0 {
		return -1
	}
	return 0
}
func RankingComparator_PreferLowerCp(a, b *PvPRankingStats) int {
	d := RankingComparator_Default(a, b)
	if d > 0 {
		return 1
	}
	if d < 0 {
		return -1
	}
	d = a.Cp - b.Cp
	if d > 0 {
		return 1
	}
	if d < 0 {
		return -1
	}
	return 0
}

// calculateRanksCompact is optimized (for cache) core method used to calculate PvP ranks for provided Pokemon data.
func calculateRanksCompact(stats *PokemonStats, cpCap int, lvCap float64, comparator RankingComparator, ivFloor int) (*[4096]int16, *[4096]PvPRankingStats) {
	combinations := new([4096]int16)
	sortedRanks := new([4096]PvPRankingStats)

	count := 0
	for a := ivFloor; a <= 15; a++ {
		for d := ivFloor; d <= 15; d++ {
			for s := ivFloor; s <= 15; s++ {
				if calculatePvPStat(&sortedRanks[count], stats, a, d, s, cpCap, lvCap, 1) == nil {
					sortedRanks[count].Index = (a*16+d)*16 + s
					count++
				}
			}
		}
	}

	// use Slice over SliceStable for performance since we have index available to us
	sort.Slice(sortedRanks[:count], func(i, j int) bool {
		d := comparator(&sortedRanks[i], &sortedRanks[j])
		return d < 0 || d == 0 && sortedRanks[i].Index < sortedRanks[j].Index
	})

	for i, j := 0, 0; i < len(sortedRanks); i++ {
		entry := &sortedRanks[i]
		if comparator(&sortedRanks[j], entry) < 0 {
			j = i
		}
		combinations[entry.Index] = int16(j + 1)
	}
	return combinations, sortedRanks
}
