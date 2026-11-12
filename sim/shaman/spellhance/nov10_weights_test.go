package spellhance

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/wowsims/forever/sim/core/stats"
	"github.com/wowsims/forever/sim/shaman"
)

// TestNov10Weights prices a point of each stat for 29/22 under the hotfix: the bolt-chain and weave rotations on
// the 29/22 BiS set, DPS gained per point, from one bump of each stat on the same seed. It writes weights.csv to
// $SPELLHANCE_NOV10_WEIGHTS_OUT; SPELLHANCE_BUFFS=full picks the full raid buffs.
func TestNov10Weights(t *testing.T) {
	out := os.Getenv("SPELLHANCE_NOV10_WEIGHTS_OUT")
	if out == "" {
		t.Skip("set SPELLHANCE_NOV10_WEIGHTS_OUT to run the stat weights")
	}
	iterations := matrixIterations(t)
	shaman.SetNov10Hotfix(true)
	defer shaman.SetNov10Hotfix(false)
	bumps := []struct {
		name  string
		bonus stats.Stats
		by    float64
	}{
		{"Intellect", stats.Stats{stats.Intellect: 50}, 50},
		{"Spell power", stats.Stats{stats.SpellDamage: 50}, 50},
		{"Crit rating (melee and spell)", stats.Stats{stats.SpellCritRating: 28, stats.MeleeCritRating: 28}, 28},
		{"Spell crit rating alone", stats.Stats{stats.SpellCritRating: 28}, 28},
		{"Hit rating (melee and spell)", stats.Stats{stats.SpellHitRating: 20, stats.MeleeHitRating: 20}, 20},
		{"Spell hit rating alone", stats.Stats{stats.SpellHitRating: 20}, 20},
		{"Haste rating (melee and spell)", stats.Stats{stats.SpellHasteRating: 20, stats.MeleeHasteRating: 20}, 20},
		{"Strength", stats.Stats{stats.Strength: 50}, 50},
		{"Agility", stats.Stats{stats.Agility: 50}, 50},
		{"Attack power", stats.Stats{stats.AttackPower: 100}, 100},
		{"Mana per 5", stats.Stats{stats.MP5: 20}, 20},
		{"Spell piercing", stats.Stats{stats.SpellPiercing: 20}, 20},
	}
	var lines string
	for _, b := range []build{
		{"29/22 AK3 bolt-chain", weave2922AK.Talents, "apls/chain.apl.json", false},
		{"29/22 AK3 weave", weave2922AK.Talents, "apls/weave.apl.json", false},
	} {
		for _, d := range []time.Duration{2 * time.Minute, 5 * time.Minute, 10 * time.Minute} {
			base := run{Build: b, Gear: foreverHybBisGear(), Duration: d, ReactionMs: 150, Iterations: iterations, Seed: 7, Air: "own"}
			baseDPS := simulate(t, base).RaidMetrics.Dps.Avg
			for _, bump := range bumps {
				r := base
				r.Bonus = bump.bonus
				dps := simulate(t, r).RaidMetrics.Dps.Avg
				lines += fmt.Sprintf("%s,%d,%.1f,%s,%.0f,%.1f,%.3f\n", b.Label, int(d.Seconds()), baseDPS, bump.name, bump.by, dps, (dps-baseDPS)/bump.by)
			}
		}
	}
	if err := os.MkdirAll(out, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(out, "weights.csv"), []byte("build,duration_s,base_dps,stat,bump,dps,dps_per_point\n"+lines), 0o644); err != nil {
		t.Fatal(err)
	}
}
