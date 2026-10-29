package spellhance

import (
	"testing"
	"time"
)

// Enchanting one run's SoD gear must leave the next run's alone. Before request() copied the shared
// set's items, TestEnchantedGear's "unenchanted" SoD rows came out enchanted.
func TestRequestsDoNotShareGear(t *testing.T) {
	r := run{Build: elePreset, Gear: "caster_mirage_2h_sets", Duration: time.Minute, ReactionMs: 150, Iterations: 1, Seed: 7}
	first := r.request().Raid.Parties[0].Players[0].Equipment
	casterEnchants.apply(first)
	for i, item := range r.request().Raid.Parties[0].Players[0].Equipment.Items {
		if item.Enchant != 0 {
			t.Fatalf("slot %d of a fresh request carries enchant %d from an earlier one", i, item.Enchant)
		}
	}
}
