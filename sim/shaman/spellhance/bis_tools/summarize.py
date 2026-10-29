"""Print a searched gear set with names, enchants and Wowhead Forever sources; write the gear file.

usage: summarize.py <data dir> <py dir> <engine db.json> <search_out.json> [gear out]
"""
import json, sys, re
sys.path.insert(0, sys.argv[2])
from load import load_planner

planner = load_planner(sys.argv[1] + '/gearplanner.txt')['item']
db = json.load(open(sys.argv[3], encoding='utf-8'))
res = json.load(open(sys.argv[4], encoding='utf-8'))
items = {it['id']: it for it in db['items']}
enchants = {}
for e in db['enchants']:
    enchants.setdefault(e['effectId'], e.get('name'))
zones = {z['id']: z['name'] for z in db.get('zones', [])}
npcs = {n['id']: n['name'] for n in db.get('npcs', [])}
SLOTS = ['Head', 'Neck', 'Shoulder', 'Back', 'Chest', 'Wrist', 'Hands', 'Waist', 'Legs', 'Feet', 'Ring', 'Ring', 'Trinket',
         'Trinket', 'Main hand', 'Off hand', 'Totem']
SRC = {1: 'crafted', 2: 'drop', 3: 'PvP', 4: 'quest', 5: 'vendor', 16: 'world drop'}


def source(i):
    it, pl = items.get(i, {}), planner.get(str(i), {})
    out = []
    for s in it.get('sources', []):
        d = s.get('drop')
        if d:
            out.append(f"{npcs.get(d.get('npcId'), 'drop')}, {zones.get(d.get('zoneId'), d.get('zoneId'))}")
    if not out:
        kinds = [SRC.get(k, str(k)) for k in pl.get('source', [])]
        more = [sm.get('n', '') + (f", {zones.get(sm['z'], sm['z'])}" if 'z' in sm else '') for sm in pl.get('sourcemore') or []]
        out.append(' / '.join(kinds) + (': ' + '; '.join(more) if more else '') if kinds else 'no source on Wowhead')
    return '; '.join(out)


gear = res['gear']['items']
print(f"DPS {res['dps']:.1f}")
for line in res['log']:
    print('  ', line)
for slot, g in zip(SLOTS, gear):
    i = g.get('id')
    if not i:
        print(f"{slot:9} -")
        continue
    it = items[i]
    print(f"{slot:9} {i:>7} {it['name']:40} ilvl{it['scalingOptions']['0'].get('ilvl')} | {enchants.get(g.get('enchant'), '')} | {source(i)}")
if len(sys.argv) > 5:
    json.dump(res['gear'], open(sys.argv[5], 'w', encoding='utf-8'))
