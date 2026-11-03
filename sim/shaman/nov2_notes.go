package shaman

import (
	"math"

	"github.com/wowsims/forever/sim/core/spelldata"
)

// The shaman changes in Kaivax's "WoW Forever Beta Development Notes - Updated November 2" (Blizzard
// forums topic 2360696, post 9, 3 November 2026 03:20 UTC). Client builds 1.60.1.70707 and 1.60.1.70753
// still state the old numbers for every one of them, so they ship as server hotfixes, and are patched
// over the store here. Where the notes give a ratio rather than a number, the ratio is applied to the
// client row named. When a build or a hotfix cache carries them, regenerate the store, drop this file,
// and let TestNov2Notes name each number the client states differently.
//
// Not modelled: Windfury Weapon and Windfury Totem only animate their extra swing; Totemic Call, Elemental
// Reach, Improved Ghost Wolf and Riptide change nothing a damage sim reads. Lava Burst's removal, Call of
// Thunder, Stormstrike's 150% and Maelstrom Weapon's rate are in the files that use them, from the
// constants below.

// "Lightning Bolt: Base damage of all ranks increased, with larger percentage increases on the higher
// ranks. Max rank was increased by 80%."
const nov2LightningBoltMaxRankGain = 1.8

// "Each rank's new base damage is roughly 78% of the base damage as the Classic Era version." The Era rows
// (1.15.9.69722) of ranks 4 to 9: the base at the rank's own level and the gain per level. Ranks 1 to 3 keep
// Forever's numbers: 78% of Era is below what they already deal, and the notes say every rank went up, by an
// amount they don't give. The Spellhance comparisons never cast them, but the repo's forever_lowrank Elemental
// preset fills with rank 2, so it trails the others by more than it should.
const nov2LightningBoltShareOfEra = 0.78

var nov2LightningBoltEra = map[int32]struct{ base, perLevel float64 }{
	915: {89, 1}, 943: {134, 1.2}, 6041: {183, 1.5}, 10391: {241, 1.7}, 10392: {299, 1.9}, 15207: {368, 2.1},
}

// "Chain Lightning cast time increased to 2.5 seconds. Primary target base damage increased to approximately
// 145% of comparable rank Lightning Bolt base damage. Mana cost increased to approximately 225% of comparable
// rank Lightning Bolt Mana cost." The comparable rank is read as the Lightning Bolt rank of the same level, or
// the closest below it. The coefficient (0.571 on rank 4) and the 6 sec cooldown stay the client's.
const (
	nov2ChainLightningCastMs     = 2500
	nov2ChainLightningDamageOfLB = 1.45
	nov2ChainLightningCostOfLB   = 2.25
)

var nov2ChainLightningComparable = map[int32]int32{421: 6041, 930: 10391, 2860: 10392, 10605: 15208}

// "Shocks: All shock base damage reverted to match Classic Era values." The Era rows' low and high at the
// rank's level and gain per level; Flame Shock's tick has no spread. Coefficients, costs and cooldowns stay
// Forever's: the notes name base damage only.
type nov2EraDamage struct{ low, high, perLevel float64 }

var nov2ShockEra = map[int32]nov2EraDamage{
	// Earth Shock
	8042: {17, 19, 0.5}, 8044: {32, 34, 0.7}, 8045: {60, 64, 1}, 8046: {119, 127, 1.4},
	10412: {225, 239, 2}, 10413: {359, 381, 2.6}, 10414: {517, 545, 3.1},
	// Flame Shock, the hit
	8050: {21, 21, 0.8}, 8052: {45, 45, 1.2}, 8053: {86, 86, 1.7}, 10447: {152, 152, 2.3}, 10448: {230, 230, 2.9},
	29228: {292, 292, 3.5},
	// Frost Shock
	8056: {89, 95, 1.2}, 8058: {206, 220, 1.9}, 10472: {333, 353, 2.5}, 10473: {486, 514, 3},
}

// Flame Shock's tick in the Era rows: 4 ticks over 12 sec, so rank 6 is 320 over the dot.
var nov2FlameShockTickEra = map[int32]float64{8050: 7, 8052: 12, 8053: 24, 10447: 42, 10448: 64, 29228: 80}

// "Lightning Overload: Talent now grants a 10/20/30% chance to trigger a duplicate spell, up from 3/7/10%."
var nov2LightningOverloadCurve = [][]float64{{10, 20, 30}}

// "Lava Burst: Removed. Replaced with Call of Thunder, which increases critical strike chance of Lightning
// Bolt and Chain Lightning by 5%." The client's Call of Thunder (16120) still states 3.
const nov2CallOfThunderCrit = 5

// "Maelstrom Weapon: Chance to trigger doubled." The client states no rate; the sim's 10 per minute is
// measured from beta logs (see applyMaelstromWeapon), so doubled is 20.
const nov2MaelstromWeaponPPM = 20

// "Stormstrike: Now deals 150% normalized weapon damage, up from 100%. Cooldown reduced to 6 seconds from 8."
const (
	nov2StormstrikeWeaponDamage = 1.5
	nov2StormstrikeCooldownMs   = 6000
)

// "Mental Dexterity: Now increases the Shaman's Attack Power by 50/100/150% of Intellect, up from 33/67/100%."
var nov2MentalDexterityCurve = [][]float64{{50, 100, 150}}

// "Flurry: Now increases attack speed by 6/12/18/24/30%, up from 5/10/15/20/25%."
var nov2FlurryCurve = [][]float64{{6, 12, 18, 24, 30}}

func init() {
	// Lightning Bolt first: Chain Lightning and both overloads are priced off its new rows.
	type gain struct{ base, perLevel float64 }
	boltGain := map[int32]gain{}
	for rank := int32(1); rank <= spellData.LightningBolt.Len(); rank++ {
		spelldata.Patch(spellData.LightningBolt.Rank(rank).ID, func(row *spelldata.Spell) {
			damage := row.DamageEffect()
			oldBase, oldPerLevel := damage.BasePoints, damage.PPL
			if era, ok := nov2LightningBoltEra[row.ID]; ok {
				damage.BasePoints = nov2LightningBoltShareOfEra * era.base
				damage.PPL = nov2LightningBoltShareOfEra * era.perLevel
			} else if rank == spellData.LightningBolt.Len() {
				damage.BasePoints *= nov2LightningBoltMaxRankGain
				damage.PPL *= nov2LightningBoltMaxRankGain
			}
			boltGain[rank] = gain{damage.BasePoints / oldBase, damage.PPL / oldPerLevel}
		})
	}
	// An overload keeps its share of its own bolt (half, on rank 10).
	for rank := int32(1); rank <= lightningBoltOverloadRanks.Len(); rank++ {
		spelldata.Patch(lightningBoltOverloadRanks.Rank(rank).ID, func(row *spelldata.Spell) {
			damage := row.DamageEffect()
			damage.BasePoints *= boltGain[rank].base
			damage.PPL *= boltGain[rank].perLevel
		})
	}

	for rank := int32(1); rank <= spellData.ChainLightning.Len(); rank++ {
		spelldata.Patch(spellData.ChainLightning.Rank(rank).ID, func(row *spelldata.Spell) {
			bolt := spelldata.MustFind(nov2ChainLightningComparable[row.ID])
			boltDamage, damage := bolt.DamageEffect(), row.DamageEffect()
			oldBase, oldPerLevel := damage.BasePoints, damage.PPL
			boltAtLevel := boltDamage.BasePoints + boltDamage.PPL*float64(damage.SpellLevel-boltDamage.SpellLevel)
			damage.BasePoints = nov2ChainLightningDamageOfLB * boltAtLevel
			damage.PPL = nov2ChainLightningDamageOfLB * boltDamage.PPL
			row.CastTimeMs = nov2ChainLightningCastMs
			row.Powers[0].Cost = int32(math.Round(nov2ChainLightningCostOfLB * bolt.Cost()))

			spelldata.Patch(chainLightningOverloadRanks.Rank(rank).ID, func(overload *spelldata.Spell) {
				overload.DamageEffect().BasePoints *= damage.BasePoints / oldBase
				overload.DamageEffect().PPL *= damage.PPL / oldPerLevel
			})
		})
	}

	for id, era := range nov2ShockEra {
		spelldata.Patch(id, func(row *spelldata.Spell) {
			damage := row.DamageEffect()
			damage.BasePoints = (era.low + era.high) / 2
			damage.Variance = (era.high - era.low) / damage.BasePoints
			damage.PPL = era.perLevel
		})
	}
	for id, tick := range nov2FlameShockTickEra {
		spelldata.Patch(id, func(row *spelldata.Spell) { row.PeriodicEffect().BasePoints = tick })
	}

	spelldata.Patch(spellData.CallOfThunder.Highest().ID, func(row *spelldata.Spell) {
		row.Effects[0].BasePoints = nov2CallOfThunderCrit
	})
	spelldata.Patch(spellData.Stormstrike.Highest().ID, func(row *spelldata.Spell) {
		row.CooldownMs = nov2StormstrikeCooldownMs
	})

	spelldata.SetCurve(spellData.LightningOverload.Highest().ID, nov2LightningOverloadCurve)
	spellData.LightningOverload = spelldata.Talent(spellData.LightningOverload.Highest().ID, 3)
	spelldata.SetCurve(spellData.MentalDexterity.Highest().ID, nov2MentalDexterityCurve)
	spellData.MentalDexterity = spelldata.Talent(spellData.MentalDexterity.Highest().ID, 3)
	spelldata.SetCurve(spellData.Flurry.Highest().ID, nov2FlurryCurve)
	spellData.Flurry = spelldata.Talent(spellData.Flurry.Highest().ID, 5)
}
