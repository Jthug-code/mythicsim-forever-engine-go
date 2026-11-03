package balance

import (
	"testing"

	"github.com/wowsims/forever/sim/classcompare"
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
)

// TestClassCompare runs the Balance preset (TestBalance's build and launch rotation) on the launch gear set;
// see sim/classcompare.
func TestClassCompare(t *testing.T) {
	classcompare.Run(t, "druid-balance", []classcompare.Spec{{
		Name: "Balance", Role: classcompare.Caster,
		Gear: "../../../ui/specs/druid/balance/gear_sets/launch.gear.json",
		Suite: core.CharacterSuiteConfig{
			Class: proto.Class_ClassDruid,
			SpecOptions: core.SpecOptionsCombo{Label: "Standard", SpecOptions: &proto.Player_BalanceDruid{
				BalanceDruid: &proto.BalanceDruid{Options: &proto.BalanceDruid_Options{ClassOptions: &proto.DruidOptions{}}},
			}},
			Talents:  DefaultTalents,
			Rotation: core.GetAplRotation("../../../ui/specs/druid/balance/apls", "launch"),
		},
	}})
}
