package spellhance

import (
	"fmt"
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

// The same builds on two to five targets. Melee hits one target; Chain Lightning hits up to three, and Magma
// Totem and Fire Nova hit them all. The *_aoe rotations (gen_apls.py) drop Magma Totem and press Fire Nova on
// cooldown; the Elemental and Enhancement presets switch on the target count themselves.
var (
	aoe8367        = build{"8/36/7 no hard casts, Magma + Fire Nova", talents8367, "apls/melee_maelstrom_aoe.apl.json", false}
	aoe8367CL      = build{"8/36/7 weave Chain Lightning, Magma + Fire Nova", talents8367, "apls/weave_maelstrom_cl_aoe.apl.json", false}
	aoeEnhPreset   = build{"Enh preset 19/32/0, Magma + Fire Nova", talentsEnhPreset, "apls/melee_maelstrom_aoe.apl.json", false}
	aoe2922Weave   = build{"29/22 AK3 weave Chain Lightning, Magma + Fire Nova", weave2922AK.Talents, "apls/weave_cl_aoe.apl.json", false}
	aoe2922Chain   = build{"29/22 AK3 chain Chain Lightning, Magma + Fire Nova", weave2922AK.Talents, "apls/chain_cl_aoe.apl.json", false}
	aoe2526        = build{"25/26 MW1 no hard casts, Magma + Fire Nova", mw2526Bolt.Talents, "apls/melee_maelstrom_aoe.apl.json", false}
	aoe2526CL      = build{"25/26 MW1 weave Chain Lightning, Magma + Fire Nova", mw2526Bolt.Talents, "apls/weave_maelstrom_cl_aoe.apl.json", false}
	aoe2526ChainCL = build{"25/26 MW1 chain Chain Lightning, Magma + Fire Nova", mw2526Bolt.Talents, "apls/chain_cl_aoe.apl.json", false}
)

func TestNov10Targets(t *testing.T) {
	out := os.Getenv("SPELLHANCE_NOV10_TARGETS_OUT")
	if out == "" {
		t.Skip("set SPELLHANCE_NOV10_TARGETS_OUT to run the comparison")
	}
	iterations := matrixIterations(t)
	shaman.SetNov10Hotfix(true)
	defer shaman.SetNov10Hotfix(false)
	rows := []struct {
		b         build
		gear, air string
	}{
		{aoe8367, foreverBisGear(), "own"}, {aoe8367CL, foreverBisGear(), "own"},
		{enhPreset, foreverBisGear(), "own"}, {aoeEnhPreset, foreverBisGear(), "own"},
		{aoe2922Weave, foreverHybBisGear(), "own"}, {aoe2922Chain, foreverHybBisGear(), "own"},
		{aoe2526, foreverBisGear(), "own"}, {aoe2526CL, foreverBisGear(), "own"}, {aoe2526ChainCL, foreverBisGear(), "own"},
		{aoe2526CL, foreverHybBisGear(), "own"}, {aoe2526ChainCL, foreverHybBisGear(), "own"},
		{elePreset, foreverEleBisGear(), ""}, {eleGate5, foreverEleBisGear(), ""}, {eleGate10, foreverEleBisGear(), ""},
	}
	for targets := 1; targets <= 5; targets++ {
		var runs []run
		for _, row := range rows {
			for _, d := range []time.Duration{time.Minute, 2 * time.Minute, 5 * time.Minute} {
				runs = append(runs, run{Build: row.b, Gear: row.gear, Duration: d, ReactionMs: 150, Iterations: iterations, Seed: 7, Air: row.air, Targets: targets})
			}
		}
		runMatrix(t, runs, filepath.Join(out, fmt.Sprint(targets)))
	}
}

// TestNov10MW1Bound is 25/26 again with one point of Maelstrom Weapon rolling the five-point chance: the most the
// build could gain if the rank's lower chance turned out not to apply.
func TestNov10MW1Bound(t *testing.T) {
	out := os.Getenv("SPELLHANCE_NOV10_MW1BOUND_OUT")
	if out == "" {
		t.Skip("set SPELLHANCE_NOV10_MW1BOUND_OUT to run the comparison")
	}
	iterations := matrixIterations(t)
	shaman.SetNov10Hotfix(true)
	shaman.MaelstromWeaponRateAsRank = 5
	defer func() { shaman.SetNov10Hotfix(false); shaman.MaelstromWeaponRateAsRank = 0 }()
	var runs []run
	for _, d := range []time.Duration{2 * time.Minute, 5 * time.Minute, 10 * time.Minute} {
		runs = append(runs, run{Build: mw2526Bolt, Gear: foreverBisGear(), Duration: d, ReactionMs: 150, Iterations: iterations, Seed: 7, Air: "own"})
	}
	runMatrix(t, runs, out)
}

// The melee_maelstrom rotations press an instant only when its GCD ends before the next swing, the weave light's
// rule. With no hard cast to protect, that only wastes GCDs. The melee_free rotations press instants whenever the
// GCD is free, and the 5-stack bolt first.
var (
	free8367      = build{"8/36/7 free instants", talents8367, "apls/melee_free.apl.json", false}
	freeEnhPreset = build{"Enh preset 19/32/0 talents, free instants", talentsEnhPreset, "apls/melee_free.apl.json", false}
	free83211WS   = build{"8/32/11 Water Shield, free instants", melee83211WS.Talents, "apls/melee_free.apl.json", false}
	free2526      = build{"25/26 MW1, free instants", mw2526Bolt.Talents, "apls/melee_free.apl.json", false}
	free3021      = build{"30/21 shock melee, free instants", shock3021.Talents, "apls/melee_free.apl.json", false}
	free2922      = build{"29/22 AK3 no hard casts, free instants", weave2922AK.Talents, "apls/melee_free.apl.json", false}
	free8367AoE   = build{"8/36/7 free instants, Magma + Fire Nova", talents8367, "apls/melee_free_aoe.apl.json", false}
	freePresetAoE = build{"Enh preset 19/32/0 talents, free instants, Magma + Fire Nova", talentsEnhPreset, "apls/melee_free_aoe.apl.json", false}
	free2526AoE   = build{"25/26 MW1, free instants, Magma + Fire Nova", mw2526Bolt.Talents, "apls/melee_free_aoe.apl.json", false}
)

func TestNov10FreeInstants(t *testing.T) {
	out := os.Getenv("SPELLHANCE_NOV10_FREE_OUT")
	if out == "" {
		t.Skip("set SPELLHANCE_NOV10_FREE_OUT to run the comparison")
	}
	iterations := matrixIterations(t)
	shaman.SetNov10Hotfix(true)
	defer shaman.SetNov10Hotfix(false)
	var runs []run
	for _, b := range []build{free8367, melee8367, freeEnhPreset, enhPreset, free83211WS, free2526, free3021, free2922} {
		for _, g := range []string{foreverBisGear(), "enh_launch_1h"} {
			for _, d := range []time.Duration{2 * time.Minute, 5 * time.Minute, 10 * time.Minute} {
				runs = append(runs, run{Build: b, Gear: g, Duration: d, ReactionMs: 150, Iterations: iterations, Seed: 7, Air: "own"})
			}
		}
	}
	runMatrix(t, runs, filepath.Join(out, "1"))
	for targets := 2; targets <= 5; targets++ {
		runs = nil
		for _, b := range []build{free8367AoE, freePresetAoE, free2526AoE} {
			for _, d := range []time.Duration{time.Minute, 2 * time.Minute, 5 * time.Minute} {
				runs = append(runs, run{Build: b, Gear: foreverBisGear(), Duration: d, ReactionMs: 150, Iterations: iterations, Seed: 7, Air: "own", Targets: targets})
			}
		}
		runMatrix(t, runs, filepath.Join(out, fmt.Sprint(targets)))
	}
}

// TestNov10Loose settles what run 8 left open: the shocks under the chain rotation, the Elemental rotations on the
// 2 November state, and what Windfury Totem is worth to 8/36/7 with free instants. It writes nov2/ and nov10/ under
// $SPELLHANCE_NOV10_LOOSE_OUT.
func TestNov10Loose(t *testing.T) {
	out := os.Getenv("SPELLHANCE_NOV10_LOOSE_OUT")
	if out == "" {
		t.Skip("set SPELLHANCE_NOV10_LOOSE_OUT to run the comparison")
	}
	iterations := matrixIterations(t)
	durations := []time.Duration{2 * time.Minute, 5 * time.Minute, 10 * time.Minute}
	var runs []run
	for _, b := range []build{elePreset, eleGate5, eleGate10} {
		for _, d := range durations {
			runs = append(runs, run{Build: b, Gear: foreverEleBisGear(), Duration: d, ReactionMs: 150, Iterations: iterations, Seed: 7})
		}
	}
	runMatrix(t, runs, filepath.Join(out, "nov2"))

	shaman.SetNov10Hotfix(true)
	defer shaman.SetNov10Hotfix(false)
	runs = nil
	talents := weave2922AK.Talents
	for _, b := range []build{
		{"29/22 AK3 bolt-chain", talents, "apls/chain.apl.json", false},
		{"29/22 AK3 bolt-chain, no Earth Shock", talents, "apls/chain_no_es.apl.json", false},
		{"29/22 AK3 bolt-chain, no Flame Shock", talents, "apls/chain_no_fs.apl.json", false},
		{"29/22 AK3 bolt-chain, no shocks", talents, "apls/chain_no_shocks.apl.json", false},
	} {
		for _, g := range []string{"ele_launch", foreverHybBisGear()} {
			for _, d := range durations {
				runs = append(runs, run{Build: b, Gear: g, Duration: d, ReactionMs: 150, Iterations: iterations, Seed: 7, Air: "own"})
			}
		}
	}
	for _, air := range []string{"", "own", "party"} {
		b := free8367
		b.Label += " [air: " + air + "]"
		for _, d := range durations {
			runs = append(runs, run{Build: b, Gear: foreverBisGear(), Duration: d, ReactionMs: 150, Iterations: iterations, Seed: 7, Air: air})
		}
	}
	runMatrix(t, runs, filepath.Join(out, "nov10"))
}

// TestNov10Final is the run 8 summary under the hotfix as nov10_notes.go has it now, Windfury Totem rolling off
// autos alone: each spec's best rotation on its BiS set and launch gear, with each air totem for 8/36/7, and on
// three targets. It writes 1/ and 3/ under $SPELLHANCE_NOV10_FINAL_OUT.
func TestNov10Final(t *testing.T) {
	out := os.Getenv("SPELLHANCE_NOV10_FINAL_OUT")
	if out == "" {
		t.Skip("set SPELLHANCE_NOV10_FINAL_OUT to run the comparison")
	}
	iterations := matrixIterations(t)
	shaman.SetNov10Hotfix(true)
	defer shaman.SetNov10Hotfix(false)
	durations := []time.Duration{2 * time.Minute, 5 * time.Minute, 10 * time.Minute}
	chain := build{"29/22 AK3 bolt-chain", weave2922AK.Talents, "apls/chain.apl.json", false}
	weave := build{"29/22 AK3 weave", weave2922AK.Talents, "apls/weave.apl.json", false}
	var runs []run
	add := func(b build, gear, air string, targets int, ds []time.Duration) {
		if air != "own" {
			b.Label += " [air: " + air + "]"
		}
		for _, d := range ds {
			runs = append(runs, run{Build: b, Gear: gear, Duration: d, ReactionMs: 150, Iterations: iterations, Seed: 7, Air: air, Targets: targets})
		}
	}
	for _, b := range []build{free8367, freeEnhPreset, free83211WS, free2526} {
		add(b, foreverBisGear(), "own", 1, durations)
		add(b, "enh_launch_1h", "own", 1, durations)
	}
	add(free8367, foreverBisGear(), "", 1, durations)
	add(free8367, foreverBisGear(), "party", 1, durations)
	for _, b := range []build{chain, weave} {
		add(b, foreverHybBisGear(), "own", 1, durations)
		add(b, "ele_launch", "own", 1, durations)
		add(b, foreverHybBisGear(), "", 1, durations)
	}
	runMatrix(t, runs, filepath.Join(out, "1"))
	runs = nil
	short := []time.Duration{time.Minute, 2 * time.Minute, 5 * time.Minute}
	for _, b := range []build{free8367AoE, freePresetAoE, aoe2922Weave, aoe2922Chain} {
		gear := foreverBisGear()
		if b.Talents == weave2922AK.Talents {
			gear = foreverHybBisGear()
		}
		add(b, gear, "own", 3, short)
	}
	runMatrix(t, runs, filepath.Join(out, "3"))
}
