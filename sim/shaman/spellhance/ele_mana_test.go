package spellhance

import (
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/stats"
)

// TestEleManaSweep measures how much mana is worth to Elemental, and what each race does for it. It
// adds MP5 as a bonus stat (not gear; 100000 stands for unlimited mana) to the Water Shield + Mana Tide
// build and the preset, on the launch gear and the SoD caster set, and writes ele-mana.csv to
// $SPELLHANCE_ELE_OUT.
func TestEleManaSweep(t *testing.T) {
	out := os.Getenv("SPELLHANCE_ELE_OUT")
	if out == "" {
		t.Skip("set SPELLHANCE_ELE_OUT to run the sweep")
	}
	type row struct {
		b    build
		race proto.Race
		mp5  float64
		gear string
		d    time.Duration
	}
	var rows []row
	for _, gear := range []string{"ele_launch", "caster_mirage_2h_sets"} {
		for _, d := range []time.Duration{2 * time.Minute, 5 * time.Minute} {
			for _, race := range []proto.Race{proto.Race_RaceWindshaperSkyborne, proto.Race_RaceTroll, proto.Race_RaceOrc, proto.Race_RaceDwarf} {
				rows = append(rows, row{eleResto, race, 0, gear, d})
			}
			for _, mp5 := range []float64{25, 50, 100, 100000} {
				rows = append(rows, row{eleResto, proto.Race_RaceWindshaperSkyborne, mp5, gear, d})
			}
			for _, mp5 := range []float64{0, 100000} {
				rows = append(rows, row{elePreset, proto.Race_RaceWindshaperSkyborne, mp5, gear, d})
			}
		}
	}
	lines := make([][]string, len(rows))
	var wg sync.WaitGroup
	for i, rw := range rows {
		wg.Add(1)
		go func() {
			defer wg.Done()
			r := run{Build: rw.b, Gear: rw.gear, Duration: rw.d, ReactionMs: 150, Iterations: matrixIterations(t), Seed: 7, Race: rw.race}
			req := r.request()
			if rw.mp5 != 0 {
				req.Raid.Parties[0].Players[0].BonusStats = &proto.UnitStats{Stats: stats.Stats{stats.MP5: rw.mp5}.ToProtoArray()}
			}
			res := core.RunRaidSim(req)
			if res.Error != nil {
				t.Errorf("%s: %s", rw.b.Label, res.Error.Message)
				return
			}
			lines[i] = []string{rw.b.Label, rw.gear, strings.TrimPrefix(rw.race.String(), "Race"), fmt.Sprint(rw.mp5), fmt.Sprint(rw.d.Seconds()),
				fmt.Sprintf("%.1f", res.RaidMetrics.Dps.Avg), fmt.Sprintf("%.1f", res.RaidMetrics.Parties[0].Players[0].SecondsOomAvg)}
		}()
	}
	wg.Wait()
	if err := os.MkdirAll(out, 0o755); err != nil {
		t.Fatal(err)
	}
	f, err := os.Create(filepath.Join(out, "ele-mana.csv"))
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	w := csv.NewWriter(f)
	w.Write([]string{"build", "gear", "race", "bonus_mp5", "duration_s", "dps", "seconds_oom"})
	w.WriteAll(lines)
}
