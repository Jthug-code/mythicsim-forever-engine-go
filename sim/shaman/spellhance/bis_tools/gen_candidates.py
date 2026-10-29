"""Candidate items and enchants per gear slot for the sim-driven BiS search.

usage: gen_candidates.py <data dir> <py dir> <engine db.json> <weights.json> <out.json>
The pool is Wowhead's Forever gear planner data, filtered as in rank.py.
"""
import json, re, sys, csv
sys.path.insert(0, sys.argv[2])
from load import load_planner

planner_all = load_planner(sys.argv[1] + '/gearplanner.txt')
planner = planner_all['item']
planner_enchants = {e['enchant'] for e in planner_all['enchant'].values() if e.get('enchant')}
db = json.load(open(sys.argv[3], encoding='utf-8'))
w = json.load(open(sys.argv[4], encoding='utf-8'))

TIPS = {}
with open(sys.argv[1] + '/tooltips.csv', encoding='utf-8') as f:
    for line in f:
        i, _, body = line.partition(',')
        if i.isdigit() and i in planner:
            TIPS[int(i)] = re.sub(r'\s+', ' ', re.sub(r'<[^>]*>', ' ', json.loads(body).get('tooltip', '')))
OTHER_CLASS_SETS = ('Plagueheart', 'Frostfire', 'Cryptstalker', 'Dreadnaught', 'Bonescythe', 'Redemption', 'Dreamwalker',
                    'Vestments of Faith')


def allowed(it, pl):
    tip = TIPS.get(it['id'], '')
    m = re.search(r'Classes: ([A-Za-z ,]+?) (Requires|Durability|Equip|Use|Chance|Sell|$)', tip)
    if m and 'Shaman' not in m.group(1):
        return False
    if any(re.search(re.escape(s) + r"[^()]*\(0/\d\)", tip) for s in OTHER_CLASS_SETS):
        return False
    classes = pl.get('stats', {}).get('classes') or pl.get('classMask')
    return not classes or classes & 64


def ep_of(stats, weights):
    hit = weights['PseudoStatSpellHitPercent'] / 10 + weights['PseudoStatMeleeHitPercent'] / 10
    crit = weights['StatMeleeCritRating'] + weights['StatSpellCritRating']
    table = {0: weights['StatStrength'], 1: weights['StatAgility'], 3: weights['StatIntellect'], 5: weights['StatSpellDamage'],
             10: weights['StatNatureDamage'], 7: weights['StatFireDamage'], 17: weights['StatAttackPower'], 20: hit, 12: hit,
             21: crit, 13: crit, 22: weights['StatMeleeHasteRating'], 24: weights['StatExpertiseRating'],
             23: weights['StatArmorPenetration'], 34: weights['StatMP5']}
    return sum(table.get(int(k), 0) * v for k, v in stats.items())


# A second ranking that discounts Intellect and spell power, so the search also sees melee pieces.
w_phys = dict(w, StatIntellect=w['StatIntellect'] * 0.4, StatSpellDamage=w['StatSpellDamage'] * 0.4,
              StatNatureDamage=w['StatNatureDamage'] * 0.4, StatMP5=w['StatMP5'] * 0.4)

SLOT_OF_TYPE = {1: [0], 2: [1], 3: [2], 4: [3], 5: [4], 6: [5], 7: [6], 8: [7], 9: [8], 10: [9], 11: [10, 11], 12: [12, 13]}
by_slot = {s: [] for s in range(17)}
for it in db['items']:
    pl = planner.get(str(it['id']))
    if pl is None or it.get('quality', 0) < 3 or pl.get('requiredLevel', 0) > 60 or 'Test' in it['name']:
        continue
    if not allowed(it, pl):
        continue
    so = it.get('scalingOptions', {}).get('0', {})
    st = so.get('stats', {})
    t = it.get('type')
    fx = bool(it.get('itemEffects'))
    dps = 0
    if t in SLOT_OF_TYPE:
        if it.get('armorType', 0) > 3:
            continue
        slots = SLOT_OF_TYPE[t]
    elif t == 13:
        wt, ht = it.get('weaponType'), it.get('handType')
        if wt == 7 or (wt == 5 and ht == 3):
            slots = [15]
        elif wt in (1, 2, 3, 4, 8) and ht in (1, 2, 4):
            slots = [14]
            dps = (so.get('weaponDamageMin', 0) + so.get('weaponDamageMax', 0)) / 2 / it['weaponSpeed']
        else:
            continue
    elif t == 14 and it.get('rangedWeaponType') == 8:
        slots = [16]
    else:
        continue
    e1 = ep_of(st, w) + w['PseudoStatMainHandDps'] * dps
    e2 = ep_of(st, w_phys) + w['PseudoStatMainHandDps'] * dps
    for s in slots:
        by_slot[s].append({'id': it['id'], 'name': it['name'], 'ep': e1, 'ep_phys': e2, 'fx': fx,
                           'two_hand': it.get('handType') == 4, 'unique': bool(it.get('unique'))})

candidates, start = {}, []
for s in range(17):
    items = by_slot[s]
    pick = {}
    n = 8 if s not in (14,) else 14
    for key in ('ep', 'ep_phys'):
        for c in sorted(items, key=lambda c: -c[key])[:n]:
            pick[c['id']] = c
    if s in (12, 13, 16):
        for c in items:
            if c['fx']:
                pick[c['id']] = c
    if s == 14:
        for c in sorted([c for c in items if not c['two_hand']], key=lambda c: -c['ep'])[:6]:
            pick[c['id']] = c
    candidates[str(s)] = sorted(pick.values(), key=lambda c: -c['ep'])

# EP-greedy start: best two-hander, no off-hand; rings and trinkets not doubled.
used = set()
for s in range(17):
    if s == 15:
        start.append({})
        continue
    for c in candidates[str(s)]:
        if s == 14 and not c['two_hand']:
            continue
        if c['id'] in used:
            continue
        start.append({'id': c['id']})
        used.add(c['id'])
        break
    else:
        start.append({})

ENCHANT_SLOT_TYPE = {0: 1, 1: 2, 2: 3, 3: 4, 4: 5, 5: 6, 6: 7, 8: 1, 9: 10, 14: 13}
enchants = {}
for s, t in ENCHANT_SLOT_TYPE.items():
    opts = []
    for e in db['enchants']:
        if e.get('effectId') not in planner_enchants:
            continue
        if e.get('type') != t and t not in (e.get('extraTypes') or []):
            continue
        st = {str(i): v for i, v in enumerate(e.get('stats') or []) if v}
        score = ep_of(st, w)
        named = any(k in e.get('name', '') for k in ('Crusader', 'Rapidity', 'Haste', 'Signet of Might', 'Scourge'))
        opts.append((score + (1000 if named else 0), e['effectId'], e.get('name')))
    opts.sort(reverse=True)
    enchants[str(s)] = [{'id': eid, 'name': name} for _, eid, name in opts[:6]]

json.dump({'start': {'items': start}, 'candidates': {s: [{'id': c['id'], 'name': c['name'], 'two_hand': c['two_hand']}
                                                         for c in cs] for s, cs in candidates.items()},
           'enchants': enchants}, open(sys.argv[5], 'w', encoding='utf-8'), indent=1)
print({s: len(c) for s, c in candidates.items()})
print('start', [x.get('id') for x in start])
for s, es in enchants.items():
    print('enchants', s, [(e['id'], e['name']) for e in es])
