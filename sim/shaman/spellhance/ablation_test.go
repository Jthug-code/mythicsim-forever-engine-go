package spellhance

import (
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"testing"
	"time"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/shaman"
)

// TestNov2Ablation prices each 2 November change on its own: every build runs with all the changes, then once
// with each change undone and the rest kept. The DPS a build loses with a change undone is what that change is
// worth to it. Launch gear and the Forever BiS sets, 2 and 5 minutes, as the default Skyborne. It writes
// ablation.csv to $SPELLHANCE_ABLATE_OUT; SPELLHANCE_BUFFS=full picks the full raid buffs.
func TestNov2Ablation(t *testing.T) {
	out := os.Getenv("SPELLHANCE_ABLATE_OUT")
	if out == "" {
		t.Skip("set SPELLHANCE_ABLATE_OUT to run the ablation")
	}
	iterations := matrixIterations(t)
	type row struct {
		b    build
		gear string
	}
	rows := []row{
		{enhPreset, "enh_launch_1h"}, {melee8367, "enh_launch_1h"}, {elePreset, "ele_launch"}, {weave2922, "ele_launch"},
		{enhPreset, foreverBisGear()}, {melee8367, foreverBisGear()}, {elePreset, foreverEleBisGear()},
		{weave2922, foreverHybBisGear()},
	}
	durations := []time.Duration{2 * time.Minute, 5 * time.Minute}

	var lines [][]string
	for _, undone := range append([]string{"none"}, shaman.Nov2Changes()...) {
		if undone != "none" {
			shaman.SetNov2Change(undone, false)
		}
		var runs []run
		for _, rw := range rows {
			for _, d := range durations {
				runs = append(runs, run{Build: rw.b, Gear: rw.gear, Duration: d, ReactionMs: 150, Iterations: iterations, Seed: 7})
			}
		}
		results := make([][]string, len(runs))
		var wg sync.WaitGroup
		jobs := make(chan int)
		for w := 0; w < runtime.NumCPU(); w++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for i := range jobs {
					r := runs[i]
					res := core.RunRaidSim(r.request())
					if res.Error != nil {
						t.Errorf("%s, %s on %s: %s", undone, r.Build.Label, r.Gear, res.Error.Message)
						continue
					}
					results[i] = []string{undone, r.Build.Label, r.Gear, fmt.Sprint(r.Duration.Seconds()),
						fmt.Sprintf("%.1f", res.RaidMetrics.Dps.Avg), fmt.Sprintf("%.1f", res.RaidMetrics.Parties[0].Players[0].SecondsOomAvg)}
				}
			}()
		}
		for i := range runs {
			jobs <- i
		}
		close(jobs)
		wg.Wait()
		lines = append(lines, results...)
		if undone != "none" {
			shaman.SetNov2Change(undone, true)
		}
	}

	if err := os.MkdirAll(out, 0o755); err != nil {
		t.Fatal(err)
	}
	f, err := os.Create(filepath.Join(out, "ablation.csv"))
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	w := csv.NewWriter(f)
	w.Write([]string{"undone", "build", "gear", "duration_s", "dps", "seconds_oom"})
	w.WriteAll(lines)
}
