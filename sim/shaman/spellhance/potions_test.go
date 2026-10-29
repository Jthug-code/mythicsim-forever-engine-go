package spellhance

import (
	"testing"
	"time"

	"github.com/wowsims/forever/sim/core"
)

// A shaman that runs out of mana drinks its Major Mana Potion (13444) and Demonic Rune (12662). Before
// the harness filled the Potions and ConjuredItems lists, it drank neither.
func TestManaPotionsAreDrunk(t *testing.T) {
	r := run{Build: melee8367, Gear: "enh_launch_1h", Duration: 5 * time.Minute, ReactionMs: 150, Iterations: 50, Seed: 7}
	res := core.RunRaidSim(r.request())
	if res.Error != nil {
		t.Fatal(res.Error.Message)
	}
	uses := map[int32]int32{}
	for _, a := range res.RaidMetrics.Parties[0].Players[0].Actions {
		for _, tg := range a.Targets {
			uses[a.Id.GetItemId()] += tg.Casts
		}
	}
	for id, name := range map[int32]string{13444: "Major Mana Potion", 12662: "Demonic Rune"} {
		if uses[id] == 0 {
			t.Errorf("no %s drunk in %d out-of-mana fights", name, r.Iterations)
		}
	}
}
