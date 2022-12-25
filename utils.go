package ohbemgo

import (
	"encoding/json"
	"errors"
	"math"
	"net/http"
)

const masterFileUrl = "https://raw.githubusercontent.com/WatWowMap/Masterfile-Generator/master/master-latest-basics.json"

func roundFloat(val float64, precision uint) float64 {
	ratio := math.Pow(10, float64(precision))
	return math.Round(val*ratio) / ratio
}

func containsFloat64(slice []float64, value float64) bool {
	for _, v := range slice {
		if v == value {
			return true
		}
	}
	return false
}

func containsInt(slice []int, value int) bool {
	for _, v := range slice {
		if v == value {
			return true
		}
	}
	return false
}

func fetchMasterFile() (PokemonData, error) {
	resp, err := http.Get(masterFileUrl)
	if err != nil {
		return PokemonData{}, errors.New("can't fetch remote MasterFile")
	}
	//goland:noinspection GoUnhandledErrorResult
	defer resp.Body.Close()

	var data PokemonData
	err = json.NewDecoder(resp.Body).Decode(&data)
	if err != nil {
		return PokemonData{}, errors.New("can't decode remote MasterFile")
	}
	return data, nil
}

func UNUSED(x ...interface{}) {}
