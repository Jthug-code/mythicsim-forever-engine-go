package feralcat

import (
	"testing"

	"github.com/wowsims/forever/sim/classcompare"
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
)

// TestClassCompare runs the Feral cat preset on the launch gear set; see sim/classcompare.
func TestClassCompare(t *testing.T) {
	classcompare.Run(t, "druid-feralcat", []classcompare.Spec{{
		Name: "FeralCat", Agility: true,
		Gear: "../../../ui/specs/druid/feralcat/gear_sets/launch.gear.json",
		Suite: core.CharacterSuiteConfig{
			Class:       proto.Class_ClassDruid,
			Talents:     DefaultTalents,
			SpecOptions: core.SpecOptionsCombo{Label: "Standard", SpecOptions: DefaultSpecOptions},
			Rotation:    core.GetAplRotation("../../../ui/specs/druid/feralcat/apls", "default"),
			Consumables: DefaultConsumables,
		},
	}})
}
