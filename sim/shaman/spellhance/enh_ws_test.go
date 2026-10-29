package spellhance

import (
	"os"
	"testing"
	"time"
)

// 8/32/11 (2026-10-29): 8/36/7 with Ancestral Knowledge 5 -> 1, and Mindfulness 3 plus Water
// Shield, cast and kept up. https://www.wowhead.com/forever/talent-calc/shaman/v2053-051032031005112251-053002001
var melee83211WS = build{"8/32/11 Water Shield, no hard casts", "053-051032031005112251-053002001", "apls/melee_maelstrom_ws.apl.json", false}

// TestEnhWaterShield compares 8/36/7 and 8/32/11 with no hard casts on the launch gear, the SoD
// Enhancement set and the Forever BiS set, and writes results.csv and sources.csv to
// $SPELLHANCE_ENHWS_OUT. SPELLHANCE_BUFFS=full picks the run 5 buffs and that run's BiS set.
func TestEnhWaterShield(t *testing.T) {
	out := os.Getenv("SPELLHANCE_ENHWS_OUT")
	if out == "" {
		t.Skip("set SPELLHANCE_ENHWS_OUT to run the comparison")
	}
	iterations := matrixIterations(t)
	var runs []run
	for _, gear := range []string{"enh_launch_1h", "enh_smashbringer_2h_sets", foreverBisGear()} {
		for _, b := range []build{melee8367, melee83211WS} {
			for _, d := range []time.Duration{2 * time.Minute, 5 * time.Minute} {
				runs = append(runs, run{Build: b, Gear: gear, Duration: d, ReactionMs: 150, Iterations: iterations, Seed: 7})
			}
		}
	}
	runMatrix(t, runs, out)
}
