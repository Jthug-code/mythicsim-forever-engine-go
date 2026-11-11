package shaman

// The shaman changes in "WoW: Forever Hotfixes - November 10, 2026" (Blizzard news 24315281), as far as a
// damage sim reads them. Off by default, so every earlier run keeps its numbers; SetNov10Hotfix(true)
// applies them over the 2 November state.
//
//   - "Stormstrike damage reduced to 100% weapon damage (was 150%), and cooldown increased to 8 seconds
//     (was 6 seconds)": the two 2 November Stormstrike changes undone.
//   - Mental Dexterity back to 33/67/100% of Intellect (reported by mirrors of the notes; the wording there
//     is garbled, so check the tooltip): the 2 November change undone.
//   - "Maelstrom Weapon now grants 20% reduced cast time and Mana cost per application at all ranks (was
//     4/8/12/16/20%)", and its rate is 6/7/8/9/10 a minute by rank ("5 points ... 3.6 speed weapon ... 60%
//     chance to activate on swing"): the 2 November doubling undone, and both read from the rank.
//   - "Weapon imbues and Weapon totems can now be used at the same time ... the Shaman can benefit from
//     Windfury Weapon and Windfury Totem at the same time, granting their attacks a chance to activate
//     both": Windfury Weapon no longer takes the Windfury Totem category.
//
// Not modelled: "Lightning Overload now deals the correct amount of damage for Lightning Bolt" (the notes
// don't say what was wrong; the sim's overload is half a bolt), and Rage of the Storm's stats.
var nov10Hotfix bool

// SetNov10Hotfix applies or undoes the hotfix. Call it between sims, never while one is being built.
func SetNov10Hotfix(on bool) {
	nov10Hotfix = on
	for _, name := range []string{"Stormstrike 150% weapon damage", "Stormstrike 6 s cooldown",
		"Mental Dexterity 50/100/150%", "Maelstrom Weapon procs doubled"} {
		SetNov2Change(name, !on)
	}
}
