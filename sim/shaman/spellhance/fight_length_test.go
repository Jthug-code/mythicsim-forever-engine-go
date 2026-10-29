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
)

// TestFightLength runs the main builds on the Forever launch gear at 2, 5, 8 and 10 minutes, and counts
// the Major Mana Potions (13444) and Demonic Runes (12662) each fight uses. It writes fight-length.csv
// to $SPELLHANCE_LENGTH_OUT; SPELLHANCE_BUFFS=full picks the run 5 buffs.
func TestFightLength(t *testing.T) {
	out := os.Getenv("SPELLHANCE_LENGTH_OUT")
	if out == "" {
		t.Skip("set SPELLHANCE_LENGTH_OUT to run the comparison")
	}
	type row struct {
		b    build
		gear string
	}
	rows := []row{{enhPreset, "enh_launch_1h"}, {melee8367, "enh_launch_1h"}, {melee83211WS, "enh_launch_1h"},
		{elePreset, "ele_launch"}, {eleResto, "ele_launch"}}
	durations := []time.Duration{2 * time.Minute, 5 * time.Minute, 8 * time.Minute, 10 * time.Minute}
	iterations := matrixIterations(t)
	lines := make([][]string, len(rows)*len(durations))
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
				p := res.RaidMetrics.Parties[0].Players[0]
				uses := map[int32]float64{}
				for _, a := range p.Actions {
					if id := a.Id.GetItemId(); id == 13444 || id == 12662 {
						for _, tg := range a.Targets {
							uses[id] += float64(tg.Casts) / float64(iterations)
						}
					}
				}
				lines[i*len(durations)+j] = []string{rw.b.Label, rw.gear, fmt.Sprint(d.Seconds()), fmt.Sprintf("%.1f", res.RaidMetrics.Dps.Avg),
					fmt.Sprintf("%.1f", p.SecondsOomAvg), fmt.Sprintf("%.2f", uses[13444]), fmt.Sprintf("%.2f", uses[12662])}
			}()
		}
	}
	wg.Wait()
	if err := os.MkdirAll(out, 0o755); err != nil {
		t.Fatal(err)
	}
	f, err := os.Create(filepath.Join(out, "fight-length.csv"))
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	w := csv.NewWriter(f)
	w.Write([]string{"build", "gear", "duration_s", "dps", "seconds_oom", "mana_potions", "demonic_runes"})
	w.WriteAll(lines)
}
