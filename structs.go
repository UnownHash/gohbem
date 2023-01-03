package ohbemgo

import (
	"sync"
	"time"
)

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

type RankingSortable []Ranking
type RankingSortableIndexed []Ranking

func (r RankingSortable) Len() int {
	return len(r)
}

func (r RankingSortable) Less(i, j int) bool {
	return r[i].Value > r[j].Value
}

func (r RankingSortable) Swap(i, j int) {
	r[i], r[j] = r[j], r[i]
}

func (r RankingSortableIndexed) Len() int {
	return len(r)
}

func (r RankingSortableIndexed) Less(i, j int) bool {
	if r[i].Value == r[j].Value {
		return r[i].Index < r[j].Index
	} else {
		return r[i].Value > r[j].Value
	}
}

func (r RankingSortableIndexed) Swap(i, j int) {
	r[i], r[j] = r[j], r[i]
}

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
	Initialized bool            `json:"-"`
	Pokemon     map[int]Pokemon `json:"pokemon"`
	Costumes    map[int]bool    `json:"costumes"`
}

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

type League struct {
	Cap            int  `json:"cap"`
	LittleCupRules bool `json:"little_cup_rules"`
}

type Ohbem struct {
	PokemonData           PokemonData
	LevelCaps             []int
	Leagues               map[string]League
	DisableCache          bool
	IncludeHundosUnderCap bool
	WatcherInterval       time.Duration
	compactRankCache      sync.Map
	watcherChan           chan bool
}

type CompactCacheValue struct {
	Combinations [4096]int16
	TopValue     float64
}
