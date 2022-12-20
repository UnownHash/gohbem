# ohbemgo

GoLang port of https://github.com/Mygod/ohbem


# WIP

## Current State
- Complete tests only for private `pvp_core.go` core methods
- `FilterLevelCaps` not implemented & structs could be optimized/organized

### Usage
```go
package main

import (
	"github.com/Pupitar/ohbemgo"
)

func main() {
	type Leagues map[string]struct {
		Cap    int
		Little bool
	}

	var PikachuStats = ohbemgo.PokemonStats{Attack: 112, Defense: 96, Stamina: 111}

	leagues := Leagues{
		"little": {
			Cap:    500,
			Little: true,
		},
		"great": {
			Cap:    1500,
			Little: false,
		},
		"ultra": {
			Cap:    2500,
			Little: false,
		},
		"master": {
			Cap:    0,
			Little: false,
		},
	}

	levelCaps := []float64{50, 51}

	ohbem := ohbemgo.Ohbem{Leagues: ohbemgo.Leagues(leagues), LevelCaps: levelCaps} // Initialize
	_ = ohbem.FetchPokemonData()                                                    // fetch MasterFile...
	_ = ohbem.LoadPokemonData("masterfile.json")                                    // ...or load from file
	ohbem.CalculateTopRanks(...)
}
```
