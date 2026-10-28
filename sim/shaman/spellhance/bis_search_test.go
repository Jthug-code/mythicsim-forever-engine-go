package spellhance

import (
	"encoding/json"
	"fmt"
	"os"
	"runtime"
	"sync"
	"testing"
	"time"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
)

type searchCandidate struct {
	ID      int32 `json:"id"`
	Name    string
	TwoHand bool `json:"two_hand"`
}

type searchInput struct {
	Start      *proto.EquipmentSpec         `json:"start"`
	Candidates map[string][]searchCandidate `json:"candidates"`
	Enchants   map[string][]struct {
		ID   int32 `json:"id"`
		Name string
	} `json:"enchants"`
}

// TestBisSearch builds the run 4 BiS set by coordinate descent: from the EP-ranked start in
// $BIS_SEARCH_IN, it tries every candidate item and enchant one slot at a time and keeps what the sim
// says is better, until a full pass changes nothing. It scores 8/36/7 with no hard casts as an Orc on a
// 3-minute fight, every trial on the same seed, and writes the set to $BIS_SEARCH_OUT.
func TestBisSearch(t *testing.T) {
	in, out := os.Getenv("BIS_SEARCH_IN"), os.Getenv("BIS_SEARCH_OUT")
	if in == "" || out == "" {
		t.Skip("set BIS_SEARCH_IN and BIS_SEARCH_OUT to search")
	}
	var input searchInput
	if err := json.Unmarshal([]byte(readFile(in)), &input); err != nil {
		t.Fatal(err)
	}
	iterations := matrixIterations(t)
	twoHand := map[int32]bool{}
	names := map[int32]string{}
	for _, cs := range input.Candidates {
		for _, c := range cs {
			twoHand[c.ID] = c.TwoHand
			names[c.ID] = c.Name
		}
	}

	score := func(items []*proto.ItemSpec) float64 {
		gear := &proto.EquipmentSpec{Items: items}
		raw, _ := json.Marshal(gear)
		tmp, err := os.CreateTemp("", "bis-*.gear.json")
		if err != nil {
			t.Fatal(err)
		}
		tmp.Write(raw)
		tmp.Close()
		defer os.Remove(tmp.Name())
		r := run{Build: melee8367, Gear: tmp.Name(), Duration: 3 * time.Minute, ReactionMs: 150, Iterations: iterations, Seed: 7,
			Race: proto.Race_RaceOrc}
		res := core.RunRaidSim(r.request())
		if res.Error != nil {
			t.Fatalf("%v: %s", items, res.Error.Message)
		}
		return res.RaidMetrics.Dps.Avg
	}
	clone := func(items []*proto.ItemSpec) []*proto.ItemSpec {
		out := make([]*proto.ItemSpec, len(items))
		for i, it := range items {
			out[i] = &proto.ItemSpec{Id: it.Id, Enchant: it.Enchant, RandomSuffix: it.RandomSuffix}
		}
		return out
	}
	// Scores every trial set in parallel.
	scoreAll := func(trials [][]*proto.ItemSpec) []float64 {
		scores := make([]float64, len(trials))
		var wg sync.WaitGroup
		jobs := make(chan int)
		for w := 0; w < runtime.NumCPU(); w++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for i := range jobs {
					scores[i] = score(trials[i])
				}
			}()
		}
		for i := range trials {
			jobs <- i
		}
		close(jobs)
		wg.Wait()
		return scores
	}

	current := input.Start.Items
	for len(current) < 17 {
		current = append(current, &proto.ItemSpec{})
	}
	best := score(current)
	log := []string{fmt.Sprintf("start %.1f DPS", best)}
	t.Log(log[0])
	for pass := 1; pass <= 4; pass++ {
		changed := false
		for slot := 0; slot < 17; slot++ {
			var trials [][]*proto.ItemSpec
			var labels []string
			for _, c := range input.Candidates[fmt.Sprint(slot)] {
				if c.ID == current[slot].Id {
					continue
				}
				// A ring or trinket isn't worn twice.
				if (slot == 10 && current[11].Id == c.ID) || (slot == 11 && current[10].Id == c.ID) ||
					(slot == 12 && current[13].Id == c.ID) || (slot == 13 && current[12].Id == c.ID) {
					continue
				}
				// An off-hand needs a one-hander in the main hand.
				if slot == 15 && twoHand[current[14].Id] {
					continue
				}
				trial := clone(current)
				trial[slot] = &proto.ItemSpec{Id: c.ID, Enchant: current[slot].Enchant}
				if slot == 14 && c.TwoHand {
					trial[15] = &proto.ItemSpec{}
				}
				trials = append(trials, trial)
				labels = append(labels, c.Name)
			}
			for _, e := range input.Enchants[fmt.Sprint(slot)] {
				if e.ID == current[slot].Enchant || current[slot].Id == 0 {
					continue
				}
				trial := clone(current)
				trial[slot].Enchant = e.ID
				trials = append(trials, trial)
				labels = append(labels, "enchant "+e.Name)
			}
			if len(trials) == 0 {
				continue
			}
			scores := scoreAll(trials)
			bi := -1
			for i, s := range scores {
				// Keep a change only past the noise a shared seed leaves.
				if s > best*1.0015 && (bi < 0 || s > scores[bi]) {
					bi = i
				}
			}
			if bi >= 0 {
				line := fmt.Sprintf("pass %d slot %d: %s, %.1f -> %.1f DPS", pass, slot, labels[bi], best, scores[bi])
				t.Log(line)
				log = append(log, line)
				current, best, changed = trials[bi], scores[bi], true
			}
		}
		if !changed {
			break
		}
	}
	raw, _ := json.MarshalIndent(map[string]any{"gear": &proto.EquipmentSpec{Items: current}, "dps": best, "log": log,
		"names": func() []string {
			var n []string
			for _, it := range current {
				n = append(n, names[it.Id])
			}
			return n
		}()}, "", " ")
	if err := os.WriteFile(out, raw, 0o644); err != nil {
		t.Fatal(err)
	}
}
