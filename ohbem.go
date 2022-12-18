package ohbemgo

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
)

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

func (o *Ohbem) CalculateAllRanks(stats PokemonStats, cpCap int) {

}

func (o *Ohbem) CalculateTopRanks(maxRank int, pokemonId int, form int, evolution int, ivFloor int) (result []PokemonEntry) {
	return result
}

func (o *Ohbem) QueryPvPRank(pokemonId int, form int, costume int, gender int, attack int, defense int, stamina int, level float64) (result []PokemonEntry) {
	if (attack < 0 || attack > 15) || (defense < 0 || defense > 15) || (stamina < 0 || stamina > 15) || level < 1 {
		panic(fmt.Errorf("one of input arguments 'Attack, Defense, Stamina, Level' is out of range"))
	}

	return result
}

func (o *Ohbem) FindBaseStats(pokemonId int, form int, evolution int) (result []Pokemon) {
	return result
}
