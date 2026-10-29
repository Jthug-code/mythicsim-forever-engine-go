# Spellhance comparison results

All single target (`MakeSingleTargetEncounter`, no duration variation), 10,000 iterations per row, seed 7, using the
measured swing rule. Engine: `mythicsim/oct21-release` (62b7445d1c) plus this branch.

**Run 6 is the current state.** Every CSV except `results-run1.csv` now holds run 6's numbers. The run 2–5 sections
further down are kept as written; their numbers predate the potion fix.

| Run | Gear | Files | Section |
|---|---|---|---|
| 6 | Runs 2–5 again with mana potions and runes actually drunk, at both buff levels, plus Elemental Fury buffs | all CSVs; `-fullbuffs` = full raid buffs | Run 6, first below |
| 5 | Runs 2–4 again with full raid buffs and consumables | `*-fullbuffs.csv` | Run 5 |
| 4 | A Forever BiS set from Wowhead's Forever item data; Orc, Dwarf and Skyborne | `race-results.csv`, `race-sources.csv` | Run 4 |
| 3 | The repo's Forever launch-gear presets | `forever-results.csv`, `forever-sources.csv` | Run 3 |
| 2 | Season of Discovery Phase 8 BiS, set bonuses on and off | `results.csv`, `sources.csv` | Run 2 |
| 1 | SoD BiS, no set bonuses, with the old-swing-rule reference rows | `results-run1.csv` | Run 1 references, at the end |

Rerun everything at one buff level with `bis_tools/rerun_all.sh light|full <work dir>`, run from this folder. It runs
the BiS search, then the race, Elemental mana, Water Shield Enhancement, fight-length, Forever-launch and SoD matrices.
The Elemental Fury variants run with `SPELLHANCE_FURY_OUT=<dir> go test --tags=with_db . -run TestEleFuryBuff`.

# Run 6: mana potions fixed, both buff levels (2026-10-29)

**The bug, from the start of run 1:**
- The engine registers potions and conjured items only from the lists the UI sends (`Potions`, `ConjuredItems`).
  `PotId` and `ConjuredId` just pick one of them.
- The harness set only the ids, so no Major Mana Potion or Demonic Rune was ever drunk.
- `request()` now fills the lists, and `TestManaPotionsAreDrunk` fails on the old harness.
- The engine's policy is sound once the items are registered. It drinks as soon as the largest roll fits under max
  mana, so each 2-minute cooldown restarts as early as it can. It isn't perfect: it waits for room for the top roll,
  not the average, and skips a last, partly wasted potion.
- Dark Rune (bind on equip, same restore and cooldown) stands in for the bind-on-pickup Demonic Rune.

**"Light" means run 4's buffs:** Battle Shout, Leader of the Pack, Mana Spring, Moonkin, Arcane Intellect,
Fortitude, Spirit, Mark of the Wild and the main debuffs. No blessings, Judgement of Wisdom or Mana Tide.

## Run 6 findings

- **With potions, Elemental barely runs out of mana.**
  - Light buffs: 6 s of a 5-minute fight, 24 s over 10 minutes.
  - The Enhancement preset still does in long, light fights: 72 s over 5 minutes, 216 s over 10.
  - 8/36/7: 14 s and 94 s.
  - Full buffs: about 0 s for everyone.
- **The Elemental gap is a damage gap at either buff level** (Elemental preset as % of the Enhancement preset, 2 / 5 /
  10 min, launch gear):

  | Buffs | Elemental preset | Enhancement preset | Elemental's share |
  |---|---|---|---|
  | Light | 430 / 398 / 385 | 474 / 458 / 445 | 91 / 87 / 86% |
  | Full | 505 / 468 / 459 | 573 / 559 / 548 | 88 / 84 / 84% |
  | SoD proxy, full (2 / 5 min) | 1225 / 1157 | 1508 / 1486 | 81 / 78% |
- **Elemental loses about 7% from 2 to 5 minutes; Enhancement loses 3%.** Both stay flat after that. Two caveats:
  - The rows with 100,000 bonus MP5 aren't a clean "unlimited mana" reference. With free mana the preset rotation casts
    more Chain Lightning, which is weaker on one target: at full buffs the 2-minute DPS goes down, 505 → 502.
  - The fall-off comes from front-loaded cooldowns and consumables plus the rotation's mana gate.
- **Water Shield is worth little once potions are drunk.**
  - Elemental + Water Shield + Mana Tide against the preset: +2% / +0.4% (light, launch, 2 / 5 min), +3% at 10 minutes,
    +2–3% on the SoD proxy. Under full buffs it's -3% / -1%.
  - 8/32/11 against 8/36/7: -0.6% / +0.2% (light, launch), +2.4% / +2.9% at 8 / 10 minutes, and -1% under full buffs.
- **The 29/22 weave in caster gear beats the Elemental preset on launch gear with full buffs:** 519 / 517 against
  505 / 468. It's still 7–9% behind the Enhancement preset (573 / 559). On the SoD proxy it trails Elemental (964 / 972
  against 1225 / 1157).
- **Races, Forever BiS, 8/36/7:**
  - Light: Skyborne 736 / 728, Orc 729 / 720, Dwarf 724 / 718. Elemental targets: Skyborne 773 / 764. Beasts: Dwarf
    760 / 754.
  - Full: Skyborne 816 / 807, Orc 809 / 800, Dwarf 803 / 796.
- **BiS:** the light search now picks the full-buff set, except a Chest: Major Intellect enchant against Greater Stats.

## Run 6: Elemental Fury buffs (`elefury-results.csv`)

**Why Elemental Fury:** it sits on row 6 of the Elemental tree, so it needs 25 Elemental points first. Neither
Enhancement build can reach it: the preset has 19 Elemental points and 8/36/7 has 8. The 29/22 hybrid has 4/5.
Lightning Overload and Lava Burst are on row 5 (20 points), which a 20/31 Enhancement build could still reach.

**How it was tested:**
- The knobs are `shaman.ElementalFuryExtraCritBonus`, `ElementalFuryCritPerRank` and `ElementalFuryDamagePerRank`, all
  zero by default.
- The Enhancement rows moved by exactly 0.0% in every variant.

**Crit talents in Forever's client data, at max rank (crits for):**
- Every caster talent stops at 2.0×: Elemental Fury, Ruin, Ice Shards, Vengeance and Pandemic.
- The physical talents go higher: Mortal Shots and Predator's Edge 2.3×, Lethality and Impale 2.2×.

**Elemental preset, full buffs, % gain (launch / SoD proxy):**

| Elemental Fury at 5/5 | Gain (launch / SoD proxy) | 5-min DPS, launch |
|---|---|---|
| Crits for 2.25× | +3.8 / +7.3% | 486 |
| Crits for 2.5× | +7.6 / +14.6% | 504 |
| +1% crit per rank | +4.3 / +3.5% | 488 |
| +2% crit per rank | +8.5 / +7.1% | 508 |
| +2% damage per rank | +10 / +10% | 515 |
| +4% damage per rank | +20 / +20% | 562 |
| +1% crit per rank, crits for 2.25× | +9.1 / +11.7% | 511 |
| **+1% crit per rank, crits for 2.5×** | **+14.0 / +19.9%** | **534** |
| +1% crit per rank, +2% damage per rank | +14.7 / +13.9% | 537 |

- **With +1% crit per rank and 2.5× crits:**
  - Launch gear: 576 / 534 against the Enhancement preset's 573 / 559, so 101% at 2 minutes and 96% at 5.
  - SoD proxy: 1468 / 1387, which is 97% / 93%.
  - The 29/22 hybrid gains 6–11% from the combined packages.
**A change to the rotations since run 1 (all three runs' APLs have it now).** A weapon that swings faster than an
instant's GCD plus the 0.4 s window can't fit any GCD between swings. The addon then drops WAIT and shows INSTANT
throughout, and the APLs now do the same.

- **Why it mattered:** before the fix, the Elemental preset's 1.4 s dagger never got an instant.
- **What it moved in run 2:** only the 8/36/7 rows on the SoD Enhancement set, where Flurry, Rage of the Farseer and
  Abandoned Experiment push the swing under that line. Those rows rose 0.5% to 7.3%; 146 of 194 rows didn't change.
- **Run 1 is as it was run.** Its 8/36/7 Enhancement-set rows understate by up to about 3%.

# Run 5: everything again with full raid buffs and consumables (2026-10-29)

`SPELLHANCE_BUFFS=full` (`fullbuffs_test.go`); `bis_tools/run5_fullbuffs.sh` reruns it all. Files carry
`-fullbuffs`.

**What the full set is:**
- **Raid, individual and debuffs:** the engine's own `Full*` sets. That is all five Greater Blessings including Wisdom,
  Judgement of Wisdom, Improved Mana Spring, one party Mana Tide, Trueshot, Moonkin, Atiesh and the full debuff list.
- **No party Windfury Totem.** The Enhancement builds run their own Grace of Air.
- **Consumables:** each preset's flask, elixirs, food, potions, runes, Juju and Zanza, plus three additions:
  - Greater Mageblood Elixir (Forever, 20 MP5) for every build.
  - Brilliant Wizard Oil for Elemental.
  - Scrolls of Agility and Strength for Enhancement.
- **No world buffs.** The engine models none, and Forever's tooltips say they have no effect in raid instances (reported
  by wowdee.com and forum threads).
- **No Forever "Natural" flasks.** The engine only knows their +60 Stamina, so every build keeps Flask of Supreme Power.

**A new row everywhere: `Ele 32/0/19 Water Shield + Mana Tide`.**
- The Elemental preset takes Water Shield but its rotation never casts it. `apls/ele_watershield.apl.json` casts it
  before the pull and again whenever it drops.
- The talents trade Natural Grace 3 and Improved Reincarnation 2 for Totemic Focus 4 and Mana Tide Totem.

**The BiS set was searched again under full buffs** (`testdata/forever-enh-bis-fullbuffs.gear.json`, 800.0 DPS as an
Orc at 3 minutes). Four changes from run 4:
- Stormcaller's Diadem for the Black Dragonscale Helm.
- Cloak of the Fallen God for the Arcanoweave Cloak.
- Belt of Never-ending Agony for Grasp of the Fallen Emperor.
- Chest: Greater Stats for Major Intellect.

The buffs carry the mana, so the set moves toward melee stats.

## Run 5 findings

- **Full raid buffs remove mana as the limit in a 5-minute single-target fight.** Out-of-mana time is about 0 s for both
  specs on every gear set, down from 48–218 s.
- **What's left is a damage gap.** On launch gear, Elemental reaches 82–87% of the Enhancement preset. On the SoD proxy
  it reaches 75%.
- **Elemental still loses 8% from 2 to 5 minutes on launch gear** (496 → 455).
  - With unlimited mana it's 502 → 481. The rest is cooldowns and consumables front-loaded into the short fight.
  - The preset rotation's mana gate (max-rank Lightning Bolt only while mana ≥ 35 × seconds left) still costs about
    5% at 5 minutes. That gate is a rotation choice, not a game rule.
- **Water Shield + Mana Tide is about even with the preset under full buffs.** Launch 493 / 449 against 496 / 455; SoD
  proxy 1154 / 1119 against 1190 / 1115. Its value is all in under-buffed raids: in run 4's buff set it was +11% on
  launch gear and +32% on the SoD proxy at 5 minutes.
- **Enhancement 19/32/0 now leads 8/36/7 on launch gear** (573 / 557 against 526 / 518). The run 4 lead came from
  8/36/7's mana edge, which the full buffs make moot. On the SoD proxy they're even (1485 against 1491); Maelstrom at
  10 stacks leads with 1606.
- **The 29/22 hybrid in caster gear beats the Elemental preset over 2 minutes on launch gear** (518 against 496). It
  loses over 5 (448 against 455), when it runs out of mana (102 s). It trails Enhancement throughout.
- **Races keep the run 4 order.** On a neutral target, Skyborne 816 / 807, Orc 809 / 800, Dwarf 803 / 796.
  - Against elementals the Skyborne does 857 / 847.
  - Against beasts the Dwarf does 842 / 835.
  - Elemental by race on launch gear at 5 minutes: Skyborne 449, Dwarf 447, Orc 445, Troll 444.

| Build, 2 / 5 min DPS | Launch gear | SoD proxy (set bonuses) |
|---|---|---|
| Enhancement preset 19/32/0 | 573 / 557 | 1485 at 5 min; 1507 at 2 |
| 8/36/7 no hard casts | 526 / 518 | 1504 / 1491; 1621 / 1606 with Maelstrom at 10 |
| Elemental preset 32/0/19 | 496 / 455 | 1190 / 1115 |
| Ele 32/0/19 Water Shield + Mana Tide | 493 / 449 | 1154 / 1119 |
| 29/22 weave, caster gear | 518 / 448 | 964 / 927 |

## Run 5 addendum: the 8/32/11 Water Shield Enhancement

The build ([talent calc](https://www.wowhead.com/forever/talent-calc/shaman/v2053-051032031005112251-053002001)) is
8/36/7 with Ancestral Knowledge 5 → 1, plus Mindfulness 3 and Water Shield, cast and kept up
(`apls/melee_maelstrom_ws.apl.json`).
- **Files:** `enh-ws-results.csv` (run 4 buffs) and `enh-ws-results-fullbuffs.csv`.
- **Setup:** Windshaper Skyborne, 150 ms, 10,000 iterations, standard errors 0.2–1.3 DPS.

| Gear, 2 / 5 min DPS (out of mana at 5 min) | 8/36/7 | 8/32/11 Water Shield |
|---|---|---|
| Launch, run 4 buffs | 439 / 400 (159 s) | 459 / 419 (115 s) |
| SoD proxy, run 4 buffs | 1361 / 1269 (129 s) | 1351 / 1330 (17 s) |
| Forever BiS (run 4 set), run 4 buffs | 726 / 676 (86 s) | 719 / 710 (1 s) |
| Launch, full buffs | 526 / 518 | 522 / 515 |
| SoD proxy, full buffs | 1504 / 1491 | 1490 / 1477 |
| Forever BiS (run 5 set), full buffs | 817 / 807 | 810 / 800 |

- **With run 4's lighter buffs, Water Shield wins about 5% over 5 minutes at every gear level.**
  - On launch gear it also wins over 2 minutes, because that gear runs out of mana inside 2 minutes.
- **Under full buffs it loses about 1%.** Ancestral Knowledge multiplies Intellect, which 8/36/7 turns into attack
  power, and with full buffs the mana isn't needed.

# Run 4: Orc vs Dwarf vs Skyborne, 8/36/7 with no hard casts, Forever BiS (2026-10-28)

One build, `8/36/7 no hard casts` (`apls/melee_maelstrom.apl.json`), on one gear set worn by all three races. Buffs,
debuffs and consumables are run 3's Enhancement ones and identical for every race, so only the racials differ.

**Racials are the 26 October development notes'** (Kaivax; reported via timesaver.gg). This
branch patches `sim/core/racials.go` to match:
- **Dwarf:** Artisan's Precision, +1% hit with attacks and spells, replaces Earth Fury (2% crit damage). Beast Slaying
  (+5% damage to beasts) stays.
- **Orc:** Savage (+1% crit) removed. Blood Fury (+10% attack power and spell damage for 15 s, 3 min cooldown) stays and
  fires twice per 5-minute fight.
- **Windshaper Skyborne:** Wind Blessed drops from 3% to 2% attack and cast speed. Elemental Insight (+5% damage to
  elementals) stays. Only the Windshaper variant can be a shaman.
- Not patched: the same notes' Tauren and Human changes.

## The Forever BiS set

**The pool:** every item in Wowhead's Forever gear-planner data (`origin/data/master/wowhead-forever`, 28 October),
level 60 or lower, rare or better, mail or lighter, and usable by a shaman. Other classes' tier pieces are excluded,
including the Naxxramas tier, whose tooltips name no class.

**The data includes AQ40 and part of Naxxramas, so this is end-of-content BiS.** The data doesn't say which phase
anything arrives in.

**How it was picked:**
1. Stat weights for this build came from the sim. `TestBisStatWeights` uses attack power = 1.
2. Each slot's top candidates and enchants were ranked by those weights. Every trinket and totem with an effect was also
   included.
3. `TestBisSearch` then ran a coordinate descent: try every candidate item and enchant one slot at a time, as an Orc on a
   3-minute fight, keeping changes worth more than 0.15%.
4. It converged at 708.7 DPS.
5. A second search from that set, using weights recomputed on it, changed only the cloak (710.1).

| Slot | Item | Enchant | Source (Wowhead Forever data) |
|---|---|---|---|
| Head | Black Dragonscale Helm (252605) | Vodouisant's Vigilant Embrace | Crafted (new in Forever) |
| Neck | Gem of Trapped Innocents (23057) | Necklace: Strength | Kel'Thuzad, Naxxramas |
| Shoulder | Warlord's Mail Spaulders (16580) | Might of the Scourge | Horde PvP vendor (Lady Palanseer) |
| Back | Arcanoweave Cloak (272411) | — | Vendor, Pix Xizzix, Stranglethorn (new in Forever) |
| Chest | Stormcaller's Hauberk (21374) | Chest: Major Intellect | AQ40 quest |
| Wrist | Qiraji Execution Bracers (21602) | Bracer: Superior Intellect | Twin Emperors, AQ40 |
| Hands | Gauntlets of Kalimdor (21624) | Gloves: Superior Strength | Viscidus, AQ40 |
| Waist | Grasp of the Fallen Emperor (21607) | — | Twin Emperors, AQ40 |
| Legs | Stormcaller's Leggings (21375) | Vodouisant's Vigilant Embrace | AQ40 quest |
| Feet | Stormcaller's Footguards (21373) | Boots: Greater Agility | AQ40 quest |
| Ring | Band of Unnatural Forces (23038) | — | Loatheb, Naxxramas |
| Ring | Band of the Better Half (284716) | — | None listed (new in Forever) |
| Trinket | Kiss of the Spider (22954) | — | Maexxna, Naxxramas |
| Trinket | Slayer's Crest (23041) | — | Sapphiron, Naxxramas |
| Two-hand | Dark Edge of Insanity (21134), 3.5 s axe | Crusader | C'Thun, AQ40 |
| Totem | Totem of the Storm (23199) | — | Drop |

**Character totals (Orc):**
- About 378 Intellect and 1,500 attack power.
- 9% melee and spell hit: 4% from gear and 5% from Tidal Focus 5/5.
- 1.2% expertise.
- Three Stormcaller's pieces. The sim models the set's 3-piece Nature damage proc; the 5-piece is Chain Heal.

**Why Intellect:** in 8/36/7 Intellect feeds attack power and mana. On this set, 1 Intellect is worth 2.55 attack power,
against Strength's 2.0 and Agility's 1.55. The build is short of mana: 84–86 s of every 5-minute fight are spent out of
mana, for every race.

**Melee hit is past the cap.** The sim's base melee miss against a level-63 boss is 8%, so melee hit weighs 0 at 9%.
Spell hit is far from its 16% cap and is worth about 11 attack power per 1%.

**Caveats on the set:**
- **Band of the Better Half** has no source in the data.
  - It is half of a new two-ring set, Partners in Crime. The other ring is First Mate Band.
  - The 2-piece bonus is +22 Strength and a chance on swing for +50 Strength for 10 s. The engine doesn't model it, and
    this set wears only one ring.
- **Every piece counts as usable by both factions** (2026-10-28). The data says otherwise in two places:
  - Warlord's Mail Spaulders are Horde PvP, and the data has no Alliance shaman PvP set.
  - Stormcaller's pieces and the Earthshatter tier carry a race mask of the Classic Horde races (178), which leaves out
    both the Dwarf and the Skyborne.
  - Under this assumption every race wears the same set, and the race rows below compare racials only.

## Run 4 findings

- **Windshaper Skyborne is the best of the three on a neutral target.** It beats the Orc by 0.8–0.9% (about 6 DPS) and
  the Dwarf by 1.3–1.5%.
- **Orc is second, Dwarf third,** 0.4–0.6% apart (3–5 DPS).
  - The Dwarf's +1% hit is wasted on melee: Tidal Focus plus gear already reach the cap.
  - It only reaches spells, and Blood Fury outweighs it.
- **The slaying racials outweigh all of that against their creature type.**
  - Against elementals the Skyborne gains 5%, and leads the Orc by 5.9%.
  - Against beasts the Dwarf gains 5%, and leads the Skyborne by 3.4% and the Orc by 4.3%.
- **Every gap is real at 10,000 iterations.** Standard errors are 0.4–0.7 DPS, the gaps are 3–6 DPS on neutral
  targets, and every race shares the same seed.
- **Reaction 150 ms vs 10 ms moves nothing** (1 DPS or less), as expected for a no-hard-cast rotation.
- **Physical share is 75% at 2 minutes and 78% at 5, for every race.** The top sources are melee swings (38–40%),
  Windfury (23–24%), Stormstrike (13–14%), Maelstrom Lightning Bolts (7–8%) and Earth Shock (4–6%).
- **For reference only, since gear is assumed shared:** a Dwarf without the Horde PvP shoulders would lose another
  0.2–0.9%. The replacements, at 2 / 5 minutes, against 715.4 / 667.4 with the PvP shoulders:
  - Stormcaller's Pauldrons: 714.0 / 667.2
  - Mantle of Wicked Revenge: 712.6 / 662.5
  - Theramore Shoulderguards: 709.1 / 663.0
- **The search optimised for the Orc.** With melee hit capped for every race, the other races' best sets should differ
  little, but no per-race search was run.

## Run 4 table: DPS, reaction 150 ms

| Race | Neutral 2m | Neutral 5m | Elemental 2m | Elemental 5m | Beast 2m | Beast 5m |
|---|---:|---:|---:|---:|---:|---:|
| Windshaper Skyborne | **726.1** | **675.9** | **762.4** | **709.7** | 726.1 | 675.9 |
| Orc | 720.0 | 670.1 | 720.0 | 670.1 | 720.0 | 670.1 |
| Dwarf | 715.4 | 667.4 | 715.4 | 667.4 | **751.1** | **700.7** |

"Neutral" is the default target (Mechanical), which none of the slaying racials touch. Reaction 10 ms rows are in
`race-results.csv`.

# Run 3: Forever launch-gear presets (2026-10-27)

The repo's Forever gear presets (`ui/specs/shaman/*/gear_sets/launch.gear.json`), worn by every build. Everything else
is as in run 2: Windshaper Skyborne, both presets' buffs and debuffs together, the consumables of the preset the gear
comes from, and 150 ms and 10 ms reactions. 52 rows.

- **Elemental launch gear:** Dark Whisper Blade (20720, a 1.4 s dagger) with Scepter of Interminable Focus (22329)
  held in the off hand, and Totem of the Storm.
- **Enhancement launch gear, main hand only:** the Enhancement preset pairs Premier High Warlord's Cleaver (272592,
  2.9 s one-hand axe) with Premier High Warlord's Left Claw (272598), and the engine dual-wields them. Shamans can't
  dual wield in Forever (2026-10-27), although `docs/forever_rules.md` says they can. So run 3 drops the
  Left Claw and swings the Cleaver alone. The preset as shipped, dual-wielding, is kept as one reference row: about
  12% above main hand only, and not a set a Forever shaman can wear.
- **No Forever set bonus is active.** Every set in these presets is represented by one piece: Darkmantle Armor,
  Champion's Earthshaker, Bloodmail Regalia, Augur's Regalia and Sorcerer's Regalia.
- **Neither preset has a two-hander.** The weave the addon is built for is therefore not what these presets play. On
  the dagger, every instant is always INSTANT (see the rotation change at the top).

**No Forever best-in-slot set: skipped, because it can't be built credibly from the engine's item data.**
- **The item DB doesn't record what is obtainable at launch.** Its `phase` field is 1 for 10,581 of 10,769 items,
  including Naxxramas, Ahn'Qiraj and Blackwing Lair drops.
- **Sources are sparse.** Most Forever-only items have no source at all; 7 of the Enhancement launch preset's 17 items
  have none.
- **The repo says the data isn't in yet.** `docs/DATA_SOURCES.md`: launch is 4 Nov 2026, and the gear planner fills in
  from Wowhead uploads over launch week.
- **The repo's own "Phase 1" and "Phase 2" sets don't fit either.** They are classic Molten Core-era lists (Sulfuras,
  Staff of Dominance), not Forever launch.

## Run 3 findings

- **The gap is much narrower on Forever launch gear than on SoD gear.** The best 29/22 in the Elemental gear reaches
  78% of the best 8/36/7 in the Enhancement gear at 2 minutes (345 against 442), and 65% at 5 (264 against 403). On SoD
  gear it was 59–60% and 44–46%. Four things drive it:
  - **No haste on the Forever items.** The SoD Enhancement set's 27% melee haste was worth 15–18% to the Enhancement
    builds.
  - **No SoD set bonuses.** Soulcrusher's Rage 6 gave 8/36/7 about 15%.
  - **No expertise, and a one-hand Enhancement weapon.** The SoD Enhancement set had 6% expertise and a 4.0 s
    two-hander; here it's a 2.9 s one-hander, which shrinks Stormstrike and Windfury.
  - **Far fewer melee stats overall.**
- **Full Elemental is still the best use of caster gear.** The Elemental preset does 400 / 326; the best 29/22 does
  345 / 264. That matches run 2 with set bonuses on. In run 2 without them, the 29/22 had led.
- **8/36/7 with no hard casts is the best Enhancement-gear option again:** 442 / 403, level with the Enhancement
  preset's 439 / 404. Hard casting in melee gear still loses: the 8/36/7 weave does 357.
- **On these presets the 29/22 does about as well in Enhancement gear as in Elemental gear at 2 minutes** (349
  against 345). At 5 minutes it does clearly better in Enhancement gear (328 against 264), because casting runs it out
  of mana: the Elemental-gear weave is out for 225 of 300 s.
- **Chain Lightning loses without the Storm set's +70%** (315 against 345 in Elemental gear), as in run 2 with the
  bonuses off.
- **Physical share follows the SoD pattern.** The 29/22 in Elemental gear is 39% physical at 2 minutes and 57% at 5
  (SoD: 34% and 55%). 8/36/7 in Enhancement gear is 74% and 78%.
- **Reaction 150 ms vs. 10 ms** moves results by 2% or less.

## Run 3 tables: DPS on the Forever launch-gear presets

| Build | Ele gear 2m | Ele gear 5m | Enh gear 2m | Enh gear 5m |
|---|---:|---:|---:|---:|
| **reaction 150 ms** | | | | |
| 29/22 weave | 345 | 264 | 349 | 328 |
| 29/22 weave + CL | 315 | 252 | 345 | 326 |
| 29/22 bolt-chain | 330 | 258 | 332 | 320 |
| 8/36/7 weave | 346 | 284 | 357 | 342 |
| 8/36/7 no hard casts | 394 | 321 | 442 | 403 |
| Enh preset 19/32/0 | — | — | 439 | 404 |
| Ele preset 32/0/19 | 400 | 326 | — | — |
| **reaction 10 ms** | | | | |
| 29/22 weave | 338 | 260 | 352 | 329 |
| 29/22 weave + CL | 315 | 251 | 347 | 328 |
| 29/22 bolt-chain | 332 | 257 | 337 | 322 |
| 8/36/7 weave | 350 | 286 | 362 | 344 |
| 8/36/7 no hard casts | 394 | 321 | 442 | 403 |
| Enh preset 19/32/0 | — | — | 439 | 404 |
| Ele preset 32/0/19 | 400 | 326 | — | — |

### Run 3: the gap, reaction 150 ms

| Fight | Best 29/22, Elemental gear | Best 29/22, Enhancement gear | Best 8/36/7, Enhancement gear | 29/22 Ele gear as % of 8/36/7 |
|---|---:|---:|---:|---:|
| 2m | 345 (29/22 weave) | 349 (29/22 weave) | 442 (8/36/7 no hard casts) | 78% |
| 5m | 264 (29/22 weave) | 328 (29/22 weave) | 403 (8/36/7 no hard casts) | 65% |

### Run 3: physical vs spell DPS

Physical DPS / spell DPS (physical share).

| Build | Gear | Reaction | 2m | 5m |
|---|---|---|---|---|
| 29/22 weave | Elemental | 150 ms | 133 / 212 (39%) | 150 / 114 (57%) |
| 29/22 weave | Elemental | 10 ms | 128 / 209 (38%) | 148 / 112 (57%) |
| 29/22 weave | Enhancement | 150 ms | 251 / 98 (72%) | 263 / 65 (80%) |
| 29/22 weave | Enhancement | 10 ms | 253 / 99 (72%) | 263 / 66 (80%) |
| 29/22 weave + CL | Elemental | 150 ms | 132 / 184 (42%) | 150 / 102 (59%) |
| 29/22 weave + CL | Elemental | 10 ms | 132 / 184 (42%) | 149 / 102 (59%) |
| 29/22 weave + CL | Enhancement | 150 ms | 255 / 90 (74%) | 264 / 62 (81%) |
| 29/22 weave + CL | Enhancement | 10 ms | 257 / 91 (74%) | 265 / 63 (81%) |
| 29/22 bolt-chain | Elemental | 150 ms | 118 / 212 (36%) | 144 / 113 (56%) |
| 29/22 bolt-chain | Elemental | 10 ms | 120 / 212 (36%) | 144 / 113 (56%) |
| 29/22 bolt-chain | Enhancement | 150 ms | 234 / 98 (70%) | 255 / 65 (80%) |
| 29/22 bolt-chain | Enhancement | 10 ms | 238 / 98 (71%) | 257 / 65 (80%) |
| 8/36/7 weave | Elemental | 150 ms | 140 / 206 (40%) | 158 / 126 (56%) |
| 8/36/7 weave | Elemental | 10 ms | 142 / 208 (41%) | 158 / 127 (55%) |
| 8/36/7 weave | Enhancement | 150 ms | 254 / 103 (71%) | 265 / 77 (78%) |
| 8/36/7 weave | Enhancement | 10 ms | 257 / 104 (71%) | 267 / 77 (78%) |
| 8/36/7 no hard casts | Elemental | 150 ms | 200 / 194 (51%) | 192 / 129 (60%) |
| 8/36/7 no hard casts | Elemental | 10 ms | 200 / 194 (51%) | 192 / 129 (60%) |
| 8/36/7 no hard casts | Enhancement | 150 ms | 329 / 113 (74%) | 317 / 87 (78%) |
| 8/36/7 no hard casts | Enhancement | 10 ms | 328 / 114 (74%) | 316 / 87 (78%) |
| Enh preset 19/32/0 | Enhancement | 150 ms | 321 / 118 (73%) | 310 / 94 (77%) |
| Enh preset 19/32/0 | Enhancement | 10 ms | 320 / 119 (73%) | 310 / 94 (77%) |
| Ele preset 32/0/19 | Elemental | 150 ms | 0 / 400 (0%) | 0 / 326 (0%) |
| Ele preset 32/0/19 | Elemental | 10 ms | 0 / 400 (0%) | 0 / 326 (0%) |

### Run 3: top 5 damage sources, reaction 150 ms

- **29/22 weave, Elemental launch gear, 2m:** Lightning Bolt 82 (24%); Melee swing 76 (22%); Windfury Attack 41 (12%); Lava Burst 33 (10%); Flame Shock (ticks) 29 (8%)
- **29/22 weave, Elemental launch gear, 5m:** Melee swing 89 (34%); Windfury Attack 46 (17%); Lightning Bolt 38 (14%); Searing Totem 25 (10%); Flame Shock (ticks) 17 (6%)
- **29/22 weave + CL, Elemental launch gear, 2m:** Melee swing 76 (24%); Lightning Bolt 48 (15%); Windfury Attack 40 (13%); Lava Burst 32 (10%); Chain Lightning 28 (9%)
- **29/22 weave + CL, Elemental launch gear, 5m:** Melee swing 90 (36%); Windfury Attack 46 (18%); Searing Totem 25 (10%); Lightning Bolt 24 (10%); Stormstrike (weapon strike) 15 (6%)
- **29/22 bolt-chain, Elemental launch gear, 2m:** Lightning Bolt 106 (32%); Melee swing 67 (20%); Windfury Attack 36 (11%); Lava Burst 32 (10%); Flame Shock (ticks) 24 (7%)
- **29/22 bolt-chain, Elemental launch gear, 5m:** Melee swing 86 (33%); Lightning Bolt 48 (19%); Windfury Attack 44 (17%); Searing Totem 25 (10%); Flame Shock (ticks) 14 (6%)
- **8/36/7 weave, Elemental launch gear, 2m:** Lightning Bolt 131 (38%); Melee swing 80 (23%); Windfury Attack 44 (13%); Earth Shock 23 (7%); Searing Totem 23 (6%)
- **8/36/7 weave, Elemental launch gear, 5m:** Melee swing 95 (34%); Lightning Bolt 95 (33%); Windfury Attack 50 (17%); Stormstrike (weapon strike) 13 (5%); Searing Totem 10 (4%)
- **8/36/7 no hard casts, Elemental launch gear, 2m:** Melee swing 120 (30%); Earth Shock 69 (18%); Windfury Attack 61 (16%); Lightning Bolt 49 (12%); Flame Shock (ticks) 31 (8%)
- **8/36/7 no hard casts, Elemental launch gear, 5m:** Melee swing 116 (36%); Windfury Attack 59 (18%); Lightning Bolt 49 (15%); Earth Shock 27 (8%); Searing Totem 22 (7%)
- **Ele preset 32/0/19, Elemental launch gear, 2m:** Lightning Bolt rank 4 178 (44%); Lava Burst 62 (16%); Lightning Bolt 39 (10%); Flame Shock (ticks) 38 (10%); Searing Totem 27 (7%)
- **Ele preset 32/0/19, Elemental launch gear, 5m:** Lightning Bolt rank 4 208 (64%); Lava Burst 36 (11%); Flame Shock (ticks) 29 (9%); Searing Totem 14 (4%); Flame Shock (hit) 12 (4%)
- **29/22 weave, Enhancement launch gear, main hand only, 2m:** Melee swing 130 (37%); Windfury Attack 83 (24%); Stormstrike (weapon strike) 39 (11%); Lightning Bolt 32 (9%); Searing Totem 23 (7%)
- **29/22 weave, Enhancement launch gear, main hand only, 5m:** Melee swing 140 (42%); Windfury Attack 87 (26%); Stormstrike (weapon strike) 37 (11%); Searing Totem 22 (7%); Lightning Bolt 19 (6%)
- **29/22 weave + CL, Enhancement launch gear, main hand only, 2m:** Melee swing 132 (38%); Windfury Attack 84 (24%); Stormstrike (weapon strike) 39 (11%); Searing Totem 22 (6%); Lightning Bolt 20 (6%)
- **29/22 weave + CL, Enhancement launch gear, main hand only, 5m:** Melee swing 141 (43%); Windfury Attack 87 (27%); Stormstrike (weapon strike) 37 (11%); Searing Totem 21 (6%); Lightning Bolt 14 (4%)
- **29/22 bolt-chain, Enhancement launch gear, main hand only, 2m:** Melee swing 122 (37%); Windfury Attack 77 (23%); Lightning Bolt 42 (13%); Stormstrike (weapon strike) 35 (11%); Searing Totem 23 (7%)
- **29/22 bolt-chain, Enhancement launch gear, main hand only, 5m:** Melee swing 136 (42%); Windfury Attack 84 (26%); Stormstrike (weapon strike) 35 (11%); Lightning Bolt 24 (7%); Searing Totem 21 (7%)
- **8/36/7 weave, Enhancement launch gear, main hand only, 2m:** Melee swing 134 (38%); Windfury Attack 84 (24%); Lightning Bolt 64 (18%); Stormstrike (weapon strike) 36 (10%); Searing Totem 13 (4%)
- **8/36/7 weave, Enhancement launch gear, main hand only, 5m:** Melee swing 146 (43%); Windfury Attack 88 (26%); Lightning Bolt 52 (15%); Stormstrike (weapon strike) 32 (9%); Searing Totem 9 (3%)
- **8/36/7 no hard casts, Enhancement launch gear, main hand only, 2m:** Melee swing 174 (39%); Windfury Attack 107 (24%); Stormstrike (weapon strike) 48 (11%); Lightning Bolt 34 (8%); Earth Shock 23 (5%)
- **8/36/7 no hard casts, Enhancement launch gear, main hand only, 5m:** Melee swing 170 (42%); Windfury Attack 104 (26%); Stormstrike (weapon strike) 43 (11%); Lightning Bolt 34 (8%); Searing Totem 20 (5%)
- **Enh preset 19/32/0, Enhancement launch gear, main hand only, 2m:** Melee swing 168 (38%); Windfury Attack 104 (24%); Earth Shock 51 (12%); Stormstrike (weapon strike) 48 (11%); Lightning Bolt 32 (7%)
- **Enh preset 19/32/0, Enhancement launch gear, main hand only, 5m:** Melee swing 165 (41%); Windfury Attack 101 (25%); Stormstrike (weapon strike) 44 (11%); Lightning Bolt 32 (8%); Earth Shock 31 (8%)

### Run 3 sanity: casts, swings, out of mana (reaction 150 ms)

| Build | Gear | Fight | LB+LvB+CL casts/min | MH swings/min | OOM s | DPS ± stderr |
|---|---|---|---:|---:|---:|---:|
| 29/22 weave | Elemental | 2m | 7.5 | 35.7 | 58 | 345 ± 0.3 |
| 29/22 weave | Elemental | 5m | 3.4 | 42.7 | 225 | 264 ± 0.1 |
| 29/22 weave | Enhancement | 2m | 4.9 | 20.7 | 66 | 349 ± 0.3 |
| 29/22 weave | Enhancement | 5m | 2.7 | 22.5 | 207 | 328 ± 0.2 |
| 29/22 weave + CL | Elemental | 2m | 8.0 | 36.0 | 65 | 315 ± 0.2 |
| 29/22 weave + CL | Elemental | 5m | 3.6 | 42.8 | 232 | 252 ± 0.1 |
| 29/22 weave + CL | Enhancement | 2m | 4.6 | 21.2 | 70 | 345 ± 0.3 |
| 29/22 weave + CL | Enhancement | 5m | 2.5 | 22.7 | 211 | 326 ± 0.2 |
| 29/22 bolt-chain | Elemental | 2m | 9.7 | 32.0 | 63 | 330 ± 0.2 |
| 29/22 bolt-chain | Elemental | 5m | 4.3 | 41.2 | 230 | 258 ± 0.1 |
| 29/22 bolt-chain | Enhancement | 2m | 6.2 | 19.6 | 67 | 332 ± 0.3 |
| 29/22 bolt-chain | Enhancement | 5m | 3.3 | 22.0 | 208 | 320 ± 0.2 |
| 8/36/7 weave | Elemental | 2m | 9.5 | 35.9 | 48 | 346 ± 0.2 |
| 8/36/7 weave | Elemental | 5m | 6.6 | 42.9 | 216 | 284 ± 0.1 |
| 8/36/7 weave | Enhancement | 2m | 7.9 | 21.2 | 61 | 357 ± 0.4 |
| 8/36/7 weave | Enhancement | 5m | 6.3 | 23.1 | 205 | 342 ± 0.2 |
| 8/36/7 no hard casts | Elemental | 2m | 3.1 | 53.5 | 3 | 394 ± 0.2 |
| 8/36/7 no hard casts | Elemental | 5m | 3.2 | 52.5 | 174 | 321 ± 0.1 |
| 8/36/7 no hard casts | Enhancement | 2m | 3.8 | 27.4 | 30 | 442 ± 0.4 |
| 8/36/7 no hard casts | Enhancement | 5m | 3.8 | 26.9 | 160 | 403 ± 0.2 |
| Enh preset 19/32/0 | Enhancement | 2m | 3.8 | 27.2 | 60 | 439 ± 0.4 |
| Enh preset 19/32/0 | Enhancement | 5m | 3.8 | 26.8 | 218 | 404 ± 0.2 |
| Ele preset 32/0/19 | Elemental | 2m | 8.8 | 0.0 | 4 | 400 ± 0.2 |
| Ele preset 32/0/19 | Elemental | 5m | 3.4 | 0.0 | 51 | 326 ± 0.2 |

### Run 3 reference: the Enhancement preset as the repo ships it, dual-wielding

| Reaction | 2m | 5m | Physical share 2m / 5m |
|---|---:|---:|---|
| 150 ms | 490 | 456 | 73% / 77% |
| 10 ms | 491 | 457 | 73% / 76% |

# Run 2: SoD Phase 8 BiS with set bonuses (2026-10-27, rerun with the fast-weapon fix)

Season of Discovery set bonuses, Chain Lightning and Maelstrom-at-10 rotations, and the physical/spell split.
194 rows. The standard error of every DPS figure is at most 1.2.

- **Dropped from run 2:** the old-swing-rule reruns and the presets on their own launch gear. Neither involves the
  set bonuses. Run 3 covers the launch gear.
- **Agreement with run 1:** with the bonuses off, run 2 matches run 1 on the rows they share, except the 8/36/7
  Enhancement-set rows the fast-weapon fix moved.

## Run 2 setup

- **Gear:** Season of Discovery Phase 8 best-in-slot (`testdata/`), as custom items. `sod_gear_test.go` has the stat
  conversion and the item effects modeled.
- **Set bonuses:** modeled in `sod_sets_test.go`, worded as Wowhead's item-set pages 1951, 1952 and 1889 have them.
  "Set bonuses on" means the set pieces name their sets.
- **Character and buffs:** Windshaper Skyborne; both presets' buffs and debuffs together; the consumables of the preset
  that matches the set.

### Set bonuses modeled

Each bonus applies to any build wearing the set, so it does nothing without the talent or spell.

- **Soulcrusher's Storm 2 (caster set):** when Lava Burst, Chain Lightning or Lightning Bolt lands on a target with
  your Flame Shock, it deals one extra Flame Shock tick (`Dot.TickOnce`).
  - **Casts only, not overloads.** Forever's overloads are separate proc spells, and the bonus names the three spells.
  - **Every rank counts.** The Elemental preset spams Lightning Bolt rank 4, and each of those bolts pulses.
  - **The pulses are booked under "Flame Shock (ticks)"**, so they get Earthshatterer's Storm's +40%.
- **Soulcrusher's Storm 4:**
  - +10% Overload chance, through a new `Shaman.BonusOverloadChance` that `GetOverloadChance` adds. It is nothing
    without the Lightning Overload talent.
  - Lava Burst damage +(your spell crit chance)%, from the character sheet.
  - Each Lightning Bolt or Chain Lightning that deals damage (casts, not overloads) gives the next Lava Burst within
    15 s +10%, stacking to 5.
  - The two Lava Burst bonuses add together into one multiplier.
- **Soulcrusher's Storm 6:** Chain Lightning +35% per target under 3, so +70% on one target. Overloads included,
  since they are Chain Lightning damage too.
- **The Earthshatterer's Storm 2 (caster set):** Flame Shock periodic damage +40%.
- **Soulcrusher's Rage 6 (Enhancement set):** affects only 8/36/7's Maelstrom Weapon.
  - Maelstrom stacks to 10, and each gain gives 2 charges with a two-hander.
  - A cast with more than 5 charges does +20% per charge above 5.
  - A cast at 10 charges spends them all and casts twice. `TestMaelstromAtTenCastsTwice` checks this: two hits per cast.
  - Forever's −20% cast time and cost per charge keeps scaling past 5; the engine floors both at zero.
- **Soulcrusher's Rage 4, as a sensitivity only.** Its wording ("Fire Nova Totem: cooldown −60%, damage +200%, mana −60%")
  is read onto Forever's Fire Nova spell, and only the Enhancement preset's APL casts Fire Nova.
- **Not modeled:** Soulcrusher's Rage 2 and Earthshatterer's Rage 2 (they need the Static Shock rune), both
  Earthshatterer 6-pieces (Undead only), and Earthshatterer's Storm 4 (the caster set has 2 pieces).

### Rotations added

- **"+ CL":** Chain Lightning (10605) as a hard cast in the CAST window, between Lava Burst and Lightning Bolt.
- **Fast weapons:** an instant whose GCD can't fit between swings may go any time (the addon's INSTANT-only light).
- **"MW at 10":** spend Maelstrom only at 10 charges. At 5 to 9 charges the build casts no bolt; below 5 a hard bolt
  is still allowed in the window. It only exists with Soulcrusher's Rage 6.

### How damage is split into physical and spell

- **Rule:** by the school of each damage event in the engine's per-action metrics. Only plain Physical counts as
  physical; every other school counts as spell. Each row's physical + spell is checked to equal the sim's DPS within 0.5%.
- **Physical:** melee swings, Windfury Attack, Stormstrike's weapon strike, and extra attacks such as Signet of the
  Earthshatterer's. Extra attacks are swings, so they are booked as "Melee swing".
- **Spell:** everything else. That includes Searing Totem (fire, cast by the shaman), Dragonbreath Chili (a fire proc
  from the Enhancement consumables), and the Storm 2P pulses (booked under "Flame Shock (ticks)").
- **Not present:** Mirage's arcane proc isn't modeled, and Flametongue isn't used (every build runs Windfury).

## Run 2 findings

- **The gap barely moves with the set bonuses.** At 2 minutes the best 29/22 in the caster set goes from 59% to 60% of
  the best 8/36/7 in the Enhancement set. At 5 minutes it goes from 46% to 44%.
  - The 29/22 gains about 140 DPS at 2 minutes (753 → 894) and 55 at 5 minutes. Its gains are mana-bound.
  - 8/36/7 gains about 190 (1286 → 1480), all of it from Soulcrusher's Rage 6.
- **With the bonuses, full Elemental beats the 29/22 in caster gear.** The Elemental preset (32/0/19, no melee) does
  973 / 720 against the best 29/22's 894 / 602. Without the bonuses the 29/22 led, 753 / 547 to 693 / 519. The Storm
  set rewards exactly what Elemental casts: Flame Shock, Lava Burst and Chain Lightning, and a pulse on every bolt.
- **Chain Lightning only helps the 29/22 with the bonuses on.** It adds +16 at 2 minutes to the weave in caster gear;
  at 5 minutes it's a tie. With the bonuses off it loses, and in bolt-chain it loses even with them on.
- **Spending Maelstrom at 10 beats spending at 5** in every Enhancement-set case, by 7% to 8%.
- **The 29/22 in caster gear is a third physical at 2 minutes (34%), and over half at 5 (55%)**, once mana runs out and
  it mostly swings. 8/36/7 in Enhancement gear is 75% to 83% physical; the best variant (Maelstrom at 10) is 75% at 2
  minutes and 77% at 5.
- **Haste is the biggest assumption.** Taking the gear's haste off costs the top Enhancement builds 15% to 18%, and the
  caster builds 1% to 14%.

### DPS, reaction 150 ms: set bonuses on (off)

| Build | Caster 2m | Caster 5m | Enh 2m | Enh 5m |
|---|---:|---:|---:|---:|
| 29/22 weave | **878** (753) | **601** (547) | **931** (931) | **941** (941) |
| 29/22 weave + CL | **894** (726) | **602** (532) | **940** (940) | **945** (945) |
| 29/22 bolt-chain | **856** (729) | **585** (532) | **853** (853) | **908** (908) |
| 29/22 bolt-chain + CL | **848** (694) | **582** (518) | **863** (863) | **914** (914) |
| 8/36/7 weave | **762** (712) | **596** (574) | **1031** (946) | **1040** (961) |
| 8/36/7 weave, MW at 10 | — | — | **1105** (—) | **1117** (—) |
| 8/36/7 no hard casts | **706** (680) | **679** (656) | **1374** (1286) | **1281** (1188) |
| 8/36/7 no hard casts, MW at 10 | — | — | **1480** (—) | **1381** (—) |
| Enh preset 19/32/0 | — | — | **1308** (1222) | **1238** (1156) |
| Ele preset 32/0/19 | **973** (693) | **720** (519) | — | — |
| Enh preset + Rage 4P on Fire Nova (sensitivity) | — | — | **1428** | **1239** |

### DPS, reaction 10 ms: set bonuses on (off)

| Build | Caster 2m | Caster 5m | Enh 2m | Enh 5m |
|---|---:|---:|---:|---:|
| 29/22 weave | **896** (763) | **607** (550) | **944** (944) | **946** (946) |
| 29/22 weave + CL | **902** (734) | **607** (537) | **947** (947) | **947** (947) |
| 29/22 bolt-chain | **884** (744) | **598** (539) | **863** (863) | **912** (912) |
| 29/22 bolt-chain + CL | **882** (704) | **597** (523) | **878** (878) | **920** (920) |
| 8/36/7 weave | **783** (729) | **601** (579) | **1035** (954) | **1042** (964) |
| 8/36/7 weave, MW at 10 | — | — | **1105** (—) | **1120** (—) |
| 8/36/7 no hard casts | **706** (681) | **680** (656) | **1374** (1288) | **1280** (1187) |
| 8/36/7 no hard casts, MW at 10 | — | — | **1480** (—) | **1382** (—) |
| Enh preset 19/32/0 | — | — | **1306** (1223) | **1238** (1156) |
| Ele preset 32/0/19 | **973** (693) | **720** (519) | — | — |
| Enh preset + Rage 4P on Fire Nova (sensitivity) | — | — | **1427** | **1239** |

### Rotation choices (reaction 150 ms)

| Choice | Gear | Bonuses | 2m | 5m | Better |
|---|---|---|---:|---:|---|
| 29/22 weave vs CL | caster | on | 878 vs 894 | 601 vs 602 | 29/22 weave + CL |
| 29/22 weave vs CL | caster | off | 753 vs 726 | 547 vs 532 | 29/22 weave |
| 29/22 weave vs CL | Enh | on | 931 vs 940 | 941 vs 945 | 29/22 weave + CL |
| 29/22 weave vs CL | Enh | off | 931 vs 940 | 941 vs 945 | 29/22 weave + CL |
| 29/22 bolt-chain vs CL | caster | on | 856 vs 848 | 585 vs 582 | 29/22 bolt-chain |
| 29/22 bolt-chain vs CL | caster | off | 729 vs 694 | 532 vs 518 | 29/22 bolt-chain |
| 29/22 bolt-chain vs CL | Enh | on | 853 vs 863 | 908 vs 914 | 29/22 bolt-chain + CL |
| 29/22 bolt-chain vs CL | Enh | off | 853 vs 863 | 908 vs 914 | 29/22 bolt-chain + CL |
| 8/36/7 weave vs MW at 10 | Enh | on | 1031 vs 1105 | 1040 vs 1117 | 8/36/7 weave, MW at 10 |
| 8/36/7 no hard casts vs MW at 10 | Enh | on | 1374 vs 1480 | 1281 vs 1381 | 8/36/7 no hard casts, MW at 10 |

### The gap: best 29/22 in the caster set vs best 8/36/7 in the Enhancement set (reaction 150 ms)

| Bonuses | Fight | Best 29/22 caster | Best 8/36/7 Enh | 29/22 as % of 8/36/7 |
|---|---|---:|---:|---:|
| off | 2m | 753 (29/22 weave) | 1286 (8/36/7 no hard casts) | 59% |
| off | 5m | 547 (29/22 weave) | 1188 (8/36/7 no hard casts) | 46% |
| on | 2m | 894 (29/22 weave + CL) | 1480 (8/36/7 no hard casts, MW at 10) | 60% |
| on | 5m | 602 (29/22 weave + CL) | 1381 (8/36/7 no hard casts, MW at 10) | 44% |

### Physical vs spell DPS, set bonuses on

Physical DPS / spell DPS (physical share).

| Build | Gear | Reaction | 2m | 5m |
|---|---|---|---|---|
| 29/22 weave | caster | 150 ms | 299 / 579 (34%) | 328 / 272 (55%) |
| 29/22 weave | caster | 10 ms | 305 / 591 (34%) | 332 / 275 (55%) |
| 29/22 weave | Enh | 150 ms | 764 / 166 (82%) | 845 / 96 (90%) |
| 29/22 weave | Enh | 10 ms | 775 / 168 (82%) | 849 / 97 (90%) |
| 29/22 weave + CL | caster | 150 ms | 310 / 584 (35%) | 333 / 269 (55%) |
| 29/22 weave + CL | caster | 10 ms | 318 / 585 (35%) | 336 / 271 (55%) |
| 29/22 weave + CL | Enh | 150 ms | 789 / 150 (84%) | 855 / 90 (90%) |
| 29/22 weave + CL | Enh | 10 ms | 797 / 150 (84%) | 858 / 90 (90%) |
| 29/22 bolt-chain | caster | 150 ms | 259 / 598 (30%) | 312 / 274 (53%) |
| 29/22 bolt-chain | caster | 10 ms | 267 / 617 (30%) | 315 / 283 (53%) |
| 29/22 bolt-chain | Enh | 150 ms | 687 / 166 (80%) | 812 / 96 (90%) |
| 29/22 bolt-chain | Enh | 10 ms | 696 / 167 (81%) | 816 / 96 (89%) |
| 29/22 bolt-chain + CL | caster | 150 ms | 271 / 577 (32%) | 317 / 266 (54%) |
| 29/22 bolt-chain + CL | caster | 10 ms | 274 / 608 (31%) | 318 / 279 (53%) |
| 29/22 bolt-chain + CL | Enh | 150 ms | 713 / 150 (83%) | 825 / 89 (90%) |
| 29/22 bolt-chain + CL | Enh | 10 ms | 727 / 152 (83%) | 830 / 90 (90%) |
| 8/36/7 weave | caster | 150 ms | 320 / 443 (42%) | 342 / 254 (57%) |
| 8/36/7 weave | caster | 10 ms | 327 / 456 (42%) | 344 / 257 (57%) |
| 8/36/7 weave | Enh | 150 ms | 802 / 229 (78%) | 858 / 182 (82%) |
| 8/36/7 weave | Enh | 10 ms | 806 / 228 (78%) | 860 / 182 (83%) |
| 8/36/7 weave, MW at 10 | Enh | 150 ms | 844 / 261 (76%) | 890 / 227 (80%) |
| 8/36/7 weave, MW at 10 | Enh | 10 ms | 841 / 264 (76%) | 891 / 228 (80%) |
| 8/36/7 no hard casts | caster | 150 ms | 437 / 269 (62%) | 426 / 254 (63%) |
| 8/36/7 no hard casts | caster | 10 ms | 437 / 269 (62%) | 426 / 254 (63%) |
| 8/36/7 no hard casts | Enh | 150 ms | 1102 / 271 (80%) | 1060 / 221 (83%) |
| 8/36/7 no hard casts | Enh | 10 ms | 1103 / 271 (80%) | 1059 / 221 (83%) |
| 8/36/7 no hard casts, MW at 10 | Enh | 150 ms | 1105 / 375 (75%) | 1060 / 321 (77%) |
| 8/36/7 no hard casts, MW at 10 | Enh | 10 ms | 1105 / 375 (75%) | 1060 / 322 (77%) |
| Enh preset 19/32/0 | Enh | 150 ms | 1036 / 271 (79%) | 1009 / 229 (82%) |
| Enh preset 19/32/0 | Enh | 10 ms | 1036 / 270 (79%) | 1010 / 228 (82%) |
| Ele preset 32/0/19 | caster | 150 ms | 0 / 973 (0%) | 0 / 720 (0%) |
| Ele preset 32/0/19 | caster | 10 ms | 0 / 973 (0%) | 0 / 720 (0%) |
| Enh preset + Rage 4P on Fire Nova | Enh | 150 ms | 1036 / 392 (73%) | 1008 / 231 (81%) |
| Enh preset + Rage 4P on Fire Nova | Enh | 10 ms | 1036 / 391 (73%) | 1008 / 231 (81%) |

### Physical vs spell DPS, set bonuses off

Physical DPS / spell DPS (physical share).

| Build | Gear | Reaction | 2m | 5m |
|---|---|---|---|---|
| 29/22 weave | caster | 150 ms | 299 / 454 (40%) | 328 / 218 (60%) |
| 29/22 weave | caster | 10 ms | 306 / 457 (40%) | 332 / 218 (60%) |
| 29/22 weave | Enh | 150 ms | 764 / 166 (82%) | 845 / 96 (90%) |
| 29/22 weave | Enh | 10 ms | 775 / 168 (82%) | 849 / 97 (90%) |
| 29/22 weave + CL | caster | 150 ms | 309 / 417 (43%) | 332 / 199 (62%) |
| 29/22 weave + CL | caster | 10 ms | 318 / 417 (43%) | 336 / 200 (63%) |
| 29/22 weave + CL | Enh | 150 ms | 789 / 150 (84%) | 855 / 90 (90%) |
| 29/22 weave + CL | Enh | 10 ms | 797 / 150 (84%) | 858 / 90 (90%) |
| 29/22 bolt-chain | caster | 150 ms | 260 / 470 (36%) | 312 / 220 (59%) |
| 29/22 bolt-chain | caster | 10 ms | 268 / 476 (36%) | 316 / 224 (58%) |
| 29/22 bolt-chain | Enh | 150 ms | 687 / 166 (80%) | 812 / 96 (90%) |
| 29/22 bolt-chain | Enh | 10 ms | 696 / 167 (81%) | 816 / 96 (89%) |
| 29/22 bolt-chain + CL | caster | 150 ms | 271 / 423 (39%) | 317 / 201 (61%) |
| 29/22 bolt-chain + CL | caster | 10 ms | 274 / 429 (39%) | 319 / 205 (61%) |
| 29/22 bolt-chain + CL | Enh | 150 ms | 713 / 150 (83%) | 825 / 89 (90%) |
| 29/22 bolt-chain + CL | Enh | 10 ms | 727 / 152 (83%) | 830 / 90 (90%) |
| 8/36/7 weave | caster | 150 ms | 320 / 392 (45%) | 342 / 232 (60%) |
| 8/36/7 weave | caster | 10 ms | 326 / 403 (45%) | 345 / 234 (60%) |
| 8/36/7 weave | Enh | 150 ms | 771 / 175 (82%) | 840 / 121 (87%) |
| 8/36/7 weave | Enh | 10 ms | 778 / 176 (82%) | 842 / 122 (87%) |
| 8/36/7 no hard casts | caster | 150 ms | 438 / 243 (64%) | 426 / 229 (65%) |
| 8/36/7 no hard casts | caster | 10 ms | 438 / 242 (64%) | 426 / 229 (65%) |
| 8/36/7 no hard casts | Enh | 150 ms | 1098 / 188 (85%) | 1050 / 138 (88%) |
| 8/36/7 no hard casts | Enh | 10 ms | 1100 / 189 (85%) | 1049 / 138 (88%) |
| Enh preset 19/32/0 | Enh | 150 ms | 1037 / 186 (85%) | 1009 / 147 (87%) |
| Enh preset 19/32/0 | Enh | 10 ms | 1037 / 186 (85%) | 1009 / 147 (87%) |
| Ele preset 32/0/19 | caster | 150 ms | 0 / 693 (0%) | 0 / 519 (0%) |
| Ele preset 32/0/19 | caster | 10 ms | 0 / 693 (0%) | 0 / 519 (0%) |

### Top 5 damage sources, 2 min, set bonuses on, reaction 150 ms

- **29/22 weave, caster set:** Lightning Bolt 167 (19%); Lava Burst 161 (18%); Melee swing 132 (15%); Flame Shock (ticks) 108 (12%); Windfury Attack 97 (11%)
- **29/22 weave + CL, caster set:** Lava Burst 150 (17%); Melee swing 140 (16%); Chain Lightning 112 (12%); Flame Shock (ticks) 111 (12%); Windfury Attack 99 (11%)
- **29/22 bolt-chain, caster set:** Lightning Bolt 205 (24%); Lava Burst 135 (16%); Melee swing 121 (14%); Flame Shock (ticks) 110 (13%); Windfury Attack 83 (10%)
- **29/22 bolt-chain + CL, caster set:** Melee swing 128 (15%); Lava Burst 128 (15%); Lightning Bolt 125 (15%); Flame Shock (ticks) 101 (12%); Chain Lightning 88 (10%)
- **8/36/7 weave, caster set:** Lightning Bolt 250 (33%); Melee swing 143 (19%); Windfury Attack 101 (13%); Flame Shock (ticks) 88 (12%); Stormstrike (weapon strike) 76 (10%)
- **8/36/7 no hard casts, caster set:** Melee swing 221 (31%); Windfury Attack 138 (20%); Lightning Bolt 82 (12%); Stormstrike (weapon strike) 78 (11%); Flame Shock (ticks) 66 (9%)
- **Ele preset 32/0/19, caster set:** Lightning Bolt rank 4 218 (22%); Flame Shock (ticks) 210 (22%); Lightning Bolt 199 (20%); Lava Burst 180 (18%); Chain Lightning 47 (5%)
- **29/22 weave, Enhancement set:** Melee swing 382 (41%); Windfury Attack 243 (26%); Stormstrike (weapon strike) 140 (15%); Lightning Bolt 50 (5%); Searing Totem 27 (3%)
- **29/22 weave + CL, Enhancement set:** Melee swing 398 (42%); Windfury Attack 251 (27%); Stormstrike (weapon strike) 141 (15%); Searing Totem 28 (3%); Lightning Bolt 25 (3%)
- **29/22 bolt-chain, Enhancement set:** Melee swing 349 (41%); Windfury Attack 217 (26%); Stormstrike (weapon strike) 121 (14%); Lightning Bolt 66 (8%); Searing Totem 27 (3%)
- **29/22 bolt-chain + CL, Enhancement set:** Melee swing 365 (42%); Windfury Attack 225 (26%); Stormstrike (weapon strike) 122 (14%); Lightning Bolt 39 (4%); Searing Totem 26 (3%)
- **8/36/7 weave, Enhancement set:** Melee swing 419 (41%); Windfury Attack 250 (24%); Lightning Bolt 154 (15%); Stormstrike (weapon strike) 132 (13%); Searing Totem 24 (2%)
- **8/36/7 weave, MW at 10, Enhancement set:** Melee swing 440 (40%); Windfury Attack 264 (24%); Lightning Bolt 180 (16%); Stormstrike (weapon strike) 140 (13%); Searing Totem 24 (2%)
- **8/36/7 no hard casts, Enhancement set:** Melee swing 598 (44%); Windfury Attack 340 (25%); Stormstrike (weapon strike) 164 (12%); Lightning Bolt 149 (11%); Earth Shock 41 (3%)
- **8/36/7 no hard casts, MW at 10, Enhancement set:** Melee swing 599 (40%); Windfury Attack 340 (23%); Lightning Bolt 249 (17%); Stormstrike (weapon strike) 166 (11%); Earth Shock 43 (3%)
- **Enh preset 19/32/0, Enhancement set:** Melee swing 562 (43%); Windfury Attack 319 (24%); Stormstrike (weapon strike) 156 (12%); Lightning Bolt 146 (11%); Earth Shock 70 (5%)
- **Enh preset 19/32/0, Enhancement set, Rage 4P on Fire Nova:** Melee swing 562 (39%); Windfury Attack 319 (22%); Stormstrike (weapon strike) 156 (11%); Lightning Bolt 152 (11%); Fire Nova 140 (10%)

### Sensitivities, set bonuses on, reaction 150 ms

| Build | Gear | Fight | Base | Haste off | Smashbringer GCD off | Both off |
|---|---|---|---:|---:|---:|---:|
| 29/22 weave | caster | 2m | 878 | 828 | — | — |
| 29/22 weave | caster | 5m | 601 | 589 | — | — |
| 29/22 weave | Enh | 2m | 931 | 812 | 925 | 804 |
| 29/22 weave | Enh | 5m | 941 | 799 | 937 | 794 |
| 29/22 weave + CL | caster | 2m | 894 | 856 | — | — |
| 29/22 weave + CL | caster | 5m | 602 | 586 | — | — |
| 29/22 weave + CL | Enh | 2m | 940 | 818 | 933 | 813 |
| 29/22 weave + CL | Enh | 5m | 945 | 800 | 940 | 798 |
| 29/22 bolt-chain | caster | 2m | 856 | 834 | — | — |
| 29/22 bolt-chain | caster | 5m | 585 | 577 | — | — |
| 29/22 bolt-chain | Enh | 2m | 853 | 772 | 835 | 754 |
| 29/22 bolt-chain | Enh | 5m | 908 | 782 | 899 | 773 |
| 29/22 bolt-chain + CL | caster | 2m | 848 | 833 | — | — |
| 29/22 bolt-chain + CL | caster | 5m | 582 | 574 | — | — |
| 29/22 bolt-chain + CL | Enh | 2m | 863 | 768 | 850 | 755 |
| 29/22 bolt-chain + CL | Enh | 5m | 914 | 781 | 906 | 773 |
| 8/36/7 weave | caster | 2m | 762 | 702 | — | — |
| 8/36/7 weave | caster | 5m | 596 | 578 | — | — |
| 8/36/7 weave | Enh | 2m | 1031 | 899 | 1058 | 900 |
| 8/36/7 weave | Enh | 5m | 1040 | 882 | 1060 | 884 |
| 8/36/7 weave, MW at 10 | Enh | 2m | 1105 | 960 | 1138 | 960 |
| 8/36/7 weave, MW at 10 | Enh | 5m | 1117 | 948 | 1123 | 945 |
| 8/36/7 no hard casts | caster | 2m | 706 | 697 | — | — |
| 8/36/7 no hard casts | caster | 5m | 679 | 674 | — | — |
| 8/36/7 no hard casts | Enh | 2m | 1374 | 1153 | 1364 | 1152 |
| 8/36/7 no hard casts | Enh | 5m | 1281 | 1057 | 1284 | 1061 |
| 8/36/7 no hard casts, MW at 10 | Enh | 2m | 1480 | 1232 | 1465 | 1231 |
| 8/36/7 no hard casts, MW at 10 | Enh | 5m | 1381 | 1137 | 1382 | 1138 |
| Enh preset 19/32/0 | Enh | 2m | 1308 | 1110 | 1309 | 1111 |
| Enh preset 19/32/0 | Enh | 5m | 1238 | 1037 | 1244 | 1043 |
| Ele preset 32/0/19 | caster | 2m | 973 | 841 | — | — |
| Ele preset 32/0/19 | caster | 5m | 720 | 662 | — | — |

### Sanity: casts, swings, out of mana (set bonuses on, reaction 150 ms)

| Build | Gear | Fight | LB+LvB+CL casts/min | MH swings/min | OOM s | DPS ± stderr |
|---|---|---|---:|---:|---:|---:|
| 29/22 weave | caster | 2m | 12.7 | 13.6 | 5 | 878 ± 0.6 |
| 29/22 weave | caster | 5m | 5.6 | 17.2 | 169 | 601 ± 0.3 |
| 29/22 weave | Enh | 2m | 6.8 | 19.2 | 52 | 931 ± 0.9 |
| 29/22 weave | Enh | 5m | 3.1 | 22.7 | 197 | 941 ± 0.6 |
| 29/22 weave + CL | caster | 2m | 11.8 | 14.5 | 16 | 894 ± 0.6 |
| 29/22 weave + CL | caster | 5m | 5.2 | 17.6 | 180 | 602 ± 0.3 |
| 29/22 weave + CL | Enh | 2m | 6.3 | 20.0 | 58 | 940 ± 0.9 |
| 29/22 weave + CL | Enh | 5m | 2.9 | 23.0 | 203 | 945 ± 0.6 |
| 29/22 bolt-chain | caster | 2m | 14.1 | 12.6 | 32 | 856 ± 0.6 |
| 29/22 bolt-chain | caster | 5m | 6.1 | 16.8 | 196 | 585 ± 0.3 |
| 29/22 bolt-chain | Enh | 2m | 8.2 | 17.7 | 57 | 853 ± 0.8 |
| 29/22 bolt-chain | Enh | 5m | 3.7 | 22.0 | 202 | 908 ± 0.6 |
| 29/22 bolt-chain + CL | caster | 2m | 13.2 | 13.5 | 40 | 848 ± 0.6 |
| 29/22 bolt-chain + CL | caster | 5m | 5.8 | 17.2 | 204 | 582 ± 0.3 |
| 29/22 bolt-chain + CL | Enh | 2m | 7.8 | 18.5 | 61 | 863 ± 0.8 |
| 29/22 bolt-chain + CL | Enh | 5m | 3.6 | 22.4 | 206 | 914 ± 0.6 |
| 8/36/7 weave | caster | 2m | 14.2 | 14.3 | 0 | 762 ± 0.5 |
| 8/36/7 weave | caster | 5m | 9.0 | 17.9 | 147 | 596 ± 0.3 |
| 8/36/7 weave | Enh | 2m | 13.9 | 20.4 | 28 | 1031 ± 1.1 |
| 8/36/7 weave | Enh | 5m | 11.8 | 23.4 | 167 | 1040 ± 0.7 |
| 8/36/7 weave, MW at 10 | Enh | 2m | 11.8 | 21.4 | 24 | 1105 ± 1.2 |
| 8/36/7 weave, MW at 10 | Enh | 5m | 9.5 | 24.1 | 160 | 1117 ± 0.8 |
| 8/36/7 no hard casts | caster | 2m | 4.0 | 22.5 | 0 | 706 ± 0.6 |
| 8/36/7 no hard casts | caster | 5m | 4.0 | 22.1 | 42 | 679 ± 0.4 |
| 8/36/7 no hard casts | Enh | 2m | 9.4 | 28.8 | 0 | 1374 ± 1.1 |
| 8/36/7 no hard casts | Enh | 5m | 9.2 | 28.2 | 134 | 1281 ± 0.7 |
| 8/36/7 no hard casts, MW at 10 | Enh | 2m | 5.8 | 28.8 | 1 | 1480 ± 1.2 |
| 8/36/7 no hard casts, MW at 10 | Enh | 5m | 5.8 | 28.2 | 138 | 1381 ± 0.8 |
| Enh preset 19/32/0 | Enh | 2m | 9.0 | 28.5 | 45 | 1308 ± 1.1 |
| Enh preset 19/32/0 | Enh | 5m | 9.1 | 28.0 | 176 | 1238 ± 0.7 |
| Ele preset 32/0/19 | caster | 2m | 18.4 | 0.0 | 8 | 973 ± 0.5 |
| Ele preset 32/0/19 | caster | 5m | 4.5 | 0.0 | 75 | 720 ± 0.4 |

## Run 1 reference: old swing rule (no set bonuses, reaction 150 ms)

| Build | Gear | 2m old | 2m reset | Δ | 5m old | 5m reset | Δ |
|---|---|---:|---:|---:|---:|---:|---:|
| 29/22 weave | Caster set (Mirage) | 840 | 753 | -10.3% | 581 | 547 | -5.9% |
| 29/22 weave | Enh set (Smashbringer) | 991 | 931 | -6.1% | 974 | 941 | -3.4% |
| 29/22 bolt-chain | Caster set (Mirage) | 840 | 729 | -13.1% | 581 | 532 | -8.4% |
| 29/22 bolt-chain | Enh set (Smashbringer) | 960 | 853 | -11.1% | 962 | 908 | -5.6% |
| 8/36/7 weave | Caster set (Mirage) | 842 | 712 | -15.5% | 630 | 574 | -8.9% |
| 8/36/7 weave | Enh set (Smashbringer) | 983 | 918 | -6.6% | 1013 | 944 | -6.9% |
| 8/36/7 no hard casts | Caster set (Mirage) | 680 | 680 | +0.0% | 656 | 656 | +0.0% |
| 8/36/7 no hard casts | Enh set (Smashbringer) | 1270 | 1270 | +0.0% | 1181 | 1181 | +0.0% |
| Enh preset 19/32/0 | Enh set (Smashbringer) | 1222 | 1222 | +0.0% | 1156 | 1156 | +0.0% |
| Ele preset 32/0/19 | Caster set (Mirage) | 693 | 693 | +0.0% | 519 | 519 | +0.0% |


## Run 1 reference: presets on their own launch gear (repo presets.ts, 200 ms reaction)

| Build | 2m reset | 2m old | 5m reset | 5m old |
|---|---:|---:|---:|---:|
| Enh preset 19/32/0 | 473 | 473 | 442 | 442 |
| Ele preset 32/0/19 | 394 | 394 | 324 | 324 |

