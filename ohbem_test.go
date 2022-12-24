package ohbemgo

import (
	"fmt"
	"testing"
)

var leagues = Leagues{
	"little": {
		Cap:    500,
		Little: true,
	},
	"great": {
		Cap:    1500,
		Little: true,
	},
	"ultra": {
		Cap:    2500,
		Little: false,
	},
}
var levelCaps = []float64{50, 51}

func TestCalculateTopRanks(t *testing.T) {
	ohbem := Ohbem{Leagues: leagues, LevelCaps: levelCaps}
	ohbem.LoadPokemonData("./test/master-test.json")

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

func TestQueryPvPRank(t *testing.T) {
	ohbem := Ohbem{Leagues: leagues, LevelCaps: levelCaps}
	ohbem.LoadPokemonData("./test/master-test.json")

	var tests = []struct {
		pokemonId int
		form      int
		costume   int
		gender    int
		a         int
		d         int
		s         int
		level     float64
	}{
		{5, 0, 0, 0, 15, 10, 5, 19.5},
	}
	// pokemonId int, form int, costume int, gender int, attack int, defense int, stamina int, level float64) (map[string][]PokemonEntry, error
	for _, test := range tests {
		testName := fmt.Sprintf("%+v, %d", test.pokemonId, test.form)
		t.Run(testName, func(t *testing.T) {
			entries, _ := ohbem.QueryPvPRank(test.pokemonId, test.form, test.costume, test.gender, test.a, test.d, test.s, test.level)
			fmt.Println(entries)
			//ans := &combinations[test.level][test.a][test.d][test.s]
			//if ans.Value != test.outValue || ans.Level != test.outLevel || ans.Cp != test.outCp || ans.Rank != test.outRank {
			//	t.Errorf("got %+v, want %+v", ans, test)
			//}
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
