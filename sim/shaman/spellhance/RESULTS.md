# Spellhance comparison results

10,000 iterations per run, seed 7, single target, Patchwerk-style fight (`MakeSingleTargetEncounter`, no
duration variation), engine `mythicsim/oct21-release` (62b7445d1c) plus this branch. Every row is in
`results.csv`; the standard error of every DPS figure is at most 1.0.

Rerun with:

    SPELLHANCE_OUT=<dir> SPELLHANCE_ITERATIONS=10000 go test --tags=with_db ./sim/shaman/spellhance/ -run TestSpellhanceMatrix -timeout 60m

- **Gear:** Season of Discovery Phase 8 best-in-slot (`testdata/`), as custom items; see `sod_gear_test.go` for the
  stat conversion and the effects modeled.
- **Character and buffs:** Windshaper Skyborne; both presets' buffs and debuffs together; the consumables of the
  preset that matches the set.
- **"reset"** is the measured Forever rule: every completed hard cast restarts the swing timer.
- **"old"** is `SimOptions.disable_cast_swing_reset`: only a swing due during the cast moves.

## Main: measured swing rule (reset), haste on, Smashbringer GCD on

| Build | Caster 2m | Caster 5m | Enh 2m | Enh 5m |
|---|---:|---:|---:|---:|
| **reaction 150 ms** | | | | |
| 29/22 weave | 753 | 547 | 931 | 941 |
| 29/22 bolt-chain | 729 | 532 | 853 | 908 |
| 8/36/7 weave | 712 | 574 | 918 | 944 |
| 8/36/7 no hard casts | 680 | 656 | 1270 | 1181 |
| Enh preset 19/32/0 | — | — | 1222 | 1156 |
| Ele preset 32/0/19 | 693 | 519 | — | — |
| **reaction 10 ms** | | | | |
| 29/22 weave | 763 | 550 | 944 | 946 |
| 29/22 bolt-chain | 744 | 539 | 863 | 912 |
| 8/36/7 weave | 729 | 579 | 927 | 947 |
| 8/36/7 no hard casts | 681 | 656 | 1271 | 1180 |
| Enh preset 19/32/0 | — | — | 1223 | 1156 |
| Ele preset 32/0/19 | 693 | 519 | — | — |

## Old rule (reference) and what the measured rule costs, reaction 150 ms

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

## Sensitivities, reset rule, reaction 150 ms

| Build | Gear | Fight | Base | Haste off | Smashbringer GCD off | Both off |
|---|---|---|---:|---:|---:|---:|
| 29/22 weave | Caster set (Mirage) | 2m | 753 | 702 | — | — |
| 29/22 weave | Caster set (Mirage) | 5m | 547 | 533 | — | — |
| 29/22 weave | Enh set (Smashbringer) | 2m | 931 | 812 | 925 | 804 |
| 29/22 weave | Enh set (Smashbringer) | 5m | 941 | 799 | 937 | 794 |
| 29/22 bolt-chain | Caster set (Mirage) | 2m | 729 | 689 | — | — |
| 29/22 bolt-chain | Caster set (Mirage) | 5m | 532 | 516 | — | — |
| 29/22 bolt-chain | Enh set (Smashbringer) | 2m | 853 | 772 | 835 | 754 |
| 29/22 bolt-chain | Enh set (Smashbringer) | 5m | 908 | 782 | 899 | 773 |
| 8/36/7 weave | Caster set (Mirage) | 2m | 712 | 655 | — | — |
| 8/36/7 weave | Caster set (Mirage) | 5m | 574 | 556 | — | — |
| 8/36/7 weave | Enh set (Smashbringer) | 2m | 918 | 816 | 908 | 816 |
| 8/36/7 weave | Enh set (Smashbringer) | 5m | 944 | 810 | 942 | 812 |
| 8/36/7 no hard casts | Caster set (Mirage) | 2m | 680 | 670 | — | — |
| 8/36/7 no hard casts | Caster set (Mirage) | 5m | 656 | 649 | — | — |
| 8/36/7 no hard casts | Enh set (Smashbringer) | 2m | 1270 | 1076 | 1234 | 1077 |
| 8/36/7 no hard casts | Enh set (Smashbringer) | 5m | 1181 | 980 | 1181 | 982 |
| Enh preset 19/32/0 | Enh set (Smashbringer) | 2m | 1222 | 1040 | 1221 | 1041 |
| Enh preset 19/32/0 | Enh set (Smashbringer) | 5m | 1156 | 970 | 1158 | 972 |
| Ele preset 32/0/19 | Caster set (Mirage) | 2m | 693 | 601 | — | — |
| Ele preset 32/0/19 | Caster set (Mirage) | 5m | 519 | 476 | — | — |

## Presets on their own launch gear (repo presets.ts, 200 ms reaction)

| Build | 2m reset | 2m old | 5m reset | 5m old |
|---|---:|---:|---:|---:|
| Enh preset 19/32/0 | 473 | 473 | 442 | 442 |
| Ele preset 32/0/19 | 394 | 394 | 324 | 324 |

## Sanity: casts, swings, OOM (reset, 150 ms, haste on)

| Build | Gear | Fight | LB+LvB/min | MH swings/min | OOM s | DPS ± stderr |
|---|---|---|---:|---:|---:|---:|
| 29/22 weave | Caster set (Mirage) | 2m | 12.7 | 13.7 | 5 | 753 ± 0.5 |
| 29/22 weave | Caster set (Mirage) | 5m | 5.5 | 17.2 | 170 | 547 ± 0.3 |
| 29/22 weave | Enh set (Smashbringer) | 2m | 6.8 | 19.2 | 52 | 931 ± 0.9 |
| 29/22 weave | Enh set (Smashbringer) | 5m | 3.1 | 22.7 | 197 | 941 ± 0.6 |
| 29/22 bolt-chain | Caster set (Mirage) | 2m | 14.1 | 12.6 | 32 | 729 ± 0.5 |
| 29/22 bolt-chain | Caster set (Mirage) | 5m | 6.1 | 16.8 | 196 | 532 ± 0.3 |
| 29/22 bolt-chain | Enh set (Smashbringer) | 2m | 8.2 | 17.7 | 57 | 853 ± 0.8 |
| 29/22 bolt-chain | Enh set (Smashbringer) | 5m | 3.7 | 22.0 | 202 | 908 ± 0.6 |
| 8/36/7 weave | Caster set (Mirage) | 2m | 14.2 | 14.3 | 0 | 712 ± 0.5 |
| 8/36/7 weave | Caster set (Mirage) | 5m | 9.0 | 17.9 | 147 | 574 ± 0.3 |
| 8/36/7 weave | Enh set (Smashbringer) | 2m | 11.1 | 19.0 | 40 | 918 ± 0.9 |
| 8/36/7 weave | Enh set (Smashbringer) | 5m | 8.6 | 22.7 | 181 | 944 ± 0.6 |
| 8/36/7 no hard casts | Caster set (Mirage) | 2m | 4.0 | 22.5 | 0 | 680 ± 0.5 |
| 8/36/7 no hard casts | Caster set (Mirage) | 5m | 4.0 | 22.1 | 42 | 656 ± 0.4 |
| 8/36/7 no hard casts | Enh set (Smashbringer) | 2m | 5.8 | 28.8 | 1 | 1270 ± 1.0 |
| 8/36/7 no hard casts | Enh set (Smashbringer) | 5m | 5.7 | 28.2 | 135 | 1181 ± 0.7 |
| Enh preset 19/32/0 | Enh set (Smashbringer) | 2m | 5.7 | 28.5 | 46 | 1222 ± 1.0 |
| Enh preset 19/32/0 | Enh set (Smashbringer) | 5m | 5.7 | 28.0 | 177 | 1156 ± 0.6 |
| Ele preset 32/0/19 | Caster set (Mirage) | 2m | 16.0 | 0.0 | 8 | 693 ± 0.4 |
| Ele preset 32/0/19 | Caster set (Mirage) | 5m | 3.1 | 0.0 | 76 | 519 ± 0.2 |

Notes:
- **The presets play their own rotation.** The Enhancement preset's APL never hard-casts, and the Elemental spec
  never melees, so the swing rule can't move either preset.
- **Bolt-chain under the old rule** only gets its chained cast when a cast pushed the swing, because a cast end
  restarts nothing. In the caster set that never happens, so it plays exactly like the weave there.
- **The caster set's 5-minute numbers are bound by mana.** The 29/22 weave is out of mana for 170 of 300 s.
- **The presets on their own launch gear don't match kami-labs' figures.** They come out at 473 (Enh) and 394 (Ele)
  at 2 min, where kami-labs reported 784 and 587. That MythicSim run used the site's own reference requests, which
  aren't in this repo, so the absolute numbers can't be compared.
