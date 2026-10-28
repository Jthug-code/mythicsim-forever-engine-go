package spellhance

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"time"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/stats"
	"github.com/wowsims/forever/sim/shaman"
)

// Season of Discovery Phase 8 best-in-slot gear (testdata/sod-p8-shaman-bis.json, pulled from
// Wowhead's SoD guides on 2026-10-27). None of it is in the Forever item database, so each piece
// becomes a custom item under its SoD id plus customItemOffset, which no Forever item uses.
// Wowhead gives crit, hit and haste as percentages; they become Forever ratings at the engine's
// 14 crit, 10 hit and 10 haste rating per percent. "Your attacks are dodged or parried N% less
// often" becomes expertise, 10 rating per percent, which Forever spends on dodge and parry alike.
// Set bonuses are left out: they need SoD runes or Undead targets.

const (
	customItemOffset = 9_000_000
	// Scarlet Smashbringer without its "GCD of your Shock, Lightning and Lava Burst spells -0.5 sec",
	// for the runs that leave that SoD effect out.
	smashbringerNoGCDOffset = 9_500_000
)

const (
	sodNaturalAlignmentCrystal  = 230273
	sodInfusionOfSouls          = 241039
	sodMirage                   = 241003
	sodTotemOfThunder           = 228176
	sodScarletSmashbringer      = 240921
	sodAbandonedExperiment      = 241037
	sodSignetOfTheEarthshatter  = 236176
	sodTotemOfPyroclasticThunde = 234480
)

type sodItem struct {
	ID      int32              `json:"id"`
	Name    string             `json:"name"`
	Type    string             `json:"type"`
	Speed   float64            `json:"speed"`
	Dmg     []float64          `json:"dmg"`
	Stats   map[string]float64 `json:"stats"`
	Effects []string           `json:"effects"`
}

type sodFile struct {
	Sets map[string]map[string]sodItem `json:"sets"`
}

// A gear set built from the file: the items as Forever custom items, and the haste they carry,
// so a run can take it back out.
type gearSet struct {
	Label      string
	Items      []*proto.ItemSpec
	SpellHaste float64 // rating
	MeleeHaste float64 // rating
	Totals     stats.Stats
}

var slotTypes = map[string]proto.ItemType{
	"Head": proto.ItemType_ItemTypeHead, "Neck": proto.ItemType_ItemTypeNeck, "Shoulders": proto.ItemType_ItemTypeShoulder,
	"Back": proto.ItemType_ItemTypeBack, "Chest": proto.ItemType_ItemTypeChest, "Wrists": proto.ItemType_ItemTypeWrist,
	"Hands": proto.ItemType_ItemTypeHands, "Waist": proto.ItemType_ItemTypeWaist, "Legs": proto.ItemType_ItemTypeLegs,
	"Feet": proto.ItemType_ItemTypeFeet, "Ring1": proto.ItemType_ItemTypeFinger, "Ring2": proto.ItemType_ItemTypeFinger,
	"Trinket1": proto.ItemType_ItemTypeTrinket, "Trinket2": proto.ItemType_ItemTypeTrinket,
	"TwoHand": proto.ItemType_ItemTypeWeapon, "Relic": proto.ItemType_ItemTypeRanged,
}

var slotOrder = []string{"Head", "Neck", "Shoulders", "Back", "Chest", "Wrists", "Hands", "Waist", "Legs", "Feet",
	"Ring1", "Ring2", "Trinket1", "Trinket2", "TwoHand", "Relic"}

var lessDodgedOrParried = regexp.MustCompile(`dodged or parried (\d+(?:\.\d+)?)% less often`)

func sodStats(item sodItem) stats.Stats {
	var s stats.Stats
	for key, v := range item.Stats {
		switch key {
		case "str":
			s[stats.Strength] += v
		case "agi":
			s[stats.Agility] += v
		case "sta":
			s[stats.Stamina] += v
		case "int":
			s[stats.Intellect] += v
		case "spldmg":
			s[stats.SpellDamage] += v
			s[stats.HealingPower] += v
		case "mleatkpwr":
			s[stats.AttackPower] += v
		case "splcritstrkpct":
			s[stats.SpellCritRating] += v * core.SpellCritRatingPerCritPercent
		case "mlecritstrkpct":
			s[stats.MeleeCritRating] += v * core.PhysicalCritRatingPerCritPercent
		case "splhitpct":
			s[stats.SpellHitRating] += v * core.SpellHitRatingPerHitPercent
		case "mlehitpct":
			s[stats.MeleeHitRating] += v * core.PhysicalHitRatingPerHitPercent
		case "splhastepct":
			s[stats.SpellHasteRating] += v * core.SpellHasteRatingPerHastePercent
		case "mlehastepct":
			s[stats.MeleeHasteRating] += v * core.PhysicalHasteRatingPerHastePercent
		default:
			panic(fmt.Sprintf("%s: stat %q has no Forever mapping", item.Name, key))
		}
	}
	for _, effect := range item.Effects {
		if m := lessDodgedOrParried.FindStringSubmatch(effect); m != nil {
			pct, _ := strconv.ParseFloat(m[1], 64)
			s[stats.ExpertiseRating] += pct * core.ExpertiseRatingPerExpertisePercent
		}
	}
	return s
}

func simItem(id int32, slot string, item sodItem) *proto.SimItem {
	si := &proto.SimItem{
		Id:   id,
		Name: item.Name,
		Type: slotTypes[slot],
		ScalingOptions: map[int32]*proto.ScalingItemProperties{
			0: {Stats: sodStats(item).ToProtoMap(), Ilvl: 60},
		},
	}
	switch slot {
	case "TwoHand":
		si.HandType = proto.HandType_HandTypeTwoHand
		si.WeaponType = proto.WeaponType_WeaponTypeStaff
		if item.Type == "Two-Handed Mace" {
			si.WeaponType = proto.WeaponType_WeaponTypeMace
		}
		si.WeaponSpeed = item.Speed
		si.ScalingOptions[0].WeaponDamageMin = item.Dmg[0]
		si.ScalingOptions[0].WeaponDamageMax = item.Dmg[1]
	case "Relic":
		si.RangedWeaponType = proto.RangedWeaponType_RangedWeaponTypeTotem
	default:
		si.ArmorType = proto.ArmorType_ArmorTypeMail
	}
	return si
}

// loadSodGear adds the file's two sets to the item database and returns them by name:
// "caster_mirage_2h" and "enh_smashbringer_2h", plus "enh_smashbringer_2h_nogcd".
func loadSodGear(path string) map[string]*gearSet {
	raw, err := os.ReadFile(path)
	if err != nil {
		panic(err)
	}
	var file sodFile
	if err := json.Unmarshal(raw, &file); err != nil {
		panic(err)
	}

	db := &proto.SimDatabase{}
	sets := map[string]*gearSet{}
	for name, items := range file.Sets {
		set := &gearSet{Label: name}
		for _, slot := range slotOrder {
			item, ok := items[slot]
			if !ok {
				panic(fmt.Sprintf("%s has no %s", name, slot))
			}
			si := simItem(item.ID+customItemOffset, slot, item)
			db.Items = append(db.Items, si)
			set.Items = append(set.Items, &proto.ItemSpec{Id: si.Id})
			s := sodStats(item)
			set.Totals = set.Totals.Add(s)
			set.SpellHaste += s[stats.SpellHasteRating]
			set.MeleeHaste += s[stats.MeleeHasteRating]
		}
		sets[name] = set
	}

	// The same Enhancement set with a Smashbringer that has no GCD effect.
	enh := sets["enh_smashbringer_2h"]
	noGCD := &gearSet{Label: enh.Label + "_nogcd", SpellHaste: enh.SpellHaste, MeleeHaste: enh.MeleeHaste, Totals: enh.Totals}
	for _, spec := range enh.Items {
		id := spec.Id
		if id == sodScarletSmashbringer+customItemOffset {
			weapon := simItem(sodScarletSmashbringer+smashbringerNoGCDOffset, "TwoHand", file.Sets["enh_smashbringer_2h"]["TwoHand"])
			db.Items = append(db.Items, weapon)
			id = weapon.Id
		}
		noGCD.Items = append(noGCD.Items, &proto.ItemSpec{Id: id})
	}
	sets[noGCD.Label] = noGCD

	core.AddToDatabase(db)
	registerSodEffects()
	return sets
}

// sodEffectItems are the custom items registerSodEffects gives an effect.
var sodEffectItems = []int32{
	sodNaturalAlignmentCrystal + customItemOffset, sodTotemOfThunder + customItemOffset,
	sodTotemOfPyroclasticThunde + customItemOffset, sodScarletSmashbringer + customItemOffset,
	sodScarletSmashbringer + smashbringerNoGCDOffset, sodAbandonedExperiment + customItemOffset,
	sodSignetOfTheEarthshatter + customItemOffset,
}

// registerSodEffects runs once the items are in the database: core.NewItemEffect skips an item
// it can't find there.
func registerSodEffects() {
	// Natural Alignment Crystal. Use: +20% spell damage and healing, +20% spell mana cost for 20 sec.
	// 2 min cooldown.
	core.NewItemEffect(sodNaturalAlignmentCrystal+customItemOffset, func(agent core.Agent) {
		character := agent.GetCharacter()
		actionID := core.ActionID{ItemID: sodNaturalAlignmentCrystal}
		magic := []stats.SchoolIndex{stats.SchoolIndexArcane, stats.SchoolIndexFire, stats.SchoolIndexFrost,
			stats.SchoolIndexHoly, stats.SchoolIndexNature, stats.SchoolIndexShadow}
		aura := character.RegisterAura(core.Aura{
			Label:    "Natural Alignment Crystal",
			ActionID: actionID,
			Duration: 20 * time.Second,
			OnGain: func(_ *core.Aura, _ *core.Simulation) {
				for _, school := range magic {
					character.PseudoStats.SchoolDamageDealtMultiplier[school] *= 1.2
				}
				character.PseudoStats.SpellCostPercentModifier += 20
			},
			OnExpire: func(_ *core.Aura, _ *core.Simulation) {
				for _, school := range magic {
					character.PseudoStats.SchoolDamageDealtMultiplier[school] /= 1.2
				}
				character.PseudoStats.SpellCostPercentModifier -= 20
			},
		})
		spell := character.RegisterSpell(core.SpellConfig{
			ActionID: actionID,
			Flags:    core.SpellFlagNoOnCastComplete,
			Cast: core.CastConfig{
				CD: core.Cooldown{Timer: character.NewTimer(), Duration: 2 * time.Minute},
			},
			ApplyEffects: func(sim *core.Simulation, _ *core.Unit, _ *core.Spell) {
				aura.Activate(sim)
			},
			RelatedSelfBuff: aura,
		})
		character.AddMajorCooldown(core.MajorCooldown{Spell: spell, Type: core.CooldownTypeDPS})
	})

	// Infusion of Souls. "The GCD of your non-weapon damaging spells can be reduced by spell haste,
	// up to 0.5 sec": Forever already hastes every spell GCD down to 1 s (core/cast.go), so the
	// effect adds nothing here.

	// Totem of Thunder. Equip: Lightning Bolt cast time -0.1 sec.
	core.NewItemEffect(sodTotemOfThunder+customItemOffset, func(agent core.Agent) {
		agent.GetCharacter().AddStaticMod(core.SpellModConfig{
			Kind:      core.SpellMod_CastTime_Flat,
			ClassMask: shaman.SpellMaskLightningBolt,
			TimeValue: -100 * time.Millisecond,
		})
	})

	// Totem of Pyroclastic Thunder. Equip: Lightning Bolt, Chain Lightning and Lava Burst damage +3%.
	core.NewItemEffect(sodTotemOfPyroclasticThunde+customItemOffset, func(agent core.Agent) {
		agent.GetCharacter().AddStaticMod(core.SpellModConfig{
			Kind: core.SpellMod_DamageDone_Pct,
			ClassMask: shaman.SpellMaskLightningBolt | shaman.SpellMaskLightningBoltOverload | shaman.SpellMaskChainLightning |
				shaman.SpellMaskChainLightningOverload | shaman.SpellMaskLavaBurst,
			FloatValue: 0.03,
		})
	})

	// Scarlet Smashbringer. Chance on hit: +200 Strength for 15 sec. Wowhead gives no rate; this
	// assumes 1 proc per minute.
	smashbringerProc := func(character *core.Character) {
		buff := character.NewTemporaryStatsAura("Scarlet Smashbringer", core.ActionID{ItemID: sodScarletSmashbringer},
			stats.Stats{stats.Strength: 200}, 15*time.Second)
		character.MakeProcTriggerAura(core.ProcTrigger{
			Name:     "Scarlet Smashbringer Trigger",
			Callback: core.CallbackOnSpellHitDealt,
			ProcMask: core.ProcMaskMelee,
			Outcome:  core.OutcomeLanded,
			DPM:      character.NewLegacyPPMManager(1, core.ProcMaskMelee),
			Handler: func(sim *core.Simulation, _ *core.Spell, _ *core.SpellResult) {
				buff.Activate(sim)
			},
		})
	}
	// Equip: GCD of your Shock, Lightning and Lava Burst spells -0.5 sec.
	core.NewItemEffect(sodScarletSmashbringer+customItemOffset, func(agent core.Agent) {
		character := agent.GetCharacter()
		smashbringerProc(character)
		character.AddStaticMod(core.SpellModConfig{
			Kind: core.SpellMod_GlobalCooldown_Flat,
			ClassMask: shaman.SpellMaskShock | shaman.SpellMaskLightningBolt | shaman.SpellMaskChainLightning |
				shaman.SpellMaskLavaBurst,
			TimeValue: -500 * time.Millisecond,
		})
	})
	core.NewItemEffect(sodScarletSmashbringer+smashbringerNoGCDOffset, func(agent core.Agent) {
		smashbringerProc(agent.GetCharacter())
	})

	// Abandoned Experiment. Use: for 30 sec, each melee or ranged attack adds 2% attack speed, up to 15
	// stacks. 2 min cooldown.
	core.NewItemEffect(sodAbandonedExperiment+customItemOffset, func(agent core.Agent) {
		character := agent.GetCharacter()
		actionID := core.ActionID{ItemID: sodAbandonedExperiment}
		aura := character.RegisterAura(core.Aura{
			Label:     "Abandoned Experiment",
			ActionID:  actionID,
			Duration:  30 * time.Second,
			MaxStacks: 15,
			OnStacksChange: func(_ *core.Aura, sim *core.Simulation, oldStacks, newStacks int32) {
				character.MultiplyAttackSpeed(sim, (1+0.02*float64(newStacks))/(1+0.02*float64(oldStacks)))
			},
			OnSpellHitDealt: func(aura *core.Aura, sim *core.Simulation, spell *core.Spell, _ *core.SpellResult) {
				if spell.ProcMask.Matches(core.ProcMaskMeleeOrRanged) {
					aura.AddStack(sim)
				}
			},
		})
		spell := character.RegisterSpell(core.SpellConfig{
			ActionID: actionID,
			Flags:    core.SpellFlagNoOnCastComplete,
			Cast: core.CastConfig{
				CD: core.Cooldown{Timer: character.NewTimer(), Duration: 2 * time.Minute},
			},
			ApplyEffects: func(sim *core.Simulation, _ *core.Unit, _ *core.Spell) {
				aura.Activate(sim)
			},
			RelatedSelfBuff: aura,
		})
		character.AddMajorCooldown(core.MajorCooldown{Spell: spell, Type: core.CooldownTypeDPS})
	})

	// Signet of the Earthshatterer. Equip: 2% chance on weapon damage to get an extra attack (200 ms
	// cooldown).
	core.NewItemEffect(sodSignetOfTheEarthshatter+customItemOffset, func(agent core.Agent) {
		character := agent.GetCharacter()
		character.MakeProcTriggerAura(core.ProcTrigger{
			Name:       "Signet of the Earthshatterer",
			ActionID:   core.ActionID{ItemID: sodSignetOfTheEarthshatter},
			Callback:   core.CallbackOnSpellHitDealt,
			ProcMask:   core.ProcMaskMelee,
			Outcome:    core.OutcomeLanded,
			ProcChance: 0.02,
			ICD:        200 * time.Millisecond,
			Handler: func(sim *core.Simulation, _ *core.Spell, _ *core.SpellResult) {
				character.AutoAttacks.ExtraMHAttack(sim)
			},
		})
	})

	// Mirage, Rod of Illusion: its Mirage proc and swap are left out, as the gear file says.
}
