package shaman

import (
	"testing"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

// Each overload row sits at its rank's level, and Chain Lightning's is not half the rank: in client
// 70425 rank 4 averaged 66 at 60 (62 + 1.2 a level), where halving 126 gave 63. The 2 November notes
// scale the row with its rank (nov2_notes.go): 62 and 1.2 times the rank's own gains, 276 at 60.
func TestLightningOverloadRows(t *testing.T) {
	for _, c := range []struct {
		name            string
		ranks, overload spelldata.Ladder
	}{
		{"Lightning Bolt", LightningBoltRankMap, lightningBoltOverloadRanks},
		{"Chain Lightning", ChainLightningRankMap, chainLightningOverloadRanks},
	} {
		if c.ranks.Len() != c.overload.Len() {
			t.Fatalf("%s: %d ranks, %d overload rows", c.name, c.ranks.Len(), c.overload.Len())
		}
		c.ranks.Each(func(rank int32, s *spelldata.Spell) {
			if a, b := s.DamageEffect().SpellLevel, c.overload.Rank(rank).DamageEffect().SpellLevel; a != b {
				t.Errorf("%s rank %d: level %v, overload row level %v", c.name, rank, a, b)
			}
		})
	}
	if avg := chainLightningOverloadRanks.Rank(4).DamageEffect().Average(core.CharacterLevel); avg != 276 {
		t.Fatalf("Chain Lightning rank 4 overload averages %v at 60, want 276", avg)
	}
}
