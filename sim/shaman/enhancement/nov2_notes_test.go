package enhancement

import (
	"math"
	"testing"
	"time"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/simsignals"
	"github.com/wowsims/forever/sim/shaman"
)

// "Stormstrike: Now deals 150% normalized weapon damage, up from 100%. Cooldown reduced to 6 seconds from 8
// seconds" (2 November 2026 notes). With crits taken away, the strike's hits average 1.5 times Sulfuras's
// normalized damage (attack power at 3.3 sec, not the axe's 3.7), times what the sim makes of any base on
// this spell against this target.
func TestStormstrikeNov2(t *testing.T) {
	sim := core.NewSim(&proto.RaidSimRequest{
		SimOptions: &proto.SimOptions{RandomSeed: 1},
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

	if cd := sham.Stormstrike.CD.Duration; cd != 6*time.Second {
		t.Errorf("Stormstrike's cooldown is %s, want 6s", cd)
	}

	hit := sham.GetSpell(shaman.StormstrikeActionID.WithTag(1))
	target := sham.CurrentTarget
	hit.BonusCritPercent -= 200
	scale := hit.CalcDamage(sim, target, 1000, hit.OutcomeAlwaysHit).Damage / 1000

	const casts = 20000
	for range casts {
		hit.Cast(sim, target)
	}
	m := hit.SpellMetrics[target.UnitIndex]
	if m.Hits < casts/2 || m.Crits != 0 || m.Blocks != 0 || m.Glances != 0 {
		t.Fatalf("%d hits, %d crits, %d blocks, %d glances from %d casts", m.Hits, m.Crits, m.Blocks, m.Glances, casts)
	}
	mh := sham.AutoAttacks.MH()
	normalized := (mh.BaseDamageMin+mh.BaseDamageMax)/2 + mh.NormalizedSwingSpeed*hit.MeleeAttackPower(target)/mh.AttackPowerPerDPS
	want := 1.5 * normalized * scale
	if got := m.TotalDamage / float64(m.Hits); math.Abs(got/want-1) > 0.01 {
		t.Errorf("Stormstrike hits average %.1f, want %.1f (1.5 x %.1f normalized x %.3f)", got, want, normalized, scale)
	}
}
