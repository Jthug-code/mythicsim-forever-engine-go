// Package classcompare runs every DPS spec on equal terms, to see where shaman stands against the other
// classes: each spec's Forever launch gear, talents and rotation, as every race its class can be (the best
// race is kept), with the same buffs and one consumable policy for everyone. Each class package has a
// TestClassCompare that hands its own specs to Run.
//
//	CLASS_COMPARE_OUT=<dir> CLASS_COMPARE_ITERATIONS=5000 go test --tags=with_db ./sim/... -run TestClassCompare -timeout 120m
//
// Results go to <dir>/<file>.csv, one row per spec, buff level, race and fight length.
package classcompare

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	googleProto "google.golang.org/protobuf/proto"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
)

// Role picks the consumables a spec gets when its suite leaves a slot empty.
type Role int

const (
	Melee  Role = iota // Strength or Agility melee
	Ranged             // hunters
	Caster
)

type Spec struct {
	Name  string                    // e.g. "Fury"
	Suite core.CharacterSuiteConfig // class, talents, spec options, rotation, consumables, distance
	Gear  string                    // a gear_sets file
	Drop  int32                     // an item to take off the gear file (the Enhancement launch preset's off-hand)
	Role  Role
	// The caster's school elixir (Fire Power, Frost Power, Shadow Power, Holy Power), or 0 for none.
	SchoolElixir int32
	Agility      bool // Agility food rather than Strength food
}

// Buff levels, the same two the Spellhance runs use.
var Levels = []string{"light", "full"}

var Durations = []time.Duration{2 * time.Minute, 5 * time.Minute, 10 * time.Minute}

func iterations() int32 {
	if n, err := strconv.Atoi(os.Getenv("CLASS_COMPARE_ITERATIONS")); err == nil && n > 0 {
		return int32(n)
	}
	return 5000
}

// Light: a party with Battle Shout, Leader of the Pack, Mana Spring and Moonkin, the raid's Intellect,
// Wild, Fortitude and Spirit, and the main debuffs. No blessings, Windfury Totem or Mana Tide.
func lightBuffs() (*proto.PartyBuffs, *proto.RaidBuffs, *proto.Debuffs, *proto.IndividualBuffs) {
	return &proto.PartyBuffs{BattleShout: proto.TristateEffect_TristateEffectImproved, LeaderOfThePack: true,
			ManaSpringTotem: proto.TristateEffect_TristateEffectRegular, MoonkinAura: true},
		&proto.RaidBuffs{ArcaneBrilliance: true, FireResistanceTotem: true, GiftOfTheWild: true, PrayerOfFortitude: true, PrayerOfSpirit: true},
		&proto.Debuffs{CurseOfRecklessness: true, ExposeArmor: true, FaerieFire: true, SunderArmor: true, CurseOfElements: true},
		&proto.IndividualBuffs{}
}

// Full: the engine's full raid. A shaman drops the party's Windfury Totem, since it is the shaman.
func fullBuffs(class proto.Class) (*proto.PartyBuffs, *proto.RaidBuffs, *proto.Debuffs, *proto.IndividualBuffs) {
	party := googleProto.Clone(core.FullPartyBuffs).(*proto.PartyBuffs)
	if class == proto.Class_ClassShaman {
		party.WindfuryTotem = false
	}
	return party, googleProto.Clone(core.FullRaidBuffs).(*proto.RaidBuffs), googleProto.Clone(core.FullDebuffs).(*proto.Debuffs),
		googleProto.Clone(core.FullIndividualBuffs).(*proto.IndividualBuffs)
}

// consumables fills every slot the suite leaves empty with its role's standard: Mongoose, the Jujus,
// R.O.I.D.S., Dragonbreath Chili and stat food for melee and hunters; Flask of Supreme Power, Greater
// Arcane Elixir, the school elixir, Mageblood, Cerebral Cortex Compound and Runn Tum Tuber Surprise for
// casters; Major Mana Potions and Demonic Runes for everyone with mana. Full buffs add Greater
// Mageblood for mana users, Brilliant Wizard Oil for casters and the Agility and Strength scrolls for the
// rest.
func consumables(spec Spec, level string) *proto.ConsumesSpec {
	c := &proto.ConsumesSpec{}
	if spec.Suite.Consumables != nil {
		c = googleProto.Clone(spec.Suite.Consumables).(*proto.ConsumesSpec)
	}
	fill := func(slot *int32, id int32) {
		if *slot == 0 {
			*slot = id
		}
	}
	mana := spec.Suite.Class != proto.Class_ClassWarrior && spec.Suite.Class != proto.Class_ClassRogue
	switch spec.Role {
	case Caster:
		fill(&c.FlaskId, 13512)
		fill(&c.SpellPowerElixirId, 13454)
		fill(&c.SchoolElixirId, spec.SchoolElixir)
		fill(&c.GuardianElixirId, 20007)
		fill(&c.ZanzaId, 8423)
		fill(&c.FoodId, 18254)
	default:
		fill(&c.BattleElixirId, 13452)
		fill(&c.StrengthBuffId, 12451)
		fill(&c.AttackPowerBuffId, 12460)
		fill(&c.ZanzaId, 8410)
		c.DragonbreathChili = true
		fill(&c.FoodId, core.TernaryInt32(spec.Agility, 13928, 13810))
	}
	if mana {
		fill(&c.PotId, 13444)
		fill(&c.ConjuredId, 12662)
	}
	if level == "full" {
		if mana {
			c.GuardianElixirId = 250341
		}
		if spec.Role == Caster {
			fill(&c.MhTempEnchantId, 25122)
		} else {
			c.ScrollAgi, c.ScrollStr = true, true
		}
	}
	// The engine registers a potion or conjured item only from the lists the UI sends; PotId and ConjuredId
	// just pick one of them (the Spellhance harness's run 1-5 bug).
	if c.PotId != 0 && len(c.Potions) == 0 {
		c.Potions = []int32{c.PotId}
	}
	if c.ConjuredId != 0 && len(c.ConjuredItems) == 0 {
		c.ConjuredItems = []int32{c.ConjuredId}
	}
	return c
}

func request(spec Spec, level string, race proto.Race, d time.Duration, n int32) (*proto.RaidSimRequest, error) {
	raw, err := os.ReadFile(filepath.FromSlash(spec.Gear))
	if err != nil {
		return nil, err
	}
	gear := core.EquipmentSpecFromJsonString(string(raw))
	for i, item := range gear.Items {
		if spec.Drop != 0 && item.Id == spec.Drop {
			gear.Items[i] = &proto.ItemSpec{}
		}
	}
	s := spec.Suite
	party, raid, debuffs, individual := lightBuffs()
	if level == "full" {
		party, raid, debuffs, individual = fullBuffs(s.Class)
	}
	player := core.WithSpec(&proto.Player{
		Name:               spec.Name,
		Class:              s.Class,
		Race:               race,
		TalentsString:      s.Talents,
		Rotation:           s.Rotation.Rotation,
		Equipment:          gear,
		Consumables:        consumables(spec, level),
		Buffs:              individual,
		Cooldowns:          s.Cooldowns,
		DistanceFromTarget: s.StartingDistance,
		ReactionTimeMs:     150,
		Profession1:        s.Profession1,
		Profession2:        s.Profession2,
	}, s.SpecOptions.SpecOptions)
	// The Spellhance runs' professions for anyone whose suite names none.
	if player.Profession1 == proto.Profession_ProfessionUnknown && player.Profession2 == proto.Profession_ProfessionUnknown {
		player.Profession1, player.Profession2 = proto.Profession_Enchanting, proto.Profession_Alchemy
	}

	encounter := core.MakeSingleTargetEncounter(0)
	encounter.Duration = d.Seconds()
	return &proto.RaidSimRequest{
		Raid:       core.SinglePlayerRaidProto(player, party, raid, debuffs),
		Encounter:  encounter,
		SimOptions: &proto.SimOptions{Iterations: n, RandomSeed: 7},
	}, nil
}

type row struct {
	spec, level string
	race        proto.Race
	duration    time.Duration
	dps, oom    float64
}

// Run sims every spec at both buff levels as every race its class can be, at every fight length, and
// writes <CLASS_COMPARE_OUT>/<file>.csv. Without CLASS_COMPARE_OUT it skips.
func Run(t *testing.T, file string, specs []Spec) {
	out := os.Getenv("CLASS_COMPARE_OUT")
	if out == "" {
		t.Skip("set CLASS_COMPARE_OUT to run the class comparison")
	}
	n := iterations()
	bySpec := map[string]Spec{}
	var rows []row
	for _, spec := range specs {
		bySpec[spec.Name] = spec
		races := core.ClassRaceCapabilities[spec.Suite.Class]
		if len(races) == 0 {
			t.Fatalf("%s: no races listed for %s", spec.Name, spec.Suite.Class)
		}
		for _, level := range Levels {
			for _, race := range races {
				for _, d := range Durations {
					rows = append(rows, row{spec: spec.Name, level: level, race: race, duration: d})
				}
			}
		}
	}

	var wg sync.WaitGroup
	var mu sync.Mutex
	jobs := make(chan int)
	for w := 0; w < runtime.NumCPU(); w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := range jobs {
				r := &rows[i]
				req, err := request(bySpec[r.spec], r.level, r.race, r.duration, n)
				var res *proto.RaidSimResult
				if err == nil {
					res = core.RunRaidSim(req)
					if res.Error != nil {
						err = fmt.Errorf("%s", res.Error.Message)
					}
				}
				if err != nil {
					mu.Lock()
					t.Errorf("%s %s %s: %v", r.spec, r.level, r.race, err)
					mu.Unlock()
					continue
				}
				r.dps, r.oom = res.RaidMetrics.Dps.Avg, res.RaidMetrics.Parties[0].Players[0].SecondsOomAvg
			}
		}()
	}
	for i := range rows {
		jobs <- i
	}
	close(jobs)
	wg.Wait()

	var csv strings.Builder
	csv.WriteString("spec,buffs,race,duration_s,iterations,dps,seconds_oom\n")
	for _, r := range rows {
		fmt.Fprintf(&csv, "%s,%s,%s,%d,%d,%.1f,%.1f\n", r.spec, r.level, strings.TrimPrefix(r.race.String(), "Race"),
			int(r.duration.Seconds()), n, r.dps, r.oom)
	}
	if err := os.MkdirAll(out, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(out, file+".csv"), []byte(csv.String()), 0o644); err != nil {
		t.Fatal(err)
	}
}

// WithoutSpell drops every rotation entry that casts spellID. The warrior APLs Slam when Slam's cast time is
// under 1.5 s, meant for Improved Slam; in a Fury build without it only a cast-speed racial gets there, and
// each completed cast restarts the swing timer, so the line costs those races damage the others don't lose.
func WithoutSpell(rotation *proto.APLRotation, spellID int32) *proto.APLRotation {
	kept := rotation.PriorityList[:0:0]
	for _, item := range rotation.PriorityList {
		if item.GetAction().GetCastSpell().GetSpellId().GetSpellId() == spellID {
			continue
		}
		kept = append(kept, item)
	}
	rotation.PriorityList = kept
	return rotation
}
