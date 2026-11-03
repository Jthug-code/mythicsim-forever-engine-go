package retribution

import (
	"testing"

	"github.com/wowsims/forever/sim/classcompare"
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
)

// TestClassCompare runs both Retribution builds TestRetribution tests (the Seal of Command twist build and
// deep Retribution) on the launch gear set, as Human, Dwarf or Undead; see sim/classcompare.
func TestClassCompare(t *testing.T) {
	suite := func(talents string) core.CharacterSuiteConfig {
		return core.CharacterSuiteConfig{
			Class:       proto.Class_ClassPaladin,
			Talents:     talents,
			SpecOptions: core.SpecOptionsCombo{Label: "Default", SpecOptions: DefaultOptions},
			Rotation:    core.GetAplRotation("../../../ui/specs/paladin/retribution/apls", "default"),
			Consumables: DefaultConsumables,
		}
	}
	gear := "../../../ui/specs/paladin/retribution/gear_sets/launch.gear.json"
	classcompare.Run(t, "paladin", []classcompare.Spec{
		{Name: "Retribution twist", Suite: suite(RetTalents), Gear: gear},
		{Name: "Retribution deep", Suite: suite(RetDeepTalents), Gear: gear},
	})
}
