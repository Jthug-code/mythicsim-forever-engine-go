package dps

import (
	"testing"

	"github.com/wowsims/forever/sim/classcompare"
	"github.com/wowsims/forever/sim/core"
)

// TestClassCompare runs Fury and Arms on their launch gear sets; see sim/classcompare. Fury drops the APL's
// Slam line (classcompare.WithoutSpell): untalented, only cast speed from a racial ever triggers it.
func TestClassCompare(t *testing.T) {
	fury := warriorSuite(DualWieldGear, FuryTalents, nil)
	fury.Rotation.Rotation = classcompare.WithoutSpell(fury.Rotation.Rotation, 11605)
	arms := warriorSuite(TwoHandGear, ArmsTalents, nil)
	arms.Rotation = core.GetAplRotation("../../../ui/specs/warrior/dps/apls", "dps_battle")
	classcompare.Run(t, "warrior", []classcompare.Spec{
		{Name: "Fury", Suite: fury, Gear: "../../../ui/specs/warrior/dps/gear_sets/launch.gear.json"},
		{Name: "Arms", Suite: arms, Gear: "../../../ui/specs/warrior/dps/gear_sets/arms_launch.gear.json"},
	})
}
