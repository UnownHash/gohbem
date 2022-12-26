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

Provided examples are marshaled. Each method is returning defined structs.

### QueryPvPRank

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

### CalculateTopRanks

```go
    entries := ohbem.CalculateTopRanks(
        /* maxRank: */       5,
        /* pokemonId: */     605,
        /* form: */          0,
        /* evolution: */     0,
        /* ivFloor: */       0,
    )
```
```json
{
  "great": [
    {"value": 1710113, "level": 50, "cp": 1498, "percentage": 1, "rank": 1, "attack": 8, "defense": 15, "stamina": 15, "cap": 50, "capped": false},
    {"value": 1699358, "level": 48.5, "cp": 1500, "percentage": 0.99371, "rank": 2, "attack": 11, "defense": 15, "stamina": 15, "cap": 50, "capped": false},
    {"value": 1699151, "level": 50, "cp": 1489, "percentage": 0.99359, "rank": 3, "attack": 7, "defense": 15, "stamina": 15, "cap": 50, "capped": false},
    {"value": 1698809, "level": 49, "cp": 1500, "percentage": 0.99339, "rank": 4, "attack": 10, "defense": 15, "stamina": 15, "cap": 50, "capped": false},
    {"value": 1698192, "level": 49.5, "cp": 1494, "percentage": 0.99303, "rank": 5, "attack": 9, "defense": 15, "stamina": 14, "cap": 50, "capped": false},
    {"value": 1698192, "level": 49.5, "cp": 1499, "percentage": 0.99303, "rank": 5, "attack": 9, "defense": 15, "stamina": 15, "cap": 50, "capped": false},
    {"value": 1720993, "level": 51, "cp": 1497, "percentage": 1, "rank": 1, "attack": 6, "defense": 15, "stamina": 15, "cap": 51, "capped": false},
    {"value": 1717106, "level": 51, "cp": 1500, "percentage": 0.99774, "rank": 2, "attack": 7, "defense": 14, "stamina": 15, "cap": 51, "capped": false},
    {"value": 1710113, "level": 50, "cp": 1498, "percentage": 0.99368, "rank": 3, "attack": 8, "defense": 15, "stamina": 15, "cap": 51, "capped": false},
    {"value": 1709818, "level": 51, "cp": 1487, "percentage": 0.99351, "rank": 4, "attack": 5, "defense": 15, "stamina": 15, "cap": 51, "capped": false},
    {"value": 1709291, "level": 50.5, "cp": 1498, "percentage": 0.9932, "rank": 5, "attack": 7, "defense": 15, "stamina": 15, "cap": 51, "capped": false}
  ],
  "little": [
    {"value": 337248, "level": 14, "cp": 500, "percentage": 1, "rank": 1, "attack": 0, "defense": 14, "stamina": 15, "cap": 50, "capped": false},
    {"value": 335954, "level": 14, "cp": 500, "percentage": 0.99616, "rank": 2, "attack": 0, "defense": 15, "stamina": 13, "cap": 50, "capped": false},
    {"value": 334290, "level": 14, "cp": 498, "percentage": 0.99123, "rank": 3, "attack": 0, "defense": 13, "stamina": 15, "cap": 50, "capped": false},
    {"value": 333943, "level": 14, "cp": 500, "percentage": 0.9902, "rank": 4, "attack": 1, "defense": 15, "stamina": 11, "cap": 50, "capped": false},
    {"value": 333571, "level": 14, "cp": 499, "percentage": 0.98909, "rank": 5, "attack": 1, "defense": 12, "stamina": 15, "cap": 50, "capped": false}
  ]
}
```
