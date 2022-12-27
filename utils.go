package ohbemgo

import (
	"encoding/json"
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
		return PokemonData{}, ErrMasterFileFetch
	}
	//goland:noinspection GoUnhandledErrorResult
	defer resp.Body.Close()

	var data PokemonData
	err = json.NewDecoder(resp.Body).Decode(&data)
	if err != nil {
		return PokemonData{}, ErrMasterFileDecode
	}
	data.Initialized = true
	return data, nil
}

func safetyCheck(o *Ohbem) error {
	if !o.PokemonData.Initialized {
		return ErrMasterFileUnloaded
	}
	if len(o.Leagues) == 0 {
		return ErrLeaguesMissing
	}
	if len(o.LevelCaps) == 0 {
		return ErrLevelCapsMissing
	}
	return nil
}
