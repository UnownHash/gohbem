package ohbemgo

import (
	"fmt"
	"testing"
)

var leagues = map[string]League{
	"little": {
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

var levelCaps = []float64{50, 51}

func TestCalculateTopRanks(t *testing.T) {
	ohbem := Ohbem{Leagues: leagues, LevelCaps: levelCaps}
	err := ohbem.LoadPokemonData("./test/master-test.json")
	if err != nil {
		t.Errorf("can't load MasterFile")
	}

	var tests = []struct {
		stats         PokemonStats
		level         int
		cpCap         int
		a             int
		d             int
		s             int
		outValue      float64
		outLevel      float64
		outCp         int
		outPercentage float64
		outRank       int16
	}{
		{PikachuStats, 50, 300, 0, 0, 0, 155813.01965332002, 14.5, 299, 0.93235, 1105},
	}

	for _, test := range tests {
		testName := fmt.Sprintf("%+v, %d", test.stats, test.cpCap)
		t.Run(testName, func(t *testing.T) {
			combinations, _ := ohbem.CalculateAllRanks(PikachuStats, test.cpCap)
			ans := &combinations[test.level][test.a][test.d][test.s]
			if ans.Value != test.outValue || ans.Level != test.outLevel || ans.Cp != test.outCp || ans.Rank != test.outRank {
				t.Errorf("got %+v, want %+v", ans, test)
			}
		})
	}
}

func TestCalculateAllRanksCompact(t *testing.T) {
	ohbem := Ohbem{Leagues: leagues, LevelCaps: levelCaps}
	err := ohbem.LoadPokemonData("./test/master-test.json")
	if err != nil {
		t.Errorf("can't load MasterFile")
	}

	var tests = []struct {
		stats  PokemonStats
		cpCap  int
		lvCap  int
		pos    int
		value  int16
		topVal float64
	}{
		{ElgyemStats, 500, 50, 0, 2144, 337248.95363088587},
		{ElgyemStats, 500, 50, 100, 303, 337248.95363088587},
		{ElgyemStats, 500, 50, 2007, 1149, 337248.95363088587},
		{ElgyemStats, 1500, 50, 0, 4096, 1710113.5914486984},
		{ElgyemStats, 1500, 50, 540, 3551, 1710113.5914486984},
		{ElgyemStats, 1500, 51, 540, 3529, 1720993.4871909665},
	}

	for _, test := range tests {
		testName := fmt.Sprintf("%+v, %d", test.stats, test.cpCap)
		t.Run(testName, func(t *testing.T) {
			combinations, _ := ohbem.CalculateAllRanksCompact(test.stats, test.cpCap)
			comb := combinations[test.lvCap].Combinations
			topValue := combinations[test.lvCap].TopValue
			if comb[test.pos] != test.value || topValue != test.topVal {
				t.Errorf("got %d, want %d / got %f, want %f", comb[test.pos], test.value, topValue, test.topVal)
			}
		})
	}
}

func TestQueryPvPRank(t *testing.T) {
	ohbem := Ohbem{Leagues: leagues, LevelCaps: levelCaps}
	err := ohbem.LoadPokemonData("./test/master-test.json")
	if err != nil {
		t.Errorf("can't load MasterFile")
	}

	var tests = []struct {
		pokemonId     int
		form          int
		costume       int
		gender        int
		a             int
		d             int
		s             int
		level         float64
		outKey        string
		outElem       int
		outRank       int16
		outValue      float64
		outPercentage float64
		outLevel      float64
		outCp         int
		outPokemonId  int
	}{
		{661, 0, 0, 1, 15, 15, 14, 1, "little", 0, 3287, 348805, 0.89401, 21.5, 490, 661},
		{661, 0, 0, 1, 15, 15, 14, 1, "master", 0, 1, 0, 1, 51, 0, 661},
		{661, 0, 0, 1, 15, 15, 14, 1, "master", 1, 1, 0, 1, 50, 0, 663},
		{661, 0, 0, 1, 15, 15, 14, 1, "great", 0, 1087, 1743985, 0.94736, 41.5, 1493, 662},
		{661, 0, 0, 1, 15, 15, 14, 1, "great", 1, 1328, 1743985, 0.94736, 41.5, 1493, 662},
		{661, 0, 0, 1, 15, 15, 14, 1, "great", 2, 2867, 1756548, 0.94144, 23.5, 1476, 663},
		{661, 0, 0, 1, 15, 15, 14, 1, "ultra", 0, 21, 3851769, 0.99275, 50, 2486, 663},
	}
	// pokemonId int, form int, costume int, gender int, attack int, defense int, stamina int, level float64) (map[string][]PokemonEntry, error
	for _, test := range tests {
		testName := fmt.Sprintf("%+v, %d", test.pokemonId, test.form)
		t.Run(testName, func(t *testing.T) {
			entries, _ := ohbem.QueryPvPRank(test.pokemonId, test.form, test.costume, test.gender, test.a, test.d, test.s, test.level)
			if entries[test.outKey] == nil {
				t.Errorf("missing %s in entires", test.outKey)
			} else {
				ans := entries[test.outKey][test.outElem]
				if ans.Value != test.outValue || ans.Percentage != test.outPercentage || ans.Rank != test.outRank || ans.Level != test.outLevel || ans.Cp != test.outCp {
					t.Errorf("got %+v, want %+v", ans, test)
				}
			}
		})
	}
}

func BenchmarkCalculateAllRanks(b *testing.B) {
	ohbem := Ohbem{Leagues: leagues, LevelCaps: levelCaps, DisableCache: true}

	for i := 0; i < b.N; i++ {
		_, _ = ohbem.CalculateAllRanks(PikachuStats, 5000)
	}
}

func BenchmarkCalculateAllRanksCached(b *testing.B) {
	ohbem := Ohbem{Leagues: leagues, LevelCaps: levelCaps}

	for i := 0; i < b.N; i++ {
		_, _ = ohbem.CalculateAllRanks(PikachuStats, 5000)
	}
}

func BenchmarkCalculateAllRanksCompact(b *testing.B) {
	ohbem := Ohbem{Leagues: leagues, LevelCaps: levelCaps, DisableCache: true}

	for i := 0; i < b.N; i++ {
		_, _ = ohbem.CalculateAllRanksCompact(PikachuStats, 5000)
	}
}

func BenchmarkCalculateAllRanksCompactCached(b *testing.B) {
	ohbem := Ohbem{Leagues: leagues, LevelCaps: levelCaps}

	for i := 0; i < b.N; i++ {
		_, _ = ohbem.CalculateAllRanksCompact(PikachuStats, 5000)
	}
}

func BenchmarkCalculateTopRanks(b *testing.B) {
	ohbem := Ohbem{Leagues: leagues, LevelCaps: levelCaps}

	for i := 0; i < b.N; i++ {
		_ = ohbem.CalculateTopRanks(500, 257, 0, 0, 1)
	}
}

func BenchmarkQueryPvPRank(b *testing.B) {
	ohbem := Ohbem{Leagues: leagues, LevelCaps: levelCaps, DisableCache: true}

	for i := 0; i < b.N; i++ {
		_, _ = ohbem.QueryPvPRank(257, 0, 0, 0, 10, 5, 0, 22.5)
	}
}

func BenchmarkQueryPvPRankCached(b *testing.B) {
	ohbem := Ohbem{Leagues: leagues, LevelCaps: levelCaps}

	for i := 0; i < b.N; i++ {
		_, _ = ohbem.QueryPvPRank(257, 0, 0, 0, 10, 5, 0, 22.5)
	}
}

func BenchmarkIsMegaUnreleased(b *testing.B) {
	ohbem := Ohbem{Leagues: leagues, LevelCaps: levelCaps}

	for i := 0; i < b.N; i++ {
		_ = ohbem.IsMegaUnreleased(257, 1)
	}
}

func BenchmarkFilterLevelCaps(b *testing.B) {
	ohbem := Ohbem{Leagues: leagues, LevelCaps: levelCaps}

	for i := 0; i < b.N; i++ {
		_ = ohbem.FilterLevelCaps([]PokemonEntry{}, levelCaps)
	}
}
