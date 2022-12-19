# ohbemgo

## WIP - Avoid using on prod!

GoLang port of https://github.com/Mygod/ohbem

## Current State
- Node version is giving much better results in core methods. Dummy tests using 10000 iterations and results for both Go and Node provided bellow.
- Complete tests only for `pvp_core.go` private methods
- `FilterLevelCaps` not implemented
- Structs could be optimized/organized

### Dummy performance tests of core methods

Facts
```
iterations = 10000
PikachuStats = PokemonStats{Attack: 112, Defense: 96, Stamina: 111}
calculateCpMultiplier(15)
calculateHp(PikachuStats, 97, 15)
calculateStatProduct(PikachuStats, 15, 10, 5, 10)
calculateCp(PikachuStats, 15, 10, 5, 10)
calculatePvPStat(PikachuStats, 15, 10, 5, 10, 10, 1)
calculateRanksCompact(PikachuStats, 500, 30)
calculateRanks(PikachuStats, 500, 30, 1)
```

GO Timings
```
calculateCpMultiplier: 520.9µs
calculateHp: 1.0353ms
calculateStatProduct: 1.0406ms
calculateCp: 516.9µs
calculatePvPStat: 1.0344ms
calculateRanksCompact: 27.7545581s
calculateRanks: 35.7354238s
```

Node Timings
```
calculateCpMultiplier 0.5467289984226227 ms
calculateHp 1.2540239989757538 ms
calculateStatProduct 1.5217469930648804 ms
calculateCp 1.3211849927902222 ms
calculatePvPStat 0.7839210033416748 ms !!
calculateRanksCompact 17557.272655010223 ms !!
calculateRanks 21972.77894899249 ms !!
```
