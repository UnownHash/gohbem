package ohbemgo

import (
	"sync"
	"time"
)

type Ranking struct {
	Value      float64
	Level      float64
	Cp         int
	Percentage float64
	Rank       int16
	Attack     int
	Defense    int
	Stamina    int
	Cap        float64
	Capped     bool
	Index      int
}

type Pokemon struct {
	Forms                     map[int]Form         `json:"forms"`
	Attack                    int                  `json:"attack"`
	Defense                   int                  `json:"defense"`
	Stamina                   int                  `json:"stamina"`
	Evolutions                []Evolution          `json:"evolutions,omitempty"`
	TempEvolutions            map[int]PokemonStats `json:"temp_evolutions,omitempty"`
	Little                    bool                 `json:"little,omitempty"`
	CostumeOverrideEvolutions []int                `json:"costume_override_evos,omitempty"`
}

type Form struct {
	Attack                    int                  `json:"attack,omitempty"`
	Defense                   int                  `json:"defense,omitempty"`
	Stamina                   int                  `json:"stamina,omitempty"`
	Little                    bool                 `json:"little,omitempty"`
	Evolutions                []Evolution          `json:"evolutions,omitempty"`
	TempEvolutions            map[int]PokemonStats `json:"temp_evolutions,omitempty"`
	CostumeOverrideEvolutions []int                `json:"costume_override_evos,omitempty"`
}

type Evolution struct {
	Pokemon           int `json:"pokemon"`
	Form              int `json:"form,omitempty"`
	GenderRequirement int `json:"gender_requirement,omitempty"`
}

type PokemonStats struct {
	Attack     int  `json:"attack,omitempty"`
	Defense    int  `json:"defense,omitempty"`
	Stamina    int  `json:"stamina,omitempty"`
	Unreleased bool `json:"unreleased,omitempty"`
}

type PokemonData struct {
	Pokemon  map[int]Pokemon `json:"pokemon"`
	Costumes map[int]bool    `json:"costumes"`
}

type PokemonEntry struct {
	Pokemon    int     `json:"pokemon"`
	Form       int     `json:"form"`
	Cap        float64 `json:"cap"`
	Value      float64 `json:"Value"`
	Level      float64 `json:"Level"`
	Cp         int     `json:"cp"`
	Percentage float64 `json:"percentage"`
	Rank       int16   `json:"rank"`
	Capped     bool    `json:"capped"`
	Evolution  int     `json:"evolution"`
}

type PokemonEntries struct {
	Little []PokemonEntry `json:"little"`
	Great  []PokemonEntry `json:"great"`
	Ultra  []PokemonEntry `json:"ultra"`
}

type Leagues map[string]struct {
	Cap    int
	Little bool
}

type Ohbem struct {
	PokemonData      PokemonData
	LevelCaps        []float64
	Leagues          Leagues
	DisableCache     bool
	WatcherInterval  time.Duration
	compactRankCache sync.Map
	watcherChan      chan bool
}

type CompactCacheValue struct {
	Combinations [4096]int16
	TopValue     float64
}
