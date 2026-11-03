package spellhance

import (
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/shaman"
)

// A balance idea to try on top of the 2 November changes. Zero values leave that part as the notes have it.
type candidate struct {
	name             string
	overloadPerRank  float64 // extra Lightning Overload chance per point, in percent
	stormstrikePct   float64 // Stormstrike weapon damage percent (the notes: 150)
	stormstrikeSpell float64 // Stormstrike's bonus to the next bolt or Earth Shock, in percent (the client: 20)
	masteryDamage    float64 // extra bolt damage while Lightning Mastery is up, in percent
	farseerHaste     float64 // Rage of the Farseer's melee haste, in percent (the client: 30)
	dexterity        float64 // Mental Dexterity's Intellect-to-attack-power share at 3/3, in percent (the notes: 150)
	maelstromSingle  bool    // Maelstrom Weapon's procs not doubled
	furyPackage      bool    // the run 6 Elemental Fury package: +1% crit a rank, 2.5x crits, +6% bolt base damage
	noCapstones      bool    // Lightning Mastery and Rage of the Farseer untaken (the point left unspent)
}

// Batch "overload": Lightning Overload raised (Elemental and 29/22 reach it, no Enhancement build does) and
// Stormstrike's damage trimmed (Enhancement and 29/22).
// Batch "levers": each burst or hybrid lever on its own, and what the two capstones are worth.
var candidateSets = map[string][]candidate{
	"overload": {
		{name: "2 November"},
		{name: "Lightning Overload 15/30/45%", overloadPerRank: 5},
		{name: "Lightning Overload 20/40/60%", overloadPerRank: 10},
		{name: "Stormstrike 125%", stormstrikePct: 125},
		{name: "Overload 45% + Stormstrike 125%", overloadPerRank: 5, stormstrikePct: 125},
		{name: "Overload 60% + Stormstrike 125%", overloadPerRank: 10, stormstrikePct: 125},
	},
	"levers": {
		{name: "2 November"},
		{name: "No capstones", noCapstones: true},
		{name: "Maelstrom not doubled", maelstromSingle: true},
		{name: "Stormstrike spell bonus 40%", stormstrikeSpell: 40},
		{name: "Stormstrike spell bonus 60%", stormstrikeSpell: 60},
		{name: "Lightning Mastery +50% damage", masteryDamage: 50},
		{name: "Lightning Mastery +100% damage", masteryDamage: 100},
		{name: "Rage of the Farseer 60% haste", farseerHaste: 60},
		{name: "Elemental Fury package", furyPackage: true},
	},
	// Batch "verify": the package that keeps sustained power out of the deep talents, and Mental Dexterity, which
	// was expected to pay the Intellect-heavy hybrid most (it doesn't: Enhancement gains as much).
	"verify": {
		{name: "2 November"},
		{name: "Mental Dexterity 200%", dexterity: 200},
		{name: "Mental Dexterity 250%", dexterity: 250},
		{name: "A: Maelstrom single, Overload 60%, Stormstrike 125%", maelstromSingle: true, overloadPerRank: 10, stormstrikePct: 125},
		{name: "A + Mental Dexterity 200%", maelstromSingle: true, overloadPerRank: 10, stormstrikePct: 125, dexterity: 200},
		{name: "A + Mental Dexterity 250%", maelstromSingle: true, overloadPerRank: 10, stormstrikePct: 125, dexterity: 250},
		{name: "B: Maelstrom single, Overload 60%, Mental Dexterity 250%", maelstromSingle: true, overloadPerRank: 10, dexterity: 250},
	},
	// Batch "respec": package A with the 29/22 respec rows, and A keeping Stormstrike at 150%.
	"respec": {
		{name: "2 November"},
		{name: "A: Maelstrom single, Overload 60%, Stormstrike 125%", maelstromSingle: true, overloadPerRank: 10, stormstrikePct: 125},
		{name: "A2: Maelstrom single, Overload 60%", maelstromSingle: true, overloadPerRank: 10},
	},
}

// 29/22 after Improved Ghost Wolf moves from tier 2 to tier 4 (2 November notes): its two points go to Ancestral
// Knowledge, 1/5 to 3/5, +6% Intellect instead of +2%. Improved Ghost Wolf changes nothing a damage sim reads.
var weave2922AK = build{"29/22 weave, Ancestral Knowledge 3/5", "053033102310134-053030031004102", "apls/weave.apl.json", false}

// apply sets the candidate's knobs; the zero candidate puts every one back.
func (c candidate) apply() {
	shaman.LightningOverloadBonusPerRank = c.overloadPerRank
	shaman.StormstrikeWeaponPercent = c.stormstrikePct
	shaman.StormstrikeSpellBonusPercent = c.stormstrikeSpell
	shaman.LightningMasteryDamagePercent = c.masteryDamage
	shaman.RageOfTheFarseerHastePercent = c.farseerHaste
	shaman.MentalDexterityPercent = c.dexterity
	shaman.SetNov2Change("Maelstrom Weapon procs doubled", !c.maelstromSingle)
	shaman.ElementalFuryCritPerRank, shaman.ElementalFuryExtraCritBonus, shaman.LightningBaseDamageBonus = 0, 0, 0
	if c.furyPackage {
		shaman.ElementalFuryCritPerRank, shaman.ElementalFuryExtraCritBonus, shaman.LightningBaseDamageBonus = 1, 1, 0.06
	}
}

// withoutCapstone leaves the build's 31-point talent untaken: Lightning Mastery is the Elemental tree's 16th digit,
// Rage of the Farseer the Enhancement tree's 18th.
func withoutCapstone(b build) build {
	trees := strings.Split(b.Talents, "-")
	for i, at := range []int{16, 18} {
		if i < len(trees) && len(trees[i]) >= at && trees[i][at-1] == '1' {
			trees[i] = trees[i][:at-1] + "0" + trees[i][at:]
		}
	}
	b.Talents = strings.Join(trees, "-")
	return b
}

// TestCandidatePackages runs one batch of balance candidates ($SPELLHANCE_CAND_SET, "overload" unless set).
// Elemental runs both the preset and the rank 10 gate 5 rotation. Launch gear and the Forever BiS sets, 2 / 5 / 10
// min, as the default Skyborne. It writes candidates.csv to $SPELLHANCE_CAND_OUT; SPELLHANCE_BUFFS=full picks the
// full raid buffs.
func TestCandidatePackages(t *testing.T) {
	out := os.Getenv("SPELLHANCE_CAND_OUT")
	if out == "" {
		t.Skip("set SPELLHANCE_CAND_OUT to run the candidates")
	}
	set := os.Getenv("SPELLHANCE_CAND_SET")
	if set == "" {
		set = "overload"
	}
	candidates, ok := candidateSets[set]
	if !ok {
		t.Fatalf("no candidate set %q", set)
	}
	iterations := matrixIterations(t)
	type row struct {
		b    build
		gear string
	}
	rows := []row{
		{enhPreset, "enh_launch_1h"}, {melee8367, "enh_launch_1h"}, {melee83211WS, "enh_launch_1h"},
		{elePreset, "ele_launch"}, {eleGate5, "ele_launch"}, {weave2922, "ele_launch"},
		{enhPreset, foreverBisGear()}, {melee8367, foreverBisGear()}, {melee83211WS, foreverBisGear()},
		{elePreset, foreverEleBisGear()}, {eleGate5, foreverEleBisGear()}, {weave2922, foreverHybBisGear()},
	}
	if set == "respec" {
		rows = append(rows, row{weave2922AK, "ele_launch"}, row{weave2922AK, foreverHybBisGear()})
	}
	durations := []time.Duration{2 * time.Minute, 5 * time.Minute, 10 * time.Minute}

	defer candidate{}.apply()
	var lines [][]string
	for _, c := range candidates {
		c.apply()
		var runs []run
		for _, rw := range rows {
			b := rw.b
			if c.noCapstones {
				b = withoutCapstone(b)
			}
			for _, d := range durations {
				runs = append(runs, run{Build: b, Gear: rw.gear, Duration: d, ReactionMs: 150, Iterations: iterations, Seed: 7})
			}
		}
		results := make([][]string, len(runs))
		var wg sync.WaitGroup
		jobs := make(chan int)
		for w := 0; w < runtime.NumCPU(); w++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for i := range jobs {
					r := runs[i]
					res := core.RunRaidSim(r.request())
					if res.Error != nil {
						t.Errorf("%s, %s on %s: %s", c.name, r.Build.Label, r.Gear, res.Error.Message)
						continue
					}
					results[i] = []string{c.name, r.Build.Label, r.Gear, fmt.Sprint(r.Duration.Seconds()),
						fmt.Sprintf("%.1f", res.RaidMetrics.Dps.Avg), fmt.Sprintf("%.1f", res.RaidMetrics.Parties[0].Players[0].SecondsOomAvg)}
				}
			}()
		}
		for i := range runs {
			jobs <- i
		}
		close(jobs)
		wg.Wait()
		lines = append(lines, results...)
	}

	if err := os.MkdirAll(out, 0o755); err != nil {
		t.Fatal(err)
	}
	f, err := os.Create(filepath.Join(out, "candidates.csv"))
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	w := csv.NewWriter(f)
	w.Write([]string{"candidate", "build", "gear", "duration_s", "dps", "seconds_oom"})
	w.WriteAll(lines)
}

// The capstone switch takes exactly the 31-point talent: Lightning Mastery from the Elemental preset, Rage of the
// Farseer from the Enhancement builds, and nothing from 29/22, which reaches neither.
func TestWithoutCapstone(t *testing.T) {
	for _, c := range []struct {
		b    build
		want string
	}{
		{elePreset, "5505331000001350--553302001"},
		{enhPreset, "5505031-053030031005112250"},
		{melee8367, "053-055032031005112250-050002"},
		{weave2922, talents2922},
	} {
		if got := withoutCapstone(c.b).Talents; got != c.want {
			t.Errorf("%s: %s, want %s", c.b.Label, got, c.want)
		}
	}
}
