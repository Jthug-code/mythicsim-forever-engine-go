package spellhance

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/wowsims/forever/sim/shaman"
)

var melee2922AK = build{"29/22 AK3 no hard casts", weave2922AK.Talents, "apls/melee_maelstrom.apl.json", false}

// TestNov10Hotfix prices the 10 November 2026 hotfix (sim/shaman/nov10_notes.go): the 29/22 rotations and the
// Enhancement builds on launch gear and the BiS sets, at 2, 5 and 10 minutes, first as 2 November left them, then
// under the hotfix with each air totem. It writes nov2/ and nov10/ under $SPELLHANCE_NOV10_OUT; SPELLHANCE_BUFFS=full
// picks the full raid buffs.
func TestNov10Hotfix(t *testing.T) {
	out := os.Getenv("SPELLHANCE_NOV10_OUT")
	if out == "" {
		t.Skip("set SPELLHANCE_NOV10_OUT to run the comparison")
	}
	iterations := matrixIterations(t)
	rows := []struct {
		b    build
		gear string
	}{}
	for _, b := range append(weaveRotations2922AK(), melee2922AK) {
		for _, g := range []string{"ele_launch", foreverHybBisGear(), foreverBisGear()} {
			rows = append(rows, struct {
				b    build
				gear string
			}{b, g})
		}
	}
	for _, b := range []build{enhPreset, melee8367, weave8367, melee83211WS} {
		for _, g := range []string{"enh_launch_1h", foreverBisGear()} {
			rows = append(rows, struct {
				b    build
				gear string
			}{b, g})
		}
	}
	matrix := func(dir string, airs ...string) {
		var runs []run
		for _, air := range airs {
			for _, row := range rows {
				b := row.b
				if air != "" {
					b.Label += " [air: " + air + "]"
				}
				for _, d := range []time.Duration{2 * time.Minute, 5 * time.Minute, 10 * time.Minute} {
					runs = append(runs, run{Build: b, Gear: row.gear, Duration: d, ReactionMs: 150, Iterations: iterations, Seed: 7, Air: air})
				}
			}
		}
		runMatrix(t, runs, filepath.Join(out, dir))
	}
	matrix("nov2", "")
	shaman.SetNov10Hotfix(true)
	defer shaman.SetNov10Hotfix(false)
	matrix("nov10", "", "own", "party")
}

// Melee builds that go up Elemental for Elemental Fury's shock crits rather than for bolts: the bolt talents
// (Elemental Alacrity, Eye of the Storm, Lightning Overload, Call of Thunder) traded for Convection and
// Reverberation, with Elemental Warding as the filler to reach tier 6.
var (
	shock2922 = build{"29/22 shock melee (Fury 4, Reverberation)", "553503130000004-053030031004102", "apls/melee_maelstrom.apl.json", false}
	shock3021 = build{"30/21 shock melee (Fury 5, Flurry 5)", "553503130000005-0530300310051", "apls/melee_maelstrom.apl.json", false}
)

// TestNov10State is the three specs side by side under the hotfix: the shock-melee hybrids and the Elemental
// rotations, to go with TestNov10Hotfix's rows. It writes to $SPELLHANCE_NOV10_STATE_OUT.
func TestNov10State(t *testing.T) {
	out := os.Getenv("SPELLHANCE_NOV10_STATE_OUT")
	if out == "" {
		t.Skip("set SPELLHANCE_NOV10_STATE_OUT to run the comparison")
	}
	iterations := matrixIterations(t)
	shaman.SetNov10Hotfix(true)
	defer shaman.SetNov10Hotfix(false)
	var runs []run
	add := func(b build, gear, air string) {
		for _, d := range []time.Duration{2 * time.Minute, 5 * time.Minute, 10 * time.Minute} {
			runs = append(runs, run{Build: b, Gear: gear, Duration: d, ReactionMs: 150, Iterations: iterations, Seed: 7, Air: air})
		}
	}
	for _, b := range []build{shock2922, shock3021, melee2922AK, melee8367, enhPreset} {
		for _, g := range []string{foreverBisGear(), foreverHybBisGear(), "enh_launch_1h"} {
			add(b, g, "own")
		}
	}
	for _, b := range []build{elePreset, eleES, eleESGated, eleGate5, eleGate10} {
		for _, g := range []string{foreverEleBisGear(), "ele_launch"} {
			add(b, g, "")
		}
	}
	runMatrix(t, runs, out)
}

// 25/26: Elemental to Lightning Overload and Call of Thunder, Enhancement to one point of Maelstrom Weapon, which
// since the hotfix gives the full 20% a stack at 6 procs a minute. "bolt" keeps Elemental Alacrity and Elemental
// Focus; "shock" trades them for Convection.
var (
	mw2526Bolt  = build{"25/26 MW1, Alacrity + Focus", "15353310000013-05303003100511211", "apls/melee_maelstrom.apl.json", false}
	mw2526Shock = build{"25/26 MW1, Convection", "55350300000013-05303003100511211", "apls/melee_maelstrom.apl.json", false}
	mw2526At4   = build{"25/26 MW1, Alacrity + Focus, bolt at 4 stacks", "15353310000013-05303003100511211", "apls/melee_maelstrom_at4.apl.json", false}
)

func TestNov10MW1(t *testing.T) {
	out := os.Getenv("SPELLHANCE_NOV10_MW1_OUT")
	if out == "" {
		t.Skip("set SPELLHANCE_NOV10_MW1_OUT to run the comparison")
	}
	iterations := matrixIterations(t)
	shaman.SetNov10Hotfix(true)
	defer shaman.SetNov10Hotfix(false)
	var runs []run
	for _, b := range []build{mw2526Bolt, mw2526Shock, mw2526At4, melee8367, enhPreset} {
		for _, g := range []string{foreverBisGear(), "enh_launch_1h"} {
			for _, d := range []time.Duration{2 * time.Minute, 5 * time.Minute, 10 * time.Minute} {
				runs = append(runs, run{Build: b, Gear: g, Duration: d, ReactionMs: 150, Iterations: iterations, Seed: 7, Air: "own"})
			}
		}
	}
	runMatrix(t, runs, out)
}
