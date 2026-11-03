package classcompare

import (
	"slices"
	"testing"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
)

// A mana user's potion and rune reach the lists the engine registers them from, at either buff level; a
// warrior gets neither.
func TestConsumablesRegisterPotions(t *testing.T) {
	for _, level := range Levels {
		mage := consumables(Spec{Role: Caster, Suite: core.CharacterSuiteConfig{Class: proto.Class_ClassMage}}, level)
		if !slices.Contains(mage.Potions, 13444) || !slices.Contains(mage.ConjuredItems, 12662) {
			t.Errorf("%s: a mage's potions %v and conjured items %v miss Major Mana Potion and Demonic Rune", level, mage.Potions, mage.ConjuredItems)
		}
		warrior := consumables(Spec{Suite: core.CharacterSuiteConfig{Class: proto.Class_ClassWarrior}}, level)
		if len(warrior.Potions) != 0 || len(warrior.ConjuredItems) != 0 {
			t.Errorf("%s: a warrior gets mana consumables %v %v", level, warrior.Potions, warrior.ConjuredItems)
		}
	}
}
