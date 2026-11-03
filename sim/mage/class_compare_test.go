package mage

import (
	"testing"

	"github.com/wowsims/forever/sim/classcompare"
)

// TestClassCompare runs the three mage specs on their launch gear sets; see sim/classcompare.
func TestClassCompare(t *testing.T) {
	dir := "../../ui/specs/mage/dps/gear_sets/"
	classcompare.Run(t, "mage", []classcompare.Spec{
		{Name: "Arcane", Suite: mageSuite("arcane", ArcaneTalents), Gear: dir + "arcane_launch.gear.json", Role: classcompare.Caster},
		{Name: "Fire", Suite: mageSuite("fire", FireTalents), Gear: dir + "fire_launch.gear.json", Role: classcompare.Caster, SchoolElixir: 6373},
		{Name: "Frost", Suite: mageSuite("frost", FrostTalents), Gear: dir + "frost_launch.gear.json", Role: classcompare.Caster, SchoolElixir: 17708},
	})
}
