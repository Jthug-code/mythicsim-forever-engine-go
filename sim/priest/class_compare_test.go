package priest

import (
	"testing"

	"github.com/wowsims/forever/sim/classcompare"
)

// TestClassCompare runs Shadow and Smite on their launch gear sets; see sim/classcompare.
func TestClassCompare(t *testing.T) {
	dir := "../../ui/specs/priest/dps/gear_sets/"
	classcompare.Run(t, "priest", []classcompare.Spec{
		{Name: "Shadow", Suite: priestSuite("shadow", ShadowTalents, true), Gear: dir + "launch.gear.json",
			Role: classcompare.Caster, SchoolElixir: 9264},
		{Name: "Smite", Suite: priestSuite("smite", SmiteTalents, false), Gear: dir + "smite_launch.gear.json",
			Role: classcompare.Caster, SchoolElixir: 21546},
	})
}
