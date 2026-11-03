package shaman

import (
	"math"
	"testing"
	"time"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/dbcenums"
	"github.com/wowsims/forever/sim/core/spelldata"
)

// The 2 November 2026 notes as nov2_notes.go applies them (client 70425/70707 numbers in brackets). When a
// client build carries the changes, this names each number it states differently.
func TestNov2Notes(t *testing.T) {
	checkNov2Notes(t)
}

// Each change undone puts its client numbers back, and redone puts the notes' back exactly.
func TestNov2ChangesUndo(t *testing.T) {
	t.Cleanup(func() {
		for _, name := range Nov2Changes() {
			SetNov2Change(name, true)
		}
	})
	at60 := func(id int32) float64 { return spelldata.MustFind(id).DamageEffect().Average(core.CharacterLevel) }
	client := map[string]func() (string, float64, float64){
		"Lightning Bolt damage": func() (string, float64, float64) { return "Lightning Bolt rank 10 at 60", at60(15208), 200 },
		"Chain Lightning rework": func() (string, float64, float64) {
			return "Chain Lightning rank 4 cost", spelldata.MustFind(10605).Cost(), 315
		},
		"Shocks at Classic Era damage": func() (string, float64, float64) { return "Flame Shock rank 6 hit", at60(29228), 166 },
		"Lightning Overload 10/20/30%": func() (string, float64, float64) {
			return "Lightning Overload 3/3", spellData.LightningOverload.FractionAt(3), 0.10
		},
		"Call of Thunder": func() (string, float64, float64) {
			return "Call of Thunder on", core.Ternary(nov2CallOfThunder, 1.0, 0.0), 0
		},
		"Maelstrom Weapon procs doubled": func() (string, float64, float64) { return "Maelstrom Weapon PPM", nov2MaelstromWeaponPPM, 10 },
		"Stormstrike 150% weapon damage": func() (string, float64, float64) { return "Stormstrike weapon damage", nov2StormstrikeWeaponDamage, 1 },
		"Stormstrike 6 s cooldown": func() (string, float64, float64) {
			return "Stormstrike cooldown", spellData.Stormstrike.Highest().Cooldown().Seconds(), 8
		},
		"Mental Dexterity 50/100/150%": func() (string, float64, float64) {
			return "Mental Dexterity 3/3", spellData.MentalDexterity.EffectAt(1).FractionAt(3), 1
		},
		"Flurry 6-30%": func() (string, float64, float64) { return "Flurry 5/5", spellData.Flurry.MultiplierAt(5), 1.25 },
	}
	if len(client) != len(Nov2Changes()) {
		t.Fatalf("%d changes, %d checked", len(Nov2Changes()), len(client))
	}
	for _, name := range Nov2Changes() {
		SetNov2Change(name, false)
		if what, got, want := client[name](); math.Abs(got-want) > 1e-6 {
			t.Errorf("%s off: %s = %v, want the client's %v", name, what, got, want)
		}
		SetNov2Change(name, true)
	}
	checkNov2Notes(t)
}

func checkNov2Notes(t *testing.T) {
	t.Helper()
	near := func(what string, got, want float64) {
		t.Helper()
		if math.Abs(got-want) > 1e-6 {
			t.Errorf("%s = %v, want %v", what, got, want)
		}
	}
	at60 := func(id int32) float64 { return spelldata.MustFind(id).DamageEffect().Average(core.CharacterLevel) }

	// Lightning Bolt: max rank +80% (196 + 1.2 a level, 200 at 60); ranks 4-9 at 78% of Era; ranks 1-3 as they were.
	near("Lightning Bolt rank 10 at 60", at60(15208), 361)
	near("Lightning Bolt rank 9 base", spelldata.MustFind(15207).DamageEffect().BasePoints, 0.78*368)
	near("Lightning Bolt rank 4 base", spelldata.MustFind(915).DamageEffect().BasePoints, 0.78*89)
	near("Lightning Bolt rank 1 base", spelldata.MustFind(403).DamageEffect().BasePoints, 15)
	near("Lightning Bolt rank 10 overload at 60", lightningBoltOverloadRanks.Rank(10).DamageEffect().Average(core.CharacterLevel), 180) // [100]

	// Chain Lightning rank 4: 145% of rank 10's bolt [126 at 60], 225% of its 220 mana [315], 2.5 sec [2].
	cl := spelldata.MustFind(10605)
	near("Chain Lightning rank 4 at 60", at60(10605), 524)
	near("Chain Lightning rank 4 cost", cl.Cost(), 495)
	if cl.CastTime() != 2500*time.Millisecond {
		t.Errorf("Chain Lightning rank 4 casts in %s, want 2.5s", cl.CastTime())
	}

	// Shocks at their Era values [Earth Shock 293-309, Flame Shock 166 + 44 a tick, Frost Shock 275-291].
	es := spelldata.MustFind(10414).DamageEffect()
	near("Earth Shock rank 7 low", es.Min(60), 517)
	near("Earth Shock rank 7 high", es.Max(60), 545)
	near("Flame Shock rank 6 hit", at60(29228), 292)
	near("Flame Shock rank 6 tick", spelldata.MustFind(29228).PeriodicEffect().Average(core.CharacterLevel), 80)
	frost := spelldata.MustFind(10473).DamageEffect()
	near("Frost Shock rank 4 low", frost.Min(58), 486)
	near("Frost Shock rank 4 high", frost.Max(58), 514)

	// Talents [3/7/10%, 33/67/100%, 25%, 3%], and Stormstrike's cooldown [8 sec].
	near("Lightning Overload 3/3", spellData.LightningOverload.FractionAt(3), 0.30)
	near("Mental Dexterity 3/3", spellData.MentalDexterity.EffectAt(1).FractionAt(3), 1.5)
	near("Flurry 5/5", spellData.Flurry.MultiplierAt(5), 1.30)
	near("Call of Thunder", spellData.CallOfThunder.Highest().Effect(dbcenums.A_ADD_FLAT_MODIFIER, int32(dbcenums.SPELLMOD_CRITICAL_CHANCE)).BasePoints, 5)
	if cd := spellData.Stormstrike.Highest().Cooldown(); cd != 6*time.Second {
		t.Errorf("Stormstrike's cooldown is %s, want 6s", cd)
	}
}
