package spellhance

import (
	"os"
	"testing"
	"time"
)

// The 2 November 2026 notes put the shocks back at Classic Era damage (Earth Shock 517-545), but the repo's
// Elemental preset never casts Earth Shock: it was written when Earth Shock did 293-309. These variants add
// it after the Flame Shock refresh, so the shared shock cooldown goes to Earth Shock whenever Flame Shock is
// up; the gated one spends it only while the Lightning Bolt line's mana check passes.
//
// The preset also paces its mana: rank 10 Lightning Bolt only while mana is at least 35 a second of the fight
// left, rank 4 otherwise. The 2 November notes raised rank 10 by 80% and rank 4 by 14%, which moves where that
// line should sit, so the gate variants put it at 5, 10, 15, 25 and 50.
var (
	eleES      = build{"Ele 32/0/19 + Earth Shock", talentsElePreset, "apls/ele_nov2_es.apl.json", true}
	eleESGated = build{"Ele 32/0/19 + Earth Shock, mana-gated", talentsElePreset, "apls/ele_nov2_es_gated.apl.json", true}
	eleGate5   = build{"Ele 32/0/19, rank 10 gate 5 mana/s", talentsElePreset, "apls/ele_nov2_gate5.apl.json", true}
	eleGate10  = build{"Ele 32/0/19, rank 10 gate 10 mana/s", talentsElePreset, "apls/ele_nov2_gate10.apl.json", true}
	eleGate15  = build{"Ele 32/0/19, rank 10 gate 15 mana/s", talentsElePreset, "apls/ele_nov2_gate15.apl.json", true}
	eleGate25  = build{"Ele 32/0/19, rank 10 gate 25 mana/s", talentsElePreset, "apls/ele_nov2_gate25.apl.json", true}
	eleGate50  = build{"Ele 32/0/19, rank 10 gate 50 mana/s", talentsElePreset, "apls/ele_nov2_gate50.apl.json", true}
)

// TestEleRotationNov2 runs the preset against the variants on the launch and SoD proxy caster gear at 2, 5 and
// 10 minutes. It writes results.csv and sources.csv to $SPELLHANCE_ELEROT_OUT; SPELLHANCE_BUFFS=full picks the
// full raid buffs.
func TestEleRotationNov2(t *testing.T) {
	out := os.Getenv("SPELLHANCE_ELEROT_OUT")
	if out == "" {
		t.Skip("set SPELLHANCE_ELEROT_OUT to run the comparison")
	}
	iterations := matrixIterations(t)
	var runs []run
	for _, gear := range []string{"ele_launch", "caster_mirage_2h_sets"} {
		for _, b := range []build{elePreset, eleES, eleESGated, eleGate5, eleGate10, eleGate15, eleGate25, eleGate50} {
			for _, d := range []time.Duration{2 * time.Minute, 5 * time.Minute, 10 * time.Minute} {
				runs = append(runs, run{Build: b, Gear: gear, Duration: d, ReactionMs: 150, Iterations: iterations, Seed: 7})
			}
		}
	}
	runMatrix(t, runs, out)
}
