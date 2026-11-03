package spellhance

import (
	"testing"

	"github.com/wowsims/forever/sim/classcompare"
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
)

// TestClassCompare runs the shaman builds in sim/classcompare's comparison: the same launch gear, rotations
// and consumables as the Spellhance runs (Enhancement one-handed), every race a shaman can be.
func TestClassCompare(t *testing.T) {
	suite := func(b build) core.CharacterSuiteConfig {
		s := core.CharacterSuiteConfig{
			Class:    proto.Class_ClassShaman,
			Talents:  b.Talents,
			Rotation: core.RotationCombo{Label: b.Label, Rotation: core.APLRotationFromJsonString(readFile(b.APL))},
		}
		if b.Elemental {
			s.SpecOptions = core.SpecOptionsCombo{Label: "Elemental", SpecOptions: &proto.Player_ElementalShaman{ElementalShaman: &proto.ElementalShaman{
				Options: &proto.ElementalShaman_Options{ClassOptions: &proto.ShamanOptions{}},
			}}}
			s.StartingDistance, s.Consumables = 15, casterConsumes()
		} else {
			s.SpecOptions = core.SpecOptionsCombo{Label: "Enhancement", SpecOptions: &proto.Player_EnhancementShaman{EnhancementShaman: &proto.EnhancementShaman{
				Options: &proto.EnhancementShaman_Options{
					SyncType:     proto.ShamanSyncType_Auto,
					ClassOptions: &proto.ShamanOptions{ImbueMh: proto.ShamanImbue_WindfuryWeapon},
				},
			}}}
			s.StartingDistance, s.Consumables = 5, enhConsumes()
		}
		return s
	}
	enh := launchGear["enh_launch_1h"]
	ele := launchGear["ele_launch"]
	classcompare.Run(t, "shaman", []classcompare.Spec{
		{Name: "Enhancement preset", Suite: suite(enhPreset), Gear: enh.file, Drop: enh.drop},
		{Name: "Enhancement 8/36/7", Suite: suite(melee8367), Gear: enh.file, Drop: enh.drop},
		{Name: "Enhancement 8/32/11", Suite: suite(melee83211WS), Gear: enh.file, Drop: enh.drop},
		{Name: "Elemental preset", Suite: suite(elePreset), Gear: ele.file, Role: classcompare.Caster, SchoolElixir: 6373},
		{Name: "29/22 weave", Suite: suite(weave2922), Gear: ele.file, Role: classcompare.Caster, SchoolElixir: 6373},
	})
}
