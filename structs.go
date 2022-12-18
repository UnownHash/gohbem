package ohbemgo

type PokemonStats struct {
	Attack  int
	Defense int
	Stamina int
}

type Ranking struct {
	Value      float64
	Level      float64
	Cp         int
	Percentage float64
	Rank       int
	Attack     int
	Defense    int
	Stamina    int
	Cap        int
	Index      int
}

type BySortedRanks []Ranking

func (a BySortedRanks) Len() int {
	return len(a)
}

func (a BySortedRanks) Less(i, j int) bool {
	if a[i].Value != a[j].Value {
		return a[i].Value < a[j].Value
	} else {
		return a[i].Index < a[j].Index
	}
}

func (a BySortedRanks) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}

type Pokemon struct {
	Forms                     map[int]Form          `json:"forms"`
	Attack                    int                   `json:"Attack"`
	Defense                   int                   `json:"Defense"`
	Stamina                   int                   `json:"Stamina"`
	Evolutions                []Evolution           `json:"evolutions,omitempty"`
	TempEvolutions            map[int]TempEvolution `json:"temp_evolutions,omitempty"`
	Little                    bool                  `json:"little,omitempty"`
	CostumeOverrideEvolutions []int                 `json:"costume_override_evos,omitempty"`
}

type Form struct {
	Attack                    int                   `json:"Attack,omitempty"`
	Defense                   int                   `json:"Defense,omitempty"`
	Stamina                   int                   `json:"Stamina,omitempty"`
	Little                    bool                  `json:"little,omitempty"`
	Evolutions                []Evolution           `json:"evolutions,omitempty"`
	TempEvolutions            map[int]TempEvolution `json:"temp_evolutions,omitempty"`
	CostumeOverrideEvolutions []int                 `json:"costume_override_evos,omitempty"`
}

type Evolution struct {
	Pokemon           int `json:"pokemon"`
	Form              int `json:"form,omitempty"`
	GenderRequirement int `json:"gender_requirement,omitempty"`
}

type TempEvolution struct {
	Attack     int  `json:"Attack,omitempty"`
	Defense    int  `json:"Defense,omitempty"`
	Stamina    int  `json:"Stamina,omitempty"`
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
	Value      int     `json:"Value"`
	Level      float64 `json:"Level"`
	Cp         int     `json:"Cp"`
	Percentage float64 `json:"Percentage"`
	Rank       int     `json:"Rank"`
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
	PokemonData PokemonData
	LevelCaps   []float64
	Leagues     Leagues
}
