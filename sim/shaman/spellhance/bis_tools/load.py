import json, re


def load_planner(path):
    """Every WH.setPageData block of Wowhead's gear planner file, keyed by the last part of its name."""
    raw = open(path, encoding='utf-8').read()
    out = {}
    for m in re.finditer(r'WH\.setPageData\("([^"]+)",\s*', raw):
        txt = raw[m.end():]
        depth, instr, esc, j = 0, False, False, 0
        for j, ch in enumerate(txt):
            if instr:
                if esc:
                    esc = False
                elif ch == '\\':
                    esc = True
                elif ch == '"':
                    instr = False
                continue
            if ch == '"':
                instr = True
            elif ch in '{[':
                depth += 1
            elif ch in '}]':
                depth -= 1
                if depth == 0:
                    break
        body = re.sub(r',\s*([}\]])', r'\1', txt[:j + 1])
        out[m.group(1).split('.')[-1]] = json.loads(body)
    return out
