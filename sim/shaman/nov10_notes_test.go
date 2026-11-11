package shaman

import (
	"testing"
	"time"

	"github.com/wowsims/forever/sim/core/buffs"
)

// The 10 November 2026 hotfix as nov10_notes.go applies it, and the 2 November state back when it is undone.
func TestNov10Hotfix(t *testing.T) {
	t.Cleanup(func() { SetNov10Hotfix(false) })
	check := func(on bool, stormstrike float64, cooldown time.Duration, dexterity, maelstromPPM float64) {
		t.Helper()
		SetNov10Hotfix(on)
		if buffs.WindfuryTotemAutosOnly != on {
			t.Errorf("hotfix %v: Windfury Totem's autos-only rule reads %v", on, buffs.WindfuryTotemAutosOnly)
		}
		if nov10Hotfix != on {
			t.Errorf("hotfix %v: the switch reads %v", on, nov10Hotfix)
		}
		if nov2StormstrikeWeaponDamage != stormstrike {
			t.Errorf("hotfix %v: Stormstrike deals %v weapon damage, want %v", on, nov2StormstrikeWeaponDamage, stormstrike)
		}
		if cd := spellData.Stormstrike.Highest().Cooldown(); cd != cooldown {
			t.Errorf("hotfix %v: Stormstrike's cooldown is %s, want %s", on, cd, cooldown)
		}
		if got := spellData.MentalDexterity.EffectAt(1).FractionAt(3); got != dexterity {
			t.Errorf("hotfix %v: Mental Dexterity 3/3 is %v, want %v", on, got, dexterity)
		}
		// Under the hotfix the rate comes from the rank (applyMaelstromWeapon); this is the 2 November doubling undone.
		if nov2MaelstromWeaponPPM != maelstromPPM {
			t.Errorf("hotfix %v: Maelstrom Weapon's base rate is %v a minute, want %v", on, nov2MaelstromWeaponPPM, maelstromPPM)
		}
	}
	check(true, 1, 8*time.Second, 1, 10)
	check(false, 1.5, 6*time.Second, 1.5, 20)
}
