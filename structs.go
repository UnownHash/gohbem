package gohbem

import (
	"sync"
	"time"
)

// Ohbem struct is holding main configuration, cache and channels.
type Ohbem struct {
	PokemonData           PokemonData
	LevelCaps             []int
	Leagues               map[string]League
	DisableCache          bool
	MasterFileCachePath   string // when provided: store there latest changed version of masterfile
	RankingComparator     RankingComparator
	IncludeHundosUnderCap bool
	WatcherInterval       time.Duration
	compactRankCache      sync.Map
	watcherChan           chan bool
	Logger                Logger
}

// Logger interface
//
// Example implementation:
//
//	type CustomLogger struct{}
//	func (cl *CustomLogger) Print(message string) {
//		fmt.Println("CustomLogger:", message)
//	}
//	logger := &CustomLogger{}
//	ohbem := Ohbem{Logger: logger, ...}
//
// Notes:
//   - The implementation of the Logger interface defines the specific behavior of the logging operation.
//   - The method should handle the formatting and output of the Log message according to the logger's rules.
type Logger interface {
	Print(message string)
}

// League struct is holding one entry of League configuration passed to Ohbem struct.
type League struct {
	Cap            int  `json:"cap"`
	LittleCupRules bool `json:"little_cup_rules"`
}

// PvPRankingStats internal struct for comparison.
type PvPRankingStats struct {
	Attack float64
	Value  float64
	Level  float64
	Cp     int
	Index  int
}

// RankingComparator specifies how to sort rankings.
type RankingComparator func(a, b *PvPRankingStats) int

// Ranking entry represents PvP row for Pokemon.
type Ranking struct {
	Value      float64 `json:"value"`
	Level      float64 `json:"level"`
	Cp         int     `json:"cp"`
	Percentage float64 `json:"percentage"`
	Rank       int16   `json:"rank"`
	Attack     int     `json:"attack"`
	Defense    int     `json:"defense"`
	Stamina    int     `json:"stamina"`
	Cap        float64 `json:"cap"`
	Capped     bool    `json:"capped,omitempty"`
	Index      int     `json:"index,omitempty"`
}

// PokemonEntry is holding a row of result for QueryPvPRank and FilterLevelCaps functions.
type PokemonEntry struct {
	Pokemon    int     `json:"pokemon"`
	Form       int     `json:"form,omitempty"`
	Cap        float64 `json:"cap,omitempty"`
	Value      float64 `json:"value,omitempty"`
	Level      float64 `json:"level"`
	Cp         int     `json:"cp,omitempty"`
	Percentage float64 `json:"percentage"`
	Rank       int16   `json:"rank"`
	Capped     bool    `json:"capped,omitempty"`
	Evolution  int     `json:"evolution,omitempty"`
}

// Pokemon entry represents row of Pokemon data from MasterFile
type Pokemon struct {
	Attack                    int                  `json:"attack"`
	Defense                   int                  `json:"defense"`
	Stamina                   int                  `json:"stamina"`
	Little                    bool                 `json:"little,omitempty"`
	Evolutions                []Evolution          `json:"evolutions,omitempty"`
	TempEvolutions            map[int]PokemonStats `json:"temp_evolutions,omitempty"`
	CostumeOverrideEvolutions []int                `json:"costume_override_evos,omitempty"`
	Forms                     map[int]Form         `json:"forms"`
}

// Form entry represents row of Pokemon -> Form.
type Form struct {
	Attack                    int                  `json:"attack,omitempty"`
	Defense                   int                  `json:"defense,omitempty"`
	Stamina                   int                  `json:"stamina,omitempty"`
	Little                    bool                 `json:"little,omitempty"`
	Evolutions                []Evolution          `json:"evolutions,omitempty"`
	TempEvolutions            map[int]PokemonStats `json:"temp_evolutions,omitempty"`
	CostumeOverrideEvolutions []int                `json:"costume_override_evos,omitempty"`
}

// Evolution entry represents row of Pokemon -> Evolution.
type Evolution struct {
	Pokemon           int `json:"pokemon"`
	Form              int `json:"form,omitempty"`
	GenderRequirement int `json:"gender_requirement,omitempty"`
}

// PokemonStats entry represents basic Pokemon stats and mega release state.
type PokemonStats struct {
	Attack     int  `json:"attack,omitempty"`
	Defense    int  `json:"defense,omitempty"`
	Stamina    int  `json:"stamina,omitempty"`
	Unreleased bool `json:"unreleased,omitempty"`
}

// PokemonData is a struct holding MasterFile data.
type PokemonData struct {
	Initialized bool            `json:"-"`
	Pokemon     map[int]Pokemon `json:"pokemon"`
	Costumes    map[int]bool    `json:"costumes"`
}

// compactCacheValue is holding Combinations and TopValue for provided stats and cpCap.
type compactCacheValue struct {
	Combinations *[4096]int16
	TopValue     float64
}
