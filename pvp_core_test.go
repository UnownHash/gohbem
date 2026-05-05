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
		stats   *PokemonStats
		stamina int
		level   float64
		output  int
	}{
		{&PikachuStats, 10, 10, 51},
		{&PikachuStats, 12, 25.5, 82},
		{&PikachuStats, 98, 10, 88},
		{&PikachuStats, 100, 30, 154},
		{&PikachuStats, 97, 35.5, 159},

		{&ElgyemStats, 10, 10, 65},
		{&ElgyemStats, 12, 25.5, 106},
		{&ElgyemStats, 98, 10, 103},
		{&ElgyemStats, 100, 30, 179},
		{&ElgyemStats, 97, 35.5, 185},
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
		_ = calculateHp(&PikachuStats, 95, 30)
	}
}

func TestCalculateCp(t *testing.T) {
	var tests = []struct {
		stats   *PokemonStats
		attack  int
		defense int
		stamina int
		level   float64
		output  int
	}{
		{&PikachuStats, 0, 0, 0, 1, 10},
		{&PikachuStats, 15, 15, 15, 30, 804},
		{&PikachuStats, 10, 2, 15, 30, 725},
		{&PikachuStats, 15, 15, 15, 34.5, 864},

		{&ElgyemStats, 0, 0, 0, 1, 15},
		{&ElgyemStats, 15, 15, 15, 30, 1187},
		{&ElgyemStats, 10, 2, 15, 30, 1084},
		{&ElgyemStats, 15, 15, 15, 34.5, 1276},
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
		_ = calculateCp(&PikachuStats, 10, 5, 2, 15)
	}
}

func TestCalculatePvPStat(t *testing.T) {
	var tests = []struct {
		stats       *PokemonStats
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
		{&PikachuStats, 0, 0, 0, 40, 0, 1, 950.0466549389662, 1, 10},
		{&PikachuStats, 5, 10, 15, 300, 20, 1, 154064.78899667264, 12, 289},
		{&PikachuStats, 0, 0, 0, 100, 10, 1, 28985.670041102363, 5, 97},
		{&PikachuStats, 15, 15, 15, 5000, 50, 1, 1045164.7410539213, 50, 1060},

		{&ElgyemStats, 0, 0, 0, 40, 0, 1, 1700.04628357754, 1, 15},
		{&ElgyemStats, 5, 10, 15, 300, 20, 1, 142181.60313125828, 8, 286},
		{&ElgyemStats, 0, 0, 0, 100, 10, 1, 28162.402883172115, 3.5, 100},
		{&ElgyemStats, 15, 15, 15, 5000, 50, 1, 1786849.4577316528, 50, 1566},
	}

	for ix, test := range tests {
		testName := fmt.Sprintf("%d", ix)
		testOutput := Ranking{Value: test.outputValue, Level: test.outputLevel, Cp: test.outputCp}
		t.Run(testName, func(t *testing.T) {
			var ans PvPRankingStats
			_ = calculatePvPStat(&ans, test.stats, test.attack, test.defense, test.stamina, test.cap, test.lvCap, test.minLevel)
			if ans.Value != test.outputValue || ans.Level != test.outputLevel || ans.Cp != test.outputCp {
				t.Errorf("got %+v, want %+v", ans, testOutput)
			}
		})
	}
}

func BenchmarkCalculatePvPStat(b *testing.B) {
	var stat PvPRankingStats
	for i := 0; i < b.N; i++ {
		_ = calculatePvPStat(&stat, &PikachuStats, 10, 5, 2, 15, 40, 1)
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
		{1500, 30, 1, 0, 694353.519051347, 30, 804, 4095},
		{1500, 30, 1, 15, 675259.5521701364, 30, 791, 3822},
		{1500, 30, 1, 2547, 549931.3021919342, 30, 677, 349},
	}

	for ix, test := range combinationTests {
		testName := fmt.Sprintf("combinations/%d", ix)
		t.Run(testName, func(t *testing.T) {
			combinations, _ := calculateRanksCompact(&PikachuStats, test.cpCap, test.lvCap, RankingComparatorDefault, test.ivFloor)
			ans := combinations[test.pos]
			if ans != test.rank {
				t.Errorf("got %d, want %d", ans, test.rank)
			}
		})
	}

	for ix, test := range sortedTests {
		testName := fmt.Sprintf("sortedRanks/%d", ix)
		t.Run(testName, func(t *testing.T) {
			_, sortedRanks := calculateRanksCompact(&PikachuStats, test.cpCap, test.lvCap, RankingComparatorDefault, test.ivFloor)
			ans := sortedRanks[test.pos]
			if ans.Value != test.value || ans.Level != test.level || ans.Cp != test.cp || ans.Index != test.index {
				t.Errorf("got %+v, want %+v", ans, test)
			}
		})
	}
}

func BenchmarkCalculateRanksCompact(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = calculateRanksCompact(&PikachuStats, 1500, 50, RankingComparatorDefault, 0)
	}
}
