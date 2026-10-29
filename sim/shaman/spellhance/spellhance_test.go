package spellhance

import (
	"math"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/wowsims/forever/sim/common"
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/simsignals"
	"github.com/wowsims/forever/sim/core/stats"
	"github.com/wowsims/forever/sim/shaman"
	"github.com/wowsims/forever/sim/shaman/elemental"
	"github.com/wowsims/forever/sim/shaman/enhancement"
)

var gear map[string]*gearSet

func init() {
	enhancement.RegisterEnhancementShaman()
	elemental.RegisterElementalShaman()
	common.RegisterAllEffects()
	gear = loadSodGear("testdata/sod-p8-shaman-bis.json")
}

// Talent strings are Wowhead Forever's, without its "v2" version prefix.
const (
	// https://www.wowhead.com/forever/talent-calc/shaman/v2053033102310134-051032031004102
	talents2922 = "053033102310134-051032031004102"
	// https://www.wowhead.com/forever/talent-calc/shaman/v2053-055032031005112251-050002
	talents8367 = "053-055032031005112251-050002"
	// The repo's Enhancement and Elemental "Level 60" presets (ui/specs/shaman/*/presets.ts).
	talentsEnhPreset = "5505031-053030031005112251"
	talentsElePreset = "5505331000001351--553302001"
)

type build struct {
	Label     string
	Talents   string
	APL       string
	Elemental bool // Elemental spec, which does not melee; otherwise Enhancement
}

var (
	weave2922       = build{"29/22 weave", talents2922, "apls/weave.apl.json", false}
	chain2922       = build{"29/22 bolt-chain", talents2922, "apls/chain.apl.json", false}
	weave2922CL     = build{"29/22 weave + CL", talents2922, "apls/weave_cl.apl.json", false}
	chain2922CL     = build{"29/22 bolt-chain + CL", talents2922, "apls/chain_cl.apl.json", false}
	weave8367       = build{"8/36/7 weave", talents8367, "apls/weave_maelstrom.apl.json", false}
	melee8367       = build{"8/36/7 no hard casts", talents8367, "apls/melee_maelstrom.apl.json", false}
	weave8367at10   = build{"8/36/7 weave, MW at 10", talents8367, "apls/weave_maelstrom10.apl.json", false}
	melee8367at10   = build{"8/36/7 no hard casts, MW at 10", talents8367, "apls/melee_maelstrom10.apl.json", false}
	enhPreset       = build{"Enh preset 19/32/0", talentsEnhPreset, "../../../ui/specs/shaman/enhancement/apls/forever.apl.json", false}
	elePreset       = build{"Ele preset 32/0/19", talentsElePreset, "../../../ui/specs/shaman/elemental/apls/forever.apl.json", true}
	enhPresetLaunch = "../../../ui/specs/shaman/enhancement/gear_sets/launch.gear.json"
	elePresetLaunch = "../../../ui/specs/shaman/elemental/gear_sets/launch.gear.json"
)

// The repo's Forever launch-gear presets by name, with the consumables of the preset they come from.
// The Enhancement preset pairs two one-handers, and the engine dual-wields them, but shamans can't
// dual wield in Forever (2026-10-27; docs/forever_rules.md says otherwise). "enh_launch_1h"
// is that preset without its off-hand weapon, Premier High Warlord's Left Claw (272598).
var launchGear = map[string]struct {
	file     string
	consumes func() *proto.ConsumesSpec
	drop     int32
}{
	"enh_launch":    {enhPresetLaunch, enhConsumes, 0},
	"enh_launch_1h": {enhPresetLaunch, enhConsumes, 272598},
	"ele_launch":    {elePresetLaunch, casterConsumes, 0},
}

// The preset consumables (presets.ts): the Elemental ones go with the caster set, the Enhancement
// ones with the Enhancement set.
func casterConsumes() *proto.ConsumesSpec {
	return &proto.ConsumesSpec{FlaskId: 13512, SpellPowerElixirId: 13454, SchoolElixirId: 6373, GuardianElixirId: 20007,
		StrengthBuffId: 12451, AttackPowerBuffId: 12460, ZanzaId: 8423, FoodId: 18254, PotId: 13444, ConjuredId: 12662}
}

func enhConsumes() *proto.ConsumesSpec {
	return &proto.ConsumesSpec{FlaskId: 13512, BattleElixirId: 13452, SpellPowerElixirId: 13454, SchoolElixirId: 6373,
		GuardianElixirId: 20007, StrengthBuffId: 12451, AttackPowerBuffId: 12460, ZanzaId: 8410, DragonbreathChili: true,
		FoodId: 13810, PotId: 13444, ConjuredId: 12662}
}

// The comparison runs use both presets' buffs and debuffs at once, so no build is short a buff
// the other build's preset gives.
func unionBuffs() (*proto.PartyBuffs, *proto.RaidBuffs, *proto.Debuffs) {
	return &proto.PartyBuffs{BattleShout: proto.TristateEffect_TristateEffectImproved, LeaderOfThePack: true,
			ManaSpringTotem: proto.TristateEffect_TristateEffectRegular, MoonkinAura: true},
		&proto.RaidBuffs{ArcaneBrilliance: true, FireResistanceTotem: true, GiftOfTheWild: true, PrayerOfFortitude: true, PrayerOfSpirit: true},
		&proto.Debuffs{CurseOfRecklessness: true, ExposeArmor: true, FaerieFire: true, SunderArmor: true, CurseOfElements: true}
}

type run struct {
	Build    build
	Gear     string // a gearSet name, or a preset gear file
	Duration time.Duration
	// The measured rule (every completed hard cast restarts the swing) unless set.
	DisableCastSwingReset bool
	ReactionMs            int32
	NoHaste               bool
	Preset                bool // the preset's own race, buffs, consumables and distance
	Iterations            int32
	Seed                  int64
	Debug                 bool
	Race                  proto.Race    // Windshaper Skyborne unless set
	MobType               proto.MobType // the default target's (Mechanical) unless set
}

func (r run) request() *proto.RaidSimRequest {
	b := r.Build
	player := &proto.Player{
		Name:           "Shaman",
		Class:          proto.Class_ClassShaman,
		Race:           proto.Race_RaceWindshaperSkyborne,
		TalentsString:  b.Talents,
		Rotation:       core.APLRotationFromJsonString(readFile(b.APL)),
		ReactionTimeMs: r.ReactionMs,
		Buffs:          &proto.IndividualBuffs{},
		Profession1:    proto.Profession_Enchanting,
		Profession2:    proto.Profession_Alchemy,
	}
	if b.Elemental {
		player.Spec = &proto.Player_ElementalShaman{ElementalShaman: &proto.ElementalShaman{
			Options: &proto.ElementalShaman_Options{ClassOptions: &proto.ShamanOptions{}},
		}}
		player.DistanceFromTarget = 15
	} else {
		player.Spec = &proto.Player_EnhancementShaman{EnhancementShaman: &proto.EnhancementShaman{
			Options: &proto.EnhancementShaman_Options{
				SyncType:     proto.ShamanSyncType_Auto,
				ClassOptions: &proto.ShamanOptions{ImbueMh: proto.ShamanImbue_WindfuryWeapon},
			},
		}}
		player.DistanceFromTarget = 5
	}

	partyBuffs, raidBuffs, debuffs := unionBuffs()
	if set, ok := gear[r.Gear]; ok {
		player.Equipment = &proto.EquipmentSpec{Items: set.Items}
		player.Consumables = casterConsumes()
		if !strings.HasPrefix(r.Gear, "caster_mirage_2h") {
			player.Consumables = enhConsumes()
		}
		if r.NoHaste {
			player.BonusStats = &proto.UnitStats{Stats: stats.Stats{
				stats.SpellHasteRating: -set.SpellHaste,
				stats.MeleeHasteRating: -set.MeleeHaste,
			}.ToProtoArray()}
		}
	} else if launch, ok := launchGear[r.Gear]; ok {
		player.Equipment = core.EquipmentSpecFromJsonString(readFile(launch.file))
		for i, item := range player.Equipment.Items {
			if launch.drop != 0 && item.Id == launch.drop {
				player.Equipment.Items[i] = &proto.ItemSpec{}
			}
		}
		player.Consumables = launch.consumes()
	} else {
		// A gear file: the Forever BiS set of run 4, with the Enhancement consumables.
		player.Equipment = core.EquipmentSpecFromJsonString(readFile(r.Gear))
		player.Consumables = enhConsumes()
	}

	if r.Preset {
		// presets.ts: race, buffs, debuffs and consumables as each spec opens.
		if b.Elemental {
			player.Race = proto.Race_RaceTroll
			player.Consumables = casterConsumes()
			partyBuffs = &proto.PartyBuffs{ManaSpringTotem: proto.TristateEffect_TristateEffectRegular, MoonkinAura: true}
			raidBuffs = &proto.RaidBuffs{ArcaneBrilliance: true, GiftOfTheWild: true, PrayerOfFortitude: true, PrayerOfSpirit: true}
			debuffs = &proto.Debuffs{CurseOfElements: true}
		} else {
			player.Race = proto.Race_RaceOrc
			player.Consumables = enhConsumes()
			partyBuffs = &proto.PartyBuffs{BattleShout: proto.TristateEffect_TristateEffectImproved, LeaderOfThePack: true,
				ManaSpringTotem: proto.TristateEffect_TristateEffectRegular}
			raidBuffs = &proto.RaidBuffs{ArcaneBrilliance: true, FireResistanceTotem: true, GiftOfTheWild: true, PrayerOfFortitude: true, PrayerOfSpirit: true}
			debuffs = &proto.Debuffs{CurseOfRecklessness: true, ExposeArmor: true, FaerieFire: true, SunderArmor: true}
		}
	}

	if r.Race != proto.Race_RaceUnknown {
		player.Race = r.Race
	}
	if fullBuffs() {
		partyBuffs, raidBuffs, debuffs = applyFullBuffs(b, player)
	}

	encounter := core.MakeSingleTargetEncounter(0)
	encounter.Duration = r.Duration.Seconds()
	if r.MobType != proto.MobType_MobTypeUnknown {
		target := core.FreshDefaultTargetConfig()
		target.MobType = r.MobType
		encounter.Targets = []*proto.Target{target}
	}
	return &proto.RaidSimRequest{
		Raid:      core.SinglePlayerRaidProto(player, partyBuffs, raidBuffs, debuffs),
		Encounter: encounter,
		SimOptions: &proto.SimOptions{
			Iterations:            max(r.Iterations, 1),
			RandomSeed:            max(r.Seed, 1),
			DebugFirstIteration:   r.Debug,
			DisableCastSwingReset: r.DisableCastSwingReset,
		},
	}
}

func readFile(path string) string {
	raw, err := os.ReadFile(filepath.FromSlash(path))
	if err != nil {
		panic(err)
	}
	return string(raw)
}

func simulate(t testing.TB, r run) *proto.RaidSimResult {
	t.Helper()
	result := core.RunRaidSim(r.request())
	if result.Error != nil {
		t.Fatalf("%s on %s: %s", r.Build.Label, r.Gear, result.Error.Message)
	}
	return result
}

// The custom items add up to the gear file's own totals.
func TestSodGearMatchesTheFileTotals(t *testing.T) {
	want := map[string]stats.Stats{
		"caster_mirage_2h": {stats.Intellect: 404, stats.SpellDamage: 436, stats.SpellCritRating: 11 * 14, stats.MeleeCritRating: 11 * 14,
			stats.SpellHitRating: 5 * 10, stats.MeleeHitRating: 5 * 10, stats.SpellHasteRating: 28 * 10, stats.Stamina: 138},
		"enh_smashbringer_2h": {stats.Strength: 437, stats.Agility: 26, stats.Intellect: 130, stats.AttackPower: 24,
			stats.SpellCritRating: 15 * 14, stats.MeleeCritRating: 15 * 14, stats.SpellHitRating: 6 * 10, stats.MeleeHitRating: 6 * 10,
			stats.MeleeHasteRating: 27 * 10, stats.Stamina: 131},
	}
	for name, totals := range want {
		got := gear[name].Totals
		for stat, v := range totals {
			if v != 0 && math.Abs(got[stat]-v) > 1e-6 {
				t.Errorf("%s: %s = %v, the file says %v", name, stats.Stat(stat).StatName(), got[stat], v)
			}
		}
	}
	if got := gear["enh_smashbringer_2h"].Totals[stats.ExpertiseRating]; got != 60 {
		t.Errorf("Enhancement set: expertise %v, want 60 (6%% less dodged or parried)", got)
	}
}

// core.NewItemEffect drops an effect for an item the database does not have yet, without an error.
func TestSodEffectsAreRegistered(t *testing.T) {
	for _, id := range sodEffectItems {
		if !core.HasItemEffect(id) {
			t.Errorf("custom item %d has no effect", id)
		}
	}
}

type logCast struct {
	at       float64
	spellID  int32
	castTime float64
	gcd      float64
}

var (
	logCasting = regexp.MustCompile(`^\[([-0-9.]+)\] \[Shaman \(#1\)\] Casting \{SpellID: (\d+)\} \(Cost = [-0-9.]+, Cast Time = ([0-9.a-zµ]+), GCD = ([0-9.a-zµ]+),`)
	logSwing   = regexp.MustCompile(`^\[([-0-9.]+)\] \[Shaman \(#1\)\] Completed cast \{OtherID: 3, Tag: 1\}`)
	logDone    = regexp.MustCompile(`^\[([-0-9.]+)\] \[Shaman \(#1\)\] Completed cast \{SpellID: (\d+)\}$`)
)

func parseLog(t *testing.T, logs string) (casts []logCast, swings []float64, hardEnds []float64) {
	t.Helper()
	seconds := func(s string) float64 {
		d, err := time.ParseDuration(s)
		if err != nil {
			t.Fatal(err)
		}
		return d.Seconds()
	}
	hard := map[int32]bool{}
	for _, line := range strings.Split(logs, "\n") {
		line = strings.TrimRight(line, "\r")
		if m := logCasting.FindStringSubmatch(line); m != nil {
			at, _ := strconv.ParseFloat(m[1], 64)
			id, _ := strconv.Atoi(m[2])
			c := logCast{at: at, spellID: int32(id), castTime: seconds(m[3]), gcd: seconds(m[4])}
			casts = append(casts, c)
			if c.castTime > 0 {
				hard[c.spellID] = true
			}
		} else if m := logSwing.FindStringSubmatch(line); m != nil {
			at, _ := strconv.ParseFloat(m[1], 64)
			swings = append(swings, at)
		} else if m := logDone.FindStringSubmatch(line); m != nil {
			id, _ := strconv.Atoi(m[2])
			if hard[int32(id)] {
				at, _ := strconv.ParseFloat(m[1], 64)
				hardEnds = append(hardEnds, at)
			}
		}
	}
	return casts, swings, hardEnds
}

// The weave rotations play the SpellhanceWeaveHelper light: a hard cast starts in the 0.4 s after
// the swing timer restarts, no sooner than the reaction time (after a swing for the weave, after a
// swing or a completed hard cast for the bolt chain), and an instant goes only when its GCD ends
// before the next swing, unless the weapon swings faster than a GCD plus the window.
func TestWeaveRotationsFollowTheLight(t *testing.T) {
	const eps = 0.011 // the log's two decimals
	for _, c := range []struct {
		b    build
		gear string
	}{{weave2922, "caster_mirage_2h"}, {chain2922, "caster_mirage_2h"}, {weave2922CL, "caster_mirage_2h_sets"}, {chain2922CL, "caster_mirage_2h_sets"}} {
		b := c.b
		for _, reactionMs := range []int32{150, 10} {
			res := simulate(t, run{Build: b, Gear: c.gear, Duration: 90 * time.Second, ReactionMs: reactionMs, Iterations: 1, Debug: true})
			casts, swings, hardEnds := parseLog(t, res.Logs)
			reaction := float64(reactionMs) / 1000
			last := func(times []float64, at float64) float64 {
				prev := math.Inf(-1)
				for _, x := range times {
					if x <= at+eps {
						prev = x
					}
				}
				return prev
			}
			hard, instants := 0, 0
			for _, c := range casts {
				if c.at < 0 {
					continue
				}
				if c.castTime > 0 {
					hard++
					sinceSwing := c.at - last(swings, c.at)
					ok := sinceSwing >= reaction-eps && sinceSwing <= 0.4+eps
					if (b == chain2922 || b == chain2922CL) && !ok {
						sinceCast := c.at - last(hardEnds, c.at)
						ok = sinceCast >= reaction-eps && sinceCast <= 0.4+eps
					}
					if !ok {
						t.Errorf("%s, %d ms: %d started at %.2f, %.2f after the last swing", b.Label, reactionMs, c.spellID, c.at, sinceSwing)
					}
					continue
				}
				if c.gcd == 0 {
					continue
				}
				instants++
				for _, s := range swings {
					if s > c.at+eps {
						// A swing shorter than the GCD plus the window leaves no WAIT: any instant crosses it.
						fast := s-last(swings, c.at) < c.gcd+0.4-eps
						if c.at+c.gcd > s+eps && !fast {
							t.Errorf("%s, %d ms: %d at %.2f has a %.3f s GCD past the swing at %.2f", b.Label, reactionMs, c.spellID, c.at, c.gcd, s)
						}
						break
					}
				}
			}
			if hard < 10 || instants < 10 {
				t.Errorf("%s, %d ms: %d hard casts and %d instants in 90 s; the rotation did not run", b.Label, reactionMs, hard, instants)
			}
		}
	}
}

// The set pieces turn on the bonuses: the 4-piece's Overload chance, the 6-piece's 10-charge
// Maelstrom Weapon, and nothing without the pieces.
func TestSodSetBonusesAreActive(t *testing.T) {
	newShaman := func(b build, gearName string) *shaman.Shaman {
		sim := core.NewSim(run{Build: b, Gear: gearName, Duration: time.Minute, ReactionMs: 150}.request(), simsignals.CreateSignals())
		sim.Reset()
		return sim.Raid.Parties[0].Players[0].(shaman.ShamanAgent).GetShaman()
	}
	for gearName, want := range map[string]float64{"caster_mirage_2h": 0, "caster_mirage_2h_sets": 0.10} {
		if got := newShaman(weave2922, gearName).BonusOverloadChance; got != want {
			t.Errorf("29/22 in %s: bonus Overload chance %v, want %v", gearName, got, want)
		}
	}
	for gearName, want := range map[string]int32{"enh_smashbringer_2h": 5, "enh_smashbringer_2h_sets": 10, "enh_smashbringer_2h_sets_rage4": 10} {
		if got := newShaman(melee8367, gearName).GetAura("Maelstrom Weapon").MaxStacks; got != want {
			t.Errorf("8/36/7 in %s: Maelstrom Weapon stacks to %d, want %d", gearName, got, want)
		}
	}
	for _, label := range []string{setSoulcrushersStorm + " 6P", setEarthshatterersStorm + " 2P"} {
		if newShaman(weave2922, "caster_mirage_2h_sets").GetAura(label) == nil {
			t.Errorf("no %s aura in the caster set", label)
		}
	}
}

// Spent at 10 charges, a Maelstrom Weapon bolt casts twice: two hits a cast.
func TestMaelstromAtTenCastsTwice(t *testing.T) {
	res := simulate(t, run{Build: melee8367at10, Gear: "enh_smashbringer_2h_sets", Duration: 2 * time.Minute, ReactionMs: 150, Iterations: 20})
	for _, action := range res.RaidMetrics.Parties[0].Players[0].Actions {
		if action.Id.GetSpellId() != 15208 || action.Id.Tag != 0 {
			continue
		}
		casts, hits := 0, 0
		for _, target := range action.Targets {
			casts += int(target.Casts)
			hits += int(target.Hits + target.Crits + target.Misses)
		}
		if casts < 20 || hits != 2*casts {
			t.Errorf("Lightning Bolt: %d casts landed %d hits, want twice as many", casts, hits)
		}
		return
	}
	t.Error("no Lightning Bolt cast")
}

// The one-hand Enhancement launch gear has no off-hand weapon, so it does not dual-wield.
func TestEnhLaunchOneHandDoesNotDualWield(t *testing.T) {
	for gearName, want := range map[string]bool{"enh_launch": true, "enh_launch_1h": false} {
		sim := core.NewSim(run{Build: enhPreset, Gear: gearName, Duration: time.Minute, ReactionMs: 150}.request(), simsignals.CreateSignals())
		sim.Reset()
		if got := sim.Raid.Parties[0].Players[0].GetCharacter().AutoAttacks.IsDualWielding; got != want {
			t.Errorf("%s: dual wielding %v, want %v", gearName, got, want)
		}
	}
}
