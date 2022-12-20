package ohbemgo

import (
	"fmt"
	"testing"
)

var PikachuStats = PokemonStats{Attack: 112, Defense: 96, Stamina: 111}

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

	for _, test := range tests {
		testName := fmt.Sprintf("%f", test.input)
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
		stats  PokemonStats
		iv     float64
		level  float64
		output int
	}{
		{PikachuStats, 10, 10, 51},
		{PikachuStats, 12.5, 25.5, 83},
		{PikachuStats, 98, 10, 88},
	}

	for _, test := range tests {
		testName := fmt.Sprintf("%+v, %f, %f = %d", test.stats, test.iv, test.level, test.output)
		t.Run(testName, func(t *testing.T) {
			ans := calculateHp(test.stats, test.iv, test.level)
			if ans != test.output {
				t.Errorf("got %d, want %d", ans, test.output)
			}
		})
	}
}

func BenchmarkCalculateHp(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = calculateHp(PikachuStats, 95.5, 30)
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
		{PikachuStats, 10, 5, 2, 15, 191316.26099503902},
		{PikachuStats, 5, 0, 0, 20, 264564.4463604694},
	}

	for _, test := range tests {
		testName := fmt.Sprintf("%+v, %d, %d, %d, %f = %f", test.stats, test.attack, test.defense, test.stamina, test.level, test.output)
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
		{PikachuStats, 15, 15, 15, 30, 804},
		{PikachuStats, 10, 2, 15, 30, 725},
	}

	for _, test := range tests {
		testName := fmt.Sprintf("%+v, %d, %d, %d, %f = %d", test.stats, test.attack, test.defense, test.stamina, test.level, test.output)
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
		{PikachuStats, 5, 10, 15, 300, 20, 1, 154064.78899667264, 12, 289},
		{PikachuStats, 0, 0, 0, 100, 10, 1, 28985.670041102363, 5, 97},
	}

	for _, test := range tests {
		testOutput := Ranking{Value: test.outputValue, Level: test.outputLevel, Cp: test.outputCp}
		testName := fmt.Sprintf("%+v, %d, %d, %d, %d, %f, %f = %+v", test.stats, test.attack, test.defense, test.stamina, test.cap, test.lvCap, test.minLevel, testOutput)
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
		attack     int
		defense    int
		stamina    int
		value      float64
		level      float64
		cp         int
		percentage float64
		rank       int
	}{
		{0, 0, 0, 439598.41793819424, 29, 598, 0.92505, 1994},
		{15, 0, 0, 410559.4224700931, 25.5, 596, 0.86395, 4089},
		{15, 15, 0, 419733.0878105161, 23.5, 591, 0.88325, 3924},
		{15, 15, 15, 431674.6163042061, 22, 589, 0.90838, 2984},
	}

	var sortedTests = []struct {
		pos        int
		value      float64
		level      float64
		cp         int
		percentage float64
		rank       int
	}{
		{0, 475214.44963073777, 25.5, 600, 1, 1},
		{1, 472634.34117978957, 26, 600, 0.99457, 2},
		{15, 468041.78255510365, 25.5, 598, 0.98491, 15},
		{100, 461712.0022201541, 26.5, 600, 0.97159, 101},
		{4095, 406700.0985435657, 25, 590, 0.85582, 4096},
	}

	combinations, sortedRanks := calculateRanks(PikachuStats, 600, 30)

	for _, test := range combinationTests {
		testName := fmt.Sprintf("combinations %d/%d/%d = %f, %f, %d, %f, %d", test.attack, test.defense, test.stamina, test.value, test.level, test.cp, test.percentage, test.rank)
		t.Run(testName, func(t *testing.T) {
			ans := combinations[test.attack][test.defense][test.stamina]
			if ans.Value != test.value || ans.Level != test.level || ans.Cp != test.cp || ans.Percentage != test.percentage || ans.Rank != test.rank {
				t.Errorf("got %+v, want %+v", ans, test)
			}
		})
	}

	for _, test := range sortedTests {
		testName := fmt.Sprintf("sortedRanks[%d] = %f, %f, %d, %f, %d", test.pos, test.value, test.level, test.cp, test.percentage, test.rank)
		t.Run(testName, func(t *testing.T) {
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
		pos   int
		value int
	}{
		{0, 0},
		{170, 0},
		{1500, 770},
		{2500, 1346},
		{3500, 311},
	}

	var sortedTests = []struct {
		pos   int
		value float64
		level float64
		cp    int
		index int
	}{
		{0, 694353.519051347, 30, 804, 4095},
		{15, 675259.5521701364, 30, 791, 3822},
		{2547, 549931.3021919342, 30, 677, 349},
	}

	combinations, sortedRanks := calculateRanksCompact(PikachuStats, 1500, 30, 1)

	for _, test := range combinationTests {
		testName := fmt.Sprintf("combinations %d = %d", test.pos, test.value)
		t.Run(testName, func(t *testing.T) {
			ans := combinations[test.pos]
			if ans != test.value {
				t.Errorf("got %d, want %d", ans, test.value)
			}
		})
	}

	for _, test := range sortedTests {
		testName := fmt.Sprintf("sortedRanks[%d] = %f, %f, %d, %d", test.pos, test.value, test.level, test.cp, test.index)
		t.Run(testName, func(t *testing.T) {
			ans := sortedRanks[test.pos]
			if ans.Value != test.value || ans.Level != test.level || ans.Cp != test.cp || ans.Index != test.index {
				t.Errorf("got %+v, want %+v", ans, test)
			}
		})
	}
}

func BenchmarkCalculateRanksCompact(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = calculateRanksCompact(PikachuStats, 1500, 30, 1)
	}
}
