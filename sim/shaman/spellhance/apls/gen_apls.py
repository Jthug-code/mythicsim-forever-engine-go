"""Generates the Spellhance weave APLs (SpellhanceWeaveHelper rules). Usage: python gen_apls.py sim/shaman/spellhance/apls"""
import json
import os
import sys

OUT = sys.argv[1]

LB, LVB, SS, FS, ES, CL = 15208, 1326047, 17364, 29228, 10414, 10605
SOE, GOA, SEARING = 25361, 25359, 10438
MAELSTROM = 408505
WINDOW = "0.4s"


def const(v):
    return {"const": {"val": str(v)}}


def cmp(op, lhs, rhs):
    return {"cmp": {"op": op, "lhs": lhs, "rhs": rhs}}


def math(op, lhs, rhs):
    return {"math": {"op": op, "lhs": lhs, "rhs": rhs}}


def and_(*vals):
    return {"and": {"vals": list(vals)}}


def or_(*vals):
    return {"or": {"vals": list(vals)}}


def not_(v):
    return {"not": {"val": v}}


def spell(sid):
    return {"spellId": {"spellId": sid}}


def cast(sid, condition=None):
    action = {"castSpell": spell(sid)}
    if condition is not None:
        action["condition"] = condition
    return {"action": action}


time_to_swing = {"autoTimeToNext": {"autoType": "MainHandAuto"}}
swing_time = {"autoSwingTime": {"autoType": "MainHand"}}
since_swing = {"autoTimeSinceLast": {"autoType": "MainHandAuto"}}
# Time since the swing timer last restarted: a swing landed or a hard cast completed.
since_restart = math("OpSub", swing_time, time_to_swing)
reaction = {"inputDelay": {}}


def fire_delay(sid):
    # The engine queues a spell up to 0.4 s before its cooldown or the GCD is ready, and fires it
    # then without asking again, so every timing condition is checked at that moment.
    return {"max": {"vals": [{"gcdTimeToReady": {}}, {"spellTimeToReady": spell(sid)}]}}


def in_window(elapsed, sid=None, after_reaction=True):
    # CAST is lit for 0.4 s after the restart; a player presses after their reaction time.
    at_fire = elapsed if sid is None else math("OpAdd", elapsed, fire_delay(sid))
    if not after_reaction:
        return cmp("OpLe", elapsed, const(WINDOW))
    return and_(cmp("OpGe", elapsed, reaction), cmp("OpLe", at_fire, const(WINDOW)))


def gcd_fits(sid):
    # INSTANT: its GCD ends before the swing lands. Otherwise WAIT.
    return cmp("OpGe", time_to_swing, math("OpAdd", {"spellGcdHastedDuration": spell(sid)}, fire_delay(sid)))


def cast_window(variant, sid=None, after_reaction=True):
    if variant == "weave":
        # Hard casts only in the window after a swing lands.
        return in_window(since_swing, sid, after_reaction)
    if variant == "chain":
        # Also in the window after a hard cast completes, for one chained cast per swing: the
        # first cast after a swing ends at most 0.4 s + one bolt after it, the chained one later.
        one_bolt_after_swing = math("OpAdd", math("OpAdd", const(WINDOW), const(WINDOW)), {"spellCastTime": spell(LB)})
        return or_(
            in_window(since_swing, sid, after_reaction),
            and_(in_window(since_restart, sid, after_reaction), cmp("OpLe", since_swing, one_bolt_after_swing)),
        )
    raise ValueError(variant)


def build(variant, hard_casts=True, maelstrom=False, lava_burst=True, chain_lightning=False, spend_at=5):

    # Only name spells and auras the build has: the APL drops an action whose value it can't resolve.
    hard_ready = {"spellCanCast": spell(LB)}
    if chain_lightning:
        hard_ready = or_({"spellCanCast": spell(CL)}, hard_ready)
    if lava_burst:
        hard_ready = or_(and_({"dotIsActive": spell(FS)}, {"spellCanCast": spell(LVB)}), hard_ready)
    # While CAST is lit and a hard cast can go, the player waits for it rather than press an instant.
    cast_pending = and_(cast_window(variant, after_reaction=False), hard_ready)
    instant_ok = not_(cast_pending) if hard_casts else None

    def instant(sid, extra=None):
        conds = [gcd_fits(sid)]
        if instant_ok is not None:
            conds.append(instant_ok)
        if extra is not None:
            conds.append(extra)
        return cast(sid, and_(*conds))

    priority = [{"action": {"autocastOtherCooldowns": {}}}]
    if hard_casts:
        if lava_burst:
            priority.append(cast(LVB, and_(cast_window(variant, LVB), {"dotIsActive": spell(FS)})))
        if chain_lightning:
            priority.append(cast(CL, cast_window(variant, CL)))
        if maelstrom:
            # A 5-stack bolt is an instant; it goes with the instants below.
            priority.append(cast(LB, and_(cast_window(variant, LB), cmp("OpLt", {"auraNumStacks": {"auraId": {"spellId": MAELSTROM}}}, const(5)))))
        else:
            priority.append(cast(LB, cast_window(variant, LB)))
    priority += [
        instant(SOE, cmp("OpLe", {"totemRemainingTime": {"totemType": "Earth"}}, const("0s"))),
        instant(GOA, cmp("OpLe", {"totemRemainingTime": {"totemType": "Air"}}, const("0s"))),
        instant(SS),
    ]
    if maelstrom:
        # Spend the charges at spend_at: 5 makes the bolt instant and free; Soulcrusher's Rage 6 stacks
        # to 10, where the bolt does +100% and casts twice.
        priority.append(instant(LB, cmp("OpGe", {"auraNumStacks": {"auraId": {"spellId": MAELSTROM}}}, const(spend_at))))
    priority += [
        instant(FS, not_({"dotIsActive": spell(FS)})),
        instant(ES),
        instant(SEARING, and_(cmp("OpLe", {"totemRemainingTime": {"totemType": "Fire"}}, const("0s")),
                              cmp("OpGe", {"remainingTime": {}}, const("20s")))),
    ]
    return {
        "type": "TypeAPL",
        "prepullActions": [
            {"action": {"castSpell": spell(SOE)}, "doAtValue": const("-4.5s")},
            {"action": {"castSpell": spell(GOA)}, "doAtValue": const("-3s")},
            {"action": {"castSpell": spell(SEARING)}, "doAtValue": const("-1.5s")},
        ],
        "priorityList": priority,
    }


os.makedirs(OUT, exist_ok=True)
for name, apl in {
    "weave": build("weave"),
    "chain": build("chain"),
    "weave_maelstrom": build("weave", maelstrom=True, lava_burst=False),
    "melee_maelstrom": build("weave", hard_casts=False, maelstrom=True, lava_burst=False),
    "weave_cl": build("weave", chain_lightning=True),
    "chain_cl": build("chain", chain_lightning=True),
    "weave_maelstrom10": build("weave", maelstrom=True, lava_burst=False, spend_at=10),
    "melee_maelstrom10": build("weave", hard_casts=False, maelstrom=True, lava_burst=False, spend_at=10),
}.items():
    with open(os.path.join(OUT, name + ".apl.json"), "w", newline="\n") as f:
        json.dump(apl, f, indent=2)
        f.write("\n")
    print("wrote", name)
