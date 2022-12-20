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
