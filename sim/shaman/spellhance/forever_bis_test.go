package spellhance

import (
	"os"
	"testing"
	"time"
)

// TestForeverBisMatrix is run 7's high-end comparison on Forever's own items: the Elemental preset and the
// 29/22 weave on the Elemental BiS set, the Enhancement builds on the Enhancement one, both searched from
// Wowhead's Forever item data, at 2, 5 and 10 minutes as the default Windshaper Skyborne. It writes
// results.csv and sources.csv to $SPELLHANCE_BIS_OUT; SPELLHANCE_BUFFS=full picks the full raid buffs and
// the sets searched under them.
func TestForeverBisMatrix(t *testing.T) {
	out := os.Getenv("SPELLHANCE_BIS_OUT")
	if out == "" {
		t.Skip("set SPELLHANCE_BIS_OUT to run the comparison")
	}
	iterations := matrixIterations(t)
	var runs []run
	for _, row := range []struct {
		b    build
		gear string
	}{
		{elePreset, foreverEleBisGear()}, {weave2922, foreverEleBisGear()},
		{enhPreset, foreverBisGear()}, {melee8367, foreverBisGear()}, {melee83211WS, foreverBisGear()},
	} {
		for _, d := range []time.Duration{2 * time.Minute, 5 * time.Minute, 10 * time.Minute} {
			runs = append(runs, run{Build: row.b, Gear: row.gear, Duration: d, ReactionMs: 150, Iterations: iterations, Seed: 7})
		}
	}
	runMatrix(t, runs, out)
}
