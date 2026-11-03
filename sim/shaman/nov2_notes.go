package shaman

import (
	"fmt"
	"math"
	"slices"

	"github.com/wowsims/forever/sim/core"
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
// values below.
//
// Each change can be undone on its own (SetNov2Change), so a test can price it: sim/shaman/spellhance's
// TestNov2Ablation. init applies them all.

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
// Bolt and Chain Lightning by 5%." The client's Call of Thunder (16120) still states 3. Off, the node gives
// nothing (Lava Burst stays removed).
const nov2CallOfThunderCrit = 5

var nov2CallOfThunder = true

// "Maelstrom Weapon: Chance to trigger doubled." The client states no rate; the sim's 10 per minute is
// measured from beta logs (see applyMaelstromWeapon), so doubled is 20.
var nov2MaelstromWeaponPPM = 20.0

// "Stormstrike: Now deals 150% normalized weapon damage, up from 100%. Cooldown reduced to 6 seconds from 8."
var nov2StormstrikeWeaponDamage = 1.5

const nov2StormstrikeCooldownMs = 6000

// "Mental Dexterity: Now increases the Shaman's Attack Power by 50/100/150% of Intellect, up from 33/67/100%."
var nov2MentalDexterityCurve = [][]float64{{50, 100, 150}}

// "Flurry: Now increases attack speed by 6/12/18/24/30%, up from 5/10/15/20/25%."
var nov2FlurryCurve = [][]float64{{6, 12, 18, 24, 30}}

// The client's curves for the three talents (70425, unchanged through 70753).
var (
	clientLightningOverloadCurve = [][]float64{{3, 7, 10}}
	clientMentalDexterityCurve   = [][]float64{{33, 67, 100}}
	clientFlurryCurve            = [][]float64{{5, 10, 15, 20, 25}}
)

// One change of the notes, with what applies it and what undoes it.
type nov2Change struct {
	name    string
	on, off func()
}

var nov2Changes []nov2Change

// Nov2Changes names every change this file applies, in the notes' order.
func Nov2Changes() []string {
	names := make([]string, len(nov2Changes))
	for i, c := range nov2Changes {
		names[i] = c.name
	}
	return names
}

// SetNov2Change applies or undoes one change. Spells read these values when a sim registers them, so call it
// between sims, never while one is being built.
func SetNov2Change(name string, on bool) {
	for _, c := range nov2Changes {
		if c.name == name {
			if on {
				c.on()
			} else {
				c.off()
			}
			return
		}
	}
	panic(fmt.Sprintf("shaman: no 2 November change %q", name))
}

// The parts of a store row these changes write: effect amounts, power costs, cast time and cooldown.
type nov2Row struct {
	effects    []spelldata.Effect
	powers     []spelldata.Power
	castTimeMs int32
	cooldownMs int32
}

func snapshotRows(ids []int32) map[int32]nov2Row {
	rows := make(map[int32]nov2Row, len(ids))
	for _, id := range ids {
		s := spelldata.MustFind(id)
		rows[id] = nov2Row{slices.Clone(s.Effects), slices.Clone(s.Powers), s.CastTimeMs, s.CooldownMs}
	}
	return rows
}

func restoreRows(rows map[int32]nov2Row) {
	for id, r := range rows {
		spelldata.Patch(id, func(s *spelldata.Spell) {
			copy(s.Effects, r.effects)
			copy(s.Powers, r.powers)
			s.CastTimeMs, s.CooldownMs = r.castTimeMs, r.cooldownMs
		})
	}
}

// addRowChange patches the rows and records both states, so the change can be undone and redone exactly.
func addRowChange(name string, ids []int32, patch func()) {
	before := snapshotRows(ids)
	patch()
	after := snapshotRows(ids)
	nov2Changes = append(nov2Changes, nov2Change{name, func() { restoreRows(after) }, func() { restoreRows(before) }})
}

func addVarChange(name string, set func(on bool)) {
	set(true)
	nov2Changes = append(nov2Changes, nov2Change{name, func() { set(true) }, func() { set(false) }})
}

func setCurve(ladder *spelldata.Ladder, ranks int32, curve [][]float64) {
	id := ladder.Highest().ID
	spelldata.SetCurve(id, curve)
	*ladder = spelldata.Talent(id, ranks)
}

func addCurveChange(name string, ladder *spelldata.Ladder, ranks int32, client, notes [][]float64) {
	addVarChange(name, func(on bool) { setCurve(ladder, ranks, core.Ternary(on, notes, client)) })
}

func ladderIDs(ladders ...spelldata.Ladder) []int32 {
	var ids []int32
	for _, l := range ladders {
		l.Each(func(_ int32, s *spelldata.Spell) { ids = append(ids, s.ID) })
	}
	return ids
}

func init() {
	// Lightning Bolt first: Chain Lightning is priced off its new rows.
	addRowChange("Lightning Bolt damage", ladderIDs(spellData.LightningBolt, lightningBoltOverloadRanks), func() {
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
	})

	addRowChange("Chain Lightning rework", ladderIDs(spellData.ChainLightning, chainLightningOverloadRanks), func() {
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
	})

	addRowChange("Shocks at Classic Era damage", ladderIDs(spellData.EarthShock, spellData.FlameShock, spellData.FrostShock), func() {
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
	})

	addCurveChange("Lightning Overload 10/20/30%", &spellData.LightningOverload, 3, clientLightningOverloadCurve, nov2LightningOverloadCurve)

	spelldata.Patch(spellData.CallOfThunder.Highest().ID, func(row *spelldata.Spell) {
		row.Effects[0].BasePoints = nov2CallOfThunderCrit
	})
	addVarChange("Call of Thunder", func(on bool) { nov2CallOfThunder = on })

	addVarChange("Maelstrom Weapon procs doubled", func(on bool) { nov2MaelstromWeaponPPM = core.Ternary(on, 20.0, 10.0) })
	addVarChange("Stormstrike 150% weapon damage", func(on bool) { nov2StormstrikeWeaponDamage = core.Ternary(on, 1.5, 1.0) })
	addRowChange("Stormstrike 6 s cooldown", ladderIDs(spellData.Stormstrike), func() {
		spelldata.Patch(spellData.Stormstrike.Highest().ID, func(row *spelldata.Spell) {
			row.CooldownMs = nov2StormstrikeCooldownMs
		})
	})

	addCurveChange("Mental Dexterity 50/100/150%", &spellData.MentalDexterity, 3, clientMentalDexterityCurve, nov2MentalDexterityCurve)
	addCurveChange("Flurry 6-30%", &spellData.Flurry, 5, clientFlurryCurve, nov2FlurryCurve)
}
