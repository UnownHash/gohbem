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

// battleFormMaxLevel is where the CP multiplier stops growing for battle-only form scaling: a 15/15/13 Minior's
// Meteor stopped at its level 55 CP of 1837 in Ultra League however far it was powered up (pvpoke/pvpoke#404).
const battleFormMaxLevel = 55

// battleFormUncappedLevel is the cap level of a form that is still under the CP cap at battleFormMaxLevel.
const battleFormUncappedLevel = 80

// calculateUnflooredCp is calculateCp before flooring and the minimum of 10.
func calculateUnflooredCp(stats *PokemonStats, attack, defense, stamina int, level float64) float64 {
	multiplier := calculateCpMultiplier(level)
	return multiplier * multiplier * float64(stats.Attack+attack) *
		math.Sqrt(float64((stats.Defense+defense)*(stats.Stamina+stamina))) / 10
}

// findBattleFormCapLevel returns the highest half-level (from 1) whose CP before flooring fits the cap,
// or battleFormUncappedLevel if the form fits at battleFormMaxLevel.
//
// Unfloored on purpose, as pvpoke/pvpoke#404 defines it. A Core showing exactly 1500 CP is over the cap here but not
// when it enters, so its Meteor lands above Meteor's own cap level. In-game readings in jgmcelwain/go-iv#48 show
// exactly that: 12/1/7 and 12/14/12 Cores at 1500 CP fought as 1502 and 1508 CP Meteors. jgmcelwain/go-iv#50 floors
// instead and expects 1493 and 1499, but cites no reading for those.
func findBattleFormCapLevel(stats *PokemonStats, attack, defense, stamina, cap int) float64 {
	if calculateUnflooredCp(stats, attack, defense, stamina, battleFormMaxLevel) <= float64(cap) {
		return battleFormUncappedLevel
	}
	lowest, highest := 1.0, float64(battleFormMaxLevel)
	for mid := math.Ceil(lowest+highest) / 2; lowest < highest; mid = math.Ceil(lowest+highest) / 2 {
		if calculateUnflooredCp(stats, attack, defense, stamina, mid) <= float64(cap) {
			lowest = mid
		} else {
			highest = mid - 0.5
		}
	}
	return lowest
}

// calculateBattleFormLevel returns the level a battle-only form fights at when its stored form is at storedLevel,
// using the formula fitted to player readings by u/eli5questions (pvpoke/pvpoke#404):
// 1 + floor(2 * (M - 1) * L / C) / 2, where M and C are the battle and stored form cap levels.
func calculateBattleFormLevel(battle, stored *PokemonStats, attack, defense, stamina int, storedLevel float64, cap int) float64 {
	battleCapLevel := findBattleFormCapLevel(battle, attack, defense, stamina, cap)
	storedCapLevel := findBattleFormCapLevel(stored, attack, defense, stamina, cap)
	halfLevels := math.Floor(2 * (battleCapLevel - 1) * storedLevel / storedCapLevel)
	return math.Min(1+halfLevels/2, battleFormMaxLevel)
}

// calculateBattleFormStats fills out with how a battle-only form fights when its stored form is at storedLevel:
// Attack and Defense come from the battle form at its scaled level, and so do Level and Cp, while HP stays the
// stored form's. HP: a 9/3/1 Minior's shell broke after 52 one-damage Tackles, half its Core HP of 104 (go-iv#48).
func calculateBattleFormStats(out *PvPRankingStats, battle, stored *PokemonStats, attack, defense, stamina int, storedLevel float64, cap int) {
	level := calculateBattleFormLevel(battle, stored, attack, defense, stamina, storedLevel, cap)
	multiplier := calculateCpMultiplier(level)
	out.Attack = float64(attack+battle.Attack) * multiplier
	hp := math.Floor(float64(stamina+stored.Stamina) * calculateCpMultiplier(storedLevel))
	if hp < 10 {
		hp = 10.0
	}
	out.Value = out.Attack * float64(defense+battle.Defense) * multiplier * hp
	out.Level = level
	out.Cp = calculateCp(battle, attack, defense, stamina, level)
}

// calculateFormPvPStat is calculatePvPStat for a form that is battle-only when stored is not nil: the stored form's
// level is maximized under the caps as usual, and stats is the form it changes into in battle.
func calculateFormPvPStat(out *PvPRankingStats, stats, stored *PokemonStats, attack, defense, stamina, cap int, lvCap, minLevel float64) error {
	if stored == nil {
		return calculatePvPStat(out, stats, attack, defense, stamina, cap, lvCap, minLevel)
	}
	if err := calculatePvPStat(out, stored, attack, defense, stamina, cap, lvCap, minLevel); err != nil {
		return err
	}
	calculateBattleFormStats(out, stats, stored, attack, defense, stamina, out.Level, cap)
	return nil
}

// hundosFitUnderCap reports whether every IV combination fits under the CP cap at the level cap,
// so the level cap ranks the same as having no level cap at all.
func hundosFitUnderCap(stats, stored *PokemonStats, cpCap int, lvCap float64) bool {
	if stored == nil {
		return calculateCp(stats, 15, 15, 15, lvCap) <= cpCap
	}
	// Battle-only forms keep scaling with the CP cap until both forms fit at battleFormMaxLevel.
	return calculateCp(stored, 15, 15, 15, lvCap) <= cpCap &&
		findBattleFormCapLevel(stored, 15, 15, 15, cpCap) == battleFormUncappedLevel &&
		findBattleFormCapLevel(stats, 15, 15, 15, cpCap) == battleFormUncappedLevel
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
	d := RankingComparatorDefault(a, b)
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

// RankingComparatorPreferLowerCp in addition to the default rules, also compare by CP ascending in the end.
// While ties are not meaningfully different most of the time,
// the rationale here is that you can flex beating your opponent using one with a lower CP.
func RankingComparatorPreferLowerCp(a, b *PvPRankingStats) int {
	d := RankingComparatorDefault(a, b)
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

type compactRankSorter struct {
	ranks      *[4096]PvPRankingStats
	count      int
	comparator RankingComparator
}

func (sorter compactRankSorter) Len() int {
	return sorter.count
}

func (sorter compactRankSorter) Less(i, j int) bool {
	d := sorter.comparator(&sorter.ranks[i], &sorter.ranks[j])
	return d < 0 || d == 0 && sorter.ranks[i].Index < sorter.ranks[j].Index
}

func (sorter compactRankSorter) Swap(i, j int) {
	sorter.ranks[i], sorter.ranks[j] = sorter.ranks[j], sorter.ranks[i]
}

// calculateRanksCompact is optimized (for cache) core method used to calculate PvP ranks for provided Pokemon data.
// stored is the stored form when stats is a battle-only form, or nil otherwise.
func calculateRanksCompact(stats, stored *PokemonStats, cpCap int, lvCap float64, comparator RankingComparator, ivFloor int) (*[4096]int16, *[4096]PvPRankingStats) {
	combinations := new([4096]int16)
	sorter := compactRankSorter{ranks: new([4096]PvPRankingStats), comparator: comparator}

	for a := ivFloor; a <= 15; a++ {
		for d := ivFloor; d <= 15; d++ {
			for s := ivFloor; s <= 15; s++ {
				if calculateFormPvPStat(&sorter.ranks[sorter.count], stats, stored, a, d, s, cpCap, lvCap, 1) == nil {
					sorter.ranks[sorter.count].Index = (a*16+d)*16 + s
					sorter.count++
				}
			}
		}
	}

	sort.Sort(sorter)

	for i, j := 0, 0; i < sorter.count; i++ {
		entry := &sorter.ranks[i]
		if comparator(&sorter.ranks[j], entry) < 0 {
			j = i
		}
		combinations[entry.Index] = int16(j + 1)
	}
	return combinations, sorter.ranks
}
