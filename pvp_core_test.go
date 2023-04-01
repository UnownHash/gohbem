package gohbem

import (
	"fmt"
	"testing"
)

var PikachuStats = PokemonStats{Attack: 112, Defense: 96, Stamina: 111}
var ElgyemStats = PokemonStats{Attack: 148, Defense: 100, Stamina: 146}

func TestCalculateCpMultiplier(t *testing.T) {
	var tests = []struct {
		input  float64
		output float64
	}{
		{0, 0},
		{1, 0.0939999967813492},
		{10, 0.422500014305115},
		{40, 0.790300011634826},
		{50, 0.840300023555755},
		{55, 0.865299999713897},
	}

	for ix, test := range tests {
		testName := fmt.Sprintf("%d", ix)
		t.Run(testName, func(t *testing.T) {
			ans := calculateCpMultiplier(test.input)
			if ans != test.output {
				t.Errorf("got %f, want %f", ans, test.output)
			}
		})
	}
}

func BenchmarkCalculateCpMultiplier(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = calculateCpMultiplier(10.65)
	}
}

func TestCalculateHp(t *testing.T) {
	var tests = []struct {
		stats   PokemonStats
		stamina int
		level   float64
		output  int
	}{
		{PikachuStats, 0, 0, 0},
		{PikachuStats, 10, 10, 51},
		{PikachuStats, 12, 25.5, 82},
		{PikachuStats, 98, 10, 88},
		{PikachuStats, 100, 30, 154},
		{PikachuStats, 97, 35.5, 159},

		{ElgyemStats, 0, 0, 0},
		{ElgyemStats, 10, 10, 65},
		{ElgyemStats, 12, 25.5, 106},
		{ElgyemStats, 98, 10, 103},
		{ElgyemStats, 100, 30, 179},
		{ElgyemStats, 97, 35.5, 185},
	}

	for ix, test := range tests {
		testName := fmt.Sprintf("%d", ix)
		t.Run(testName, func(t *testing.T) {
			ans := calculateHp(test.stats, test.stamina, test.level)
			if ans != test.output {
				t.Errorf("got %d, want %d", ans, test.output)
			}
		})
	}
}

func BenchmarkCalculateHp(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = calculateHp(PikachuStats, 95, 30)
	}
}

func TestCalculateStatProduct(t *testing.T) {
	var tests = []struct {
		stats   PokemonStats
		attack  int
		defense int
		stamina int
		level   float64
		output  float64
	}{
		{PikachuStats, 0, 0, 0, 0, 0},
		{PikachuStats, 10, 5, 2, 15, 191316.26099503902},
		{PikachuStats, 5, 0, 0, 20, 264564.4463604694},
		{PikachuStats, 15, 15, 15, 30.5, 700137.150494098},

		{ElgyemStats, 0, 0, 0, 0, 0},
		{ElgyemStats, 10, 5, 2, 15, 337522.4500514709},
		{ElgyemStats, 5, 0, 0, 20, 475051.98155489296},
		{ElgyemStats, 15, 15, 15, 30.5, 1194087.212935685},
	}

	for ix, test := range tests {
		testName := fmt.Sprintf("%d", ix)
		t.Run(testName, func(t *testing.T) {
			ans := calculateStatProduct(test.stats, test.attack, test.defense, test.stamina, test.level)
			if ans != test.output {
				t.Errorf("got %f, want %f", ans, test.output)
			}
		})
	}
}

func BenchmarkCalculateStatProduct(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = calculateStatProduct(PikachuStats, 10, 5, 2, 15)
	}
}

func TestCalculateCp(t *testing.T) {
	var tests = []struct {
		stats   PokemonStats
		attack  int
		defense int
		stamina int
		level   float64
		output  int
	}{
		{PikachuStats, 0, 0, 0, 0, 0},
		{PikachuStats, 0, 0, 0, 1, 10},
		{PikachuStats, 15, 15, 15, 30, 804},
		{PikachuStats, 10, 2, 15, 30, 725},
		{PikachuStats, 15, 15, 15, 34.5, 864},

		{ElgyemStats, 0, 0, 0, 0, 0},
		{ElgyemStats, 0, 0, 0, 1, 15},
		{ElgyemStats, 15, 15, 15, 30, 1187},
		{ElgyemStats, 10, 2, 15, 30, 1084},
		{ElgyemStats, 15, 15, 15, 34.5, 1276},
	}

	for ix, test := range tests {
		testName := fmt.Sprintf("%d", ix)
		t.Run(testName, func(t *testing.T) {
			ans := calculateCp(test.stats, test.attack, test.defense, test.stamina, test.level)
			if ans != test.output {
				t.Errorf("got %d, want %d", ans, test.output)
			}
		})
	}
}

func BenchmarkCalculateCp(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = calculateCp(PikachuStats, 10, 5, 2, 15)
	}
}

func TestCalculatePvPStat(t *testing.T) {
	var tests = []struct {
		stats       PokemonStats
		attack      int
		defense     int
		stamina     int
		cap         int
		lvCap       float64
		minLevel    float64
		outputValue float64
		outputLevel float64
		outputCp    int
	}{
		{PikachuStats, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		{PikachuStats, 0, 0, 0, 40, 0, 1, 950.0466549389662, 1, 10},
		{PikachuStats, 5, 10, 15, 300, 20, 1, 154064.78899667264, 12, 289},
		{PikachuStats, 0, 0, 0, 100, 10, 1, 28985.670041102363, 5, 97},
		{PikachuStats, 15, 15, 15, 5000, 50, 1, 1045164.7410539213, 50, 1060},

		{ElgyemStats, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		{ElgyemStats, 0, 0, 0, 40, 0, 1, 1700.04628357754, 1, 15},
		{ElgyemStats, 5, 10, 15, 300, 20, 1, 142181.60313125828, 8, 286},
		{ElgyemStats, 0, 0, 0, 100, 10, 1, 28162.402883172115, 3.5, 100},
		{ElgyemStats, 15, 15, 15, 5000, 50, 1, 1786849.4577316528, 50, 1566},
	}

	for ix, test := range tests {
		testName := fmt.Sprintf("%d", ix)
		testOutput := Ranking{Value: test.outputValue, Level: test.outputLevel, Cp: test.outputCp}
		t.Run(testName, func(t *testing.T) {
			ans, _ := calculatePvPStat(test.stats, test.attack, test.defense, test.stamina, test.cap, test.lvCap, test.minLevel)
			if ans.Value != test.outputValue || ans.Level != test.outputLevel || ans.Cp != test.outputCp {
				t.Errorf("got %+v, want %+v", ans, testOutput)
			}
		})
	}
}

func BenchmarkCalculatePvPStat(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = calculatePvPStat(PikachuStats, 10, 5, 2, 15, 40, 1)
	}
}

func TestCalculateRanks(t *testing.T) {
	var combinationTests = []struct {
		stats      PokemonStats
		cpCap      int
		lvCap      float64
		attack     int
		defense    int
		stamina    int
		value      float64
		level      float64
		cp         int
		percentage float64
		rank       int16
	}{
		{PikachuStats, 100, 0, 0, 0, 0, 950.0466549389662, 1, 10, 0.69338, 4090},
		{PikachuStats, 10, 0, 0, 0, 0, 950.0466549389662, 1, 10, 0, 497},
		{PikachuStats, 600, 30, 0, 0, 0, 439598.41793819424, 29, 598, 0.92505, 1994},
		{PikachuStats, 600, 30, 15, 0, 0, 410559.4224700931, 25.5, 596, 0.86395, 4089},
		{PikachuStats, 600, 30, 15, 15, 0, 419733.0878105161, 23.5, 591, 0.88325, 3924},
		{PikachuStats, 600, 30, 15, 15, 15, 431674.6163042061, 22, 589, 0.90838, 2984},

		{ElgyemStats, 100, 0, 0, 0, 0, 1700.04628357754, 1, 15, 0.68427, 4094},
		{ElgyemStats, 600, 30, 0, 0, 0, 405531.30261898035, 18.5, 590, 0.9177, 2959},
		{ElgyemStats, 600, 30, 15, 0, 0, 395597.85979182937, 17, 597, 0.89522, 3886},
		{ElgyemStats, 600, 30, 15, 15, 0, 394072.0167082542, 15.5, 584, 0.89177, 3951},
		{ElgyemStats, 600, 30, 15, 15, 15, 416491.5778971401, 15, 593, 0.9425, 1315},
	}

	var sortedTests = []struct {
		stats      PokemonStats
		cpCap      int
		lvCap      float64
		pos        int
		value      float64
		level      float64
		cp         int
		percentage float64
		rank       int16
	}{
		{PikachuStats, 100, 10, 0, 32189.76897186037, 4.5, 100, 1, 1},
		{PikachuStats, 100, 10, 4095, 23253.65960055367, 4, 88, 0.72239, 4096},
		{PikachuStats, 600, 30, 1, 472634.34117978957, 26, 600, 0.99457, 2},
		{PikachuStats, 600, 30, 15, 468041.78255510365, 25.5, 598, 0.98491, 15},
		{PikachuStats, 600, 30, 100, 461712.0022201541, 26.5, 600, 0.97159, 101},
		{PikachuStats, 600, 30, 4095, 406700.0985435657, 25, 590, 0.85582, 4096},

		{ElgyemStats, 100, 10, 0, 29115.735973629493, 3, 100, 1, 1},
		{ElgyemStats, 100, 10, 4095, 20145.311247780945, 2.5, 80, 0.6919, 4096},
		{ElgyemStats, 600, 30, 0, 441901.18212997954, 16.5, 600, 1, 1},
		{ElgyemStats, 600, 30, 4095, 382959.52940267365, 16.5, 584, 0.86662, 4096},
	}

	for ix, test := range combinationTests {
		testName := fmt.Sprintf("combinations/%d", ix)
		t.Run(testName, func(t *testing.T) {
			combinations, _ := calculateRanks(test.stats, test.cpCap, test.lvCap)
			ans := combinations[test.attack][test.defense][test.stamina]
			if ans.Value != test.value || ans.Level != test.level || ans.Cp != test.cp || ans.Percentage != test.percentage || ans.Rank != test.rank {
				t.Errorf("got %+v, want %+v", ans, test)
			}
		})
	}

	for ix, test := range sortedTests {
		testName := fmt.Sprintf("sortedRanks/%d", ix)
		t.Run(testName, func(t *testing.T) {
			_, sortedRanks := calculateRanks(test.stats, test.cpCap, test.lvCap)
			ans := sortedRanks[test.pos]
			if ans.Value != test.value || ans.Level != test.level || ans.Cp != test.cp || ans.Percentage != test.percentage || ans.Rank != test.rank {
				t.Errorf("got %+v, want %+v", ans, test)
			}
		})
	}
}

func BenchmarkCalculateRanks(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = calculateRanks(PikachuStats, 600, 30)
	}
}

func TestCalculateRanksCompact(t *testing.T) {
	var combinationTests = []struct {
		cpCap   int
		lvCap   float64
		ivFloor int
		pos     int
		rank    int16
	}{
		{40, 0, 0, 0, 4090},
		{40, 0, 0, 454, 3237},
		{40, 0, 1, 272, 0},
		{40, 0, 1, 273, 3370},
		{40, 0, 1, 279, 2983},
		{1500, 30, 1, 1500, 770},
		{1500, 30, 1, 2500, 1346},
		{1500, 30, 1, 3500, 311},
	}

	var sortedTests = []struct {
		cpCap   int
		lvCap   float64
		ivFloor int
		pos     int
		value   float64
		level   float64
		cp      int
		index   int
	}{
		{40, 0, 1, 0, 1370.171918167975, 1, 12, 4087},
		{1500, 30, 1, 0, 694353.519051347, 30, 804, 4095},
		{1500, 30, 1, 15, 675259.5521701364, 30, 791, 3822},
		{1500, 30, 1, 2547, 549931.3021919342, 30, 677, 349},
	}

	for ix, test := range combinationTests {
		testName := fmt.Sprintf("combinations/%d", ix)
		t.Run(testName, func(t *testing.T) {
			combinations, _ := calculateRanksCompact(PikachuStats, test.cpCap, test.lvCap, test.ivFloor)
			ans := combinations[test.pos]
			if ans != test.rank {
				t.Errorf("got %d, want %d", ans, test.rank)
			}
		})
	}

	for ix, test := range sortedTests {
		testName := fmt.Sprintf("sortedRanks/%d", ix)
		t.Run(testName, func(t *testing.T) {
			_, sortedRanks := calculateRanksCompact(PikachuStats, test.cpCap, test.lvCap, test.ivFloor)
			ans := sortedRanks[test.pos]
			if ans.Value != test.value || ans.Level != test.level || ans.Cp != test.cp || ans.Index != test.index {
				t.Errorf("got %+v, want %+v", ans, test)
			}
		})
	}
}

func BenchmarkCalculateRanksCompact(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = calculateRanksCompact(PikachuStats, 40, 0, 1)
	}
}
