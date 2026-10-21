package hunter

import (
	"testing"
	"time"

	"github.com/wowsims/forever/sim/core"
)

// Patch 192: a spell's range check reads where a walking hunter is now. The hunter walks in from 12
// yards at 7 a second and its rotation checks every 500 ms, so nothing brings the position up to
// date during the first half second of the walk, as on a GCD. At 0.25 s it is 10.25 yards out, inside the 10.5 yard deadzone: a cast that
// fires then outside the rotation (a spell queued at the end of the GCD) cannot be an Arcane Shot.
// It could from the 12 yards the walk started at.
func TestCastRangeReadsThePositionMidWalk(t *testing.T) {
	for _, reads := range []bool{true, false} {
		core.CastRangeReadsCurrentPosition = reads
		sim, hunter := rangedMinimumSim(t, 12)
		hunter.ReactionTime = 500 * time.Millisecond
		hunter.MoveTo(5, sim)
		var castable bool
		var at float64
		check := &core.PendingAction{NextActionAt: 250 * time.Millisecond, OnAction: func(sim *core.Simulation) {
			castable = hunter.ArcaneShot.CanCast(sim, hunter.CurrentTarget)
			at = hunter.DistanceFromTarget
		}}
		sim.AddPendingAction(check)
		for i := 0; i < 100000 && sim.CurrentTime < 300*time.Millisecond; i++ {
			sim.Step()
		}
		if at != 12 {
			t.Fatalf("the position was read during the walk (%.2f yards)", at)
		}
		if castable != !reads {
			t.Errorf("reads=%v: Arcane Shot castable %v at 10.25 yards (last read %.2f), want %v", reads, castable, at, !reads)
		}
	}
	core.CastRangeReadsCurrentPosition = true
}
