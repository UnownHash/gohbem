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

	for ix, test := range tests {
		testName := fmt.Sprintf("%d", ix)
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

func BenchmarkCalculateAllRanksCompact(b *testing.B) {
	ohbem := Ohbem{Leagues: leagues, LevelCaps: levelCaps, DisableCache: true}
	_ = ohbem.LoadPokemonData("./test/master-test.json")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = ohbem.CalculateAllRanksCompact(PikachuStats, 5000)
	}
}

func BenchmarkCalculateAllRanksCompactCached(b *testing.B) {
	ohbem := Ohbem{Leagues: leagues, LevelCaps: levelCaps}
	_ = ohbem.LoadPokemonData("./test/master-test.json")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = ohbem.CalculateAllRanksCompact(PikachuStats, 5000)
	}
}

func TestCalculateAllRanks(t *testing.T) {
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

	for ix, test := range tests {
		testName := fmt.Sprintf("%d", ix)
		t.Run(testName, func(t *testing.T) {
			combinations, _ := ohbem.CalculateAllRanks(PikachuStats, test.cpCap)
			ans := &combinations[test.level][test.a][test.d][test.s]
			if ans.Value != test.outValue || ans.Level != test.outLevel || ans.Cp != test.outCp || ans.Rank != test.outRank {
				t.Errorf("got %+v, want %+v", ans, test)
			}
		})
	}
}

func BenchmarkCalculateAllRanks(b *testing.B) {
	ohbem := Ohbem{Leagues: leagues, LevelCaps: levelCaps}
	_ = ohbem.LoadPokemonData("./test/master-test.json")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = ohbem.CalculateAllRanks(PikachuStats, 5000)
	}
}

func TestCalculateTopRanks(t *testing.T) {
	t.SkipNow()
}

func BenchmarkCalculateTopRanks(b *testing.B) {
	ohbem := Ohbem{Leagues: leagues, LevelCaps: levelCaps}
	_ = ohbem.LoadPokemonData("./test/master-test.json")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ohbem.CalculateTopRanks(500, 257, 0, 0, 1)
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
	for ix, test := range tests {
		testName := fmt.Sprintf("%d", ix)
		t.Run(testName, func(t *testing.T) {
			var found bool
			entries, _ := ohbem.QueryPvPRank(test.pokemonId, test.form, test.costume, test.gender, test.a, test.d, test.s, test.level)
			if entries[test.outKey] == nil {
				t.Errorf("missing %s in entires", test.outKey)
			} else {
				for _, ans := range entries[test.outKey] {
					if ans.Value == test.outValue && ans.Percentage == test.outPercentage && ans.Rank == test.outRank && ans.Level == test.outLevel && ans.Cp == test.outCp {
						found = true
					}
				}
				if !found {
					t.Errorf("entries are missing %+v", test)
				}
			}
		})
	}
}

func BenchmarkQueryPvPRank(b *testing.B) {
	ohbem := Ohbem{Leagues: leagues, LevelCaps: levelCaps, DisableCache: true}
	_ = ohbem.LoadPokemonData("./test/master-test.json")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = ohbem.QueryPvPRank(257, 0, 0, 0, 10, 5, 0, 22.5)
	}
}

func BenchmarkQueryPvPRankCached(b *testing.B) {
	ohbem := Ohbem{Leagues: leagues, LevelCaps: levelCaps}
	_ = ohbem.LoadPokemonData("./test/master-test.json")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = ohbem.QueryPvPRank(257, 0, 0, 0, 10, 5, 0, 22.5)
	}
}

func TestFindBaseStats(t *testing.T) {
	ohbem := Ohbem{Leagues: leagues, LevelCaps: levelCaps}
	err := ohbem.LoadPokemonData("./test/master-test.json")
	if err != nil {
		t.Errorf("can't load MasterFile")
	}

	var tests = []struct {
		pokemonId int
		form      int
		evolution int
		attack    int
		defense   int
		stamina   int
	}{
		{0, 1, 1, 0, 0, 0},
		{15, 1, 1, 303, 148, 163},
		{15, 1, 0, 169, 130, 163},
		{255, 0, 0, 130, 87, 128},
	}

	for ix, test := range tests {
		testName := fmt.Sprintf("%d", ix)
		t.Run(testName, func(t *testing.T) {
			output, _ := ohbem.FindBaseStats(test.pokemonId, test.form, test.evolution)
			if output.Attack != test.attack || output.Defense != test.defense || output.Stamina != test.stamina {
				t.Errorf("got %+v, want %+v", output, test)
			}
		})
	}
}

func BenchmarkFindBaseStats(b *testing.B) {
	ohbem := Ohbem{Leagues: leagues, LevelCaps: levelCaps}
	_ = ohbem.LoadPokemonData("./test/master-test.json")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = ohbem.FindBaseStats(257, 1, 0)
	}
}

func TestIsMegaUnreleased(t *testing.T) {
	ohbem := Ohbem{Leagues: leagues, LevelCaps: levelCaps}
	err := ohbem.LoadPokemonData("./test/master-test.json")
	if err != nil {
		t.Errorf("can't load MasterFile")
	}

	var tests = []struct {
		pokemonId int
		form      int
		output    bool
	}{
		{0, 0, false},
		{127, 1, true},
		{127, 0, false},
		{150, 0, false},
		{150, 1, false},
		{150, 2, true},
	}

	for ix, test := range tests {
		testName := fmt.Sprintf("%d", ix)
		t.Run(testName, func(t *testing.T) {
			output := ohbem.IsMegaUnreleased(test.pokemonId, test.form)
			if output != test.output {
				t.Errorf("got %t, want %t", output, test.output)
			}
		})
	}
}

func BenchmarkIsMegaUnreleased(b *testing.B) {
	ohbem := Ohbem{Leagues: leagues, LevelCaps: levelCaps}
	_ = ohbem.LoadPokemonData("./test/master-test.json")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ohbem.IsMegaUnreleased(127, 1)
	}
}

func TestFilterLevelCaps(t *testing.T) {
	ohbem := Ohbem{Leagues: leagues, LevelCaps: levelCaps}
	err := ohbem.LoadPokemonData("./test/master-test.json")
	if err != nil {
		t.Errorf("can't load MasterFile")
	}
	entries, _ := ohbem.QueryPvPRank(661, 0, 0, 1, 15, 15, 14, 1)

	var tests = []struct {
		league string
		caps   []float64
		count  int
	}{
		{"master", []float64{51}, 2},
		{"ultra", []float64{51}, 1},
		{"master", []float64{50, 51}, 3},
		{"great", []float64{50, 51}, 3},
		{"ultra", []float64{50, 51}, 1},
	}

	for ix, test := range tests {
		testName := fmt.Sprintf("%d", ix)
		t.Run(testName, func(t *testing.T) {
			output := ohbem.FilterLevelCaps(entries[test.league], test.caps)
			if len(output) != test.count {
				t.Errorf("got %d, want %d", len(output), test.count)
			}
		})
	}
}

func BenchmarkFilterLevelCaps(b *testing.B) {
	ohbem := Ohbem{Leagues: leagues, LevelCaps: levelCaps}
	_ = ohbem.LoadPokemonData("./test/master-test.json")
	entries, _ := ohbem.QueryPvPRank(361, 0, 0, 1, 15, 15, 14, 1)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ohbem.FilterLevelCaps(entries["great"], []float64{51})
	}
}
