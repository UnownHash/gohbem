package ohbemgo

import (
	"encoding/json"
	"errors"
	"log"
	"math"
	"os"
	"reflect"
	"sync"
	"time"
)

const maxLevel = 100

// FetchPokemonData Fetch remote MasterFile and keep it in memory.
func (o *Ohbem) FetchPokemonData() error {
	var err error

	o.PokemonData, err = fetchMasterFile()
	if err != nil {
		return err
	}
	o.ClearCache()
	return nil
}

// LoadPokemonData Load MasterFile from provided filePath and keep it in memory.
func (o *Ohbem) LoadPokemonData(filePath string) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return errors.New("can't open MasterFile")
	}
	if err := json.Unmarshal(data, &o.PokemonData); err != nil {
		return errors.New("can't unmarshal MasterFile")
	}
	return nil
}

// SavePokemonData Save MasterFile from memory to provided location.
func (o *Ohbem) SavePokemonData(filePath string) error {
	data, err := json.Marshal(o.PokemonData)
	if err != nil {
		return errors.New("can't marshal MasterFile")
	}
	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return errors.New("can't save MasterFile")
	}
	return nil
}

// WatchPokemonData Watch for remote MasterFile changes. When new, auto-update and clean cache.
func (o *Ohbem) WatchPokemonData() {
	if o.watcherChan != nil {
		log.Printf("PokemonData Watcher Already Started")
		return
	}

	log.Printf("PokemonData Watcher Started")
	o.watcherChan = make(chan bool)
	var interval time.Duration

	// if interval is not provided, use 60 minutes
	if o.WatcherInterval == 0 {
		interval = 60
	} else {
		interval = o.WatcherInterval
	}

	go func() {
		ticker := time.NewTicker(interval * time.Minute)

		for {
			select {
			case <-o.watcherChan:
				log.Printf("PokemonData Watcher Stopped")
				ticker.Stop()
				return
			case <-ticker.C:
				log.Printf("Checking remote MasterFile")
				pokemonData, err := fetchMasterFile()
				if err != nil {
					log.Printf("Remote MasterFile fetch failed")
					continue
				}
				if reflect.DeepEqual(o.PokemonData, pokemonData) {
					continue
				} else {
					log.Printf("New MasterFile found! Updating PokemonData & Cleaning cache if in use")
					o.PokemonData = pokemonData // overwrite PokemonData using new MasterFile
					o.ClearCache()              // clean compactRankCache cache
				}
			}
		}
	}()
}

// StopWatchingPokemonData Stop watching for remote MasterFile changes.
func (o *Ohbem) StopWatchingPokemonData() {
	close(o.watcherChan)
}

func (o *Ohbem) ClearCache() {
	if !o.DisableCache {
		o.compactRankCache = sync.Map{}
		log.Printf("Cache cleaned")
	}
}

// CalculateAllRanksCompact Calculate all PvP ranks for a specific base stats with the specified CP cap. Compact version intended to be used with cache.
func (o *Ohbem) CalculateAllRanksCompact(stats PokemonStats, cpCap int) (map[int]CompactCacheValue, bool) {
	cacheKey := cpCap*4096 + stats.Attack*256 + stats.Defense*16 + stats.Stamina

	if !o.DisableCache {
		if obj, ok := o.compactRankCache.Load(cacheKey); ok {
			return obj.(map[int]CompactCacheValue), true
		}
	}

	maxed, filled := false, false
	result := make(map[int]CompactCacheValue)

	for _, lvCap := range o.LevelCaps {
		if !o.IncludeHundosUnderCap && calculateCp(stats, 15, 15, 15, lvCap) <= cpCap {
			continue
		}

		combinations, sortedRanks := calculateRanksCompact(stats, cpCap, lvCap, 0)
		res := CompactCacheValue{
			Combinations: combinations,
			TopValue:     sortedRanks[0].Value,
		}
		result[int(lvCap)] = res
		filled = true
		if calculateCp(stats, 0, 0, 0, float64(lvCap)+0.5) > cpCap {
			maxed = true
			break
		}
	}
	if filled && !maxed {
		combinations, sortedRanks := calculateRanksCompact(stats, cpCap, maxLevel, 0)

		res := CompactCacheValue{
			Combinations: combinations,
			TopValue:     sortedRanks[0].Value,
		}
		result[maxLevel] = res
	}
	if !o.DisableCache && filled {
		o.compactRankCache.Store(cacheKey, result)
	}
	return result, filled
}

// CalculateAllRanks Calculate all PvP ranks for a specific base stats with the specified CP cap.
func (o *Ohbem) CalculateAllRanks(stats PokemonStats, cpCap int) ([101][16][16][16]Ranking, bool) {
	var filled bool
	var result [101][16][16][16]Ranking

	for _, lvCap := range o.LevelCaps {
		if !o.IncludeHundosUnderCap && calculateCp(stats, 15, 15, 15, lvCap) <= cpCap {
			continue
		}
		result[int(lvCap)], _ = calculateRanks(stats, cpCap, lvCap)
		filled = true
		if calculateCp(stats, 0, 0, 0, float64(lvCap)+0.5) > cpCap {
			break
		} else {
			filled = true
			result[maxLevel], _ = calculateRanks(stats, cpCap, float64(maxLevel))
		}
	}
	return result, filled
}

// CalculateTopRanks Return ranked list of PVP statistics for a given Pokémon.
func (o *Ohbem) CalculateTopRanks(maxRank int16, pokemonId int, form int, evolution int, ivFloor int) map[string][]Ranking {
	var masterPokemon = o.PokemonData.Pokemon[pokemonId]
	var stats PokemonStats
	var masterForm Form
	var masterEvolution PokemonStats

	result := make(map[string][]Ranking)

	if masterPokemon.Attack == 0 {
		return result
	}

	if form != 0 {
		if _, ok := masterPokemon.Forms[form]; ok {
			masterForm = masterPokemon.Forms[form]
		} else {
			masterForm = Form{
				Attack:  masterPokemon.Attack,
				Defense: masterPokemon.Defense,
				Stamina: masterPokemon.Stamina,
				Little:  masterPokemon.Little,
			}
		}
	} else {
		masterForm = Form{
			Attack:  masterPokemon.Attack,
			Defense: masterPokemon.Defense,
			Stamina: masterPokemon.Stamina,
			Little:  masterPokemon.Little,
		}
	}

	if evolution != 0 {
		if _, ok := masterForm.TempEvolutions[evolution]; ok {
			masterEvolution = masterForm.TempEvolutions[evolution]
		} else {
			masterEvolution = PokemonStats{
				Attack:  masterForm.Attack,
				Defense: masterForm.Defense,
				Stamina: masterForm.Stamina,
			}
		}
	}

	if masterEvolution.Attack != 0 {
		stats = PokemonStats{
			Attack:  masterEvolution.Attack,
			Defense: masterEvolution.Defense,
			Stamina: masterEvolution.Stamina,
		}
	} else {
		if masterForm.Attack != 0 {
			stats = PokemonStats{
				Attack:  masterForm.Attack,
				Defense: masterForm.Defense,
				Stamina: masterForm.Stamina,
			}
		} else {
			stats = PokemonStats{
				Attack:  masterPokemon.Attack,
				Defense: masterPokemon.Defense,
				Stamina: masterPokemon.Stamina,
			}
		}
	}

	for leagueName, leagueOptions := range o.Leagues {
		var rankings, lastRank []Ranking
		var lastStat Ranking

		processLevelCap := func(lvCap float64, setOnDup bool) {
			combinations, sortedRanks := calculateRanksCompact(stats, leagueOptions.Cap, lvCap, ivFloor)

			for i := 0; i < len(sortedRanks); i++ {
				var stat = &sortedRanks[i]
				var rank = combinations[stat.Index]
				if rank > maxRank {
					for len(lastRank) > i {
						lastRank = lastRank[:len(lastRank)-1]
					}
					break
				}
				var attack = stat.Index >> 8 % 16
				var defense = stat.Index >> 4 % 16
				var stamina = stat.Index % 16

				if len(lastRank) > i {
					lastStat = lastRank[i]
				}

				if lastStat.Value != 0 && stat.Level == lastStat.Level && rank == lastStat.Rank && attack == lastStat.Attack && defense == lastStat.Defense && stamina == lastStat.Stamina {
					if setOnDup {
						lastStat.Capped = true
					}
				} else if !setOnDup {
					lastStat = Ranking{
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
					rankings = append(rankings, lastStat)
				}
			}
		}

		if leagueOptions.LittleCupRules && !(masterForm.Little || masterPokemon.Little) {
			continue
		} else if leagueName == "master" {
			for _, lvCap := range o.LevelCaps {
				var maxHp = calculateHp(stats, 15, lvCap)
				for stamina := ivFloor; stamina < 15; stamina++ {
					if calculateHp(stats, stamina, lvCap) == maxHp {
						entry := Ranking{
							Attack:     15,
							Defense:    15,
							Stamina:    stamina,
							Level:      lvCap,
							Percentage: 1,
							Rank:       1,
						}
						rankings = append(rankings, entry)
					}
				}
			}
		} else {
			var maxed bool
			for _, lvCap := range o.LevelCaps {
				if !o.IncludeHundosUnderCap && calculateCp(stats, 15, 15, 15, lvCap) <= leagueOptions.Cap {
					continue
				}
				processLevelCap(lvCap, false)
				if calculateCp(stats, ivFloor, ivFloor, ivFloor, lvCap+0.5) > leagueOptions.Cap {
					maxed = true
					for ix := range lastRank {
						lastRank[ix].Capped = true
					}
					break
				}
			}
			if len(rankings) != 0 && !maxed {
				processLevelCap(maxLevel, true)
			}
		}
		if len(rankings) != 0 {
			result[leagueName] = rankings
		}
	}

	return result
}

// QueryPvPRank Query all ranks for a specific Pokémon, including its possible evolutions.
func (o *Ohbem) QueryPvPRank(pokemonId int, form int, costume int, gender int, attack int, defense int, stamina int, level float64) (map[string][]PokemonEntry, error) {
	result := make(map[string][]PokemonEntry)

	if (attack < 0 || attack > 15) || (defense < 0 || defense > 15) || (stamina < 0 || stamina > 15) || level < 1 {
		return result, errors.New("one of input arguments 'Attack, Defense, Stamina, Level' is out of range")
	}

	var masterForm Form
	var masterPokemon = o.PokemonData.Pokemon[pokemonId]

	if masterPokemon.Attack == 0 {
		return result, errors.New("missing Pokemon data")
	}

	if form != 0 {
		masterForm = masterPokemon.Forms[form]
	}

	if masterForm.Attack == 0 && masterForm.Defense == 0 {
		masterForm = Form{
			Attack:                    masterPokemon.Attack,
			Defense:                   masterPokemon.Defense,
			Stamina:                   masterPokemon.Stamina,
			Little:                    masterPokemon.Little,
			Evolutions:                masterPokemon.Evolutions,
			TempEvolutions:            masterPokemon.TempEvolutions,
			CostumeOverrideEvolutions: masterPokemon.CostumeOverrideEvolutions,
		}
	}

	var baseEntry = PokemonEntry{Pokemon: pokemonId}

	if form != 0 {
		baseEntry.Form = form
	}

	pushAllEntries := func(stats PokemonStats, evolution int) {
		for leagueName, leagueOptions := range o.Leagues {
			var entries []PokemonEntry

			if leagueName == "master" {
				if evolution == 0 && attack == 15 && defense == 15 && stamina < 15 {
					for _, lvCap := range o.LevelCaps {
						if calculateHp(stats, stamina, lvCap) == calculateHp(stats, 15, lvCap) {
							entry := PokemonEntry{
								Pokemon:    baseEntry.Pokemon,
								Form:       baseEntry.Form,
								Level:      lvCap,
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
				result[leagueName] = entries
			} else {
				if leagueOptions.LittleCupRules && !(masterForm.Little || masterPokemon.Little) {
					continue
				}
				combinationIndex, filled := o.CalculateAllRanksCompact(stats, leagueOptions.Cap)
				if !filled {
					continue
				}
				for lvCap, combinations := range combinationIndex {
					pCap := float64(lvCap)
					stat, err := calculatePvPStat(stats, attack, defense, stamina, leagueOptions.Cap, pCap, level)
					if err != nil {
						continue
					}
					entry := PokemonEntry{
						Pokemon:    baseEntry.Pokemon,
						Form:       baseEntry.Form,
						Cap:        pCap,
						Value:      stat.Value,
						Level:      stat.Level,
						Cp:         stat.Cp,
						Percentage: roundFloat(stat.Value/combinations.TopValue, 5),
						Rank:       combinations.Combinations[(attack*16+defense)*16+stamina],
					}

					if evolution != 0 {
						entry.Evolution = evolution
					}
					entry.Value = math.Floor(entry.Value)
					entries = append(entries, entry)
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
				if last.Cap < maxLevel {
					last.Capped = true
				} else {
					if len(entries) == 1 {
						continue
					}
					entries = entries[:len(entries)-1]
				}
				if result[leagueName] == nil {
					result[leagueName] = entries
				} else {
					result[leagueName] = append(result[leagueName], entries...)
				}
			}
		}
	}

	pushAllEntries(PokemonStats{masterForm.Attack, masterForm.Defense, masterForm.Stamina, true}, 0)
	var canEvolve = true
	if costume != 0 {
		canEvolve = !o.PokemonData.Costumes[costume] || containsInt(masterForm.CostumeOverrideEvolutions, costume)
	}
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
			pushRecursively := func(form int) {
				evolvedRanks, _ := o.QueryPvPRank(evolution.Pokemon, form, costume, gender, attack, defense, stamina, level)
				for leagueName, results := range evolvedRanks {
					if result[leagueName] == nil {
						result[leagueName] = results
					} else {
						result[leagueName] = append(result[leagueName], results...)
					}
				}
			}
			pushRecursively(evolution.Form)
			switch evolution.Pokemon {
			case 26:
				pushRecursively(50) // RAICHU_ALOLA
			case 103:
				pushRecursively(78) // EXEGGUTOR_ALOLA
			case 105:
				pushRecursively(80) // MAROWAK_ALOLA
			case 110:
				pushRecursively(944) // WEEZING_GALARIAN
			}
		}
	}

	if len(masterForm.TempEvolutions) != 0 {
		for tempEvoId, tempEvo := range masterForm.TempEvolutions {
			if tempEvo.Attack != 0 {
				pushAllEntries(tempEvo, tempEvoId)
			} else {
				pushAllEntries(masterPokemon.TempEvolutions[tempEvoId], tempEvoId)
			}

		}
	}

	return result, nil
}

// FindBaseStats Look up base stats of a Pokémon.
func (o *Ohbem) FindBaseStats(pokemonId int, form int, evolution int) (PokemonStats, error) {
	masterPokemon, ok := o.PokemonData.Pokemon[pokemonId]
	if !ok {
		return PokemonStats{}, errors.New("missing pokemonId in MasterFile")
	}

	var masterForm Form
	var masterEvolution PokemonStats

	if form != 0 {
		if _, ok := masterPokemon.Forms[form]; ok {
			masterForm = masterPokemon.Forms[form]
		} else {
			masterForm = Form{
				Attack:  masterPokemon.Attack,
				Defense: masterPokemon.Defense,
				Stamina: masterPokemon.Stamina,
			}
		}
	} else {
		masterForm = Form{
			Attack:  masterPokemon.Attack,
			Defense: masterPokemon.Defense,
			Stamina: masterPokemon.Stamina,
		}
	}

	if evolution != 0 {
		if _, ok := masterPokemon.TempEvolutions[evolution]; ok {
			masterEvolution = masterPokemon.TempEvolutions[evolution]
		} else {
			masterForm = Form{
				Attack:  masterPokemon.Attack,
				Defense: masterPokemon.Defense,
				Stamina: masterPokemon.Stamina,
			}
		}
	} else {
		masterForm = Form{
			Attack:  masterPokemon.Attack,
			Defense: masterPokemon.Defense,
			Stamina: masterPokemon.Stamina,
		}
	}

	if masterEvolution.Attack != 0 {
		return masterEvolution, nil
	} else if masterForm.Attack != 0 {
		return PokemonStats{
			Attack:  masterForm.Attack,
			Defense: masterForm.Defense,
			Stamina: masterForm.Stamina,
		}, nil
	} else {
		return PokemonStats{
			Attack:  masterPokemon.Attack,
			Defense: masterPokemon.Defense,
			Stamina: masterPokemon.Stamina,
		}, nil
	}
}

// IsMegaUnreleased Check whether the stats for a given mega is speculated.
func (o *Ohbem) IsMegaUnreleased(pokemonId int, evolution int) bool {
	masterPokemon := o.PokemonData.Pokemon[pokemonId]
	if masterPokemon.Attack != 0 {
		evo := masterPokemon.TempEvolutions[evolution]
		return evo.Unreleased
	}
	return false
}

// FilterLevelCaps Filter the output of queryPvPRank with a subset of interested level caps.
func (o *Ohbem) FilterLevelCaps(entries []PokemonEntry, interestedLevelCaps []float64) []PokemonEntry {
	var result []PokemonEntry
	var last PokemonEntry

	for _, entry := range entries {
		if entry.Cap == 0 { // functionally perfect, fast route
			for _, interested := range interestedLevelCaps {
				if interested == entry.Level {
					result = append(result, entry)
					break
				}
			}
			continue
		}
		if (entry.Capped && interestedLevelCaps[len(interestedLevelCaps)-1] < entry.Cap) || (!entry.Capped && !containsFloat64(interestedLevelCaps, entry.Cap)) {
			continue
		}
		if last.Pokemon != 0 && last.Pokemon == entry.Pokemon && last.Form == entry.Form && last.Evolution == entry.Evolution && last.Level == entry.Level && last.Rank == entry.Rank {
			last.Cap = entry.Cap
			if entry.Capped {
				last.Capped = true
			}
		} else {
			result = append(result, entry)
			last = result[len(result)-1]
		}
	}
	return result
}
