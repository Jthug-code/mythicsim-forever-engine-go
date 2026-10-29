package spellhance

import (
	"os"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
	googleProto "google.golang.org/protobuf/proto"
)

// Run 5 re-runs everything with SPELLHANCE_BUFFS=full: the engine's full raid, party, individual and
// debuff sets, and full consumables, for every build alike.
func fullBuffs() bool {
	return os.Getenv("SPELLHANCE_BUFFS") == "full"
}

// The Elemental preset takes Water Shield but its rotation never casts it. This build casts it, and
// trades Natural Grace 3 and Improved Reincarnation 2 for Totemic Focus 4 and Mana Tide Totem.
var eleResto = build{"Ele 32/0/19 Water Shield + Mana Tide", "5505331000001351--553040001001", "apls/ele_watershield.apl.json", true}

// applyFullBuffs sets run 5's buffs and tops up the player's consumables.
//   - Party: the engine's full set without Windfury Totem. The Enhancement builds run Grace of Air
//     (2026-10-29), and casters don't melee.
//   - Consumables: Greater Mageblood Elixir (Forever's 20 MP5) for every build, Brilliant Wizard Oil
//     for the Elemental spec, and Scrolls of Agility and Strength for the Enhancement spec, whose
//     weapon carries Windfury Weapon instead of an oil. Flasks, elixirs, food, potions, runes, Juju
//     and Zanza stay as the gear's preset has them. The engine models no world buffs.
func applyFullBuffs(b build, player *proto.Player) (*proto.PartyBuffs, *proto.RaidBuffs, *proto.Debuffs) {
	party := googleProto.Clone(core.FullPartyBuffs).(*proto.PartyBuffs)
	party.WindfuryTotem = false
	player.Buffs = googleProto.Clone(core.FullIndividualBuffs).(*proto.IndividualBuffs)

	c := player.Consumables
	if c == nil {
		c = enhConsumes()
		if b.Elemental {
			c = casterConsumes()
		}
	} else {
		c = googleProto.Clone(c).(*proto.ConsumesSpec)
	}
	c.GuardianElixirId = 250341
	if b.Elemental {
		c.MhTempEnchantId = 25122
	} else {
		c.ScrollAgi, c.ScrollStr = true, true
	}
	player.Consumables = c

	return party, googleProto.Clone(core.FullRaidBuffs).(*proto.RaidBuffs), googleProto.Clone(core.FullDebuffs).(*proto.Debuffs)
}
