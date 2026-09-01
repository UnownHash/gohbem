package gohbem

import (
	"sync"
	"testing"
)

// TestReloadConcurrentWithQueries pins that reloading the masterfile
// (LoadPokemonData, as WatchPokemonData does on a changed remote file) and
// clearing the rank cache are safe while other goroutines query. Run under
// -race. Before the atomic snapshot change, LoadPokemonData unmarshaled into
// the live PokemonData maps and ClearCache overwrote the live sync.Map,
// crashing production consumers with "concurrent map read and map write"
// (UnownHash/Golbat#403).
func TestReloadConcurrentWithQueries(t *testing.T) {
	ohbem := Ohbem{Leagues: leagues, LevelCaps: levelCaps}
	if err := ohbem.LoadPokemonData("./test/master-test.json"); err != nil {
		t.Fatalf("LoadPokemonData: %v", err)
	}

	// Sanity checks: the queries below must reach the masterfile maps and the
	// rank cache, otherwise the concurrent phase is vacuous.
	if pvp, err := ohbem.QueryPvPRank(661, 0, 0, 1, 15, 15, 14, 1); err != nil || len(pvp) == 0 {
		t.Fatalf("QueryPvPRank sanity check failed: pvp=%v err=%v", pvp, err)
	}
	if _, err := ohbem.CalculateCp(661, 0, 0, 15, 15, 14, 40); err != nil {
		t.Fatalf("CalculateCp sanity check failed: %v", err)
	}

	stop := make(chan struct{})
	var wg sync.WaitGroup
	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-stop:
					return
				default:
				}
				if _, err := ohbem.QueryPvPRank(661, 0, 0, 1, 15, 15, 14, 1); err != nil {
					t.Errorf("QueryPvPRank during reload: %v", err)
					return
				}
				if _, err := ohbem.CalculateCp(661, 0, 0, 15, 15, 14, 40); err != nil {
					t.Errorf("CalculateCp during reload: %v", err)
					return
				}
			}
		}()
	}

	for i := 0; i < 5; i++ {
		if err := ohbem.LoadPokemonData("./test/master-test.json"); err != nil {
			t.Errorf("LoadPokemonData during queries: %v", err)
		}
		ohbem.ClearCache()
	}
	close(stop)
	wg.Wait()
}
