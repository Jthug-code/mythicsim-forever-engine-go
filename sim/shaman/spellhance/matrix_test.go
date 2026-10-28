package spellhance

import (
	"encoding/csv"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
)

// TestSpellhanceMatrix runs the full comparison and writes results.csv to $SPELLHANCE_OUT.
// SPELLHANCE_ITERATIONS sets the iterations per run (default 3000).
func TestSpellhanceMatrix(t *testing.T) {
	out := os.Getenv("SPELLHANCE_OUT")
	if out == "" {
		t.Skip("set SPELLHANCE_OUT to run the comparison")
	}
	iterations := int32(3000)
	if s := os.Getenv("SPELLHANCE_ITERATIONS"); s != "" {
		n, err := strconv.Atoi(s)
		if err != nil {
			t.Fatal(err)
		}
		iterations = int32(n)
	}

	var runs []run
	add := func(r run) {
		r.Iterations = iterations
		r.Seed = 7
		runs = append(runs, r)
	}
	durations := []time.Duration{2 * time.Minute, 5 * time.Minute}
	sodRows := []struct {
		b    build
		gear []string
	}{
		{weave2922, []string{"caster_mirage_2h", "enh_smashbringer_2h"}},
		{chain2922, []string{"caster_mirage_2h", "enh_smashbringer_2h"}},
		{weave8367, []string{"caster_mirage_2h", "enh_smashbringer_2h"}},
		{melee8367, []string{"caster_mirage_2h", "enh_smashbringer_2h"}},
		{enhPreset, []string{"enh_smashbringer_2h"}},
		{elePreset, []string{"caster_mirage_2h"}},
	}
	for _, row := range sodRows {
		for _, g := range row.gear {
			for _, d := range durations {
				for _, disable := range []bool{false, true} {
					for _, reaction := range []int32{150, 10} {
						add(run{Build: row.b, Gear: g, Duration: d, DisableCastSwingReset: disable, ReactionMs: reaction})
					}
				}
				// Sensitivities, measured rule and 150 ms only: no haste on the gear, and Scarlet
				// Smashbringer without its SoD GCD effect.
				add(run{Build: row.b, Gear: g, Duration: d, ReactionMs: 150, NoHaste: true})
				if g == "enh_smashbringer_2h" {
					add(run{Build: row.b, Gear: g + "_nogcd", Duration: d, ReactionMs: 150})
					add(run{Build: row.b, Gear: g + "_nogcd", Duration: d, ReactionMs: 150, NoHaste: true})
				}
			}
		}
	}
	// The presets as MythicSim opens them: their own launch gear, race, buffs, and 200 ms reaction.
	for _, d := range durations {
		for _, disable := range []bool{false, true} {
			add(run{Build: enhPreset, Gear: enhPresetLaunch, Duration: d, DisableCastSwingReset: disable, ReactionMs: 200, Preset: true})
			add(run{Build: elePreset, Gear: elePresetLaunch, Duration: d, DisableCastSwingReset: disable, ReactionMs: 200, Preset: true})
		}
	}

	type row struct {
		r                       run
		dps, stdev              float64
		boltsPerMin, swingPerMin float64
		oomSeconds              float64
	}
	rows := make([]row, len(runs))
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
				// Action metrics count casts over every iteration.
				minutes := r.Duration.Minutes() * float64(r.Iterations)
				bolts, swings := 0.0, 0.0
				for _, action := range player.Actions {
					id := action.Id
					casts := 0.0
					for _, target := range action.Targets {
						casts += float64(target.Casts)
					}
					switch {
					case id.GetOtherId() == proto.OtherAction_OtherActionAttack && id.Tag == 1:
						swings += casts
					case id.GetSpellId() == 1326047 || (id.GetSpellId() == 15208 && id.Tag == 0):
						// Lava Burst and Lightning Bolt, Maelstrom-instant bolts included; not overloads.
						bolts += casts
					}
				}
				rows[i] = row{r: r, dps: res.RaidMetrics.Dps.Avg, stdev: res.RaidMetrics.Dps.Stdev,
					boltsPerMin: bolts / minutes, swingPerMin: swings / minutes, oomSeconds: player.SecondsOomAvg}
			}
		}()
	}
	start := time.Now()
	for i := range runs {
		jobs <- i
	}
	close(jobs)
	wg.Wait()
	t.Logf("%d runs of %d iterations in %s", len(runs), iterations, time.Since(start).Round(time.Second))

	if err := os.MkdirAll(out, 0o755); err != nil {
		t.Fatal(err)
	}
	f, err := os.Create(filepath.Join(out, "results.csv"))
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	w := csv.NewWriter(f)
	w.Write([]string{"build", "gear", "duration_s", "swing_rule", "reaction_ms", "haste", "preset", "iterations",
		"dps", "dps_stdev", "dps_stderr", "lb_lvb_casts_per_min", "mh_swings_per_min", "seconds_oom"})
	for _, x := range rows {
		rule := "reset"
		if x.r.DisableCastSwingReset {
			rule = "old"
		}
		haste := "on"
		if x.r.NoHaste {
			haste = "off"
		}
		gearLabel := x.r.Gear
		if x.r.Preset {
			gearLabel = "preset launch"
		}
		w.Write([]string{x.r.Build.Label, gearLabel, fmt.Sprint(x.r.Duration.Seconds()), rule, fmt.Sprint(x.r.ReactionMs), haste,
			fmt.Sprint(x.r.Preset), fmt.Sprint(x.r.Iterations), fmt.Sprintf("%.1f", x.dps), fmt.Sprintf("%.1f", x.stdev),
			fmt.Sprintf("%.1f", x.stdev/math.Sqrt(float64(x.r.Iterations))), fmt.Sprintf("%.2f", x.boltsPerMin),
			fmt.Sprintf("%.2f", x.swingPerMin), fmt.Sprintf("%.1f", x.oomSeconds)})
	}
	w.Flush()
	if err := w.Error(); err != nil {
		t.Fatal(err)
	}
}
