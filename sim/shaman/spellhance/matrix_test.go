package spellhance

import (
	"encoding/csv"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
)

// TestSpellhanceMatrix runs the full comparison and writes results.csv and sources.csv to
// $SPELLHANCE_OUT. SPELLHANCE_ITERATIONS sets the iterations per run (default 3000). Every run uses
// the measured swing rule; results-run1.csv has the first run (no set bonuses), with the old-rule
// reference rows.
func TestSpellhanceMatrix(t *testing.T) {
	out := os.Getenv("SPELLHANCE_OUT")
	if out == "" {
		t.Skip("set SPELLHANCE_OUT to run the comparison")
	}
	iterations := matrixIterations(t)

	var runs []run
	add := func(r run) {
		r.Iterations = iterations
		r.Seed = 7
		runs = append(runs, r)
	}
	type gearRow struct {
		gear   string
		builds []build
	}
	durations := []time.Duration{2 * time.Minute, 5 * time.Minute}
	hybrids := []build{weave2922, weave2922CL, chain2922, chain2922CL, weave8367, melee8367}
	for _, sets := range []bool{false, true} {
		suffix := ""
		if sets {
			suffix = "_sets"
		}
		rows := []gearRow{
			{"caster_mirage_2h" + suffix, append(append([]build{}, hybrids...), elePreset)},
			{"enh_smashbringer_2h" + suffix, append(append([]build{}, hybrids...), enhPreset)},
		}
		if sets {
			// Only Soulcrusher's Rage 6 stacks Maelstrom Weapon past 5.
			rows[1].builds = append(rows[1].builds, weave8367at10, melee8367at10)
			// Soulcrusher's Rage 4 read onto Fire Nova, which only the Enhancement preset casts.
			rows = append(rows, gearRow{"enh_smashbringer_2h_sets_rage4", []build{enhPreset}})
		}
		for _, row := range rows {
			for _, b := range row.builds {
				for _, d := range durations {
					for _, reaction := range []int32{150, 10} {
						add(run{Build: b, Gear: row.gear, Duration: d, ReactionMs: reaction})
					}
					if !sets {
						continue
					}
					// Sensitivities at 150 ms: no haste on the gear, and Scarlet Smashbringer
					// without its SoD GCD effect.
					add(run{Build: b, Gear: row.gear, Duration: d, ReactionMs: 150, NoHaste: true})
					if row.gear == "enh_smashbringer_2h_sets" {
						add(run{Build: b, Gear: "enh_smashbringer_2h_nogcd_sets", Duration: d, ReactionMs: 150})
						add(run{Build: b, Gear: "enh_smashbringer_2h_nogcd_sets", Duration: d, ReactionMs: 150, NoHaste: true})
					}
				}
			}
		}
	}

	runMatrix(t, runs, out)
}

// TestSpellhanceForeverMatrix runs the builds on the repo's Forever launch-gear presets
// (ui/specs/shaman/*/gear_sets/launch.gear.json; the Enhancement one without its off-hand) and writes results.csv and sources.csv to
// $SPELLHANCE_FOREVER_OUT. These are Forever items: no SoD set bonuses, and the engine's own item
// sets apply as it knows them.
func TestSpellhanceForeverMatrix(t *testing.T) {
	out := os.Getenv("SPELLHANCE_FOREVER_OUT")
	if out == "" {
		t.Skip("set SPELLHANCE_FOREVER_OUT to run the comparison")
	}
	iterations := matrixIterations(t)
	var runs []run
	builds := []build{weave2922, weave2922CL, chain2922, weave8367, melee8367}
	for _, row := range []struct {
		gear   string
		builds []build
	}{
		{"ele_launch", append(append([]build{}, builds...), elePreset)},
		// Shamans can't dual wield in Forever, so the Enhancement gear goes without its off-hand.
		{"enh_launch_1h", append(append([]build{}, builds...), enhPreset)},
		// The Enhancement preset as the repo ships it, dual-wielding, for reference only.
		{"enh_launch", []build{enhPreset}},
	} {
		for _, b := range row.builds {
			for _, d := range []time.Duration{2 * time.Minute, 5 * time.Minute} {
				for _, reaction := range []int32{150, 10} {
					runs = append(runs, run{Build: b, Gear: row.gear, Duration: d, ReactionMs: reaction, Iterations: iterations, Seed: 7})
				}
			}
		}
	}
	runMatrix(t, runs, out)
}

func matrixIterations(t *testing.T) int32 {
	if s := os.Getenv("SPELLHANCE_ITERATIONS"); s != "" {
		n, err := strconv.Atoi(s)
		if err != nil {
			t.Fatal(err)
		}
		return int32(n)
	}
	return 3000
}

// runMatrix runs every run on all CPUs and writes results.csv and sources.csv to out.
func runMatrix(t *testing.T, runs []run, out string) {
	type source struct {
		name     string
		physical bool
		dps      float64
	}
	type result struct {
		r                        run
		dps, stdev               float64
		physicalDPS, spellDPS    float64
		boltsPerMin, swingPerMin float64
		oomSeconds               float64
		sources                  []source
	}
	results := make([]result, len(runs))
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
				// Action metrics total every iteration.
				perSecond := 1 / (r.Duration.Seconds() * float64(r.Iterations))
				perMinute := 60 * perSecond
				x := result{r: r, dps: res.RaidMetrics.Dps.Avg, stdev: res.RaidMetrics.Dps.Stdev, oomSeconds: player.SecondsOomAvg}
				for _, unit := range append([]*proto.UnitMetrics{player}, player.Pets...) {
					for _, action := range unit.Actions {
						id := action.Id
						casts, damage := 0.0, 0.0
						for _, target := range action.Targets {
							casts += float64(target.Casts)
							damage += target.Damage
						}
						switch {
						case unit != player:
						case id.GetOtherId() == proto.OtherAction_OtherActionAttack && id.Tag == 1:
							x.swingPerMin += casts * perMinute
						case id.GetSpellId() == lavaBurstID || id.GetSpellId() == 10605 || id.GetSpellId() == 15208:
							// Lava Burst, Chain Lightning and Lightning Bolt casts, Maelstrom bolts included;
							// overloads (tag 6) are not casts.
							if id.Tag == 0 {
								x.boltsPerMin += casts * perMinute
							}
						}
						if damage == 0 {
							continue
						}
						// By the school of the damage event: only plain Physical counts as physical.
						physical := action.SpellSchool == int32(core.SpellSchoolPhysical)
						if physical {
							x.physicalDPS += damage * perSecond
						} else {
							x.spellDPS += damage * perSecond
						}
						x.sources = append(x.sources, source{name: sourceName(id, unit != player), physical: physical, dps: damage * perSecond})
					}
				}
				if sum := x.physicalDPS + x.spellDPS; math.Abs(sum-x.dps) > 0.005*x.dps {
					t.Errorf("%s on %s: the damage sources add up to %.1f DPS, the sim says %.1f", r.Build.Label, r.Gear, sum, x.dps)
				}
				sort.Slice(x.sources, func(a, b int) bool { return x.sources[a].dps > x.sources[b].dps })
				results[i] = x
			}
		}()
	}
	start := time.Now()
	for i := range runs {
		jobs <- i
	}
	close(jobs)
	wg.Wait()
	t.Logf("%d runs of %d iterations in %s", len(runs), runs[0].Iterations, time.Since(start).Round(time.Second))

	if err := os.MkdirAll(out, 0o755); err != nil {
		t.Fatal(err)
	}
	write := func(name string, header []string, rows [][]string) {
		f, err := os.Create(filepath.Join(out, name))
		if err != nil {
			t.Fatal(err)
		}
		defer f.Close()
		w := csv.NewWriter(f)
		w.Write(header)
		w.WriteAll(rows)
		if err := w.Error(); err != nil {
			t.Fatal(err)
		}
	}
	var main, sources [][]string
	for _, x := range results {
		haste := "on"
		if x.r.NoHaste {
			haste = "off"
		}
		sets := strings.Contains(x.r.Gear, "_sets")
		gear := strings.NewReplacer("_sets_rage4", "+rage4", "_sets", "").Replace(x.r.Gear)
		key := []string{x.r.Build.Label, gear, fmt.Sprint(sets), fmt.Sprint(x.r.Duration.Seconds()), fmt.Sprint(x.r.ReactionMs), haste}
		total := x.physicalDPS + x.spellDPS
		main = append(main, append(append([]string{}, key...),
			fmt.Sprint(x.r.Iterations), fmt.Sprintf("%.1f", x.dps), fmt.Sprintf("%.1f", x.stdev/math.Sqrt(float64(x.r.Iterations))),
			fmt.Sprintf("%.1f", x.physicalDPS), fmt.Sprintf("%.1f", x.spellDPS), fmt.Sprintf("%.3f", x.physicalDPS/total),
			fmt.Sprintf("%.2f", x.boltsPerMin), fmt.Sprintf("%.2f", x.swingPerMin), fmt.Sprintf("%.1f", x.oomSeconds)))
		for _, s := range x.sources {
			school := "spell"
			if s.physical {
				school = "physical"
			}
			sources = append(sources, append(append([]string{}, key...), s.name, school, fmt.Sprintf("%.1f", s.dps), fmt.Sprintf("%.3f", s.dps/total)))
		}
	}
	keyHeader := []string{"build", "gear", "set_bonuses", "duration_s", "reaction_ms", "haste"}
	write("results.csv", append(append([]string{}, keyHeader...), "iterations", "dps", "dps_stderr", "physical_dps", "spell_dps",
		"physical_share", "lb_lvb_cl_casts_per_min", "mh_swings_per_min", "seconds_oom"), main)
	write("sources.csv", append(append([]string{}, keyHeader...), "source", "school", "dps", "share"), sources)
}

var sourceNames = map[string]string{
	"o3/1":       "Melee swing",
	"s439440/0":  "Windfury Attack",
	"s17364/1":   "Stormstrike (weapon strike)",
	"s15208/0":   "Lightning Bolt",
	"s15208/6":   "Lightning Bolt overload",
	"s10605/0":   "Chain Lightning",
	"s10605/6":   "Chain Lightning overload",
	"s1326047/0": "Lava Burst",
	"s10414/0":   "Earth Shock",
	"s29228/0":   "Flame Shock (hit)",
	"s29228/1":   "Flame Shock (ticks)",
	"s10436/0":   "Searing Totem",
	"s408345/0":  "Fire Nova",
	"s915/0":     "Lightning Bolt rank 4",
	"s915/6":     "Lightning Bolt rank 4 overload",
	"s15851/0":   "Dragonbreath Chili",
	"s10587/0":   "Magma Totem",
}

func sourceName(id *proto.ActionID, pet bool) string {
	key := fmt.Sprintf("s%d/%d", id.GetSpellId(), id.Tag)
	if id.GetOtherId() != 0 {
		key = fmt.Sprintf("o%d/%d", id.GetOtherId(), id.Tag)
	} else if id.GetItemId() != 0 {
		key = fmt.Sprintf("i%d/%d", id.GetItemId(), id.Tag)
	}
	name, ok := sourceNames[key]
	if !ok {
		name = key
	}
	if pet {
		name = "pet: " + name
	}
	return name
}
