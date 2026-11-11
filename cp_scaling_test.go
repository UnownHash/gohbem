package gohbem

import (
	"fmt"
	"math"
	"testing"
)

var MiniorCoreStats = PokemonStats{Attack: 218, Defense: 131, Stamina: 155}
var MiniorMeteorStats = PokemonStats{Attack: 116, Defense: 194, Stamina: 155}
var AegislashShieldStats = PokemonStats{Attack: 97, Defense: 272, Stamina: 155}
var AegislashBladeStats = PokemonStats{Attack: 272, Defense: 97, Stamina: 155}
var MegaCharizardYStats = PokemonStats{Attack: 319, Defense: 212, Stamina: 186}

const (
	miniorId        = 774
	miniorCoreRed   = 2712
	miniorMeteor    = 3380
	aegislashId     = 681
	aegislashShield = 2639
	aegislashBlade  = 2640
	charizardId     = 6
	megaY           = 3
)

// battleFormHp recovers the HP that calculateBattleFormStats put into the stat product.
func battleFormHp(stat *PvPRankingStats, battle *PokemonStats, defense int) float64 {
	return math.Round(stat.Value / (stat.Attack * float64(defense+battle.Defense) * calculateCpMultiplier(stat.Level)))
}

// Player readings collected in pvpoke/pvpoke#404 and jgmcelwain/go-iv#50.
func TestBattleFormReadings(t *testing.T) {
	var tests = []struct {
		name                     string
		stored, battle           PokemonStats
		attack, defense, stamina int
		storedLevel              float64
		cap                      int
		wantBattleLevel          float64
		wantBattleCp             int
	}{
		{"Minior 2/8/13 Lv25 in Great League", MiniorCoreStats, MiniorMeteorStats, 2, 8, 13, 25, 1500, 48, 1498},
		{"Minior 2/8/13 Lv25 in Ultra League", MiniorCoreStats, MiniorMeteorStats, 2, 8, 13, 25, 2500, 37.5, 1309},
		{"Minior 14/15/9 Lv34 in Ultra League", MiniorCoreStats, MiniorMeteorStats, 14, 15, 9, 34, 2500, 55, 1802},
		{"Minior 15/15/13 Lv50 stopping at the level 55 multiplier", MiniorCoreStats, MiniorMeteorStats, 15, 15, 13, 50, 2500, 55, 1837},
		{"Aegislash 15/15/15 Lv20 in Great League", AegislashShieldStats, AegislashBladeStats, 15, 15, 15, 20, 1500, 11.5, 812},
		{"Aegislash 15/15/15 Lv20 in Ultra League", AegislashShieldStats, AegislashBladeStats, 15, 15, 15, 20, 2500, 10.5, 742},
		{"Aegislash 15/15/15 Lv20 in Master League", AegislashShieldStats, AegislashBladeStats, 15, 15, 15, 20, 10000, 20.5, 1448},
		// The theoretical Little Cup row from pvpoke#404.
		{"Aegislash 15/15/15 Lv11 in Little Cup", AegislashShieldStats, AegislashBladeStats, 15, 15, 15, 11, 500, 7, 482},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var stat PvPRankingStats
			calculateBattleFormStats(&stat, &test.battle, &test.stored, test.attack, test.defense, test.stamina, test.storedLevel, test.cap)
			if stat.Level != test.wantBattleLevel || stat.Cp != test.wantBattleCp {
				t.Errorf("got Lv%v %d CP, want Lv%v %d CP", stat.Level, stat.Cp, test.wantBattleLevel, test.wantBattleCp)
			}
			wantHp := float64(calculateHp(&test.stored, test.stamina, test.storedLevel))
			if hp := battleFormHp(&stat, &test.battle, test.defense); hp != wantHp {
				t.Errorf("got %v HP, want the stored form's %v HP", hp, wantHp)
			}
		})
	}
}

// In-game readings from jgmcelwain/go-iv#48: Core IVs and level, CP cap, Core CP, then Meteor CP in battle.
func TestBattleFormGoIv48Readings(t *testing.T) {
	var tests = []struct {
		attack, defense, stamina int
		level                    float64
		cap                      int
		coreCp                   int
		meteorCp                 int
	}{
		{2, 12, 14, 24.5, 1500, 1495, 1499},
		{10, 6, 10, 24.5, 1500, 1498, 1493},
		{8, 6, 5, 25, 1500, 1492, 1492},
		{3, 14, 14, 24, 1500, 1481, 1492},
		{7, 13, 13, 24, 1500, 1498, 1496},
		{1, 3, 13, 25.5, 1500, 1495, 1494},
		// Cores showing exactly 1500 CP fight as Meteors over the cap, so cap levels use CP before flooring.
		{12, 1, 7, 25, 1500, 1500, 1502},
		{12, 14, 12, 23.5, 1500, 1500, 1508},
		{7, 13, 13, 24, 10000, 1498, 1002},
		{10, 6, 10, 24.5, 10000, 1498, 1021},
	}

	for _, test := range tests {
		testName := fmt.Sprintf("%d/%d/%d Lv%v under %d CP", test.attack, test.defense, test.stamina, test.level, test.cap)
		t.Run(testName, func(t *testing.T) {
			if cp := calculateCp(&MiniorCoreStats, test.attack, test.defense, test.stamina, test.level); cp != test.coreCp {
				t.Errorf("got Core %d CP, want %d", cp, test.coreCp)
			}
			var stat PvPRankingStats
			calculateBattleFormStats(&stat, &MiniorMeteorStats, &MiniorCoreStats, test.attack, test.defense, test.stamina, test.level, test.cap)
			if stat.Cp != test.meteorCp {
				t.Errorf("got Meteor %d CP, want %d", stat.Cp, test.meteorCp)
			}
		})
	}
}

// A 9/3/1 Lv25 Minior's shell broke after 52 one-damage Tackles in go-iv#48, half its Core HP of 104.
// The same reading saw 1492 CP as a Meteor, the one known miss in both pvpoke#404 and go-iv#50: the formula gives 1474.
func TestBattleFormGoIv48Minior931(t *testing.T) {
	var stat PvPRankingStats
	calculateBattleFormStats(&stat, &MiniorMeteorStats, &MiniorCoreStats, 9, 3, 1, 25, 1500)
	if hp := battleFormHp(&stat, &MiniorMeteorStats, 3); hp != 104 {
		t.Errorf("got %v HP, want 104", hp)
	}
	if stat.Cp != 1474 {
		t.Errorf("got %d CP, want the formula's 1474 (in-game reading: 1492)", stat.Cp)
	}
}

func TestMegaDownscaling(t *testing.T) {
	// The Mega Charizard Y example from the report: 5037 CP at Lv50 scales down to the Great League cap.
	if cp := calculateCp(&MegaCharizardYStats, 15, 15, 15, 50); cp != 5037 {
		t.Errorf("got %d CP at Lv50, want 5037", cp)
	}
	var stat PvPRankingStats
	if err := calculatePvPStat(&stat, &MegaCharizardYStats, 15, 15, 15, 1500, 50, 1); err != nil || stat.Level != 11.5 || stat.Cp != 1464 {
		t.Errorf("got Lv%v %d CP (%v), want Lv11.5 1464 CP", stat.Level, stat.Cp, err)
	}
}

// findEntries returns the entries of a league for a form and temp evolution.
func findEntries(entries []PokemonEntry, form, evolution int) []PokemonEntry {
	var result []PokemonEntry
	for _, entry := range entries {
		if entry.Form == form && entry.Evolution == evolution {
			result = append(result, entry)
		}
	}
	return result
}

// expectedRank ranks one IV combination among all of them, independently of the cached rank tables.
func expectedRank(stats, stored *PokemonStats, cpCap int, lvCap float64, attack, defense, stamina int) int16 {
	var mine PvPRankingStats
	if calculateFormPvPStat(&mine, stats, stored, attack, defense, stamina, cpCap, lvCap, 1) != nil {
		return 0
	}
	rank := int16(1)
	for a := 0; a <= 15; a++ {
		for d := 0; d <= 15; d++ {
			for s := 0; s <= 15; s++ {
				var other PvPRankingStats
				if calculateFormPvPStat(&other, stats, stored, a, d, s, cpCap, lvCap, 1) == nil && RankingComparatorDefault(&other, &mine) < 0 {
					rank++
				}
			}
		}
	}
	return rank
}

func TestQueryPvPRankBattleForms(t *testing.T) {
	ohbem := Ohbem{Leagues: leagues, LevelCaps: levelCaps}
	if err := ohbem.LoadPokemonData("./test/master-test.json"); err != nil {
		t.Fatalf("can't load MasterFile")
	}

	t.Run("Minior Core 2/8/13 Lv25 in Great League", func(t *testing.T) {
		attack, defense, stamina := 2, 8, 13
		entries, err := ohbem.QueryPvPRank(miniorId, miniorCoreRed, 0, 0, attack, defense, stamina, 25)
		if err != nil {
			t.Fatal(err)
		}
		core := findEntries(entries["great"], miniorCoreRed, 0)
		if len(core) != 1 || core[0].Level != 25 || core[0].Cp != 1499 || core[0].Cap != 50 || !core[0].Capped {
			t.Errorf("got Core entries %+v, want one capped Lv25 1499 CP entry", core)
		}
		meteor := findEntries(entries["great"], miniorMeteor, 0)
		if len(meteor) != 1 || meteor[0].Level != 48 || meteor[0].Cp != 1498 || meteor[0].Cap != 50 || !meteor[0].Capped {
			t.Fatalf("got Meteor entries %+v, want one capped Lv48 1498 CP entry", meteor)
		}
		if want := expectedRank(&MiniorMeteorStats, &MiniorCoreStats, 1500, 50, attack, defense, stamina); meteor[0].Rank != want {
			t.Errorf("got Meteor rank %d, want %d", meteor[0].Rank, want)
		}
		if len(entries["master"]) != 0 {
			t.Errorf("got master entries %+v, want none", entries["master"])
		}
	})

	t.Run("Aegislash ranks Blade in Ultra League where Shield never reaches the cap", func(t *testing.T) {
		attack, defense, stamina := 0, 15, 15
		entries, err := ohbem.QueryPvPRank(aegislashId, aegislashShield, 0, 0, attack, defense, stamina, 1)
		if err != nil {
			t.Fatal(err)
		}
		if shield := findEntries(entries["ultra"], aegislashShield, 0); len(shield) != 0 {
			t.Errorf("got Shield entries %+v, want none", shield)
		}
		blade := findEntries(entries["ultra"], aegislashBlade, 0)
		if len(blade) != 2 || blade[1].Cap != 51 || blade[1].Level != 29 || blade[1].Cp != 1942 {
			t.Fatalf("got Blade entries %+v, want a Lv51 cap entry at Lv29 1942 CP", blade)
		}
		for _, entry := range blade {
			if want := expectedRank(&AegislashBladeStats, &AegislashShieldStats, 2500, entry.Cap, attack, defense, stamina); entry.Rank != want {
				t.Errorf("got Blade rank %d at Lv%v cap, want %d", entry.Rank, entry.Cap, want)
			}
		}
	})

	t.Run("Doublade evolves into both Aegislash forms", func(t *testing.T) {
		entries, err := ohbem.QueryPvPRank(680, 0, 0, 0, 15, 15, 15, 20)
		if err != nil {
			t.Fatal(err)
		}
		if len(findEntries(entries["great"], aegislashShield, 0)) == 0 || len(findEntries(entries["great"], aegislashBlade, 0)) == 0 {
			t.Errorf("got great entries %+v, want both Shield and Blade", entries["great"])
		}
	})

	t.Run("Meteor queries rank as before", func(t *testing.T) {
		entries, err := ohbem.QueryPvPRank(miniorId, miniorMeteor, 0, 0, 2, 8, 13, 25)
		if err != nil {
			t.Fatal(err)
		}
		for league, results := range entries {
			for _, entry := range results {
				if entry.Form != miniorMeteor {
					t.Errorf("got %s entry %+v, want only Meteor entries", league, entry)
				}
			}
		}
	})
}

// A Meteor can fight at Lv55 under several level caps while its Core's level, and so its HP, still grows.
func TestQueryPvPRankBattleFormLevelCaps(t *testing.T) {
	ohbem := Ohbem{Leagues: leagues, LevelCaps: []int{40, 50, 51}}
	if err := ohbem.LoadPokemonData("./test/master-test.json"); err != nil {
		t.Fatalf("can't load MasterFile")
	}
	entries, err := ohbem.QueryPvPRank(miniorId, miniorCoreRed, 0, 0, 9, 14, 15, 1)
	if err != nil {
		t.Fatal(err)
	}
	meteor := findEntries(entries["ultra"], miniorMeteor, 0)
	var tests = []struct {
		lvCap     int
		wantValue float64
	}{
		{40, 2608624},
		{50, 2764363},
		{51, 2764363},
	}
	for _, test := range tests {
		filtered := ohbem.FilterLevelCaps(meteor, []int{test.lvCap})
		if len(filtered) != 1 || filtered[0].Level != 55 || filtered[0].Value != test.wantValue {
			t.Errorf("got %+v under Lv%d cap, want one Lv55 entry with value %v", filtered, test.lvCap, test.wantValue)
		}
	}
}

func TestQueryPvPRankMegaDownscaling(t *testing.T) {
	ohbem := Ohbem{Leagues: leagues, LevelCaps: levelCaps}
	if err := ohbem.LoadPokemonData("./test/master-test.json"); err != nil {
		t.Fatalf("can't load MasterFile")
	}

	// A maxed Mega is ranked in capped leagues instead of being called overleveled.
	entries, err := ohbem.QueryPvPRank(charizardId, 0, 0, 0, 1, 15, 15, 50)
	if err != nil {
		t.Fatal(err)
	}
	mega := findEntries(entries["ultra"], 0, megaY)
	if len(mega) != 1 || mega[0].Level != 20.5 || mega[0].Cp != 2500 || mega[0].Rank != 1 || !mega[0].Capped {
		t.Errorf("got Mega Charizard Y entries %+v, want one capped rank 1 Lv20.5 2500 CP entry", mega)
	}
	if plain := findEntries(entries["ultra"], 0, 0); len(plain) != 0 {
		t.Errorf("got Charizard entries %+v, want none since it is overleveled", plain)
	}

	// The level cap still makes a Mega overleveled when the CP cap does not bind.
	huge := Ohbem{Leagues: map[string]League{"huge": {Cap: 50000}}, LevelCaps: levelCaps, IncludeHundosUnderCap: true}
	if err := huge.LoadPokemonData("./test/master-test.json"); err != nil {
		t.Fatalf("can't load MasterFile")
	}
	entries, err = huge.QueryPvPRank(charizardId, 0, 0, 0, 15, 15, 14, 51)
	if err != nil {
		t.Fatal(err)
	}
	mega = findEntries(entries["huge"], 0, megaY)
	if len(mega) != 1 || mega[0].Cap != 51 || mega[0].Level != 51 || mega[0].Rank != 1 {
		t.Errorf("got Mega Charizard Y entries %+v, want only a rank 1 Lv51 entry", mega)
	}
}
