# ohbemgo

## WIP - Avoid using on prod!

GoLang port of https://github.com/Mygod/ohbem

## Current State
- Node version is giving much better results in core methods. Dummy tests using 10000 iterations and results for both Go and Node provided bellow.
- Complete tests only for `pvp_core.go` private methods
- `FilterLevelCaps` not implemented
- Structs could be optimized/organized


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

used 
```go
func TimeIt() {
	var PikachuStats = PokemonStats{Attack: 112, Defense: 96, Stamina: 111}

	start := time.Now()
	for i := 0; i < 10000; i++ {
		_ = calculateCpMultiplier(15)
	}
	elapsed := time.Since(start)
	fmt.Printf("calculateCpMultiplier: %s\n", elapsed)

	start = time.Now()
	for i := 0; i < 10000; i++ {
		_ = calculateStatProduct(PikachuStats, 15, 10, 5, 10)
	}
	elapsed = time.Since(start)
	fmt.Printf("calculateStatProduct: %s\n", elapsed)

	start = time.Now()
	for i := 0; i < 10000; i++ {
		_ = calculateCp(PikachuStats, 15, 10, 5, 10)
	}
	elapsed = time.Since(start)
	fmt.Printf("calculateCp: %s\n", elapsed)

	start = time.Now()
	for i := 0; i < 10000; i++ {
		_, _ = calculatePvPStat(PikachuStats, 15, 10, 5, 10, 10, 1)
	}
	elapsed = time.Since(start)
	fmt.Printf("calculatePvPStat: %s\n", elapsed)

	start = time.Now()
	for i := 0; i < 10000; i++ {
		_, _ = calculateRanksCompact(PikachuStats, 500, 30, 1)
	}
	elapsed = time.Since(start)
	fmt.Printf("calculateRanksCompact: %s\n", elapsed)

	start = time.Now()
	for i := 0; i < 10000; i++ {
		_, _ = calculateRanks(PikachuStats, 500, 30)
	}
	elapsed = time.Since(start)
	fmt.Printf("calculateRanks: %s\n", elapsed)
}
```

and
```js
var PikachuStats = {'attack': 112, 'defense': 96, 'stamina': 111}

var hrtime = process.hrtime();
const { performance } = require('perf_hooks');

var startTime = performance.now()
var endTime = performance.now()

startTime = performance.now()
for (let i = 0;i<10000;i++) {calculateCpMultiplier(15)}
endTime = performance.now()
console.log(`calculateCpMultiplier ${endTime - startTime} ms`)
startTime = performance.now()
for (let i = 0;i<10000;i++) {calculateHp(PikachuStats, 97, 15)}
endTime = performance.now()
console.log(`calculateHp ${endTime - startTime} ms`)
startTime = performance.now()
for (let i = 0;i<10000;i++) {calculateStatProduct(PikachuStats, 15, 10, 5, 10)}
endTime = performance.now()
console.log(`calculateStatProduct ${endTime - startTime} ms`)
startTime = performance.now()
for (let i = 0;i<10000;i++) {calculateCp(PikachuStats, 15, 10, 5, 10)}
endTime = performance.now()
console.log(`calculateCp ${endTime - startTime} ms`)
startTime = performance.now()
for (let i = 0;i<10000;i++) {calculatePvPStat(PikachuStats, 15, 10, 5, 10, 10, 1)}
endTime = performance.now()
console.log(`calculatePvPStat ${endTime - startTime} ms`)
startTime = performance.now()
for (let i = 0;i<10000;i++) {calculateRanksCompact(PikachuStats, 500, 30)}
endTime = performance.now()
console.log(`calculateRanksCompact ${endTime - startTime} ms`)
startTime = performance.now()
for (let i = 0;i<10000;i++) {calculateRanks(PikachuStats, 500, 30, 1)}
endTime = performance.now()
console.log(`calculateRanks ${endTime - startTime} ms`)
```
