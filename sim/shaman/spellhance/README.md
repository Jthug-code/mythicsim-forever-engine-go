# Forever shaman sims

Shaman simulations for WoW: Forever: Elemental, Enhancement and a 29/22 Elemental/Enhancement hybrid. They're built on
this MythicSim engine and posted on the Forever Beta Discussion forum, in "I simmed Forever shaman a lot. Here's where
Elemental stands, and a buff that wouldn't touch Enhancement".

Everything here is a sim, not a combat log. Read the numbers as direction and rough size.

## Where things are

- **[RESULTS.md](RESULTS.md)** has every finding. Run 6, at the top, is the current state. The sections below it are
  older runs, kept as they were written.
- **The CSVs** hold the raw numbers; files ending in `-fullbuffs` use full raid buffs.
  - `forever-results*.csv`: Forever launch gear.
  - `results*.csv`: the SoD Phase 8 proxy.
  - `race-results*.csv`: races on the Forever BiS set.
  - `fight-length*.csv`: 2 to 10 minute fights.
  - `ele-mana*.csv`: Elemental mana.
  - `enh-ws-results*.csv`: Water Shield Enhancement.
  - `tuning-results.csv`, `elefury-results.csv`, `lean-results.csv`, `resist-results.csv`, `compensate-results.csv`
    and `enh-at10-results.csv`: the Elemental Fury tuning ideas, resists and hit, and Maelstrom at 10.
- **Gear:**
  - The Forever launch presets are the repo's own, `ui/specs/shaman/*/gear_sets/launch.gear.json`, with the
    Enhancement off-hand dropped because shamans can't dual wield.
  - `testdata/sod-p8-shaman-bis.json` is the SoD Phase 8 BiS proxy, with sources and the set bonuses that apply.
  - `testdata/forever-enh-bis*.gear.json` is the Forever BiS set, searched from Wowhead's Forever item data.
    `bis_tools/` holds the search scripts.
- **Rotations** are in `apls/`. `gen_apls.py` writes the weave rotations.

## What this branch changes in the engine

- **Racials** follow the 26 October 2026 development notes: Dwarf +1% hit, Orc's Savage removed, Skyborne 2% haste. The
  golden test results are regenerated to match.
- **Test-only knobs in `sim/shaman`**, all zero by default, so they change nothing unless a test sets them:
  - `ElementalFury*` adds crit, crit damage, damage, hit or Spell Piercing.
  - `LightningBaseDamageBonus` raises Lightning Bolt and Chain Lightning base damage.
- **Everything else is test code in this folder.**

## Running it

Set up the engine first, as the repo's [installation guide](../../../docs/installation.md) says (Go, protobuf,
generated protos). Then, from the repo root:

    go test --tags=with_db ./sim/shaman/spellhance/

That runs the quick checks. Each comparison is a test that only runs when its output folder is set, for example:

    SPELLHANCE_FOREVER_OUT=out SPELLHANCE_ITERATIONS=10000 go test --tags=with_db ./sim/shaman/spellhance/ -run TestSpellhanceForeverMatrix -timeout 60m

`SPELLHANCE_BUFFS=full` switches any of them to full raid buffs. `bis_tools/rerun_all.sh light|full <dir>` reruns
everything at one buff level; run it from this folder.

## License

MIT, same as the rest of the repo, which builds on [wowsims](https://github.com/wowsims/forever) through ElliotWood's
and sage3648's forks.
