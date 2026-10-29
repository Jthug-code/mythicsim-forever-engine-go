#!/usr/bin/env bash
# Every comparison again, at one buff level: "light" (run 4's buffs) or "full" (SPELLHANCE_BUFFS=full).
# usage: bis_tools/rerun_all.sh light|full <work dir>   (run from sim/shaman/spellhance)
# Steps: BiS stat weights and a search from that level's current set (which it overwrites in testdata/),
# then the race, Elemental mana, Forever-launch, SoD, Water Shield Enhancement and fight-length runs.
# Results go to <work dir>; nothing here commits.
set -euo pipefail
mode=${1:?light or full}
work=${2:?work dir}
case $mode in
light) unset SPELLHANCE_BUFFS; bis=testdata/forever-enh-bis.gear.json ;;
full) export SPELLHANCE_BUFFS=full; bis=testdata/forever-enh-bis-fullbuffs.gear.json ;;
*) echo "mode must be light or full" >&2; exit 2 ;;
esac
mkdir -p "$work/data"
export SPELLHANCE_ITERATIONS=${SPELLHANCE_ITERATIONS:-10000}
win() { cygpath -w "$1" 2>/dev/null || echo "$1"; }
py() { python -I "$@"; }
db=../../../assets/database/db.json
gotest() { go test --tags=with_db . -count=1 -timeout 300m "$@"; }

# Wowhead's Forever item data, from the engine's daily data branch.
for f in wowhead_forever_gearplanner.txt:gearplanner.txt wowhead_item_tooltips.csv:tooltips.csv; do
	git show "origin/data/master/wowhead-forever:assets/db_inputs/${f%%:*}" >"$work/data/${f##*:}"
done

echo "== $mode: weights and BiS search from $bis"
BIS_WEIGHTS_GEAR=$(win "$bis") BIS_WEIGHTS_OUT=$(win "$work/weights.json") SPELLHANCE_ITERATIONS=20000 \
	gotest -run TestBisStatWeights >/dev/null
py bis_tools/gen_candidates.py "$work/data" bis_tools "$db" "$work/weights.json" "$work/search_in.json" >/dev/null
py bis_tools/fix_enchants.py "$work/search_in.json" >/dev/null
py - "$work/search_in.json" "$bis" <<'EOF'
import json, sys
d = json.load(open(sys.argv[1], encoding='utf-8'))
d['start'] = json.load(open(sys.argv[2], encoding='utf-8'))
json.dump(d, open(sys.argv[1], 'w', encoding='utf-8'), indent=1)
EOF
BIS_SEARCH_IN=$(win "$work/search_in.json") BIS_SEARCH_OUT=$(win "$work/search_out.json") SPELLHANCE_ITERATIONS=4000 \
	gotest -run TestBisSearch | grep -E "start|pass" || true
py bis_tools/summarize.py "$work/data" bis_tools "$db" "$work/search_out.json" "$work/bis.gear.json"
py - "$work/bis.gear.json" "$bis" <<'EOF'
import json, sys
g = json.load(open(sys.argv[1], encoding='utf-8'))
open(sys.argv[2], 'w', encoding='utf-8').write(json.dumps({'items': [{k: v for k, v in it.items() if v} for it in g['items']]}) + '\n')
EOF

for step in race:SPELLHANCE_RACE_OUT:TestSpellhanceRaceMatrix ele:SPELLHANCE_ELE_OUT:TestEleManaSweep \
	enhws:SPELLHANCE_ENHWS_OUT:TestEnhWaterShield length:SPELLHANCE_LENGTH_OUT:TestFightLength \
	forever:SPELLHANCE_FOREVER_OUT:TestSpellhanceForeverMatrix sod:SPELLHANCE_OUT:TestSpellhanceMatrix; do
	IFS=: read -r name var test <<<"$step"
	echo "== $mode: $name"
	env "$var=$(win "$work/$name")" go test --tags=with_db . -count=1 -timeout 300m -run "^$test\$" | tail -1
done
echo "== $mode: done"
