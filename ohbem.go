package ohbemgo

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"os"
	"reflect"
)

const maxLevel = 100

func (o *Ohbem) FetchPokemonData() error {
	var err error

	o.PokemonData, err = fetchMasterFile()
	if err != nil {
		return err
	}
	return nil
}

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

func (o *Ohbem) CalculateAllRanks(stats PokemonStats, cpCap int) (result [101][16][16][16]Ranking, filled bool) {
	for _, lvCap := range o.LevelCaps {
		if calculateCp(stats, 15, 15, 15, lvCap) <= int(lvCap) {
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

func (o *Ohbem) CalculateTopRanks(maxRank int, pokemonId int, form int, evolution int, ivFloor int) map[string][]Ranking {
	var masterPokemon = o.PokemonData.Pokemon[pokemonId]
	var stats PokemonStats
	var masterForm Form
	var masterEvolution PokemonStats

	result := make(map[string][]Ranking)

	if masterPokemon.Attack == 0 {
		return result
	}

	if form != 0 {
		masterForm = masterPokemon.Forms[form]
	}

	if masterForm.Attack == 0 {
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

	if evolution != 0 {
		masterEvolution = masterForm.TempEvolutions[evolution]
	}

	if masterEvolution.Attack == 0 {
		masterEvolution = PokemonStats{
			Attack:     masterForm.Attack,
			Defense:    masterForm.Defense,
			Stamina:    masterForm.Stamina,
			Unreleased: false,
		}
	}

	if masterEvolution.Attack != 0 {
		stats = PokemonStats{
			Attack:  masterEvolution.Attack,
			Defense: masterEvolution.Defense,
			Stamina: masterEvolution.Stamina,
		}
	} else if masterForm.Attack != 0 {
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

	for leagueName, leagueOptions := range o.Leagues {
		var rankings []Ranking
		var lastRank []Ranking // TODO fix

		if leagueOptions.Little && !(masterForm.Little || masterPokemon.Little) {
			continue
		}

		processLevelCap := func(lvCap float64, setOnDup bool) {
			combinations, sortedRanks := calculateRanksCompact(stats, leagueOptions.Cap, lvCap, ivFloor)

			for i := 0; i < len(sortedRanks); i++ {
				var stat = sortedRanks[i]
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

				var lastStat *Ranking
				if len(lastRank) > i {
					lastStat = &lastRank[i]
				}

				if lastStat != nil && stat.Level == lastStat.Level && rank == lastStat.Rank && attack == lastStat.Attack && defense == lastStat.Defense && stamina == lastStat.Stamina {
					if setOnDup {
						lastStat.Capped = true
					}
				} else if !setOnDup {
					lastStat = &Ranking{
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
					rankings = append(rankings, *lastStat)
				}
			}
		}
		// TODO check is this section is correct
		var maxed bool
		for _, lvCap := range o.LevelCaps {
			if calculateCp(stats, 15, 15, 15, lvCap) <= leagueOptions.Cap {
				continue
			}
			processLevelCap(lvCap, false)
			if calculateCp(stats, ivFloor, ivFloor, ivFloor, lvCap+0.5) > leagueOptions.Cap {
				maxed = true
				for _, entry := range lastRank {
					entry.Capped = true
				}
				break
			}
		}
		if len(rankings) != 0 && !maxed {
			processLevelCap(maxLevel, true)
		}
		if len(rankings) != 0 {
			result[leagueName] = rankings
		}
	}

	return result
}

func (o *Ohbem) QueryPvPRank(pokemonId int, form int, costume int, gender int, attack int, defense int, stamina int, level float64) (map[string][]PokemonEntry, error) {
	result := make(map[string][]PokemonEntry)

	if (attack < 0 || attack > 15) || (defense < 0 || defense > 15) || (stamina < 0 || stamina > 15) || level < 1 {
		return result, errors.New("one of input arguments 'Attack, Defense, Stamina, Level' is out of range")
	}

	var masterForm Form
	var masterPokemon = o.PokemonData.Pokemon[pokemonId]

	if masterPokemon.Attack == 0 {
		return result, fmt.Errorf("missing Pokemon %d data", pokemonId)
	}

	if form != 0 {
		masterForm = masterPokemon.Forms[form]
	}

	if masterForm.Attack == 0 {
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
			if leagueOptions.Little && !(masterForm.Little || masterPokemon.Little) {
				continue
			}
			combinationIndex, filled := o.CalculateAllRanks(stats, leagueOptions.Cap)
			if !filled {
				continue
			}
			for lvCap, combinations := range combinationIndex {
				var entry PokemonEntry
				var ivEntry = combinations[attack][defense][stamina]
				if level > ivEntry.Level {
					continue
				}
				entry = baseEntry
				entry.Cap = float64(lvCap)
				entry.Value = ivEntry.Value
				entry.Level = ivEntry.Level
				entry.Cp = ivEntry.Cp
				entry.Percentage = ivEntry.Percentage
				entry.Rank = ivEntry.Rank
				entry.Capped = ivEntry.Capped

				if evolution != 0 {
					entry.Evolution = evolution
				}
				entry.Value = math.Floor(entry.Value)
				entries = append(entries, entry)
			}
			if len(entries) == 0 {
				continue
			}
			last := entries[len(entries)-1]
			for len(entries) >= 2 {
				secondLast := entries[len(entries)-2]
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
				pushAllEntries(tempEvo, int(tempEvoId))
			} else {
				pushAllEntries(masterPokemon.TempEvolutions[tempEvoId], int(tempEvoId))
			}

		}
	}

	return result, nil
}

func (o *Ohbem) FindBaseStats(pokemonId int, form int, evolution int) (result []Pokemon) {
	return result
}

func (o *Ohbem) FilterLevelCaps(entries []PokemonEntry, interestedLevelCaps []float64) (result []PokemonEntry) {
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
		if !reflect.DeepEqual(last, PokemonEntry{}) && last.Pokemon == entry.Pokemon && last.Form == entry.Form && last.Evolution == entry.Evolution && last.Level == entry.Level && last.Rank == entry.Rank {
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
