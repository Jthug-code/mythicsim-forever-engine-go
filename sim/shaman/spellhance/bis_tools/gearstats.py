import json, sys
db = json.load(open(sys.argv[1], encoding='utf-8'))
byid = {it['id']: it for it in db['items']}
gear = json.load(open(sys.argv[2], encoding='utf-8'))['items']
names = {0: 'str', 1: 'agi', 2: 'sta', 3: 'int', 5: 'sp', 17: 'ap', 20: 'hit', 21: 'crit', 22: 'haste', 24: 'exp', 12: 'shit', 13: 'scrit'}
tot = {}
for g in gear:
    it = byid.get(g.get('id'))
    if not it:
        continue
    st = it.get('scalingOptions', {}).get('0', {}).get('stats', {})
    short = {names[int(k)]: v for k, v in st.items() if int(k) in names}
    for k, v in short.items():
        tot[k] = tot.get(k, 0) + v
    print(it['id'], it['name'], short, [e.get('buffName') for e in it.get('itemEffects', [])])
print('total', tot)
