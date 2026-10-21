package hunter

import (
	"testing"
	"time"

	"github.com/wowsims/forever/sim/core"
)

// Patch 190's switch: with RangedMinimumFromTargetCenter off the hunter shoots from 10 yd again, the client's 8 yd
// read from the target's center.
func TestRangedDeadzoneSwitchOffShootsFromTenYards(t *testing.T) {
	core.RangedMinimumFromTargetCenter = false
	defer func() { core.RangedMinimumFromTargetCenter = true }()
	sim, hunter := rangedMinimumSim(t, 10)
	for i := 0; i < 100000 && sim.CurrentTime < 5*time.Second; i++ {
		sim.Step()
	}
	if hunter.AutoAttacks.PreviousRangedAttack() <= 0 {
		t.Errorf("standing at 10 yd with the switch off: no Auto Shot in 5s")
	}
}
