package gohbem

import (
	"math"
	"sort"
	"sync"
)

// rankArenaPool reuses the 4096-entry rank scratch buffer between calls.
// calculateRanksCompact returns the buffer to the caller; callers that don't
// need it long-term should release it via releaseRankArena. Cache stored
// values only need TopValue, so the buffer is short-lived in cached paths.
var rankArenaPool = sync.Pool{
	New: func() any { return new([4096]PvPRankingStats) },
}

func releaseRankArena(a *[4096]PvPRankingStats) {
	*a = [4096]PvPRankingStats{}
	rankArenaPool.Put(a)
}

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

// RankingComparatorDefault ranks everything by stat product descending then by attack descending.
// This is the default behavior, since in general, a higher stat product is usually preferable;
// and in case of tying stat products, higher attack means that you would be more likely to win CMP ties.
func RankingComparatorDefault(a, b *PvPRankingStats) int {
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

// RankingComparatorPreferHigherCp in addition to the default rules, also compare by CP descending in the end.
// While ties are not meaningfully different most of the time,
// the rationale here is that a higher CP looks more intimidating.
func RankingComparatorPreferHigherCp(a, b *PvPRankingStats) int {
	if d := RankingComparatorDefault(a, b); d != 0 {
		return d
	}
	switch {
	case b.Cp > a.Cp:
		return 1
	case b.Cp < a.Cp:
		return -1
	}
	return 0
}

// RankingComparatorPreferLowerCp in addition to the default rules, also compare by CP ascending in the end.
// While ties are not meaningfully different most of the time,
// the rationale here is that you can flex beating your opponent using one with a lower CP.
func RankingComparatorPreferLowerCp(a, b *PvPRankingStats) int {
	if d := RankingComparatorDefault(a, b); d != 0 {
		return d
	}
	switch {
	case a.Cp > b.Cp:
		return 1
	case a.Cp < b.Cp:
		return -1
	}
	return 0
}

// compactRankSorter is a sort.Interface adapter over a fixed-size rank arena.
// Using sort.Sort over this avoids per-comparison closure escapes that
// slices.SortFunc(..., func(a, b T) int) triggers when the comparator takes
// pointers to the value parameters.
type compactRankSorter struct {
	ranks      *[4096]PvPRankingStats
	count      int
	comparator RankingComparator
}

func (sorter compactRankSorter) Len() int { return sorter.count }

func (sorter compactRankSorter) Less(i, j int) bool {
	d := sorter.comparator(&sorter.ranks[i], &sorter.ranks[j])
	return d < 0 || d == 0 && sorter.ranks[i].Index < sorter.ranks[j].Index
}

func (sorter compactRankSorter) Swap(i, j int) {
	sorter.ranks[i], sorter.ranks[j] = sorter.ranks[j], sorter.ranks[i]
}

// calculateRanksCompact is optimized (for cache) core method used to calculate PvP ranks for provided Pokemon data.
// The returned [4096]PvPRankingStats array is drawn from a pool; callers that
// don't retain it should pass it back via releaseRankArena.
func calculateRanksCompact(stats *PokemonStats, cpCap int, lvCap float64, comparator RankingComparator, ivFloor int) (*[4096]int16, *[4096]PvPRankingStats) {
	combinations := new([4096]int16)
	ranks := rankArenaPool.Get().(*[4096]PvPRankingStats)
	sorter := compactRankSorter{ranks: ranks, comparator: comparator}

	for a := ivFloor; a <= 15; a++ {
		for d := ivFloor; d <= 15; d++ {
			for s := ivFloor; s <= 15; s++ {
				if calculatePvPStat(&ranks[sorter.count], stats, a, d, s, cpCap, lvCap, 1) == nil {
					ranks[sorter.count].Index = (a*16+d)*16 + s
					sorter.count++
				}
			}
		}
	}

	sort.Sort(sorter)

	for i, j := 0, 0; i < sorter.count; i++ {
		entry := &ranks[i]
		if comparator(&ranks[j], entry) < 0 {
			j = i
		}
		combinations[entry.Index] = int16(j + 1)
	}
	return combinations, ranks
}
