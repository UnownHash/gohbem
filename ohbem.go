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

func (o *Ohbem) CalculateAllRanks(stats PokemonStats, cpCap int) (result [101][16][16][16]Ranking) {
	maxed := false
	for _, lvCap := range o.LevelCaps {
		if calculateCp(stats, 15, 15, 15, lvCap) <= int(lvCap) {
			continue
		}
		result[int(lvCap)], _ = calculateRanks(stats, cpCap, lvCap)
		if calculateCp(stats, 0, 0, 0, float64(lvCap)+0.5) > cpCap {
			maxed = true
			break
		}
		if !maxed {
			result[maxLevel], _ = calculateRanks(stats, cpCap, float64(maxLevel))
		}
	}
	return result
}

func (o *Ohbem) CalculateTopRanks(maxRank int, pokemonId int, form int, evolution int, ivFloor int) (result map[string][]Ranking) {
	var masterPokemon = o.PokemonData.Pokemon[pokemonId]
	var stats PokemonStats
	var masterForm Form
	var masterEvolution TempEvolution

	result = make(map[string][]Ranking)

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
		masterEvolution = TempEvolution{
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
		var maxed = false
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

func (o *Ohbem) QueryPvPRank(pokemonId int, form int, costume int, gender int, attack int, defense int, stamina int, level float64) (result []PokemonEntry) {
	if (attack < 0 || attack > 15) || (defense < 0 || defense > 15) || (stamina < 0 || stamina > 15) || level < 1 {
		panic(fmt.Errorf("one of input arguments 'Attack, Defense, Stamina, Level' is out of range"))
	}

	var masterForm Form
	var masterPokemon = o.PokemonData.Pokemon[pokemonId]

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

	var baseEntry = PokemonEntry{Pokemon: pokemonId}

	if form != 0 {
		baseEntry.Form = form
	}

	//pushAllEntries := func(stats PokemonStats, evolution int) {
	//	for leagueName, leagueOptions := range o.Leagues {
	//		var entries []PokemonEntry
	//		if leagueOptions.Little && !(masterForm.Little || masterPokemon.Little) {
	//			continue
	//		}
	//		var combinationIndex = o.CalculateAllRanks(stats, leagueOptions.Cap)
	//		if len(combinationIndex) == 0 {
	//			continue
	//		}
	//		for lvCap, combinations := range combinationIndex {
	//			var entry PokemonEntry
	//			var ivEntry = combinations[attack][defense][stamina]
	//			if level > ivEntry.Level {
	//				continue
	//			}
	//			entry =
	//		}
	//	}
	//
	//}

	return result
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
