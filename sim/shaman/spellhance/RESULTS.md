# Spellhance comparison results

Run 2 (2026-10-27): Season of Discovery set bonuses, Chain Lightning and Maelstrom-at-10 rotations, and the
physical/spell split. 194 runs × 10,000 iterations, seed 7. Single target, `MakeSingleTargetEncounter`, no duration
variation. Every run uses the measured swing rule. Engine: `mythicsim/oct21-release` (62b7445d1c) plus this branch.
The standard error of every DPS figure is at most 1.2.

- **Files:** `results.csv` has every run; `sources.csv` has every run's damage by source and school.
- **Run 1:** `results-run1.csv`, no set bonuses, with the old-swing-rule reference rows (summarized at the end).
- **Run 1 rows reproduce exactly:** with the bonuses off, run 2 matches run 1 on all 40 rows they share.
- **Dropped from run 2:** the old-swing-rule reruns and the presets on their own launch gear. Neither involves the
  set bonuses, so run 1's numbers stand for both.

Rerun with:

    SPELLHANCE_OUT=<dir> SPELLHANCE_ITERATIONS=10000 go test --tags=with_db ./sim/shaman/spellhance/ -run TestSpellhanceMatrix -timeout 60m

## Setup

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
- **"MW at 10":** spend Maelstrom only at 10 charges. At 5 to 9 charges the build casts no bolt; below 5 a hard bolt
  is still allowed in the window. It only exists with Soulcrusher's Rage 6.

### How damage is split into physical and spell

- **Rule:** by the school of each damage event in the engine's per-action metrics. Only plain Physical counts as
  physical; every other school counts as spell. Each row's physical + spell is checked to equal the sim's DPS within 0.5%.
- **Physical:** melee swings, Windfury Attack, Stormstrike's weapon strike, and extra attacks such as Signet of the
  Earthshatterer's. Extra attacks are swings, so they are booked as "Melee swing".
- **Spell:** everything else. That includes Searing Totem (fire, cast by the shaman), Dragonbreath Chili (a fire proc
  from the Enhancement consumables), and the Storm 2P pulses.
- **Not present:** Mirage's arcane proc isn't modeled, and Flametongue isn't used (every build runs Windfury).

## Findings

- **The gap barely moves.** At 2 minutes the best 29/22 in the caster set goes from 59% to 61% of the best 8/36/7 in
  the Enhancement set, a slight narrowing. At 5 minutes it goes from 46% to 44%, a slight widening.
  - The 29/22 gains about 140 DPS at 2 minutes (753 → 894) and 55 at 5 minutes. Its gains are mana-bound.
  - 8/36/7 gains about 190 (1270 → 1458), all of it from Soulcrusher's Rage 6.
- **With the bonuses, full Elemental beats the 29/22 in caster gear.** The Elemental preset (32/0/19, no melee) does
  973 / 720 against the best 29/22's 894 / 602. Without the bonuses the 29/22 led, 753 / 547 to 693 / 519. The Storm
  set rewards exactly what Elemental casts: Flame Shock, Lava Burst and Chain Lightning, and a pulse on every bolt.
- **Chain Lightning only helps the 29/22 with the bonuses on.** It adds +16 at 2 minutes to the weave in caster gear;
  at 5 minutes it's a tie. With the bonuses off, Chain Lightning loses. In bolt-chain it loses even with the bonuses on.
- **Spending Maelstrom at 10 beats spending at 5** in every Enhancement-set case, by 7% to 8%.
- **The 29/22 in caster gear is a third physical at 2 minutes (34%), and over half at 5 minutes (55%)**, once mana
  runs out and it mostly swings. 8/36/7 in Enhancement gear is 74% to 83% physical; the best variant (Maelstrom at 10) is 74% at
  2 minutes and 77% at 5.
- **Haste is still the biggest assumption.** Taking the gear's haste off costs the top Enhancement builds 15% to 17%,
  and the caster builds 1% to 14%.

### DPS, reaction 150 ms: set bonuses on (off)

| Build | Caster 2m | Caster 5m | Enh 2m | Enh 5m |
|---|---:|---:|---:|---:|
| 29/22 weave | **878** (753) | **601** (547) | **931** (931) | **941** (941) |
| 29/22 weave + CL | **894** (726) | **602** (532) | **940** (940) | **945** (945) |
| 29/22 bolt-chain | **856** (729) | **585** (532) | **853** (853) | **908** (908) |
| 29/22 bolt-chain + CL | **848** (694) | **582** (518) | **863** (863) | **914** (914) |
| 8/36/7 weave | **762** (712) | **596** (574) | **1002** (918) | **1024** (944) |
| 8/36/7 weave, MW at 10 | — | — | **1068** (—) | **1095** (—) |
| 8/36/7 no hard casts | **706** (680) | **679** (656) | **1354** (1270) | **1271** (1181) |
| 8/36/7 no hard casts, MW at 10 | — | — | **1458** (—) | **1371** (—) |
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
| 8/36/7 weave | **783** (729) | **601** (579) | **1009** (927) | **1026** (947) |
| 8/36/7 weave, MW at 10 | — | — | **1071** (—) | **1098** (—) |
| 8/36/7 no hard casts | **706** (681) | **680** (656) | **1355** (1271) | **1271** (1180) |
| 8/36/7 no hard casts, MW at 10 | — | — | **1460** (—) | **1372** (—) |
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
| 8/36/7 weave vs MW at 10 | Enh | on | 1002 vs 1068 | 1024 vs 1095 | 8/36/7 weave, MW at 10 |
| 8/36/7 no hard casts vs MW at 10 | Enh | on | 1354 vs 1458 | 1271 vs 1371 | 8/36/7 no hard casts, MW at 10 |

### The gap: best 29/22 in the caster set vs best 8/36/7 in the Enhancement set (reaction 150 ms)

| Bonuses | Fight | Best 29/22 caster | Best 8/36/7 Enh | 29/22 as % of 8/36/7 |
|---|---|---:|---:|---:|
| off | 2m | 753 (29/22 weave) | 1270 (8/36/7 no hard casts) | 59% |
| off | 5m | 547 (29/22 weave) | 1181 (8/36/7 no hard casts) | 46% |
| on | 2m | 894 (29/22 weave + CL) | 1458 (8/36/7 no hard casts, MW at 10) | 61% |
| on | 5m | 602 (29/22 weave + CL) | 1371 (8/36/7 no hard casts, MW at 10) | 44% |

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
| 8/36/7 weave | Enh | 150 ms | 777 / 225 (78%) | 844 / 180 (82%) |
| 8/36/7 weave | Enh | 10 ms | 784 / 225 (78%) | 846 / 180 (82%) |
| 8/36/7 weave, MW at 10 | Enh | 150 ms | 812 / 256 (76%) | 869 / 226 (79%) |
| 8/36/7 weave, MW at 10 | Enh | 10 ms | 814 / 257 (76%) | 872 / 226 (79%) |
| 8/36/7 no hard casts | caster | 150 ms | 437 / 269 (62%) | 426 / 254 (63%) |
| 8/36/7 no hard casts | caster | 10 ms | 437 / 269 (62%) | 426 / 254 (63%) |
| 8/36/7 no hard casts | Enh | 150 ms | 1085 / 269 (80%) | 1051 / 220 (83%) |
| 8/36/7 no hard casts | Enh | 10 ms | 1087 / 269 (80%) | 1050 / 220 (83%) |
| 8/36/7 no hard casts, MW at 10 | Enh | 150 ms | 1086 / 372 (74%) | 1051 / 320 (77%) |
| 8/36/7 no hard casts, MW at 10 | Enh | 10 ms | 1088 / 372 (74%) | 1052 / 320 (77%) |
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
| 8/36/7 weave | Enh | 150 ms | 745 / 173 (81%) | 823 / 121 (87%) |
| 8/36/7 weave | Enh | 10 ms | 753 / 174 (81%) | 825 / 122 (87%) |
| 8/36/7 no hard casts | caster | 150 ms | 438 / 243 (64%) | 426 / 229 (65%) |
| 8/36/7 no hard casts | caster | 10 ms | 438 / 242 (64%) | 426 / 229 (65%) |
| 8/36/7 no hard casts | Enh | 150 ms | 1081 / 189 (85%) | 1042 / 139 (88%) |
| 8/36/7 no hard casts | Enh | 10 ms | 1082 / 189 (85%) | 1040 / 139 (88%) |
| Enh preset 19/32/0 | Enh | 150 ms | 1037 / 186 (85%) | 1009 / 147 (87%) |
| Enh preset 19/32/0 | Enh | 10 ms | 1037 / 186 (85%) | 1009 / 147 (87%) |
| Ele preset 32/0/19 | caster | 150 ms | 0 / 693 (0%) | 0 / 519 (0%) |
| Ele preset 32/0/19 | caster | 10 ms | 0 / 693 (0%) | 0 / 519 (0%) |

### Top 5 damage sources, 2 min, set bonuses on, reaction 150 ms

- **29/22 weave, caster set:** Lightning Bolt 167 (19%); Lava Burst 161 (18%); Melee swing 132 (15%); Flame Shock (ticks, with Storm 2P pulses) 108 (12%); Windfury Attack 97 (11%)
- **29/22 weave + CL, caster set:** Lava Burst 150 (17%); Melee swing 140 (16%); Chain Lightning 112 (12%); Flame Shock (ticks, with Storm 2P pulses) 111 (12%); Windfury Attack 99 (11%)
- **29/22 bolt-chain, caster set:** Lightning Bolt 205 (24%); Lava Burst 135 (16%); Melee swing 121 (14%); Flame Shock (ticks, with Storm 2P pulses) 110 (13%); Windfury Attack 83 (10%)
- **29/22 bolt-chain + CL, caster set:** Melee swing 128 (15%); Lava Burst 128 (15%); Lightning Bolt 125 (15%); Flame Shock (ticks, with Storm 2P pulses) 101 (12%); Chain Lightning 88 (10%)
- **8/36/7 weave, caster set:** Lightning Bolt 250 (33%); Melee swing 143 (19%); Windfury Attack 101 (13%); Flame Shock (ticks, with Storm 2P pulses) 88 (12%); Stormstrike (weapon strike) 76 (10%)
- **8/36/7 no hard casts, caster set:** Melee swing 221 (31%); Windfury Attack 138 (20%); Lightning Bolt 82 (12%); Stormstrike (weapon strike) 78 (11%); Flame Shock (ticks, with Storm 2P pulses) 66 (9%)
- **Ele preset 32/0/19, caster set:** Lightning Bolt rank 4 218 (22%); Flame Shock (ticks, with Storm 2P pulses) 210 (22%); Lightning Bolt 199 (20%); Lava Burst 180 (18%); Chain Lightning 47 (5%)
- **29/22 weave, Enhancement set:** Melee swing 382 (41%); Windfury Attack 243 (26%); Stormstrike (weapon strike) 140 (15%); Lightning Bolt 50 (5%); Searing Totem 27 (3%)
- **29/22 weave + CL, Enhancement set:** Melee swing 398 (42%); Windfury Attack 251 (27%); Stormstrike (weapon strike) 141 (15%); Searing Totem 28 (3%); Lightning Bolt 25 (3%)
- **29/22 bolt-chain, Enhancement set:** Melee swing 349 (41%); Windfury Attack 217 (26%); Stormstrike (weapon strike) 121 (14%); Lightning Bolt 66 (8%); Searing Totem 27 (3%)
- **29/22 bolt-chain + CL, Enhancement set:** Melee swing 365 (42%); Windfury Attack 225 (26%); Stormstrike (weapon strike) 122 (14%); Lightning Bolt 39 (4%); Searing Totem 26 (3%)
- **8/36/7 weave, Enhancement set:** Melee swing 408 (41%); Windfury Attack 247 (25%); Lightning Bolt 151 (15%); Stormstrike (weapon strike) 123 (12%); Searing Totem 24 (2%)
- **8/36/7 weave, MW at 10, Enhancement set:** Melee swing 427 (40%); Windfury Attack 256 (24%); Lightning Bolt 177 (16%); Stormstrike (weapon strike) 128 (12%); Searing Totem 24 (2%)
- **8/36/7 no hard casts, Enhancement set:** Melee swing 598 (44%); Windfury Attack 336 (25%); Stormstrike (weapon strike) 151 (11%); Lightning Bolt 145 (11%); Earth Shock 42 (3%)
- **8/36/7 no hard casts, MW at 10, Enhancement set:** Melee swing 598 (41%); Windfury Attack 336 (23%); Lightning Bolt 244 (17%); Stormstrike (weapon strike) 152 (10%); Earth Shock 44 (3%)
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
| 8/36/7 weave | Enh | 2m | 1002 | 897 | 1008 | 899 |
| 8/36/7 weave | Enh | 5m | 1024 | 882 | 1029 | 884 |
| 8/36/7 weave, MW at 10 | Enh | 2m | 1068 | 956 | 1060 | 953 |
| 8/36/7 weave, MW at 10 | Enh | 5m | 1095 | 947 | 1077 | 942 |
| 8/36/7 no hard casts | caster | 2m | 706 | 697 | — | — |
| 8/36/7 no hard casts | caster | 5m | 679 | 674 | — | — |
| 8/36/7 no hard casts | Enh | 2m | 1354 | 1152 | 1325 | 1151 |
| 8/36/7 no hard casts | Enh | 5m | 1271 | 1058 | 1274 | 1060 |
| 8/36/7 no hard casts, MW at 10 | Enh | 2m | 1458 | 1232 | 1406 | 1232 |
| 8/36/7 no hard casts, MW at 10 | Enh | 5m | 1371 | 1136 | 1366 | 1138 |
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
| 8/36/7 weave | Enh | 2m | 14.2 | 19.9 | 29 | 1002 ± 1.1 |
| 8/36/7 weave | Enh | 5m | 11.9 | 23.2 | 161 | 1024 ± 0.7 |
| 8/36/7 weave, MW at 10 | Enh | 2m | 12.2 | 20.8 | 26 | 1068 ± 1.2 |
| 8/36/7 weave, MW at 10 | Enh | 5m | 9.7 | 23.8 | 155 | 1095 ± 0.8 |
| 8/36/7 no hard casts | caster | 2m | 4.0 | 22.5 | 0 | 706 ± 0.6 |
| 8/36/7 no hard casts | caster | 5m | 4.0 | 22.1 | 42 | 679 ± 0.4 |
| 8/36/7 no hard casts | Enh | 2m | 9.3 | 28.8 | 0 | 1354 ± 1.1 |
| 8/36/7 no hard casts | Enh | 5m | 9.2 | 28.2 | 129 | 1271 ± 0.7 |
| 8/36/7 no hard casts, MW at 10 | Enh | 2m | 5.8 | 28.8 | 1 | 1458 ± 1.2 |
| 8/36/7 no hard casts, MW at 10 | Enh | 5m | 5.7 | 28.2 | 134 | 1371 ± 0.8 |
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

