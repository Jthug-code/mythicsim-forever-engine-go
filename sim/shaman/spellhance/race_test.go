package spellhance

import (
	"encoding/json"
	"os"
	"testing"
	"time"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
)

// Run 4 compares races for 8/36/7 with no hard casts, on a Forever BiS set built from Wowhead's
// Forever item data (testdata/forever-enh-bis.gear.json; RESULTS.md says how it was chosen).
const foreverBisGear = "testdata/forever-enh-bis.gear.json"

var raceRows = []proto.Race{proto.Race_RaceOrc, proto.Race_RaceDwarf, proto.Race_RaceWindshaperSkyborne}

// The stats the BiS search weighs, by their index in the item database.
var weighedStats = []proto.Stat{
	proto.Stat_StatStrength, proto.Stat_StatAgility, proto.Stat_StatIntellect, proto.Stat_StatSpellDamage,
	proto.Stat_StatNatureDamage, proto.Stat_StatFireDamage, proto.Stat_StatSpellHitRating, proto.Stat_StatSpellCritRating,
	proto.Stat_StatAttackPower, proto.Stat_StatMeleeHitRating, proto.Stat_StatMeleeCritRating,
	proto.Stat_StatMeleeHasteRating, proto.Stat_StatExpertiseRating, proto.Stat_StatArmorPenetration, proto.Stat_StatMP5,
}

// TestBisStatWeights writes the 8/36/7 no-hard-cast stat weights (EP, attack power = 1) on the gear in
// $BIS_WEIGHTS_GEAR to $BIS_WEIGHTS_OUT, for the BiS search.
func TestBisStatWeights(t *testing.T) {
	gearFile, out := os.Getenv("BIS_WEIGHTS_GEAR"), os.Getenv("BIS_WEIGHTS_OUT")
	if gearFile == "" || out == "" {
		t.Skip("set BIS_WEIGHTS_GEAR and BIS_WEIGHTS_OUT to compute the weights")
	}
	r := run{Build: melee8367, Gear: gearFile, Duration: 3 * time.Minute, ReactionMs: 150, Iterations: matrixIterations(t), Seed: 7,
		Race: proto.Race_RaceOrc}
	request := r.request()
	result := core.StatWeights(&proto.StatWeightsRequest{
		Player:             request.Raid.Parties[0].Players[0],
		RaidBuffs:          request.Raid.Buffs,
		PartyBuffs:         request.Raid.Parties[0].Buffs,
		Debuffs:            request.Raid.Debuffs,
		Encounter:          request.Encounter,
		SimOptions:         request.SimOptions,
		StatsToWeigh:       weighedStats,
		// Hit rating as a bonus stat moves nothing here, so hit is weighed in percent.
		PseudoStatsToWeigh: []proto.PseudoStat{proto.PseudoStat_PseudoStatMainHandDps,
			proto.PseudoStat_PseudoStatMeleeHitPercent, proto.PseudoStat_PseudoStatSpellHitPercent},
		EpReferenceStat:    proto.Stat_StatAttackPower,
	})
	if result.Error != nil {
		t.Fatal(result.Error.Message)
	}
	ep := map[string]float64{}
	for _, s := range weighedStats {
		ep[s.String()] = result.Dps.EpValues.Stats[s]
	}
	for _, s := range []proto.PseudoStat{proto.PseudoStat_PseudoStatMainHandDps, proto.PseudoStat_PseudoStatMeleeHitPercent,
		proto.PseudoStat_PseudoStatSpellHitPercent} {
		ep[s.String()] = result.Dps.EpValues.PseudoStats[s]
	}
	raw, _ := json.MarshalIndent(ep, "", "  ")
	if err := os.WriteFile(out, raw, 0o644); err != nil {
		t.Fatal(err)
	}
	t.Logf("%s", raw)
}

// TestSpellhanceRaceMatrix runs 8/36/7 with no hard casts as an Orc, a Dwarf and a Windshaper Skyborne
// on the Forever BiS set, and writes results.csv and sources.csv to $SPELLHANCE_RACE_OUT. Racials are
// the 26 October notes'. Every race gets the same buffs, so only the racials differ.
func TestSpellhanceRaceMatrix(t *testing.T) {
	out := os.Getenv("SPELLHANCE_RACE_OUT")
	if out == "" {
		t.Skip("set SPELLHANCE_RACE_OUT to run the comparison")
	}
	iterations := matrixIterations(t)
	var runs []run
	for _, mob := range []proto.MobType{proto.MobType_MobTypeUnknown, proto.MobType_MobTypeElemental, proto.MobType_MobTypeBeast} {
		for _, race := range raceRows {
			for _, d := range []time.Duration{2 * time.Minute, 5 * time.Minute} {
				for _, reaction := range []int32{150, 10} {
					if mob != proto.MobType_MobTypeUnknown && reaction != 150 {
						continue
					}
					runs = append(runs, run{Build: melee8367, Gear: foreverBisGear, Duration: d, ReactionMs: reaction,
						Iterations: iterations, Seed: 7, Race: race, MobType: mob})
				}
			}
		}
	}
	runMatrix(t, runs, out)
}
