import json, sys
d = json.load(open(sys.argv[1], encoding='utf-8'))
# ZG head and leg arcana are class-locked; the shaman's is Vodouisant's. Libram arcana are open to all.
for s in ('0', '8'):
    d['enchants'][s] = [{'id': i, 'name': n} for i, n in ((2543, 'Arcanum of Rapidity'), (2587, "Vodouisant's Vigilant Embrace"),
                        (1506, 'Lesser Arcanum of Voracity (Str)'), (1508, 'Lesser Arcanum of Voracity (Agi)'),
                        (1509, 'Lesser Arcanum of Voracity (Int)'))]
d['enchants']['14'].append({'id': 2646, 'name': 'Enchant 2H Weapon - Agility'})
d['enchants']['5'].append({'id': 7656, 'name': 'Enchant Bracer - Superior Agility'})
json.dump(d, open(sys.argv[1], 'w', encoding='utf-8'), indent=1)
print({s: [e['id'] for e in es] for s, es in d['enchants'].items()})
