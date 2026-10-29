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

// TestCompensation (2026-10-29): on top of the tuning package, two more Elemental Fury riders that
// answer misses and resists where gear can't, +1% spell hit a rank or Spell Piercing a rank, against a
// plain boss, a raid without Curse of the Elements, the same with 75 nature resistance, and a 300 fire
// resistance boss. Launch gear, 5 minutes; compensate.csv goes to $SPELLHANCE_COMP_OUT.
func TestCompensation(t *testing.T) {
	out := os.Getenv("SPELLHANCE_COMP_OUT")
	if out == "" {
		t.Skip("set SPELLHANCE_COMP_OUT to run the riders")
	}
	type rider struct {
		name          string
		hit, pierce   float64
	}
	riders := []rider{{"package", 0, 0}, {"package + 1% hit/rank", 1, 0}, {"package + 5 pierce/rank", 0, 5}, {"package + 10 pierce/rank", 0, 10}}
	type boss struct {
		name         string
		nature, fire float64
		noCoE        bool
	}
	bosses := []boss{{"plain", 0, 0, false}, {"no CoE", 0, 0, true}, {"no CoE, nature 75", 75, 0, true}, {"fire 300", 0, 300, false}}
	rows := []struct {
		b    build
		gear string
	}{{elePreset, "ele_launch"}, {enhPreset, "enh_launch_1h"}, {melee8367, "enh_launch_1h"}}
	iterations := matrixIterations(t)
	var lines [][]string
	for _, rd := range riders {
		shaman.ElementalFuryCritPerRank, shaman.ElementalFuryExtraCritBonus, shaman.LightningBaseDamageBonus = 1, 1, 0.06
		shaman.ElementalFuryHitPerRank, shaman.ElementalFuryPiercePerRank = rd.hit, rd.pierce
		part := make([][]string, len(rows)*len(bosses))
		var wg sync.WaitGroup
		for i, rw := range rows {
			for j, bs := range bosses {
				wg.Add(1)
				go func() {
					defer wg.Done()
					r := run{Build: rw.b, Gear: rw.gear, Duration: 5 * time.Minute, ReactionMs: 150, Iterations: iterations, Seed: 7}
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
					part[i*len(bosses)+j] = []string{rd.name, rw.b.Label, bs.name, fmt.Sprintf("%.1f", res.RaidMetrics.Dps.Avg)}
				}()
			}
		}
		wg.Wait()
		lines = append(lines, part...)
	}
	shaman.ElementalFuryCritPerRank, shaman.ElementalFuryExtraCritBonus, shaman.LightningBaseDamageBonus = 0, 0, 0
	shaman.ElementalFuryHitPerRank, shaman.ElementalFuryPiercePerRank = 0, 0
	if err := os.MkdirAll(out, 0o755); err != nil {
		t.Fatal(err)
	}
	f, err := os.Create(filepath.Join(out, "compensate.csv"))
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	w := csv.NewWriter(f)
	w.Write([]string{"rider", "build", "boss", "dps"})
	w.WriteAll(lines)
}
