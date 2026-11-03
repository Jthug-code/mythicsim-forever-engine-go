package rogue

import (
	"testing"

	"github.com/wowsims/forever/sim/classcompare"
)

// TestClassCompare runs the three rogue suites on the launch gear their rotations want; see sim/classcompare.
func TestClassCompare(t *testing.T) {
	dir := "../../ui/specs/rogue/dps/gear_sets/"
	classcompare.Run(t, "rogue", []classcompare.Spec{
		{Name: "Combat", Suite: rogueSuite("combat_sinister_strike", CombatTalents), Gear: dir + "sinister_strike_launch.gear.json", Agility: true},
		{Name: "Assassination", Suite: rogueSuite("forever_mutilate", AssassinationTalents), Gear: dir + "backstab_launch.gear.json", Agility: true},
		{Name: "Subtlety", Suite: rogueSuite("forever_hemorrhage", SubtletyTalents), Gear: dir + "sinister_strike_launch.gear.json", Agility: true},
	})
}
