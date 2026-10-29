package spellhance

import (
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
)

// The launch presets and the SoD sets carry no enchants (the Forever BiS set does). These are a standard
// set of Forever enchants per spec, by item type. legs uses the head enchant, as both take arcana.
type enchantSet struct {
	byType         map[proto.ItemType]int32
	oneHand, twoHand int32
}

var (
	// The caster set: spell power where Forever has it, and Fire Power on gloves (fire is about 45% of
	// the Elemental preset's damage).
	casterEnchants = enchantSet{map[proto.ItemType]int32{
		proto.ItemType_ItemTypeHead: 2587, proto.ItemType_ItemTypeLegs: 2587, proto.ItemType_ItemTypeNeck: 8218,
		proto.ItemType_ItemTypeShoulder: 2605, proto.ItemType_ItemTypeChest: 1893, proto.ItemType_ItemTypeWrist: 1901,
		proto.ItemType_ItemTypeHands: 2616,
	}, 2504, 8211}
	// The melee set: the run 6 BiS search's picks for 8/36/7, and Crusader.
	meleeEnchants = enchantSet{map[proto.ItemType]int32{
		proto.ItemType_ItemTypeHead: 2587, proto.ItemType_ItemTypeLegs: 2587, proto.ItemType_ItemTypeNeck: 856,
		proto.ItemType_ItemTypeShoulder: 2717, proto.ItemType_ItemTypeBack: 849, proto.ItemType_ItemTypeChest: 1891,
		proto.ItemType_ItemTypeWrist: 1901, proto.ItemType_ItemTypeHands: 2563, proto.ItemType_ItemTypeFeet: 1887,
	}, 1900, 1900}
)

func (set enchantSet) apply(equipment *proto.EquipmentSpec) {
	for _, spec := range equipment.Items {
		item := core.GetItemByID(spec.Id)
		if item == nil {
			continue
		}
		if item.Type == proto.ItemType_ItemTypeWeapon {
			switch {
			case item.HandType == proto.HandType_HandTypeTwoHand:
				spec.Enchant = set.twoHand
			case item.HandType != proto.HandType_HandTypeOffHand && item.WeaponType != proto.WeaponType_WeaponTypeShield:
				spec.Enchant = set.oneHand
			}
			continue
		}
		if e, ok := set.byType[item.Type]; ok {
			spec.Enchant = e
		}
	}
}

// TestEnchantedGear reruns the main launch and SoD rows with and without the enchant sets above, at 2,
// 5 and 10 minutes, and writes enchants.csv to $SPELLHANCE_ENCH_OUT.
func TestEnchantedGear(t *testing.T) {
	out := os.Getenv("SPELLHANCE_ENCH_OUT")
	if out == "" {
		t.Skip("set SPELLHANCE_ENCH_OUT to run the comparison")
	}
	rows := []struct {
		b    build
		gear string
		set  enchantSet
	}{
		{elePreset, "ele_launch", casterEnchants}, {eleResto, "ele_launch", casterEnchants}, {weave2922, "ele_launch", casterEnchants},
		{enhPreset, "enh_launch_1h", meleeEnchants}, {melee8367, "enh_launch_1h", meleeEnchants},
		{elePreset, "caster_mirage_2h_sets", casterEnchants},
		{enhPresetAt10, "enh_smashbringer_2h_sets", meleeEnchants}, {melee8367at10, "enh_smashbringer_2h_sets", meleeEnchants},
	}
	durations := []time.Duration{2 * time.Minute, 5 * time.Minute, 10 * time.Minute}
	iterations := matrixIterations(t)
	var mu sync.Mutex
	var lines [][]string
	var wg sync.WaitGroup
	for _, rw := range rows {
		for _, enchanted := range []bool{false, true} {
			for _, d := range durations {
				wg.Add(1)
				go func() {
					defer wg.Done()
					r := run{Build: rw.b, Gear: rw.gear, Duration: d, ReactionMs: 150, Iterations: iterations, Seed: 7}
					req := r.request()
					if enchanted {
						rw.set.apply(req.Raid.Parties[0].Players[0].Equipment)
					}
					res := core.RunRaidSim(req)
					if res.Error != nil {
						t.Errorf("%s: %s", rw.b.Label, res.Error.Message)
						return
					}
					mu.Lock()
					lines = append(lines, []string{rw.b.Label, rw.gear, fmt.Sprint(enchanted), fmt.Sprint(d.Seconds()), fmt.Sprintf("%.1f", res.RaidMetrics.Dps.Avg)})
					mu.Unlock()
				}()
			}
		}
	}
	wg.Wait()
	if err := os.MkdirAll(out, 0o755); err != nil {
		t.Fatal(err)
	}
	f, err := os.Create(filepath.Join(out, "enchants.csv"))
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	w := csv.NewWriter(f)
	w.Write([]string{"build", "gear", "enchanted", "duration_s", "dps"})
	w.WriteAll(lines)
}
