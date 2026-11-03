#!/usr/bin/env bash
# The Forever BiS search for one build at one buff level: stat weights, an EP-ranked candidate pool from Wowhead's
# Forever item data, then the sim-driven slot-by-slot search. "enh" scores 8/36/7 with no hard casts, "ele" the
# Elemental preset (BIS_BUILD=ele), "hyb" the 29/22 weave (BIS_BUILD=hyb). The set goes to
# testdata/forever-<build>-bis[-fullbuffs].gear.json. An existing one is the start; otherwise the optional start gear
# file is, or failing that the pool's EP-greedy pick.
# usage: bis_tools/search_bis.sh light|full enh|ele|hyb <work dir> [start gear]   (run from sim/shaman/spellhance)
set -euo pipefail
mode=${1:?light or full}
which=${2:?enh, ele or hyb}
work=${3:?work dir}
case $mode in
light) unset SPELLHANCE_BUFFS; suffix= ;;
full) export SPELLHANCE_BUFFS=full; suffix=-fullbuffs ;;
*) echo "mode must be light or full" >&2; exit 2 ;;
esac
case $which in
enh) unset BIS_BUILD; caster= ;;
ele) export BIS_BUILD=ele; caster=caster ;;
hyb) export BIS_BUILD=hyb; caster=caster ;;
*) echo "build must be enh, ele or hyb" >&2; exit 2 ;;
esac
bis=testdata/forever-$which-bis$suffix.gear.json
mkdir -p "$work/data"
win() { cygpath -w "$1" 2>/dev/null || echo "$1"; }
py() { python -I "$@"; }
db=../../../assets/database/db.json
gotest() { go test --tags=with_db . -count=1 -timeout 300m "$@"; }

for f in wowhead_forever_gearplanner.txt:gearplanner.txt wowhead_item_tooltips.csv:tooltips.csv; do
	git show "origin/data/master/wowhead-forever:assets/db_inputs/${f%%:*}" >"$work/data/${f##*:}"
done

given=${4:-}
if [ -f "$bis" ]; then start=$bis; elif [ -n "$given" ]; then start=$given; else
	# No set yet: weigh on the launch preset, then start from the pool's EP-greedy pick.
	start=../../../ui/specs/shaman/elemental/gear_sets/launch.gear.json
	[ "$which" = enh ] && start=../../../ui/specs/shaman/enhancement/gear_sets/launch.gear.json
fi
echo "== $mode $which: weights on $start"
BIS_WEIGHTS_GEAR=$(win "$start") BIS_WEIGHTS_OUT=$(win "$work/weights.json") SPELLHANCE_ITERATIONS=20000 \
	gotest -run '^TestBisStatWeights$' >/dev/null
py bis_tools/gen_candidates.py "$work/data" bis_tools "$db" "$work/weights.json" "$work/search_in.json" >/dev/null
py bis_tools/fix_enchants.py "$work/search_in.json" $caster >/dev/null
if [ -f "$bis" ] || [ -n "$given" ]; then
	py - "$work/search_in.json" "$start" <<'EOF'
import json, sys
d = json.load(open(sys.argv[1], encoding='utf-8'))
d['start'] = json.load(open(sys.argv[2], encoding='utf-8'))
json.dump(d, open(sys.argv[1], 'w', encoding='utf-8'), indent=1)
EOF
fi
echo "== $mode $which: search"
BIS_SEARCH_IN=$(win "$work/search_in.json") BIS_SEARCH_OUT=$(win "$work/search_out.json") SPELLHANCE_ITERATIONS=4000 \
	gotest -run '^TestBisSearch$' -v | grep -E "start|pass" || true
py bis_tools/summarize.py "$work/data" bis_tools "$db" "$work/search_out.json" "$work/bis.gear.json"
py - "$work/bis.gear.json" "$bis" <<'EOF'
import json, sys
g = json.load(open(sys.argv[1], encoding='utf-8'))
open(sys.argv[2], 'w', encoding='utf-8').write(json.dumps({'items': [{k: v for k, v in it.items() if v} for it in g['items']]}) + '\n')
EOF
echo "== $mode $which: wrote $bis"
