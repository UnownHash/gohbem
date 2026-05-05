package gohbem

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"slices"
	"sync"
	"time"
)

// MaxLevel handled by gohbem.
const MaxLevel = 100

// VERSION of gohbem, follows Semantic Versioning. (http://semver.org/)
const VERSION = "0.13.0"

// FetchPokemonData Fetch remote MasterFile and keep it in memory.
func (o *Ohbem) FetchPokemonData() error {
	data, err := fetchMasterFile(o.MasterFileURL)
	if err != nil {
		return err
	}
	o.mu.Lock()
	o.PokemonData = data
	if o.RankingComparator == nil {
		o.RankingComparator = RankingComparatorDefault
	}
	o.mu.Unlock()
	o.initialized.Store(true)
	o.ClearCache()
	return nil
}

// LoadPokemonData Load MasterFile from provided filePath and keep it in memory.
func (o *Ohbem) LoadPokemonData(filePath string) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return ErrMasterFileOpen
	}
	var pd PokemonData
	if err := json.Unmarshal(data, &pd); err != nil {
		return ErrMasterFileUnmarshall
	}
	pd.Initialized = true
	o.mu.Lock()
	o.PokemonData = pd
	if o.RankingComparator == nil {
		o.RankingComparator = RankingComparatorDefault
	}
	o.mu.Unlock()
	o.initialized.Store(true)
	o.ClearCache()
	return nil
}

// SavePokemonData Save MasterFile from memory to provided location.
func (o *Ohbem) SavePokemonData(filePath string) error {
	o.mu.RLock()
	data, err := json.Marshal(o.PokemonData)
	o.mu.RUnlock()
	if err != nil {
		return ErrMasterFileMarshall
	}
	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return ErrMasterFileSave
	}
	return nil
}

// WatchPokemonData Watch for remote MasterFile changes. When new, auto-update and clean cache.
func (o *Ohbem) WatchPokemonData() error {
	o.mu.Lock()
	if o.watcherChan != nil {
		o.mu.Unlock()
		return ErrWatcherStarted
	}
	o.watcherChan = make(chan bool)
	stopCh := o.watcherChan
	o.mu.Unlock()

	o.log("MasterFile Watcher Started")
	interval := o.WatcherInterval
	if interval == 0 {
		interval = 60 * time.Minute
	}

	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-stopCh:
				o.log("MasterFile Watcher Stopped")
				return
			case <-ticker.C:
				o.log("Checking remote MasterFile")
				pokemonData, err := fetchMasterFile(o.MasterFileURL)
				if err != nil {
					o.log("Remote MasterFile fetch failed")
					continue
				}
				newData, mErr := json.Marshal(pokemonData)
				if mErr != nil {
					o.log("Remote MasterFile marshal failed")
					continue
				}
				o.mu.RLock()
				oldData, mErr := json.Marshal(o.PokemonData)
				o.mu.RUnlock()
				if mErr != nil {
					o.log("Current MasterFile marshal failed")
					continue
				}
				if bytes.Equal(newData, oldData) {
					continue
				}
				o.log("New MasterFile found! Updating PokemonData")
				// Save first to disk; only swap in-memory if save succeeded so a crash
				// before swap doesn't leave cache file pointing at the old version while
				// the running process serves the new one.
				if o.MasterFileCachePath != "" {
					tmp := o.MasterFileCachePath + ".tmp"
					if err := os.WriteFile(tmp, newData, 0644); err != nil {
						o.log(fmt.Sprintf("Storing MasterFile cache under %s has failed!", o.MasterFileCachePath))
						continue
					}
					if err := os.Rename(tmp, o.MasterFileCachePath); err != nil {
						o.log(fmt.Sprintf("Renaming MasterFile cache to %s has failed!", o.MasterFileCachePath))
						continue
					}
				}
				o.mu.Lock()
				o.PokemonData = pokemonData
				o.mu.Unlock()
				o.initialized.Store(true)
				o.ClearCache()
			}
		}
	}()
	return nil
}

// StopWatchingPokemonData Stop watching for remote MasterFile changes.
func (o *Ohbem) StopWatchingPokemonData() error {
	o.mu.Lock()
	defer o.mu.Unlock()
	if o.watcherChan == nil {
		return ErrNilChannel
	}
	close(o.watcherChan)
	o.watcherChan = nil
	return nil
}

// ClearCache empties the compact rank cache by atomically swapping in a fresh map.
func (o *Ohbem) ClearCache() {
	if !o.DisableCache {
		o.compactRankCache.Store(&sync.Map{})
		o.log("Cache cleaned")
	}
}

// cacheKey packs (cpCap, attack, defense, stamina) into one int64.
// Stats fit in 16 bits and cpCap in <= 14 bits, so no overflow / collisions.
func cacheKey(cpCap int, stats *PokemonStats) int64 {
	return int64(cpCap)<<48 | int64(stats.Attack)<<32 | int64(stats.Defense)<<16 | int64(stats.Stamina)
}

// loadCache returns the current cache map, allocating one lazily if needed.
func (o *Ohbem) loadCache() *sync.Map {
	m := o.compactRankCache.Load()
	if m == nil {
		fresh := &sync.Map{}
		if o.compactRankCache.CompareAndSwap(nil, fresh) {
			return fresh
		}
		return o.compactRankCache.Load()
	}
	return m
}

// calculateAllRanksCompact Calculate all PvP ranks for a specific base stats with the specified CP cap. Compact version intended to be used with cache.
func (o *Ohbem) calculateAllRanksCompact(stats *PokemonStats, cpCap int) (map[int]compactCacheValue, bool) {
	key := cacheKey(cpCap, stats)

	var cache *sync.Map
	if !o.DisableCache {
		cache = o.loadCache()
		if obj, ok := cache.Load(key); ok {
			return obj.(map[int]compactCacheValue), true
		}
	}

	o.mu.RLock()
	comparator := o.RankingComparator
	o.mu.RUnlock()
	if comparator == nil {
		comparator = RankingComparatorDefault
	}

	filled := false
	maxed := false
	result := make(map[int]compactCacheValue)

	for _, lvCap := range o.LevelCaps {
		lvCapFloat := float64(lvCap)
		if !o.IncludeHundosUnderCap && calculateCp(stats, 15, 15, 15, lvCapFloat) <= cpCap {
			continue
		}

		combinations, sortedRanks := calculateRanksCompact(stats, cpCap, lvCapFloat, comparator, 0)
		result[lvCap] = compactCacheValue{
			Combinations: combinations,
			TopValue:     sortedRanks[0].Value,
		}
		releaseRankArena(sortedRanks)
		filled = true
		if calculateCp(stats, 0, 0, 0, lvCapFloat+0.5) > cpCap {
			maxed = true
			break
		}
	}
	if filled && !maxed {
		combinations, sortedRanks := calculateRanksCompact(stats, cpCap, MaxLevel, comparator, 0)
		result[MaxLevel] = compactCacheValue{
			Combinations: combinations,
			TopValue:     sortedRanks[0].Value,
		}
		releaseRankArena(sortedRanks)
	}
	if !o.DisableCache && filled {
		cache.Store(key, result)
	}
	return result, filled
}

// resolveStats returns the stats, form, pokemon, and existence for a (pokemon, form, evolution) tuple.
// Lookup order for stats: form's TempEvolution, pokemon's TempEvolution, form base, pokemon base.
// Caller must already hold o.mu (read lock) when accessing PokemonData fields.
func resolveStats(pd *PokemonData, pokemonId, form, evolution int) (PokemonStats, Form, Pokemon, bool) {
	mp, ok := pd.Pokemon[pokemonId]
	if !ok {
		return PokemonStats{}, Form{}, Pokemon{}, false
	}
	mf, hasForm := mp.Forms[form]
	if !hasForm || form == 0 {
		mf = Form{
			Attack:                    mp.Attack,
			Defense:                   mp.Defense,
			Stamina:                   mp.Stamina,
			Little:                    mp.Little,
			Evolutions:                mp.Evolutions,
			TempEvolutions:            mp.TempEvolutions,
			CostumeOverrideEvolutions: mp.CostumeOverrideEvolutions,
		}
	}
	var stats PokemonStats
	if evolution != 0 {
		if me, ok := mf.TempEvolutions[evolution]; ok && me.Attack != 0 {
			stats = me
			return stats, mf, mp, true
		}
		if me, ok := mp.TempEvolutions[evolution]; ok && me.Attack != 0 {
			stats = me
			return stats, mf, mp, true
		}
	}
	if mf.Attack != 0 {
		stats = PokemonStats{Attack: mf.Attack, Defense: mf.Defense, Stamina: mf.Stamina}
	} else {
		stats = PokemonStats{Attack: mp.Attack, Defense: mp.Defense, Stamina: mp.Stamina}
	}
	return stats, mf, mp, true
}

// CalculateTopRanks Return ranked list of PVP statistics for a given Pokémon.
func (o *Ohbem) CalculateTopRanks(maxRank int16, pokemonId int, form int, evolution int, ivFloor int) (map[string][]Ranking, error) {
	result := make(map[string][]Ranking)

	if err := safetyCheck(o); err != nil {
		return result, err
	}

	o.mu.RLock()
	stats, masterForm, masterPokemon, ok := resolveStats(&o.PokemonData, pokemonId, form, evolution)
	comparator := o.RankingComparator
	o.mu.RUnlock()
	if !ok || masterPokemon.Attack == 0 {
		return result, nil
	}
	if comparator == nil {
		comparator = RankingComparatorDefault
	}

	type lastEntry struct {
		ranking Ranking
		idx     int
	}

	for leagueName, leagueOptions := range o.Leagues {
		var rankings []Ranking
		var last []lastEntry

		processLevelCap := func(lvCap float64, setOnDup bool) {
			combinations, sortedRanks := calculateRanksCompact(&stats, leagueOptions.Cap, lvCap, comparator, ivFloor)
			defer releaseRankArena(sortedRanks)

			for i := range 4096 {
				stat := &sortedRanks[i]
				if stat.Value == 0 {
					break
				}
				rank := combinations[stat.Index]
				if rank > maxRank {
					if len(last) > i {
						last = last[:i]
					}
					break
				}
				attack := stat.Index >> 8 % 16
				defense := stat.Index >> 4 % 16
				stamina := stat.Index % 16

				var lastStat *Ranking
				if i < len(last) {
					lastStat = &last[i].ranking
				}

				if lastStat != nil && stat.Level == lastStat.Level && rank == lastStat.Rank &&
					attack == lastStat.Attack && defense == lastStat.Defense &&
					stamina == lastStat.Stamina {
					if setOnDup {
						rankings[last[i].idx].Capped = true
					}
				} else if !setOnDup {
					entry := Ranking{
						Rank:       rank,
						Attack:     attack,
						Defense:    defense,
						Stamina:    stamina,
						Cap:        lvCap,
						Value:      math.Floor(stat.Value),
						Level:      stat.Level,
						Cp:         stat.Cp,
						Percentage: roundFloat(stat.Value/sortedRanks[0].Value, 5),
					}
					rankingsIdx := len(rankings)
					rankings = append(rankings, entry)
					for len(last) <= i {
						last = append(last, lastEntry{})
					}
					last[i] = lastEntry{ranking: entry, idx: rankingsIdx}
				}
			}
		}

		if leagueOptions.LittleCupRules && !(masterForm.Little || masterPokemon.Little) {
			continue
		} else if leagueName == "master" {
			for _, lvCap := range o.LevelCaps {
				lvCapFloat := float64(lvCap)
				maxHp := calculateHp(&stats, 15, lvCapFloat)
				for stamina := ivFloor; stamina <= 15; stamina++ {
					if calculateHp(&stats, stamina, lvCapFloat) == maxHp {
						entry := Ranking{
							Attack:     15,
							Defense:    15,
							Stamina:    stamina,
							Level:      lvCapFloat,
							Percentage: 1,
							Rank:       1,
						}
						rankings = append(rankings, entry)
					}
				}
			}
		} else {
			maxed := false
			for _, lvCap := range o.LevelCaps {
				lvCapFloat := float64(lvCap)
				if !o.IncludeHundosUnderCap && calculateCp(&stats, 15, 15, 15, lvCapFloat) <= leagueOptions.Cap {
					continue
				}
				processLevelCap(lvCapFloat, false)
				if calculateCp(&stats, ivFloor, ivFloor, ivFloor, lvCapFloat+0.5) > leagueOptions.Cap {
					maxed = true
					for _, le := range last {
						rankings[le.idx].Capped = true
					}
					break
				}
			}
			if len(rankings) != 0 && !maxed {
				processLevelCap(MaxLevel, true)
			}
		}
		if len(rankings) != 0 {
			result[leagueName] = rankings
		}
	}

	return result, nil
}

// CalculateCp calculates CP for your pokemon. Errors if pokemon cannot be found in master.
func (o *Ohbem) CalculateCp(pokemonId, form, evolution, attack, defense, stamina int, level float64) (int, error) {
	if (attack < 0 || attack > 15) || (defense < 0 || defense > 15) || (stamina < 0 || stamina > 15) || level < 1 {
		return 0, ErrQueryInputOutOfRange
	}
	o.mu.RLock()
	stats, _, _, ok := resolveStats(&o.PokemonData, pokemonId, form, evolution)
	o.mu.RUnlock()
	if !ok {
		return 0, ErrMissingPokemon
	}
	return calculateCp(&stats, attack, defense, stamina, level), nil
}

// maxEvolutionDepth caps recursion in QueryPvPRank to guard against
// malformed (cyclic) MasterFile data. Real evolution chains are <= 3.
const maxEvolutionDepth = 8

// queryPvPRankInternal walks the evolution graph with a depth bound to avoid
// stack overflow on malformed (cyclic) MasterFile data.
func (o *Ohbem) queryPvPRankInternal(depth int, pokemonId, form, costume, gender, attack, defense, stamina int, level float64) (map[string][]PokemonEntry, error) {
	result := make(map[string][]PokemonEntry)

	if depth > maxEvolutionDepth {
		return result, nil
	}

	o.mu.RLock()
	stats, masterForm, masterPokemon, ok := resolveStats(&o.PokemonData, pokemonId, form, 0)
	costumeBlock := false
	if costume != 0 {
		costumeBlock = o.PokemonData.Costumes[costume] && !slices.Contains(masterForm.CostumeOverrideEvolutions, costume)
	}
	o.mu.RUnlock()
	if !ok {
		return result, ErrMissingPokemon
	}

	var baseEntry = PokemonEntry{Pokemon: pokemonId}
	if _, hasForm := masterPokemon.Forms[form]; hasForm && form != 0 {
		baseEntry.Form = form
	}

	pushAllEntries := func(s *PokemonStats, evolution int) {
		for leagueName, leagueOptions := range o.Leagues {
			var entries []PokemonEntry

			if leagueName != "master" {
				if leagueOptions.LittleCupRules && !(masterForm.Little || masterPokemon.Little) {
					continue
				}
				combinationIndex, filled := o.calculateAllRanksCompact(s, leagueOptions.Cap)
				if !filled {
					continue
				}

				processCombinations := func(pCap float64, combinations compactCacheValue) {
					var stat PvPRankingStats
					if err := calculatePvPStat(&stat, s, attack, defense, stamina, leagueOptions.Cap, pCap, level); err != nil {
						return
					}
					entry := PokemonEntry{
						Pokemon:    baseEntry.Pokemon,
						Form:       baseEntry.Form,
						Cap:        pCap,
						Value:      math.Floor(stat.Value),
						Level:      stat.Level,
						Cp:         stat.Cp,
						Percentage: roundFloat(stat.Value/combinations.TopValue, 5),
						Rank:       combinations.Combinations[(attack*16+defense)*16+stamina],
					}

					if evolution != 0 {
						entry.Evolution = evolution
					}
					entries = append(entries, entry)
				}

				// Iterate caps in ascending order using o.LevelCaps directly,
				// then the optional MaxLevel rollup. Avoids a per-call sort+alloc.
				for _, lvCap := range o.LevelCaps {
					if c, ok := combinationIndex[lvCap]; ok {
						processCombinations(float64(lvCap), c)
					}
				}
				if c, ok := combinationIndex[MaxLevel]; ok {
					processCombinations(float64(MaxLevel), c)
				}

				if len(entries) == 0 {
					continue
				}
				last := &entries[len(entries)-1]
				for len(entries) >= 2 {
					secondLast := &entries[len(entries)-2]
					if secondLast.Level != last.Level || secondLast.Rank != last.Rank {
						break
					}
					entries = entries[:len(entries)-1]
					last = secondLast
				}
				if last.Cap < MaxLevel {
					last.Capped = true
				} else {
					if len(entries) == 1 {
						continue
					}
					entries = entries[:len(entries)-1]
				}
			} else if evolution == 0 && attack == 15 && defense == 15 && stamina < 15 {
				for _, lvCap := range o.LevelCaps {
					lvCapFloat := float64(lvCap)
					if calculateHp(s, stamina, lvCapFloat) == calculateHp(s, 15, lvCapFloat) {
						entry := PokemonEntry{
							Pokemon:    baseEntry.Pokemon,
							Form:       baseEntry.Form,
							Level:      lvCapFloat,
							Percentage: 1,
							Rank:       1,
						}
						entries = append(entries, entry)
					}
				}
				if len(entries) == 0 {
					continue
				}
			} else {
				continue
			}
			if result[leagueName] == nil {
				result[leagueName] = entries
			} else {
				result[leagueName] = append(result[leagueName], entries...)
			}
		}
	}

	baseStats := stats
	pushAllEntries(&baseStats, 0)

	canEvolve := !costumeBlock
	if canEvolve && len(masterForm.Evolutions) != 0 {
		for _, evolution := range masterForm.Evolutions {
			switch evolution.Pokemon {
			case 106:
				if attack < defense || attack < stamina {
					continue
				}
			case 107:
				if defense < attack || defense < stamina {
					continue
				}
			case 237:
				if stamina < attack || stamina < defense {
					continue
				}
			}
			if evolution.GenderRequirement != 0 && gender != evolution.GenderRequirement {
				continue
			}
			evolvedRanks, _ := o.queryPvPRankInternal(depth+1, evolution.Pokemon, evolution.Form, costume, gender, attack, defense, stamina, level)
			for leagueName, results := range evolvedRanks {
				if result[leagueName] == nil {
					result[leagueName] = results
				} else {
					result[leagueName] = append(result[leagueName], results...)
				}
			}
		}
	}

	if len(masterForm.TempEvolutions) != 0 {
		for tempEvoId, tempEvo := range masterForm.TempEvolutions {
			t := tempEvo
			if t.Attack == 0 {
				t = masterPokemon.TempEvolutions[tempEvoId]
			}
			pushAllEntries(&t, tempEvoId)
		}
	}

	return result, nil
}

// QueryPvPRank Query all ranks for a specific Pokémon, including its possible evolutions.
func (o *Ohbem) QueryPvPRank(pokemonId int, form int, costume int, gender int, attack int, defense int, stamina int, level float64) (map[string][]PokemonEntry, error) {
	if err := safetyCheck(o); err != nil {
		return make(map[string][]PokemonEntry), err
	}
	if (attack < 0 || attack > 15) || (defense < 0 || defense > 15) || (stamina < 0 || stamina > 15) || level < 1 {
		return make(map[string][]PokemonEntry), ErrQueryInputOutOfRange
	}
	return o.queryPvPRankInternal(0, pokemonId, form, costume, gender, attack, defense, stamina, level)
}

// FindBaseStats Look up base stats of a Pokémon.
func (o *Ohbem) FindBaseStats(pokemonId int, form int, evolution int) (PokemonStats, error) {
	if err := safetyCheck(o); err != nil {
		return PokemonStats{}, err
	}
	o.mu.RLock()
	stats, _, _, ok := resolveStats(&o.PokemonData, pokemonId, form, evolution)
	o.mu.RUnlock()
	if !ok {
		return PokemonStats{}, ErrMissingPokemon
	}
	return stats, nil
}

// IsMegaUnreleased Check whether the stats for a given mega is speculated.
// Second arg is a TempEvolution key (e.g. 1=MegaX, 2=MegaY), not a Form ID.
func (o *Ohbem) IsMegaUnreleased(pokemonId int, tempEvolution int) (bool, error) {
	if err := safetyCheck(o); err != nil {
		return false, err
	}
	o.mu.RLock()
	defer o.mu.RUnlock()
	masterPokemon := o.PokemonData.Pokemon[pokemonId]
	if masterPokemon.Attack != 0 {
		evo := masterPokemon.TempEvolutions[tempEvolution]
		return evo.Unreleased, nil
	}
	return false, nil
}

// FilterLevelCaps Filter the output of queryPvPRank with a subset of interested level caps.
func (o *Ohbem) FilterLevelCaps(entries []PokemonEntry, interestedLevelCaps []int) []PokemonEntry {
	var result []PokemonEntry

	for _, entry := range entries {
		if entry.Cap == 0 { // functionally perfect, fast route
			for _, interested := range interestedLevelCaps {
				if float64(interested) == entry.Level {
					result = append(result, entry)
					break
				}
			}
			continue
		}
		if (entry.Capped && interestedLevelCaps[len(interestedLevelCaps)-1] < int(entry.Cap)) || (!entry.Capped && !slices.Contains(interestedLevelCaps, int(entry.Cap))) {
			continue
		}
		if len(result) > 0 {
			ref := &result[len(result)-1]
			if ref.Pokemon != 0 && ref.Pokemon == entry.Pokemon && ref.Form == entry.Form && ref.Evolution == entry.Evolution && ref.Level == entry.Level && ref.Rank == entry.Rank {
				ref.Cap = entry.Cap
				if entry.Capped {
					ref.Capped = true
				}
				continue
			}
		}
		result = append(result, entry)
	}
	return result
}

