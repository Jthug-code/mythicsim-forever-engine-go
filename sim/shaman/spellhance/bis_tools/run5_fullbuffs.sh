#!/usr/bin/env bash
# Run 5: every comparison again with full raid buffs and consumables (SPELLHANCE_BUFFS=full).
# usage: bis_tools/run5_fullbuffs.sh <work dir>   (run from sim/shaman/spellhance)
# Steps: BiS stat weights and search under full buffs, then the race, Forever-launch, Elemental mana and
# SoD matrices. Each writes into <work dir>; nothing here commits.
set -euo pipefail
work=${1:?work dir}
mkdir -p "$work/data"
export SPELLHANCE_BUFFS=full SPELLHANCE_ITERATIONS=${SPELLHANCE_ITERATIONS:-10000}
win() { cygpath -w "$1" 2>/dev/null || echo "$1"; }
py() { python -I "$@"; }
db=../../../assets/database/db.json

# Wowhead's Forever item data, from the engine's daily data branch.
for f in wowhead_forever_gearplanner.txt:gearplanner.txt wowhead_item_tooltips.csv:tooltips.csv; do
	git show "origin/data/master/wowhead-forever:assets/db_inputs/${f%%:*}" >"$work/data/${f##*:}"
done

echo "== weights on the run 4 set, full buffs"
BIS_WEIGHTS_GEAR=$(win testdata/forever-enh-bis.gear.json) BIS_WEIGHTS_OUT=$(win "$work/weights.json") SPELLHANCE_ITERATIONS=20000 \
	go test --tags=with_db . -run TestBisStatWeights -count=1 >/dev/null
py bis_tools/gen_candidates.py "$work/data" bis_tools "$db" "$work/weights.json" "$work/search_in.json"
py bis_tools/fix_enchants.py "$work/search_in.json" >/dev/null
# Start from the run 4 set.
py - "$work/search_in.json" testdata/forever-enh-bis.gear.json <<'EOF'
import json, sys
d = json.load(open(sys.argv[1], encoding='utf-8'))
d['start'] = json.load(open(sys.argv[2], encoding='utf-8'))
json.dump(d, open(sys.argv[1], 'w', encoding='utf-8'), indent=1)
EOF

echo "== BiS search, full buffs"
BIS_SEARCH_IN=$(win "$work/search_in.json") BIS_SEARCH_OUT=$(win "$work/search_out.json") SPELLHANCE_ITERATIONS=4000 \
	go test --tags=with_db . -run TestBisSearch -count=1 -timeout 120m | grep -E "start|pass" || true
py bis_tools/summarize.py "$work/data" bis_tools "$db" "$work/search_out.json" "$work/bis.gear.json"
py - "$work/bis.gear.json" testdata/forever-enh-bis-fullbuffs.gear.json <<'EOF'
import json, sys
g = json.load(open(sys.argv[1], encoding='utf-8'))
open(sys.argv[2], 'w', encoding='utf-8').write(json.dumps({'items': [{k: v for k, v in it.items() if v} for it in g['items']]}) + '\n')
EOF

echo "== race matrix"
SPELLHANCE_RACE_OUT=$(win "$work/race") go test --tags=with_db . -run TestSpellhanceRaceMatrix -count=1 -timeout 120m | tail -1
echo "== Elemental mana sweep"
SPELLHANCE_ELE_OUT=$(win "$work/ele") go test --tags=with_db . -run TestEleManaSweep -count=1 -timeout 120m | tail -1
echo "== Forever launch matrix"
SPELLHANCE_FOREVER_OUT=$(win "$work/forever") go test --tags=with_db . -run TestSpellhanceForeverMatrix -count=1 -timeout 180m | tail -1
echo "== SoD matrix"
SPELLHANCE_OUT=$(win "$work/sod") go test --tags=with_db . -run TestSpellhanceMatrix -count=1 -timeout 300m | tail -1
echo "== done"
