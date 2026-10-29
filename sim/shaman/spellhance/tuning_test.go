package spellhance

import (
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/shaman"
)

// TestTuningPackage runs an Elemental tuning package (2026-10-29) against the current game:
// Elemental Fury +1% spell crit a rank and crits for 2.5x, plus 6% more base damage on Lightning Bolt
// and Chain Lightning. Every Elemental, hybrid and Enhancement build runs on launch and SoD gear at 2,
// 5, 8 and 10 minutes, with and without the package; tuning.csv goes to $SPELLHANCE_TUNING_OUT, and
// SPELLHANCE_BUFFS=full picks the run 5 buffs.
func TestTuningPackage(t *testing.T) {
	out := os.Getenv("SPELLHANCE_TUNING_OUT")
	if out == "" {
		t.Skip("set SPELLHANCE_TUNING_OUT to run the package")
	}
	type row struct {
		b    build
		gear string
	}
	rows := []row{
		{elePreset, "ele_launch"}, {eleResto, "ele_launch"}, {weave2922, "ele_launch"}, {weave2922CL, "ele_launch"}, {chain2922, "ele_launch"},
		{enhPreset, "enh_launch_1h"}, {melee8367, "enh_launch_1h"}, {weave2922, "enh_launch_1h"},
		{elePreset, "caster_mirage_2h_sets"}, {eleResto, "caster_mirage_2h_sets"}, {weave2922, "caster_mirage_2h_sets"},
		{weave2922CL, "caster_mirage_2h_sets"}, {chain2922CL, "caster_mirage_2h_sets"},
		{enhPreset, "enh_smashbringer_2h_sets"}, {melee8367at10, "enh_smashbringer_2h_sets"}, {weave2922CL, "enh_smashbringer_2h_sets"},
	}
	durations := []time.Duration{2 * time.Minute, 5 * time.Minute, 8 * time.Minute, 10 * time.Minute}
	iterations := matrixIterations(t)
	var lines [][]string
	for _, pkg := range []bool{false, true} {
		if pkg {
			shaman.ElementalFuryCritPerRank, shaman.ElementalFuryExtraCritBonus, shaman.LightningBaseDamageBonus = 1, 1, 0.06
		}
		part := make([][]string, len(rows)*len(durations))
		var wg sync.WaitGroup
		for i, rw := range rows {
			for j, d := range durations {
				wg.Add(1)
				go func() {
					defer wg.Done()
					r := run{Build: rw.b, Gear: rw.gear, Duration: d, ReactionMs: 150, Iterations: iterations, Seed: 7}
					res := core.RunRaidSim(r.request())
					if res.Error != nil {
						t.Errorf("%s: %s", rw.b.Label, res.Error.Message)
						return
					}
					part[i*len(durations)+j] = []string{fmt.Sprint(pkg), rw.b.Label, rw.gear, fmt.Sprint(d.Seconds()),
						fmt.Sprintf("%.1f", res.RaidMetrics.Dps.Avg), fmt.Sprintf("%.1f", res.RaidMetrics.Parties[0].Players[0].SecondsOomAvg)}
				}()
			}
		}
		wg.Wait()
		lines = append(lines, part...)
	}
	shaman.ElementalFuryCritPerRank, shaman.ElementalFuryExtraCritBonus, shaman.LightningBaseDamageBonus = 0, 0, 0
	if err := os.MkdirAll(out, 0o755); err != nil {
		t.Fatal(err)
	}
	f, err := os.Create(filepath.Join(out, "tuning.csv"))
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	w := csv.NewWriter(f)
	w.Write([]string{"package", "build", "gear", "duration_s", "dps", "seconds_oom"})
	w.WriteAll(lines)
}
