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
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/stats"
	"github.com/wowsims/forever/sim/shaman"
)

// TestLeanPackage (2026-10-29): Elemental Fury +1% crit and +1% spell hit a rank with crits for 2.25x
// (the physical crit talents' range), plus 6% Lightning Bolt and Chain Lightning base damage, against
// the first package (2.5x, no hit). Launch and SoD gear (the Enhancement preset spending Maelstrom at 10
// on SoD), 2 / 5 / 10 minutes, four bosses. lean.csv goes to $SPELLHANCE_LEAN_OUT.
func TestLeanPackage(t *testing.T) {
	out := os.Getenv("SPELLHANCE_LEAN_OUT")
	if out == "" {
		t.Skip("set SPELLHANCE_LEAN_OUT to run the packages")
	}
	type pkg struct {
		name                 string
		critBonus, crit, hit float64
	}
	pkgs := []pkg{{"none", 0, 0, 0}, {"jesse 2.5x", 1, 1, 0}, {"lean 2.25x + hit", 0.5, 1, 1}}
	type boss struct {
		name         string
		nature, fire float64
		noCoE        bool
	}
	bosses := []boss{{"plain", 0, 0, false}, {"no CoE", 0, 0, true}, {"no CoE, nature 75", 75, 0, true}, {"fire 300", 0, 300, false}}
	rows := []struct {
		b    build
		gear string
	}{{elePreset, "ele_launch"}, {enhPreset, "enh_launch_1h"}, {elePreset, "caster_mirage_2h_sets"}, {enhPresetAt10, "enh_smashbringer_2h_sets"}}
	durations := []time.Duration{2 * time.Minute, 5 * time.Minute, 10 * time.Minute}
	iterations := matrixIterations(t)
	var lines [][]string
	for _, p := range pkgs {
		shaman.ElementalFuryExtraCritBonus, shaman.ElementalFuryCritPerRank, shaman.ElementalFuryHitPerRank = p.critBonus, p.crit, p.hit
		shaman.LightningBaseDamageBonus = 0
		if p.name != "none" {
			shaman.LightningBaseDamageBonus = 0.06
		}
		var mu sync.Mutex
		var wg sync.WaitGroup
		for _, rw := range rows {
			for _, bs := range bosses {
				for _, d := range durations {
					wg.Add(1)
					go func() {
						defer wg.Done()
						r := run{Build: rw.b, Gear: rw.gear, Duration: d, ReactionMs: 150, Iterations: iterations, Seed: 7}
						req := r.request()
						target := core.FreshDefaultTargetConfig()
						s := stats.FromProtoArray(target.Stats)
						s[stats.NatureResistance], s[stats.FireResistance] = bs.nature, bs.fire
						target.Stats = s.ToProtoArray()
						req.Encounter.Targets = []*proto.Target{target}
						if bs.noCoE {
							req.Raid.Debuffs.CurseOfElements = false
						}
						res := core.RunRaidSim(req)
						if res.Error != nil {
							t.Errorf("%s: %s", rw.b.Label, res.Error.Message)
							return
						}
						mu.Lock()
						lines = append(lines, []string{p.name, rw.b.Label, rw.gear, bs.name, fmt.Sprint(d.Seconds()), fmt.Sprintf("%.1f", res.RaidMetrics.Dps.Avg)})
						mu.Unlock()
					}()
				}
			}
		}
		wg.Wait()
	}
	shaman.ElementalFuryExtraCritBonus, shaman.ElementalFuryCritPerRank, shaman.ElementalFuryHitPerRank, shaman.LightningBaseDamageBonus = 0, 0, 0, 0
	if err := os.MkdirAll(out, 0o755); err != nil {
		t.Fatal(err)
	}
	f, err := os.Create(filepath.Join(out, "lean.csv"))
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	w := csv.NewWriter(f)
	w.Write([]string{"package", "build", "gear", "boss", "duration_s", "dps"})
	w.WriteAll(lines)
}
