"""Generate the Tiled maps in infinitefighters-backend/levels/.

  sprint.json  boss arena "Đường Đua Sprint" (co-op), crowded office as in the concept art.
  party.json   party hall for the PvP money phase (extension "Anh em hết mình, tiền tình hết hồn").

Both are 40×22 tiles of 32 px, viewed at 1280×720. Run: python3 scripts/gen_levels.py
"""
import json
import os

W, H, T = 40, 22, 32
FLOOR = 19          # floor top row: y = 608
FY = FLOOR * T
OUT = os.path.join(os.path.dirname(__file__), "..", "infinitefighters-backend", "levels")
# local tile ids (assets/tiles/office.json): 0-2 floor top L/M/R, 3 floor fill, 4-6 one-way desk L/M/R,
# 7 support leg (decor), 8-9 server top/bottom, 10 wall fill
COLLISION = {0: "solid", 1: "solid", 2: "solid", 3: "solid", 4: "oneway", 5: "oneway", 6: "oneway",
             8: "solid", 9: "solid", 10: "solid"}


def build(desks, objects):
    g = [[0] * W for _ in range(H)]

    def put(c, r, local):
        g[r][c] = local + 1

    for r in range(H):
        put(0, r, 10)
        put(W - 1, r, 10)
    for c in range(1, W - 1):
        put(c, FLOOR, 0 if c == 1 else 2 if c == W - 2 else 1)
        for r in range(FLOOR + 1, H):
            put(c, r, 3)
    for row, x0, x1 in desks:
        for c in range(x0, x1 + 1):
            put(c, row, 4 if c == x0 else 6 if c == x1 else 5)
    # Legs go down until they meet anything, so decoration never overwrites a collision tile.
    for row, x0, x1 in desks:
        for c in ({x0 + 1, x1 - 1} if x1 - x0 > 3 else {x0 + (x1 - x0) // 2}):
            if c in (0, W - 1):
                continue
            r = row + 1
            while r < FLOOR and g[r][c] == 0:
                put(c, r, 7)
                r += 1
    objs = [{"id": i + 1, "name": name, "type": type_, "point": True, "x": x, "y": y,
             "width": 0, "height": 0, "rotation": 0, "visible": True}
            for i, (type_, name, x, y) in enumerate(objects)]
    return {
        "type": "map", "version": "1.10", "tiledversion": "1.11.0", "orientation": "orthogonal",
        "renderorder": "right-down", "infinite": False, "width": W, "height": H, "tilewidth": T, "tileheight": T,
        "compressionlevel": -1, "nextlayerid": 3, "nextobjectid": len(objs) + 1,
        "layers": [
            {"id": 1, "name": "ground", "type": "tilelayer", "width": W, "height": H, "x": 0, "y": 0,
             "opacity": 1, "visible": True, "data": [v for row in g for v in row]},
            {"id": 2, "name": "objects", "type": "objectgroup", "draworder": "topdown", "x": 0, "y": 0,
             "opacity": 1, "visible": True, "objects": objs}],
        "tilesets": [{"firstgid": 1, "name": "office", "image": "office.png", "imagewidth": 256, "imageheight": 64,
                      "columns": 8, "tilecount": 16, "tilewidth": T, "tileheight": T, "margin": 0, "spacing": 0,
                      "tiles": [{"id": k, "properties": [{"name": "collision", "type": "string", "value": v}]}
                                for k, v in COLLISION.items()]}]}


def props(items):
    return [("prop", name, x, y) for name, x, y in items]


def sprint():
    # One-way desks: 3 levels, 4 tiles apart (jump apex ≈ 5 tiles). (top row, first col, last col)
    desks = [(15, 7, 14), (15, 25, 32), (11, 1, 4), (11, 16, 23), (11, 35, 38), (7, 6, 13), (7, 26, 33)]
    objects = [("spawn", f"P{i + 1}", 64 + i * 48, FY) for i in range(5)]
    objects.append(("boss", "boss", 1000, FY))
    objects += props([
        ("filing-cabinet", 56, FY), ("plant", 92, FY), ("desk", 160, FY), ("chair", 212, FY),
        ("desk", 300, FY), ("chair", 350, FY), ("filing-cabinet", 410, FY), ("desk", 540, FY), ("chair", 590, FY),
        ("water-cooler", 660, FY), ("plant", 700, FY), ("desk", 780, FY), ("chair", 830, FY),
        ("filing-cabinet", 900, FY), ("desk", 960, FY), ("plant", 1100, FY), ("filing-cabinet", 1140, FY),
        ("elevator", 1212, FY),
        ("monitor", 150, FY - 40), ("monitor", 300, FY - 40), ("monitor", 540, FY - 40), ("monitor", 780, FY - 40),
        ("monitor", 960, FY - 40),
        ("poster", 470, 440), ("poster", 1180, 420), ("poster", 70, 300),
        ("ceiling-light", 160, 112), ("ceiling-light", 480, 112), ("ceiling-light", 800, 112), ("ceiling-light", 1120, 112),
        ("monitor", 260, 224), ("plant", 420, 224), ("filing-cabinet", 860, 224), ("monitor", 1010, 224),
        ("plant", 740, 352), ("monitor", 60, 352), ("plant", 1200, 352),
        ("monitor", 400, 480), ("monitor", 870, 480), ("plant", 1060, 480),
    ])
    objects += [("coffee", "", x, y) for x, y in [(336, 480), (640, 352), (944, 224)]]
    return build(desks, objects)


def party():
    # Party hall: 7 desks on 3 levels, mirrored. The side ledges sit above the ATMs, so each ATM can be
    # reached along the floor and by dropping from the ledge.
    desks = [(15, 5, 11), (15, 28, 34), (12, 1, 4), (12, 35, 38), (11, 15, 24), (7, 7, 13), (7, 26, 32)]
    objects = [  # start points, about the same walk to the safe
        ("spawn", "P1", 384, FY), ("spawn", "P2", 896, FY), ("spawn", "P3", 192, FY),
        ("spawn", "P4", 1088, FY), ("spawn", "P5", 640, 352),
    ]
    objects.append(("safe", "safe", 640, FY))
    objects += [("atm", "atm", 72, FY), ("atm", "atm", 1208, FY)]
    objects += [("respawn", "", x, y) for x, y in
                [(160, FY), (1120, FY), (272, 480), (1008, 480), (320, 224), (960, 224)]]
    objects += props([  # missing party art is skipped by the client until it exists
        ("banquet-table", 300, FY), ("banquet-table", 980, FY), ("drink-tower", 520, FY), ("drink-tower", 760, FY),
        ("speaker", 140, FY), ("speaker", 1140, FY), ("plant", 440, FY), ("plant", 840, FY),
        ("balloon-cluster", 200, 480), ("balloon-cluster", 1080, 480), ("plant", 700, 352),
        ("party-banner", 640, 160), ("disco-ball", 640, 96),
        ("ceiling-light", 160, 112), ("ceiling-light", 1120, 112), ("poster", 470, 440), ("poster", 810, 440),
    ])
    return build(desks, objects)


if __name__ == "__main__":
    for name, m in (("sprint", sprint()), ("party", party())):
        path = os.path.join(OUT, name + ".json")
        with open(path, "w") as f:
            json.dump(m, f, separators=(",", ":"))
        print("wrote", os.path.normpath(path))
