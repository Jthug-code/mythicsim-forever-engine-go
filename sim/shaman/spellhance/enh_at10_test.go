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

// The Enhancement preset's own rotation casts Lightning Bolt at 5 Maelstrom stacks; this one waits
// for 10, as only Soulcrusher's Rage 6 allows (2026-10-29).
var enhPresetAt10 = build{"Enh preset 19/32/0, MW at 10", talentsEnhPreset, "apls/enh_preset_mw10.apl.json", false}

// TestEnhPresetAt10 runs the Enhancement preset spending Maelstrom at 5 and at 10, beside 8/36/7 at 10,
// on the SoD Enhancement set at 2, 5, 8 and 10 minutes, with and without the tuning package
// (TestTuningPackage). It writes enh-at10.csv to $SPELLHANCE_AT10_OUT.
func TestEnhPresetAt10(t *testing.T) {
	out := os.Getenv("SPELLHANCE_AT10_OUT")
	if out == "" {
		t.Skip("set SPELLHANCE_AT10_OUT to run the comparison")
	}
	builds := []build{enhPreset, enhPresetAt10, melee8367at10}
	durations := []time.Duration{2 * time.Minute, 5 * time.Minute, 8 * time.Minute, 10 * time.Minute}
	iterations := matrixIterations(t)
	var lines [][]string
	for _, pkg := range []bool{false, true} {
		if pkg {
			shaman.ElementalFuryCritPerRank, shaman.ElementalFuryExtraCritBonus, shaman.LightningBaseDamageBonus = 1, 1, 0.06
		}
		part := make([][]string, len(builds)*len(durations))
		var wg sync.WaitGroup
		for i, b := range builds {
			for j, d := range durations {
				wg.Add(1)
				go func() {
					defer wg.Done()
					r := run{Build: b, Gear: "enh_smashbringer_2h_sets", Duration: d, ReactionMs: 150, Iterations: iterations, Seed: 7}
					res := core.RunRaidSim(r.request())
					if res.Error != nil {
						t.Errorf("%s: %s", b.Label, res.Error.Message)
						return
					}
					part[i*len(durations)+j] = []string{fmt.Sprint(pkg), b.Label, fmt.Sprint(d.Seconds()),
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
	f, err := os.Create(filepath.Join(out, "enh-at10.csv"))
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	w := csv.NewWriter(f)
	w.Write([]string{"package", "build", "duration_s", "dps", "seconds_oom"})
	w.WriteAll(lines)
}
