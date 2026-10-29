package spellhance

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/wowsims/forever/sim/shaman"
)

// TestEleFuryBuff tries hypothetical Elemental Fury buffs (2026-10-29): a talent no
// Enhancement build reaches (row 6, 25 Elemental points first). Each variant writes results.csv to
// $SPELLHANCE_FURY_OUT/<variant>; the Enhancement rows are controls that must not move.
func TestEleFuryBuff(t *testing.T) {
	out := os.Getenv("SPELLHANCE_FURY_OUT")
	if out == "" {
		t.Skip("set SPELLHANCE_FURY_OUT to run the variants")
	}
	type variant struct {
		name                    string
		critBonus, crit, damage float64
	}
	variants := []variant{
		{"base", 0, 0, 0},
		{"critbonus+50", 0.5, 0, 0},
		{"critbonus+100", 1, 0, 0},
		{"crit+1pr", 0, 1, 0},
		{"crit+2pr", 0, 2, 0},
		{"dmg+2pr", 0, 0, 0.02},
		{"dmg+4pr", 0, 0, 0.04},
		// +1% crit a rank, paired with a second change.
		{"crit+1pr_critbonus+50", 0.5, 1, 0},
		{"crit+1pr_critbonus+100", 1, 1, 0},
		{"crit+1pr_dmg+2pr", 0, 1, 0.02},
	}
	if only := os.Getenv("SPELLHANCE_FURY_ONLY"); only != "" {
		var keep []variant
		for _, v := range variants {
			if v.name == "base" || strings.HasPrefix(v.name, only) {
				keep = append(keep, v)
			}
		}
		variants = keep
	}
	iterations := matrixIterations(t)
	for _, v := range variants {
		shaman.ElementalFuryExtraCritBonus, shaman.ElementalFuryCritPerRank, shaman.ElementalFuryDamagePerRank = v.critBonus, v.crit, v.damage
		var runs []run
		for _, row := range []struct {
			gear   string
			builds []build
		}{
			{"ele_launch", []build{elePreset, eleResto}},
			{"caster_mirage_2h_sets", []build{elePreset, eleResto, weave2922}},
			{"enh_launch_1h", []build{enhPreset, melee8367}},
			{"enh_smashbringer_2h_sets", []build{enhPreset, melee8367}},
		} {
			for _, b := range row.builds {
				for _, d := range []time.Duration{2 * time.Minute, 5 * time.Minute} {
					runs = append(runs, run{Build: b, Gear: row.gear, Duration: d, ReactionMs: 150, Iterations: iterations, Seed: 7})
				}
			}
		}
		runMatrix(t, runs, filepath.Join(out, v.name))
	}
	shaman.ElementalFuryExtraCritBonus, shaman.ElementalFuryCritPerRank, shaman.ElementalFuryDamagePerRank = 0, 0, 0
}
