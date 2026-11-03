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
)

// The SpellhanceWeaveHelper light's open questions under the 2 November notes, for the respecced 29/22: after a
// hard cast completes, start another (bolt-chain) or let the swing come (weave); whether Chain Lightning
// (2.5 s, 145% damage, 225% mana) beats Lightning Bolt in the window on one target; and whether a shock's mana
// buys more as bolts.
func weaveRotations2922AK() []build {
	t := weave2922AK.Talents
	return []build{
		{"29/22 AK3 weave", t, "apls/weave.apl.json", false},
		{"29/22 AK3 bolt-chain", t, "apls/chain.apl.json", false},
		{"29/22 AK3 weave + CL", t, "apls/weave_cl.apl.json", false},
		{"29/22 AK3 bolt-chain + CL", t, "apls/chain_cl.apl.json", false},
		{"29/22 AK3 weave, no Earth Shock", t, "apls/weave_no_es.apl.json", false},
		{"29/22 AK3 weave, no Flame Shock", t, "apls/weave_no_fs.apl.json", false},
		{"29/22 AK3 weave, no shocks", t, "apls/weave_no_shocks.apl.json", false},
	}
}

// TestWeaveRotationsNov2 runs those on Elemental launch gear and the 29/22 BiS set, 2 / 5 / 10 min, and writes
// weave-rotations.csv to $SPELLHANCE_WEAVE_OUT; SPELLHANCE_BUFFS=full picks the full raid buffs.
func TestWeaveRotationsNov2(t *testing.T) {
	out := os.Getenv("SPELLHANCE_WEAVE_OUT")
	if out == "" {
		t.Skip("set SPELLHANCE_WEAVE_OUT to run the weave rotations")
	}
	iterations := matrixIterations(t)
	var runs []run
	for _, gear := range []string{"ele_launch", foreverHybBisGear()} {
		for _, b := range weaveRotations2922AK() {
			for _, d := range []time.Duration{2 * time.Minute, 5 * time.Minute, 10 * time.Minute} {
				runs = append(runs, run{Build: b, Gear: gear, Duration: d, ReactionMs: 150, Iterations: iterations, Seed: 7})
			}
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
					t.Errorf("%s on %s: %s", r.Build.Label, r.Gear, res.Error.Message)
					continue
				}
				player := res.RaidMetrics.Parties[0].Players[0]
				results[i] = []string{r.Build.Label, r.Gear, fmt.Sprint(r.Duration.Seconds()),
					fmt.Sprintf("%.1f", res.RaidMetrics.Dps.Avg), fmt.Sprintf("%.1f", player.SecondsOomAvg)}
			}
		}()
	}
	for i := range runs {
		jobs <- i
	}
	close(jobs)
	wg.Wait()

	if err := os.MkdirAll(out, 0o755); err != nil {
		t.Fatal(err)
	}
	f, err := os.Create(filepath.Join(out, "weave-rotations.csv"))
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	w := csv.NewWriter(f)
	w.Write([]string{"build", "gear", "duration_s", "dps", "seconds_oom"})
	w.WriteAll(results)
	w.Flush()
	if err := w.Error(); err != nil {
		t.Fatal(err)
	}
}
