// Package spellhance compares a 29/22/0 two-hand Spellhance shaman, which weaves hard casts between
// swings, against an 8/36/7 Enhancement build and the repo's Enhancement and Elemental presets.
//
// It is test-only. The rotations in apls/ follow the SpellhanceWeaveHelper addon's light: hard casts
// start only in the 0.4 s after the swing timer restarts, and an instant goes only when its GCD ends
// before the next swing. The gear is Season of Discovery Phase 8 best-in-slot (testdata/), which the
// Forever item database does not carry, so the tests add it as custom items. Run the full comparison
// with SPELLHANCE_OUT set to an output directory; without it only the quick checks run.
package spellhance
