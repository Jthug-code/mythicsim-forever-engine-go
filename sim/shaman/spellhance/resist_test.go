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

// TestResistSensitivity (2026-10-29) measures what boss resistances and spell hit do to each spec: the
// default boss has no school resistance, so the other runs carry only level-based partial resists (6%
// average at level 63) and the 17% base spell miss. Rows add nature or fire resistance to the boss, or
// move spell hit by 3%, on the launch gear at 5 minutes, with and without the tuning package. It writes
// resist.csv to $SPELLHANCE_RESIST_OUT and logs each spec's spell hit.
func TestResistSensitivity(t *testing.T) {
	out := os.Getenv("SPELLHANCE_RESIST_OUT")
	if out == "" {
		t.Skip("set SPELLHANCE_RESIST_OUT to run the sweep")
	}
	// Forever's Curse of the Elements (1311680) takes 75 off every magic resistance, Nature included, and
	// adds 10% magic damage taken; every buff set here has it, so "no CoE" rows show a raid without a
	// warlock on it.
	type variant struct {
		name              string
		nature, fire, hit float64
		noCoE             bool
	}
	variants := []variant{
		{"baseline", 0, 0, 0, false}, {"nature 50", 50, 0, 0, false}, {"nature 100", 100, 0, 0, false},
		{"nature 150", 150, 0, 0, false}, {"fire 300", 0, 300, 0, false}, {"spell hit -3%", 0, 0, -3, false},
		{"spell hit +3%", 0, 0, 3, false}, {"no CoE", 0, 0, 0, true}, {"no CoE, nature 75", 75, 0, 0, true},
	}
	rows := []struct {
		b    build
		gear string
	}{{elePreset, "ele_launch"}, {enhPreset, "enh_launch_1h"}, {melee8367, "enh_launch_1h"}, {weave2922, "ele_launch"}}

	for _, rw := range rows {
		r := run{Build: rw.b, Gear: rw.gear, Duration: time.Minute, ReactionMs: 150, Iterations: 1, Seed: 7}
		res := core.ComputeStats(&proto.ComputeStatsRequest{Raid: r.request().Raid})
		ps := res.RaidStats.Parties[0].Players[0].FinalStats.PseudoStats
		t.Logf("%s: spell hit %.1f%%, melee hit %.1f%%", rw.b.Label, ps[proto.PseudoStat_PseudoStatSpellHitPercent], ps[proto.PseudoStat_PseudoStatMeleeHitPercent])
	}

	iterations := matrixIterations(t)
	var lines [][]string
	for _, pkg := range []bool{false, true} {
		if pkg {
			shaman.ElementalFuryCritPerRank, shaman.ElementalFuryExtraCritBonus, shaman.LightningBaseDamageBonus = 1, 1, 0.06
		}
		part := make([][]string, len(rows)*len(variants))
		var wg sync.WaitGroup
		for i, rw := range rows {
			for j, v := range variants {
				wg.Add(1)
				go func() {
					defer wg.Done()
					r := run{Build: rw.b, Gear: rw.gear, Duration: 5 * time.Minute, ReactionMs: 150, Iterations: iterations, Seed: 7}
					req := r.request()
					target := core.FreshDefaultTargetConfig()
					s := stats.FromProtoArray(target.Stats)
					s[stats.NatureResistance], s[stats.FireResistance] = v.nature, v.fire
					target.Stats = s.ToProtoArray()
					req.Encounter.Targets = []*proto.Target{target}
					if v.noCoE {
						req.Raid.Debuffs.CurseOfElements = false
					}
					if v.hit != 0 {
						// A bonus SpellHitPercent pseudo stat is ignored (character.go reads only the school
						// ones), so the hit goes on Nature and Fire, every school these rotations cast.
						player := req.Raid.Parties[0].Players[0]
						bonus := &proto.UnitStats{Stats: stats.Stats{}.ToProtoArray(), PseudoStats: make([]float64, len(proto.PseudoStat_name))}
						bonus.PseudoStats[proto.PseudoStat_PseudoStatSchoolHitPercentNature] = v.hit
						bonus.PseudoStats[proto.PseudoStat_PseudoStatSchoolHitPercentFire] = v.hit
						player.BonusStats = bonus
					}
					res := core.RunRaidSim(req)
					if res.Error != nil {
						t.Errorf("%s %s: %s", rw.b.Label, v.name, res.Error.Message)
						return
					}
					part[i*len(variants)+j] = []string{fmt.Sprint(pkg), rw.b.Label, rw.gear, v.name, fmt.Sprintf("%.1f", res.RaidMetrics.Dps.Avg)}
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
	f, err := os.Create(filepath.Join(out, "resist.csv"))
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	w := csv.NewWriter(f)
	w.Write([]string{"package", "build", "gear", "variant", "dps"})
	w.WriteAll(lines)
}
