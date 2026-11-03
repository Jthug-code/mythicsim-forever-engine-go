package warlock

import (
	"testing"

	"github.com/wowsims/forever/sim/classcompare"
	"github.com/wowsims/forever/sim/core/proto"
)

// TestClassCompare runs Affliction and Destruction on the launch gear set; see sim/classcompare.
func TestClassCompare(t *testing.T) {
	gear := "../../ui/specs/warlock/dps/gear_sets/launch.gear.json"
	classcompare.Run(t, "warlock", []classcompare.Spec{
		{Name: "Affliction", Gear: gear, Role: classcompare.Caster, SchoolElixir: 9264,
			Suite: warlockSuite("affliction", AfflictionTalents, &proto.WarlockOptions{
				Summon: proto.WarlockOptions_Succubus, Armor: proto.WarlockOptions_DemonArmor, CurseOptions: proto.WarlockOptions_Elements,
			})},
		{Name: "Destruction", Gear: gear, Role: classcompare.Caster, SchoolElixir: 9264,
			Suite: warlockSuite("destruction_conflag", DestructionTalents, &proto.WarlockOptions{
				Summon: proto.WarlockOptions_Succubus, SacrificeSummon: true, Armor: proto.WarlockOptions_DemonArmor,
				CurseOptions: proto.WarlockOptions_Elements,
			})},
	})
}
