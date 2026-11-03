package hunter

import (
	"testing"

	"github.com/wowsims/forever/sim/classcompare"
)

// TestClassCompare runs Marksmanship, Beast Mastery and ranged Survival on the launch gear set; see
// sim/classcompare.
func TestClassCompare(t *testing.T) {
	gear := "../../ui/specs/hunter/dps/gear_sets/launch.gear.json"
	classcompare.Run(t, "hunter", []classcompare.Spec{
		{Name: "Marksmanship", Suite: hunterSuite("mm", MarksmanshipTalents), Gear: gear, Role: classcompare.Ranged, Agility: true},
		{Name: "BeastMastery", Suite: hunterSuite("bm", BeastMasteryTalents), Gear: gear, Role: classcompare.Ranged, Agility: true},
		{Name: "Survival", Suite: hunterSuite("sv", SurvivalTalents), Gear: gear, Role: classcompare.Ranged, Agility: true},
	})
}
