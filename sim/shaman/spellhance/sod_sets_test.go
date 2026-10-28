package spellhance

import (
	"time"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/stats"
	"github.com/wowsims/forever/sim/shaman"
)

// The Season of Discovery set bonuses that work with Forever's talents, as Wowhead's item-set pages
// (classic/item-set=1951, 1952, 1889) word them. Each one is registered for any build wearing the
// set, so it does nothing where the talent or spell is missing. Left out: Soulcrusher's Rage 2 and
// Earthshatterer's Rage 2 (they need the Static Shock rune), both Earthshatterer 6-pieces (Undead
// only), and Earthshatterer's Storm 4 (the caster set has 2 pieces).

const (
	sodSetSoulcrushersRage     = 1951
	sodSetSoulcrushersStorm    = 1952
	sodSetEarthshatterersStorm = 1889

	setSoulcrushersStorm        = "Soulcrusher's Storm (Season of Discovery)"
	setEarthshatterersStorm     = "The Earthshatterer's Storm (Season of Discovery)"
	setSoulcrushersRage         = "Soulcrusher's Rage (Season of Discovery)"
	setSoulcrushersRageFireNova = "Soulcrusher's Rage with Fire Nova (Season of Discovery)"

	lavaBurstID = 1326047
)

// Lava Burst's ApplyEffects with its damage multiplied by bonus() for the cast; Lava Burst rolls
// its damage there and only lands it after travel.
func wrapLavaBurst(sham *shaman.Shaman, bonus func(sim *core.Simulation) float64) {
	sham.Env.RegisterPreFinalizeEffect(func() {
		lavaBurst := sham.GetSpell(core.ActionID{SpellID: lavaBurstID})
		if lavaBurst == nil {
			return
		}
		apply := lavaBurst.ApplyEffects
		lavaBurst.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			multiplier := bonus(sim)
			spell.DamageMultiplier *= multiplier
			apply(sim, target, spell)
			spell.DamageMultiplier /= multiplier
		}
	})
}

// The Maelstrom Weapon half of Soulcrusher's Rage 6.
func soulcrushersRage6(agent core.Agent) {
	sham := agent.(shaman.ShamanAgent).GetShaman()
	twoHand := sham.MainHand().HandType == proto.HandType_HandTypeTwoHand
	sham.Env.RegisterPreFinalizeEffect(func() {
		maelstrom := sham.GetAura("Maelstrom Weapon")
		if maelstrom == nil {
			return
		}
		// "Maelstrom Weapon can now stack up to 10 charges." Forever's per-charge cast time and cost
		// cut keeps scaling past 5 and both bottom out at zero: the engine floors cast time and cost.
		maelstrom.MaxStacks = 10
		// "When using a two-handed weapon, gain 2 charges at a time." The second charge comes as its
		// own action at the same instant: a stack change made inside OnStacksChange reaches the
		// handlers after this one out of order, and the APL's stack count would keep the first.
		if twoHand {
			adding := false
			maelstrom.ApplyOnStacksChange(func(aura *core.Aura, sim *core.Simulation, oldStacks, newStacks int32) {
				if adding || newStacks != oldStacks+1 {
					return
				}
				sim.AddPendingAction(core.NewDelayedAction(core.DelayedActionOptions{
					DoAt:     sim.CurrentTime,
					Priority: core.ActionPriorityAuto,
					OnAction: func(sim *core.Simulation) {
						if aura.IsActive() {
							adding = true
							aura.AddStack(sim)
							adding = false
						}
					},
				}))
			})
		}
		// "Casts with over 5 charges increase that spell's damage or healing dealt by 20% per excess
		// charge. Casts while at 10 charges will consume all charges and cast twice." The charges are
		// still there in ApplyEffects: Maelstrom Weapon spends them on cast completion, which comes after.
		for _, spell := range sham.Spellbook {
			if !spell.Matches(shaman.SpellMaskLightningBolt) || spell.ApplyEffects == nil {
				continue
			}
			apply := spell.ApplyEffects
			spell.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
				charges := maelstrom.GetStacks()
				if charges <= 5 {
					apply(sim, target, spell)
					return
				}
				multiplier := 1 + 0.2*float64(charges-5)
				spell.DamageMultiplier *= multiplier
				apply(sim, target, spell)
				if charges == 10 {
					apply(sim, target, spell)
				}
				spell.DamageMultiplier /= multiplier
			}
		}
	})
}

func registerSodSets() {
	core.NewItemSet(core.ItemSet{
		Name: setSoulcrushersStorm,
		Bonuses: map[int32]core.ApplySetBonus{
			// "When your Lava Burst, Chain Lightning, or Lightning Bolt strikes a target afflicted with
			// your Flame Shock Rank 5 or Rank 6", it also deals one Flame Shock damage pulse. Only the
			// casts count, not their overloads: Forever's overloads are separate proc spells, and the
			// bonus names the three spells.
			2: func(agent core.Agent, setBonusAura *core.Aura) {
				sham := agent.(shaman.ShamanAgent).GetShaman()
				setBonusAura.AttachProcTrigger(core.ProcTrigger{
					Name:               "Soulcrusher's Storm 2P",
					Callback:           core.CallbackOnSpellHitDealt,
					ClassSpellMask:     shaman.SpellMaskLightningBolt | shaman.SpellMaskChainLightning | shaman.SpellMaskLavaBurst,
					Outcome:            core.OutcomeLanded,
					TriggerImmediately: true,
					Handler: func(sim *core.Simulation, _ *core.Spell, result *core.SpellResult) {
						if dot := sham.FlameShock.RelatedDotSpell.Dot(result.Target); dot.IsActive() {
							dot.TickOnce(sim)
						}
					},
				})
			},
			// "Chance to trigger Overload increased by an additional 10%", Lava Burst damage increased
			// by your spell critical strike chance, and Lightning Bolt and Chain Lightning hits giving
			// the next Lava Burst within 15 sec +10%, up to 5 times. The two Lava Burst bonuses add.
			4: func(agent core.Agent, setBonusAura *core.Aura) {
				sham := agent.(shaman.ShamanAgent).GetShaman()
				setBonusAura.AttachAdditivePseudoStatBuff(&sham.BonusOverloadChance, 0.10)

				stacks := sham.RegisterAura(core.Aura{
					Label:     "Soulcrusher's Storm 4P",
					ActionID:  core.ActionID{SpellID: lavaBurstID, Tag: 4},
					Duration:  15 * time.Second,
					MaxStacks: 5,
				})
				setBonusAura.AttachProcTrigger(core.ProcTrigger{
					Name:               "Soulcrusher's Storm 4P Trigger",
					Callback:           core.CallbackOnSpellHitDealt,
					ClassSpellMask:     shaman.SpellMaskLightningBolt | shaman.SpellMaskChainLightning,
					RequireDamageDealt: true,
					TriggerImmediately: true,
					Handler: func(sim *core.Simulation, _ *core.Spell, _ *core.SpellResult) {
						stacks.Activate(sim)
						stacks.AddStack(sim)
					},
				})
				wrapLavaBurst(sham, func(sim *core.Simulation) float64 {
					bonus := sham.GetStat(stats.SpellCritPercent) / 100
					if stacks.IsActive() {
						bonus += 0.1 * float64(stacks.GetStacks())
						stacks.Deactivate(sim)
					}
					return 1 + bonus
				})
			},
			// "When your Chain Lightning damages fewer than 3 targets, it deals 35% increased damage for
			// each target less than 3." Overloads included: they are Chain Lightning damage too.
			6: func(agent core.Agent, setBonusAura *core.Aura) {
				targets := len(agent.GetCharacter().Env.Encounter.ActiveTargetUnits)
				setBonusAura.AttachSpellMod(core.SpellModConfig{
					Kind:       core.SpellMod_DamageDone_Pct,
					ClassMask:  shaman.SpellMaskChainLightning | shaman.SpellMaskChainLightningOverload,
					FloatValue: 0.35 * float64(3-min(3, targets)),
				})
			},
		},
	})

	core.NewItemSet(core.ItemSet{
		Name: setEarthshatterersStorm,
		Bonuses: map[int32]core.ApplySetBonus{
			// "Increases periodic damage done by your Flame Shock ability by 40%."
			2: func(_ core.Agent, setBonusAura *core.Aura) {
				setBonusAura.AttachSpellMod(core.SpellModConfig{
					Kind:       core.SpellMod_DamageDone_Pct,
					ClassMask:  shaman.SpellMaskFlameShockDot,
					FloatValue: 0.40,
				})
			},
		},
	})

	core.NewItemSet(core.ItemSet{
		Name: setSoulcrushersRage,
		Bonuses: map[int32]core.ApplySetBonus{
			6: func(agent core.Agent, _ *core.Aura) { soulcrushersRage6(agent) },
		},
	})

	// A sensitivity only. The 4-piece is worded for Fire Nova Totem: "Reduces the cooldown on your Fire
	// Nova Totem by 60%, increases its damage by 200%, and reduces its mana cost by 60%." Forever's
	// Fire Nova (408345) is a spell, so this reads the bonus onto it. It only matters to a rotation
	// that casts Fire Nova: the Enhancement preset's does.
	core.NewItemSet(core.ItemSet{
		Name: setSoulcrushersRageFireNova,
		Bonuses: map[int32]core.ApplySetBonus{
			4: func(_ core.Agent, setBonusAura *core.Aura) {
				setBonusAura.AttachSpellMod(core.SpellModConfig{Kind: core.SpellMod_Cooldown_Multiplier, ClassMask: shaman.SpellMaskFireNova, FloatValue: 0.4})
				setBonusAura.AttachSpellMod(core.SpellModConfig{Kind: core.SpellMod_DamageDone_Pct, ClassMask: shaman.SpellMaskFireNova, FloatValue: 2})
				setBonusAura.AttachSpellMod(core.SpellModConfig{Kind: core.SpellMod_PowerCost_Pct_Add, ClassMask: shaman.SpellMaskFireNova, FloatValue: -0.6})
			},
			6: func(agent core.Agent, _ *core.Aura) { soulcrushersRage6(agent) },
		},
	})
}
