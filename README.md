# OhbemGo

OhbemGo is an optimized judgemental library that computes PvP rankings for Pokemon GO.

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
* Faster than node :)

## Current State

- `CalculateTopRanks` is broken.
- Everything else is fine.

## [Documentation](https://pkg.go.dev/github.com/Pupitar/ohbemgo)

## Usage

```go
package main

import (
    "github.com/Pupitar/ohbemgo"
)

func main() {
    var leagues = map[string]ohbemgo.League{                          // Leagues configuration & caps.
        "little": {                                                   // Cap for master is ignored.
            Cap:            500,
            LittleCupRules: true,
        },
        "great": {
            Cap:            1500,
            LittleCupRules: false,
        },
        "ultra": {
            Cap:            2500,
            LittleCupRules: false,
        },
        "master": {
            Cap:            0,
            LittleCupRules: false,
        },
    }
    levelCaps := []int{50, 51}                                        // Level caps.

    ohbem := ohbemgo.Ohbem{Leagues: leagues, LevelCaps: levelCaps}

    err = ohbem.FetchPokemonData()                                    // Fetch latest stable MasterFile...
    err = ohbem.WatchPokemonData()                                    // ...automatically watch remote for changes...
    err = ohbem.LoadPokemonData("masterfile.json")                    // ...or load from file

    // ...
}
```

## Examples

Provided examples are marshaled. Each method is returning defined structs. Read Documentation for details.

### QueryPvPRank

```go
entries, err := ohbem.QueryPvPRank(605, 0, 0, 1, 1, 4, 12, 7)
```
```json
{
  "great":[
    {"pokemon":605,"cap":50,"value":1444316,"level":50,"cp":1348,"percentage":0.84457,"rank":3158},
    {"pokemon":605,"cap":51,"value":1472627,"level":51,"cp":1364,"percentage":0.85568,"rank":3128},
    {"pokemon":606,"cap":40,"value":1639371,"level":21,"cp":1493,"percentage":0.97919,"rank":197,"capped":true}
  ],
  "little":[
    {"pokemon":605,"cap":40,"value":320801,"level":14.5,"cp":494,"percentage":0.95123,"rank":548,"capped":true},
    {"pokemon":606,"cap":40,"value":302917,"level":7,"cp":486,"percentage":0.93383,"rank":1056,"capped":true}
  ],
  "ultra":[
    {"pokemon":606,"cap":40,"value":3519629,"level":40,"cp":2489,"percentage":0.97294,"rank":651},
    {"pokemon":606,"cap":50,"value":3519629,"level":40,"cp":2489,"percentage":0.97294,"rank":745,"capped":true}
  ]
}
```

### CalculateTopRanks (broken)

```go
entries, err := ohbem.CalculateTopRanks(5, 605, 0, 0, 0)
```
```json
{
  "great":[
    {"value":1710113,"level":50,"cp":1498,"percentage":1,"rank":1,"attack":8,"defense":15,"stamina":15,"cap":50},
    {"value":1699358,"level":48.5,"cp":1500,"percentage":0.99371,"rank":2,"attack":11,"defense":15,"stamina":15,"cap":50},
    {"value":1699151,"level":50,"cp":1489,"percentage":0.99359,"rank":3,"attack":7,"defense":15,"stamina":15,"cap":50},
    {"value":1698809,"level":49,"cp":1500,"percentage":0.99339,"rank":4,"attack":10,"defense":15,"stamina":15,"cap":50},
    {"value":1698192,"level":49.5,"cp":1494,"percentage":0.99303,"rank":5,"attack":9,"defense":15,"stamina":14,"cap":50},
    {"value":1698192,"level":49.5,"cp":1499,"percentage":0.99303,"rank":5,"attack":9,"defense":15,"stamina":15,"cap":50},
    {"value":1720993,"level":51,"cp":1497,"percentage":1,"rank":1,"attack":6,"defense":15,"stamina":15,"cap":51},
    {"value":1717106,"level":51,"cp":1500,"percentage":0.99774,"rank":2,"attack":7,"defense":14,"stamina":15,"cap":51},
    {"value":1710113,"level":50,"cp":1498,"percentage":0.99368,"rank":3,"attack":8,"defense":15,"stamina":15,"cap":51},
    {"value":1709818,"level":51,"cp":1487,"percentage":0.99351,"rank":4,"attack":5,"defense":15,"stamina":15,"cap":51},
    {"value":1709291,"level":50.5,"cp":1498,"percentage":0.9932,"rank":5,"attack":7,"defense":15,"stamina":15,"cap":51}
  ],
  "little":[
    {"value":337248,"level":14,"cp":500,"percentage":1,"rank":1,"attack":0,"defense":14,"stamina":15,"cap":40},
    {"value":335954,"level":14,"cp":500,"percentage":0.99616,"rank":2,"attack":0,"defense":15,"stamina":13,"cap":40},
    {"value":334290,"level":14,"cp":498,"percentage":0.99123,"rank":3,"attack":0,"defense":13,"stamina":15,"cap":40},
    {"value":333943,"level":14,"cp":500,"percentage":0.9902,"rank":4,"attack":1,"defense":15,"stamina":11,"cap":40},
    {"value":333571,"level":14,"cp":499,"percentage":0.98909,"rank":5,"attack":1,"defense":12,"stamina":15,"cap":40}
  ]
}
```

### FilterLevelCaps

```go
entries, err := ohbem.QueryPvPRank(661, 0, 0, 1, 15, 15, 14, 1)
filter := ohbem.FilterLevelCaps(entries["great"], []int{51})
```
```json
[
  {"pokemon":662,"cap":51,"value":1743985,"level":41.5,"cp":1493,"percentage":0.94736,"rank":1328},
  {"pokemon":663,"cap":40,"value":1756548,"level":23.5,"cp":1476,"percentage":0.94144,"rank":2867,"capped":true}
]
```

## Benchmark

TL;DR 
* Go `QueryPvPRank` is `5` times faster than node with disabled cache.
* Go `QueryPvPRank` is `10` times faster than node with enabled cache.

### Specs  & versions
```
# OhbemGo 0.7.3
# Ohbem 1.4.1
# cpu: 12th Gen Intel(R) Core(TM) i9-12900KF

$ go version
go version go1.19.4 linux/amd64
$ node --version
v16.14.0
```
### QueryPvPRank

#### OhbemGo
```bash
$ time ./main  # cache disabled ; maxPokemonId = 2
QueryPvPRank iterated 13068 in 1m23.355235694s

real    1m23.406s
user    1m29.580s
sys     0m4.767s

$ time ./main  # cache enabled ; maxPokemonId = 200
QueryPvPRank iterated 1306800 in 3.821691967s

real    0m3.898s
user    0m4.094s
sys     0m0.315s
```

#### Ohbem (node)
```bash
$ time node main.js  # cache disabled ; maxPokemonId = 2
queryPvPRank iterated 13068 in 418771ms

real    6m58.854s
user    7m7.976s
sys     0m17.332s

$ time node main.js  # cache enabled ; maxPokemonId = 200
queryPvPRank iterated 1306800 in 38922ms

real    0m39.038s
user    0m46.019s
sys     0m3.972s
```

### Test scripts

#### `main.js`
```js
const Ohbem = require('ohbem');
const pokemonData = require('./master-test.json');

async function test() {
    const ohbem = new Ohbem({
        leagues: {
        little: {
            little: false,
            cap: 500,
        },
        great: {
            little: false,
            cap: 1500,
        },
        ultra: {
            little: false,
            cap: 2500,
        },
        master: null,
    },
        levelCaps: [40, 50, 51],
        pokemonData,
        cachingStrategy: Ohbem.cachingStrategies.balanced // change
    });

    const maxPokemonId = 200; // change
    const maxAttack = 10;
    const maxDefense = 5;
    const maxStamina = 10;
    const maxLevel = 5;
    let counter = 0;

    const start = Date.now();
    for (let p = 1; p <= maxPokemonId; p++) {
        for (let a = 0; a <= maxAttack; a++) {
            for (let d = 0; d <= maxDefense; d++) {
                for (let s = 0; s <= maxStamina; s++) {
                    for (let l = 1; l <= maxLevel; l += 0.5) {
                        ohbem.queryPvPRank(p, 0, 0, 0, a, d, s, l);
                        counter++;
                    }
                }
            }
        }
    }
    const elapsed = Date.now() - start;
    console.log(`queryPvPRank iterated ${counter} in ${elapsed}ms`);
}

test();
```

#### `main.go`
```go
package main

import (
	"fmt"
	"github.com/Pupitar/ohbemgo"
	"time"
)

func mainOne() {
	var leagues = map[string]ohbemgo.League{
		"little": {
			Cap:            500,
			LittleCupRules: false,
		},
		"great": {
			Cap:            1500,
			LittleCupRules: false,
		},
		"ultra": {
			Cap:            2500,
			LittleCupRules: false,
		},
		"master": {
			Cap:            0,
			LittleCupRules: false,
		},
	}

	levelCaps := []int{40, 50, 51}

	ohbem := ohbemgo.Ohbem{Leagues: leagues, LevelCaps: levelCaps, DisableCache: false}  // change
	_ = ohbem.LoadPokemonData("master-test.json")

	const (
		maxPokemonId = 200 // change
		maxAttack    = 10
		maxDefense   = 5
		maxStamina   = 10
		maxLevel     = 5
	)
	var counter uint

	start := time.Now()
	for p := 1; p <= maxPokemonId; p++ {
		for a := 0; a <= maxAttack; a++ {
			for d := 0; d <= maxDefense; d++ {
				for s := 0; s <= maxStamina; s++ {
					for l := 1.0; l <= maxLevel; l = l + 0.5 {
						ohbem.QueryPvPRank(p, 0, 0, 0, a, d, s, l)
						counter++
					}
				}
			}
		}
	}
	elapsed := time.Since(start)
	fmt.Printf("QueryPvPRank iterated %d in %s\n", counter, elapsed)

func main() {
	mainOne() // bench go
}
```
