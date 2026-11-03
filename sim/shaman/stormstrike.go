package shaman

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/dbcenums"
)

var stormstrikeRank = spellData.Stormstrike.Highest()
var StormstrikeActionID = core.ActionID{SpellID: stormstrikeRank.ID}

// StormstrikeWeaponPercent replaces the strike's weapon damage percent, for the Spellhance balance candidates.
// Zero leaves it as nov2_notes.go has it (150%).
var StormstrikeWeaponPercent float64

// StormstrikeSpellBonusPercent replaces the debuff's bonus to the next bolt or Earth Shock, for the same
// candidates. Zero leaves the client's 20%.
var StormstrikeSpellBonusPercent float64

func (shaman *Shaman) StormstrikeDebuffAura(target *core.Unit) *core.Aura {
	aura := target.GetOrRegisterAura(core.Aura{
		Label:     "Stormstrike-" + shaman.Label,
		ActionID:  StormstrikeActionID,
		Duration:  stormstrikeRank.Duration(),
		MaxStacks: int32(stormstrikeRank.ProcCharges),
		OnSpellHitTaken: func(aura *core.Aura, sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
			if spell.Unit != &shaman.Unit || !spell.Matches(stormstrikeConsumers) {
				return
			}
			if !result.Landed() || result.Damage == 0 {
				return
			}
			aura.RemoveStack(sim)
		},
	})
	// Client 17364 (aura 271): only this shaman's Lightning Bolt, Chain Lightning and Earth Shock take the bonus;
	// only the bolts spend it (stormstrikeConsumers).
	multiplier := 1 + stormstrikeRank.Effect(dbcenums.A_MOD_SPELL_DAMAGE_FROM_CASTER, 0).Percent()
	if StormstrikeSpellBonusPercent != 0 {
		multiplier = 1 + StormstrikeSpellBonusPercent/100
	}
	return aura.AttachDDBC(0, 1, &shaman.AttackTables, func(_ *core.Simulation, spell *core.Spell, _ *core.AttackTable) float64 {
		return core.Ternary(spell.Matches(stormstrikeSpells), multiplier, 1)
	})
}

const stormstrikeSpells = SpellMaskLightningBolt | SpellMaskChainLightning | SpellMaskEarthShock | SpellMaskOverload

// stormstrikeConsumers spend the charge (MythicSim patch 140). Earth Shock takes the bonus, as client row
// 17364's mask says, but does not spend it: Redfall relayed in-game beta testing on the MythicSim Discord
// (16 October 2026) that Earth Shock is not consuming Stormstrike.
const stormstrikeConsumers = SpellMaskLightningBolt | SpellMaskChainLightning | SpellMaskOverload

func (shaman *Shaman) newStormstrikeHitSpellConfig(spellID int32, isMH bool) core.SpellConfig {
	var procMask core.ProcMask
	var actionTag int32

	procMask = core.Ternary(isMH, core.ProcMaskMeleeMHSpecial, core.ProcMaskMeleeOHSpecial)
	actionTag = core.TernaryInt32(isMH, 1, 2)

	return core.SpellConfig{
		ActionID:         core.ActionID{SpellID: spellID}.WithTag(actionTag),
		SpellSchool:      core.SpellSchoolPhysical,
		DefenseType:      core.DefenseTypeMelee,
		ProcMask:         procMask,
		Flags:            core.SpellFlagMeleeMetrics,
		ClassSpellMask:   SpellMaskStormstrikeDamage,
		ThreatMultiplier: 1,
		DamageMultiplier: 1,
		// Client 17364's strike is E_NORMALIZED_WEAPON_DMG (Era's was effect 58, plain weapon damage), and the 2
		// November notes say "150% normalized weapon damage": attack power counts at 3.3 sec for a two-hander,
		// not at the weapon's own speed.
		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			weaponDamage := core.Ternary(isMH, spell.Unit.MHNormalizedWeaponDamage, spell.Unit.OHNormalizedWeaponDamage)
			baseDamage := weaponDamage(sim, spell.MeleeAttackPower(target)) * core.Ternary(StormstrikeWeaponPercent != 0, StormstrikeWeaponPercent/100, nov2StormstrikeWeaponDamage)
			spell.CalcAndDealDamage(sim, target, baseDamage, spell.OutcomeMeleeSpecialBlockAndCrit)
		},
	}
}

func (shaman *Shaman) newStormstrikeHitSpell(isMH bool) *core.Spell {
	return shaman.RegisterSpell(shaman.newStormstrikeHitSpellConfig(stormstrikeRank.ID, isMH))
}

func (shaman *Shaman) newStormstrikeSpellConfig(spellID int32, ssDebuffAuras *core.AuraArray, mhHit *core.Spell, ohHit *core.Spell) core.SpellConfig {
	stormstrikeSpellConfig := core.SpellConfig{
		ActionID:       core.ActionID{SpellID: spellID},
		SpellSchool:    core.SpellSchoolPhysical,
		DefenseType:    core.DefenseTypeMelee,
		ProcMask:       core.ProcMaskEmpty,
		Flags:          core.SpellFlagMeleeMetrics | core.SpellFlagAPL,
		ClassSpellMask: SpellMaskStormstrikeCast,
		// 8% of base mana, not the client's flat 125 (16 October 2026 beta notes).
		ManaCost: core.ManaCostOptions{
			BaseCostPercent: 8,
		},
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: core.GCDDefault,
			},
			IgnoreHaste: true,
			CD: core.Cooldown{
				Timer:    shaman.NewTimer(),
				Duration: max(stormstrikeRank.Cooldown(), stormstrikeRank.CategoryCooldown()),
			},
		},

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			shaman.StormstrikeCastResult = spell.CalcAndDealOutcome(sim, target, spell.OutcomeMeleeSpecialHitNoHitCounter)
			if shaman.StormstrikeCastResult.Landed() {
				ssDebuffAura := ssDebuffAuras.Get(target)
				ssDebuffAura.Activate(sim)
				ssDebuffAura.SetStacks(sim, ssDebuffAura.MaxStacks)

				if shaman.HasMHWeapon() {
					mhHit.Cast(sim, target)
				}

			}
			spell.DisposeResult(shaman.StormstrikeCastResult)
		},
		ExtraCastCondition: func(sim *core.Simulation, target *core.Unit) bool {
			return (shaman.HasMHWeapon() || shaman.HasOHWeapon())
		},
	}
	return stormstrikeSpellConfig
}

func (shaman *Shaman) registerStormstrikeSpell() {
	mhHit := shaman.newStormstrikeHitSpell(true)
	ohHit := shaman.newStormstrikeHitSpell(false)

	shaman.StormStrikeDebuffAuras = shaman.NewEnemyAuraArray(shaman.StormstrikeDebuffAura)

	shaman.Stormstrike = shaman.RegisterSpell(shaman.newStormstrikeSpellConfig(stormstrikeRank.ID, &shaman.StormStrikeDebuffAuras, mhHit, ohHit))
}
