package mage

import (
	"testing"
	"time"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/simsignals"
)

// lastChargeFrostboltThenIceLance casts Frostbolt on the last Fingers of Frost charge from 20 yards and
// Ice Lance as soon as the Frostbolt leaves the hand. It reports whether the charge was still up when
// Ice Lance went out, and whether it is gone once both have landed.
func lastChargeFrostboltThenIceLance(t *testing.T) (frozenForIceLance bool, activeAfter bool) {
	t.Helper()
	player := &proto.Player{
		Name: "Mage", Class: proto.Class_ClassMage, Race: proto.Race_RaceGnome,
		TalentsString: FrostTalents, Equipment: &proto.EquipmentSpec{}, DistanceFromTarget: 20,
		Spec:     &proto.Player_Mage{Mage: &proto.Mage{Options: &proto.Mage_Options{ClassOptions: &proto.MageOptions{}}}},
		Rotation: &proto.APLRotation{Type: proto.APLRotation_TypeAPL},
	}
	sim := core.NewSim(&proto.RaidSimRequest{
		SimOptions: &proto.SimOptions{RandomSeed: 42},
		Raid:       core.SinglePlayerRaidProto(player, &proto.PartyBuffs{}, &proto.RaidBuffs{}, &proto.Debuffs{}),
		Encounter:  core.MakeSingleTargetEncounter(0),
	}, simsignals.CreateSignals())
	sim.Reset()
	mage := sim.Raid.Parties[0].Players[0].(MageAgent).GetMage()
	if mage.FingersOfFrostAura == nil {
		t.Fatal("FrostTalents no longer take Fingers of Frost; pick a build that does")
	}
	// No fresh charges from the Frostbolt's chill.
	mage.GetAura("Fingers of Frost Trigger").Deactivate(sim)
	mage.FingersOfFrostAura.Activate(sim)
	mage.FingersOfFrostAura.SetStacks(sim, 1)

	frostbolt := mage.GetSpell(core.ActionID{SpellID: 25304})
	iceLance := mage.GetSpell(core.ActionID{SpellID: 1240047})
	if frostbolt.TravelTime() <= iceLance.TravelTime() || iceLance.TravelTime() <= 0 {
		t.Fatalf("travel times: Frostbolt %v, Ice Lance %v", frostbolt.TravelTime(), iceLance.TravelTime())
	}
	if !frostbolt.Cast(sim, mage.CurrentTarget) {
		t.Fatal("Frostbolt did not cast")
	}
	for mage.Hardcast.Expires > sim.CurrentTime || !iceLance.CanCast(sim, mage.CurrentTarget) {
		if sim.Step() {
			t.Fatal("simulation ended before the Frostbolt left the hand")
		}
	}
	if frostbolt.SpellMetrics[0].Casts != 1 || frostbolt.Unit.SpellsInFlight[frostbolt] != 1 {
		t.Fatal("Frostbolt is not in flight")
	}
	frozenForIceLance = mage.IsTargetFrozen()
	if !iceLance.Cast(sim, mage.CurrentTarget) {
		t.Fatal("Ice Lance did not cast")
	}
	end := sim.CurrentTime + time.Second
	for sim.CurrentTime < end {
		if sim.Step() {
			break
		}
	}
	return frozenForIceLance, mage.FingersOfFrostAura.IsActive()
}

func TestFingersOfFrostLastChargeSpentOnImpact(t *testing.T) {
	defer func(old bool) { FingersOfFrostSpentOnImpact = old }(FingersOfFrostSpentOnImpact)

	FingersOfFrostSpentOnImpact = false
	if frozen, active := lastChargeFrostboltThenIceLance(t); frozen || active {
		t.Errorf("spent on cast: Ice Lance frozen = %v, charge left = %v; want false, false", frozen, active)
	}

	FingersOfFrostSpentOnImpact = true
	if frozen, active := lastChargeFrostboltThenIceLance(t); !frozen || active {
		t.Errorf("spent on impact: Ice Lance frozen = %v, charge left = %v; want true, false", frozen, active)
	}
}
