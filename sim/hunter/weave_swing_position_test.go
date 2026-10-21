package hunter

import (
	"testing"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
	"google.golang.org/protobuf/encoding/protojson"
)

// Patch 191: a weaver kicks and steps out in the same rotation check, and its main hand comes due on
// the way out. Strider Kick puts the rotation on its 1.5 sec GCD, so no check runs during the walk and
// the hunter's position was last read at 5 yards: the swing used to land from 7.7 yards. It now reads
// where the hunter is, finds it out of melee range and does not happen.
func TestMeleeSwingReadsThePositionMidWalk(t *testing.T) {
	rotation := &proto.APLRotation{}
	// Arcanite Reaper swings standing in melee just after the pull (0.1 s on this seed) and again 3.8 s
	// later. Past the first second the kick goes out with under 0.35 s of the swing timer left, and the
	// walk to 12 yards at 9.1 a second takes 0.77 s, so the second swing comes due 2.7 yards out.
	if err := protojson.Unmarshal([]byte(`{"type": "TypeAPL", "priorityList": [
		{"action": {"condition": {"and": {"vals": [
				{"cmp": {"lhs": {"currentTime": {}}, "op": "OpGe", "rhs": {"const": {"val": "1s"}}}},
				{"cmp": {"lhs": {"autoTimeToNext": {"autoType": "MeleeAuto"}}, "op": "OpLe", "rhs": {"const": {"val": "0.35s"}}}}]}},
			"castSpell": {"spellId": {"spellId": 1317257}}}},
		{"action": {"condition": {"auraIsActive": {"auraId": {"spellId": 1317257}}}, "move": {"rangeFromTarget": {"const": {"val": "12"}}}}}
	]}`), rotation); err != nil {
		t.Fatal(err)
	}
	for _, reads := range []bool{true, false} {
		core.MeleeSwingReadsCurrentPosition = reads
		player := &proto.Player{
			Name: "weaver", Class: proto.Class_ClassHunter, Race: proto.Race_RaceOrc, TalentsString: SurvivalTalents,
			Equipment: WeaponsOnly, Consumables: &proto.ConsumesSpec{}, DistanceFromTarget: 5, ReactionTimeMs: 100,
			Spec: &proto.Player_Hunter{Hunter: &proto.Hunter{Options: &proto.Hunter_Options{ClassOptions: &proto.HunterOptions{
				Ammo: proto.HunterOptions_Doomshot, QuiverBonus: proto.HunterOptions_Speed15, PetType: proto.HunterOptions_PetNone}}}},
			Rotation: rotation,
		}
		encounter := core.MakeSingleTargetEncounter(0)
		encounter.Duration = 5
		encounter.DurationVariation = 0
		result := core.RunRaidSim(&proto.RaidSimRequest{
			SimOptions: &proto.SimOptions{RandomSeed: 1, Iterations: 1},
			Raid:       core.SinglePlayerRaidProto(player, &proto.PartyBuffs{}, &proto.RaidBuffs{}, &proto.Debuffs{}),
			Encounter:  encounter,
		})
		if result.Error != nil {
			t.Fatal(result.Error.Message)
		}
		var kicks, white int32
		for _, action := range result.RaidMetrics.Parties[0].Players[0].Actions {
			for _, target := range action.Targets {
				switch {
				case action.Id.GetSpellId() == 1317257:
					kicks += target.Casts
				case action.Id.GetOtherId() == proto.OtherAction_OtherActionAttack:
					white += target.Casts
				}
			}
		}
		want := int32(1)
		if !reads {
			want = 2
		}
		if kicks != 1 || white != want {
			t.Errorf("reads=%v: %d kicks and %d white swings, want 1 and %d", reads, kicks, white, want)
		}
	}
	core.MeleeSwingReadsCurrentPosition = true
}
