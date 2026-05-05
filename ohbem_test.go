package gohbem

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"
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

var levelCaps = []int{50, 51}

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
			combinations, _ := ohbem.calculateAllRanksCompact(&test.stats, test.cpCap)
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
		_, _ = ohbem.calculateAllRanksCompact(&PikachuStats, 5000)
	}
}

func BenchmarkCalculateAllRanksCompactCached(b *testing.B) {
	ohbem := Ohbem{Leagues: leagues, LevelCaps: levelCaps}
	_ = ohbem.LoadPokemonData("./test/master-test.json")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = ohbem.calculateAllRanksCompact(&PikachuStats, 5000)
	}
}

func TestCalculateTopRanks(t *testing.T) {
	ohbem := Ohbem{Leagues: leagues, LevelCaps: levelCaps}
	err := ohbem.LoadPokemonData("./test/master-test.json")
	if err != nil {
		t.Errorf("can't load MasterFile")
	}

	var tests = []struct {
		maxRank   int16
		pokemonId int
		form      int
		evolution int
		ivFloor   int
		league    string
		pos       int
		rank      int16
		level     float64
		value     float64
		a         int
		d         int
		s         int
		cap       float64
		capped    bool
	}{
		{5, 605, 0, 0, 0, "little", 0, 1, 14, 337248, 0, 14, 15, 50, true},
		{5, 605, 0, 0, 0, "little", 4, 5, 14, 333571, 1, 12, 15, 50, true},
		{5, 605, 0, 0, 0, "great", 0, 1, 50, 1710113, 8, 15, 15, 50, false},
		{5, 605, 0, 0, 0, "great", 10, 5, 50.5, 1709291, 7, 15, 15, 51, false},
	}

	for ix, test := range tests {
		testName := fmt.Sprintf("%d", ix)
		t.Run(testName, func(t *testing.T) {
			entries, _ := ohbem.CalculateTopRanks(test.maxRank, test.pokemonId, test.form, test.evolution, test.ivFloor)
			ans := entries[test.league][test.pos]
			if ans.Value != test.value || ans.Level != test.level || ans.Rank != test.rank || ans.Attack != test.a || ans.Defense != test.d || ans.Stamina != test.s || ans.Cap != test.cap || ans.Capped != test.capped {
				t.Errorf("got %+v, want %+v", ans, test)
			}
		})
	}
}

func BenchmarkCalculateTopRanks(b *testing.B) {
	ohbem := Ohbem{Leagues: leagues, LevelCaps: levelCaps}
	_ = ohbem.LoadPokemonData("./test/master-test.json")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = ohbem.CalculateTopRanks(500, 257, 0, 0, 1)
	}
}

func TestOhbem_CalculateCp(t *testing.T) {
	ohbem := Ohbem{}
	err := ohbem.LoadPokemonData("./test/master-test.json")
	if err != nil {
		t.Errorf("can't load MasterFile")
	}
	var tests = []struct {
		pokemonId int
		form      int
		evolution int
		a         int
		d         int
		s         int
		level     float64
		out       int
	}{
		{25, 0, 0, 1, 2, 2, 8, 167},
		{25, 2, 0, 1, 2, 2, 8, 167},
		{25, 2670, 0, 1, 2, 2, 8, 167},
		{3, 0, 1, 1, 2, 2, 8, 743},
	}
	for ix, test := range tests {
		testName := fmt.Sprintf("%d", ix)
		t.Run(testName, func(t *testing.T) {
			out, err := ohbem.CalculateCp(test.pokemonId, test.form, test.evolution, test.a, test.d, test.s, test.level)
			if out != test.out {
				t.Errorf("got %+v %+v, want %+v", out, err, test)
			}
		})
	}
}

func BenchmarkOhbem_CalculateCp(b *testing.B) {
	ohbem := Ohbem{Leagues: leagues, LevelCaps: levelCaps, DisableCache: true}
	_ = ohbem.LoadPokemonData("./test/master-test.json")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = ohbem.CalculateCp(257, 0, 0, 0, 10, 5, 22.5)
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
		isEmpty       bool
		outElem       int
		outRank       int16
		outValue      float64
		outPercentage float64
		outLevel      float64
		outCp         int
		outPokemonId  int
	}{
		{25, 0, 0, 1, 1, 2, 2, 8, "great", false, 0, 3309, 1587775, 0.9288, 27.5, 1475, 26},
		{25, 2, 0, 1, 1, 2, 2, 8, "ultra", false, 1, 4053, 2972156, 0.83553, 51, 2258, 26},
		{25, 2670, 0, 1, 1, 2, 2, 8, "great", true, 0, 3309, 1587775, 0.9288, 27.5, 1475, 26},
		{661, 0, 0, 1, 15, 15, 14, 1, "little", false, 0, 3287, 348805, 0.89401, 21.5, 490, 661},
		{661, 0, 0, 1, 15, 15, 14, 1, "master", false, 0, 1, 0, 1, 51, 0, 661},
		{661, 0, 0, 1, 15, 15, 14, 1, "master", false, 1, 1, 0, 1, 50, 0, 663},
		{661, 0, 0, 1, 15, 15, 14, 1, "great", false, 0, 1087, 1743985, 0.94736, 41.5, 1493, 662},
		{661, 0, 0, 1, 15, 15, 14, 1, "great", false, 1, 1328, 1743985, 0.94736, 41.5, 1493, 662},
		{661, 0, 0, 1, 15, 15, 14, 1, "great", false, 2, 2867, 1756548, 0.94144, 23.5, 1476, 663},
		{661, 0, 0, 1, 15, 15, 14, 1, "ultra", false, 0, 21, 3851769, 0.99275, 50, 2486, 663},
	}
	for ix, test := range tests {
		testName := fmt.Sprintf("%d", ix)
		t.Run(testName, func(t *testing.T) {
			var found bool
			entries, _ := ohbem.QueryPvPRank(test.pokemonId, test.form, test.costume, test.gender, test.a, test.d, test.s, test.level)
			if test.isEmpty && entries[test.outKey] == nil {
				return
			} else if entries[test.outKey] == nil {
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
		{127, 1, false},
		{127, 0, false},
		{150, 0, false},
		{150, 1, false},
		{150, 2, true},
		{666, 2, false},
	}

	for ix, test := range tests {
		testName := fmt.Sprintf("%d", ix)
		t.Run(testName, func(t *testing.T) {
			output, _ := ohbem.IsMegaUnreleased(test.pokemonId, test.form)
			if output != test.output {
				t.Errorf("%d:%d: got %t, want %t", test.pokemonId, test.form, output, test.output)
			}
		})
	}
}

func BenchmarkIsMegaUnreleased(b *testing.B) {
	ohbem := Ohbem{Leagues: leagues, LevelCaps: levelCaps}
	_ = ohbem.LoadPokemonData("./test/master-test.json")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = ohbem.IsMegaUnreleased(127, 1)
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
		caps   []int
		count  int
	}{
		{"master", []int{51}, 2},
		{"ultra", []int{51}, 1},
		{"master", []int{50, 51}, 3},
		{"great", []int{50, 51}, 3},
		{"ultra", []int{50, 51}, 1},
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

// TestFilterLevelCapsMerge ensures collapsed entries inherit Cap and Capped
// from later level caps (regression test for value-copy mutation bug).
func TestFilterLevelCapsMerge(t *testing.T) {
	// Synthetic entries: same pokemon/level/rank across two caps.
	// FilterLevelCaps should collapse the second into the first and update Cap+Capped.
	entries := []PokemonEntry{
		{Pokemon: 100, Level: 30, Rank: 5, Cap: 50, Capped: false},
		{Pokemon: 100, Level: 30, Rank: 5, Cap: 51, Capped: true},
	}
	ohbem := Ohbem{Leagues: leagues, LevelCaps: levelCaps}
	out := ohbem.FilterLevelCaps(entries, []int{50, 51})
	if len(out) != 1 {
		t.Fatalf("got %d entries, want 1", len(out))
	}
	if out[0].Cap != 51 {
		t.Errorf("got Cap=%v, want 51 (later cap should propagate)", out[0].Cap)
	}
	if !out[0].Capped {
		t.Errorf("got Capped=false, want true (Capped from later entry must propagate)")
	}
}

func BenchmarkFilterLevelCaps(b *testing.B) {
	ohbem := Ohbem{Leagues: leagues, LevelCaps: levelCaps}
	_ = ohbem.LoadPokemonData("./test/master-test.json")
	entries, _ := ohbem.QueryPvPRank(361, 0, 0, 1, 15, 15, 14, 1)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ohbem.FilterLevelCaps(entries["great"], []int{51})
	}
}

// TestWatchPokemonDataRestart verifies start→stop→start works without
// returning ErrWatcherStarted or panicking on a re-stop.
func TestWatchPokemonDataRestart(t *testing.T) {
	ohbem := Ohbem{Leagues: leagues, LevelCaps: levelCaps, WatcherInterval: time.Hour}
	if err := ohbem.LoadPokemonData("./test/master-test.json"); err != nil {
		t.Fatalf("load: %v", err)
	}
	if err := ohbem.WatchPokemonData(); err != nil {
		t.Fatalf("first start: %v", err)
	}
	if err := ohbem.StopWatchingPokemonData(); err != nil {
		t.Fatalf("first stop: %v", err)
	}
	if err := ohbem.WatchPokemonData(); err != nil {
		t.Fatalf("restart: %v", err)
	}
	if err := ohbem.StopWatchingPokemonData(); err != nil {
		t.Fatalf("second stop: %v", err)
	}
	// Third stop must return error, not panic on closed channel.
	if err := ohbem.StopWatchingPokemonData(); err != ErrNilChannel {
		t.Errorf("expected ErrNilChannel on extra stop, got %v", err)
	}
}

// TestFetchPokemonDataNon200 verifies FetchPokemonData surfaces non-200
// HTTP responses as ErrMasterFileFetch.
func TestFetchPokemonDataNon200(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "rate limited", http.StatusTooManyRequests)
	}))
	defer srv.Close()
	ohbem := Ohbem{Leagues: leagues, LevelCaps: levelCaps, MasterFileURL: srv.URL}
	if err := ohbem.FetchPokemonData(); err != ErrMasterFileFetch {
		t.Errorf("got %v, want ErrMasterFileFetch", err)
	}
}

// TestFetchPokemonDataMalformed verifies malformed JSON yields ErrMasterFileDecode.
func TestFetchPokemonDataMalformed(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("{not json"))
	}))
	defer srv.Close()
	ohbem := Ohbem{Leagues: leagues, LevelCaps: levelCaps, MasterFileURL: srv.URL}
	if err := ohbem.FetchPokemonData(); err != ErrMasterFileDecode {
		t.Errorf("got %v, want ErrMasterFileDecode", err)
	}
}

// TestSavePokemonDataAtomic verifies SavePokemonData round-trips and that
// a path under a fresh tmp dir works.
func TestSavePokemonDataAtomic(t *testing.T) {
	ohbem := Ohbem{Leagues: leagues, LevelCaps: levelCaps}
	if err := ohbem.LoadPokemonData("./test/master-test.json"); err != nil {
		t.Fatalf("load: %v", err)
	}
	dir := t.TempDir()
	path := filepath.Join(dir, "out.json")
	if err := ohbem.SavePokemonData(path); err != nil {
		t.Fatalf("save: %v", err)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("stat: %v", err)
	}
	other := Ohbem{Leagues: leagues, LevelCaps: levelCaps}
	if err := other.LoadPokemonData(path); err != nil {
		t.Fatalf("reload: %v", err)
	}
}

// FuzzQueryPvPRankBounds checks that out-of-range IV/level inputs return the
// documented sentinel error rather than panicking.
func FuzzQueryPvPRankBounds(f *testing.F) {
	ohbem := Ohbem{Leagues: leagues, LevelCaps: levelCaps}
	if err := ohbem.LoadPokemonData("./test/master-test.json"); err != nil {
		f.Fatalf("load: %v", err)
	}
	f.Add(25, 0, 0, 1, 5, 5, 5, 1.0)
	f.Add(25, 0, 0, 1, -1, 5, 5, 1.0)
	f.Add(25, 0, 0, 1, 16, 5, 5, 1.0)
	f.Add(25, 0, 0, 1, 5, 5, 5, 0.0)
	f.Fuzz(func(t *testing.T, pid, form, costume, gender, a, d, s int, level float64) {
		// Just ensure no panic; error vs success depends on inputs.
		_, _ = ohbem.QueryPvPRank(pid, form, costume, gender, a, d, s, level)
	})
}

// FuzzCalculateCpBounds is the equivalent fuzz for CalculateCp inputs.
func FuzzCalculateCpBounds(f *testing.F) {
	ohbem := Ohbem{Leagues: leagues, LevelCaps: levelCaps}
	if err := ohbem.LoadPokemonData("./test/master-test.json"); err != nil {
		f.Fatalf("load: %v", err)
	}
	f.Add(25, 0, 0, 5, 5, 5, 1.0)
	f.Add(25, 0, 0, -1, 5, 5, 1.0)
	f.Fuzz(func(t *testing.T, pid, form, evolution, a, d, s int, level float64) {
		_, _ = ohbem.CalculateCp(pid, form, evolution, a, d, s, level)
	})
}
