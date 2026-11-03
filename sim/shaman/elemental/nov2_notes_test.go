package elemental

import (
	"testing"

	"github.com/wowsims/forever/sim/core"
)

// "Lava Burst: Removed. Replaced with Call of Thunder, which increases critical strike chance of Lightning
// Bolt and Chain Lightning by 5%" (2 November 2026 notes). DefaultTalents takes the node; the same build
// without it is the control.
func TestCallOfThunderReplacesLavaBurst(t *testing.T) {
	withoutNode := "5505331000000351--553302001"
	_, with := newElementalShaman(t, DefaultTalents)
	_, without := newElementalShaman(t, withoutNode)

	for _, id := range []int32{408490, 1238299, 1238300, 1326047} {
		if with.GetSpell(core.ActionID{SpellID: id}) != nil {
			t.Errorf("Lava Burst %d is still registered", id)
		}
	}
	for _, id := range []int32{15208, 10605} {
		a, b := with.GetSpell(core.ActionID{SpellID: id}), without.GetSpell(core.ActionID{SpellID: id})
		if got := a.BonusCritPercent - b.BonusCritPercent; got != 5 {
			t.Errorf("%d: Call of Thunder adds %v%% crit, want 5", id, got)
		}
	}
	for rank, overload := range with.LightningBoltOverloads {
		if got := overload.BonusCritPercent - without.LightningBoltOverloads[rank].BonusCritPercent; got != 5 {
			t.Errorf("Lightning Bolt rank %d overload: Call of Thunder adds %v%% crit, want 5", rank, got)
		}
	}
}
