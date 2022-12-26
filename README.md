# OhbemGo

Ohbem is an optimized judgemental library that computes PvP rankings for Pokemon GO.

This is a rewrite of node version https://github.com/Mygod/ohbem

## Features

* Little cup/great league/ultra league rankings
* Multiple level caps (level 50/51)
* Customizable CP/level caps
* Evolutions support
* Mega evolutions support (including unreleased Mega)
* Tyrogue evolutions support
* Gender-locked evolutions support
* Unevolvable costumes support
* Tied PvP ranks
  (for example, 13/15/14 and 13/15/15 Talonflame are both UL rank 1 at L51, followed by 14/14/14 being UL rank 3)
* Functionally perfect support
* Optional built-in caching

## Current State

- Work in Progress
- Missing quite a loot of tests
- `FilterLevelCaps` and `FindBaseStats` not yet implemented

## Usage

```go
package main

import (
    "github.com/Pupitar/ohbemgo"
)

func main() {
    var leagues = map[string]int{                                     // Provide leagues configuration & caps.
        "little": 500,                                                // Cap for master is ignored.
        "great":  1500,
        "ultra":  2500,
        "master": 0,
    }
    levelCaps := []float64{50, 51}                                    // Provide level caps.

    ohbem := ohbemgo.Ohbem{Leagues: leagues, LevelCaps: levelCaps}    // Initialize Ohbem.

    err = ohbem.FetchPokemonData()                                    // Fetch latest stable MasterFile...
    err = ohbem.WatchPokemonData()                                    // ...and automatically watch for changes...
    err = ohbem.LoadPokemonData("masterfile.json")                    // ...or load from file
```

## Examples

### ohbem.QueryPvPRank
```go
    entries, err := ohbem.QueryPvPRank(
        /* pokemonId: */    605,
        /* form: */         0,
        /* costume: */      0, // costume is used to check for evolutions. To skip this check, always pass 0.
        /* gender: */       1,
        /* attack: */       1,
        /* defense: */      4,
        /* stamina: */      12,
        /* level: */        7,
    )
```
Which produces (after json.Marshal)
```json
    {
      "great": [
        {"pokemon": 605, "form": 0, "cap": 50, "value": 1444316, "level": 50, "cp": 1348, "percentage": 0.84457, "rank": 3158, "capped": false, "evolution": 0},
        {"pokemon": 605, "form": 0, "cap": 51, "value": 1472627, "level": 51, "cp": 1364, "percentage": 0.85568, "rank": 3128, "capped": false, "evolution": 0},
        {"pokemon": 606, "form": 0, "cap": 50, "value": 1639371, "level": 21, "cp": 1493, "percentage": 0.97919, "rank": 197, "capped": true, "evolution": 0}
      ],
      "little": [
        {"pokemon": 605, "form": 0, "cap": 50, "value": 320801, "level": 14.5, "cp": 494, "percentage": 0.95123, "rank": 548, "capped": true, "evolution": 0}
      ],
      "ultra": [
        {"pokemon": 606, "form": 0, "cap": 50, "value": 3519629, "level": 40, "cp": 2489, "percentage": 0.97294, "rank": 745, "capped": true, "evolution": 0}
      ]
    }
```
