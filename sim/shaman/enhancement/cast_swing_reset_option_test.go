package enhancement

import (
	"testing"
	"time"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/simsignals"
	"github.com/wowsims/forever/sim/shaman"
)

// SimOptions.DisableCastSwingReset brings back the rule before patch 17 for comparison runs: a hard
// cast that ends before the next swing leaves the swing alone, and one that runs past it pushes it a
// full swing past the cast. With the option off, the measured Forever rule holds: every completed
// hard cast restarts the swing timer (TestHardCastRestartsTheSwingTimer).
func TestDisableCastSwingReset(t *testing.T) {
	// The next swing, counted from the end and from the start of a Lightning Bolt rank 1 (1.5 s)
	// started sinceSwing after the last swing landed.
	nextSwing := func(disable bool, sinceSwing func(speed time.Duration) time.Duration) (fromCastEnd, fromStart, speed time.Duration) {
		sim := core.NewSim(&proto.RaidSimRequest{
			SimOptions: &proto.SimOptions{RandomSeed: 1, DisableCastSwingReset: disable},
			Raid: core.SinglePlayerRaidProto(&proto.Player{
				Name: "enh", Class: proto.Class_ClassShaman, Race: proto.Race_RaceOrc, TalentsString: DefaultTalents,
				Equipment: &proto.EquipmentSpec{Items: []*proto.ItemSpec{
					{}, {}, {}, {}, {}, {}, {}, {}, {}, {}, {}, {}, {}, {},
					{Id: 17182}, // Sulfuras, Hand of Ragnaros
				}},
				Spec:     &proto.Player_EnhancementShaman{EnhancementShaman: &proto.EnhancementShaman{Options: &proto.EnhancementShaman_Options{ClassOptions: &proto.ShamanOptions{}}}},
				Rotation: &proto.APLRotation{Type: proto.APLRotation_TypeAPL},
			}, &proto.PartyBuffs{}, &proto.RaidBuffs{}, &proto.Debuffs{}),
			Encounter: core.MakeSingleTargetEncounter(0),
		}, simsignals.CreateSignals())
		sim.Reset()
		sham := sim.Raid.Parties[0].Players[0].(shaman.ShamanAgent).GetShaman()
		sham.AutoAttacks.EnableAutoSwing(sim)
		speed = sham.AutoAttacks.MainhandSwingSpeed()
		sham.AutoAttacks.StopMeleeUntil(sim, sim.CurrentTime-sinceSwing(speed)) // the last swing landed sinceSwing ago
		start := sim.CurrentTime
		spell := sham.GetSpell(core.ActionID{SpellID: 403})
		if !spell.Cast(sim, sham.CurrentTarget) {
			t.Fatal("Lightning Bolt rank 1 did not cast")
		}
		castEnd := sim.CurrentTime + spell.CurCast.CastTime
		next := sham.AutoAttacks.NextAttackAt()
		return next - castEnd, next - start, speed
	}
	justLanded := func(time.Duration) time.Duration { return 0 }
	dueMidCast := func(speed time.Duration) time.Duration { return speed - time.Second }

	// A swing just landed, so the next one is a full swing away and the bolt ends before it.
	if got, _, speed := nextSwing(false, justLanded); got != speed {
		t.Errorf("reset on, bolt between swings: next swing %s after the cast ends, want %s", got, speed)
	}
	if _, got, speed := nextSwing(true, justLanded); got != speed {
		t.Errorf("reset off, bolt between swings: next swing %s after the cast started, want %s (untouched)", got, speed)
	}

	// The swing is due 1 s into the 1.5 s bolt. Both rules put it a full swing past the cast.
	for _, disable := range []bool{false, true} {
		if got, _, speed := nextSwing(disable, dueMidCast); got != speed {
			t.Errorf("reset off = %v, swing due mid-cast: next swing %s after the cast ends, want %s", disable, got, speed)
		}
	}
}
