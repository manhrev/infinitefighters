#!/usr/bin/env python3
"""Turn a raw imagegen sheet (one or more rows of poses on magenta) into a runtime-style sprite strip.

    .venv-sprites/bin/python engine-sandbox/tools/build_v2_strip.py RAW.jpg --count 6 --out OUT_DIR/idle \
        --height 120 --fps 6

Same steps as scripts/build_boss_*.py: remove the magenta background, isolate the poses, scale them with ONE shared
nearest-neighbour factor, snap to the project palette (BASE), add the 1 px outline, then paste each pose into a
FRAME x FRAME cell with the feet on the bottom row. Writes OUT.png (strip), OUT-preview.png and OUT.gif (at --fps).
"""
import argparse
import pathlib
import sys

import numpy as np
from PIL import Image

ROOT = pathlib.Path(__file__).resolve().parents[2]
sys.path.insert(0, str(ROOT / "scripts"))
from build_boss_khabib import remove_magenta_bg  # noqa: E402
from build_sprites import largest_component, outlined, quantize, segments, strip  # noqa: E402

NEAREST = Image.Resampling.NEAREST


def isolate(raw: Image.Image, count: int, rows: int) -> list[Image.Image]:
    """Poses in reading order: `rows` bands top to bottom, count // rows poses per band."""
    keyed = remove_magenta_bg(raw)
    opaque = np.array(keyed)[:, :, 3] > 0
    bands = segments(opaque.sum(axis=1) > 3, 12)
    if len(bands) != rows:
        sys.exit(f"expected {rows} rows, found {len(bands)} at {bands}; inspect the source")
    poses = []
    for y0, y1 in bands:
        cols = segments(opaque[y0:y1].sum(axis=0) > 3, 12)
        if len(cols) != count // rows:
            sys.exit(f"expected {count // rows} poses per row, found {len(cols)} at columns {cols}; inspect the source")
        poses += [largest_component(keyed.crop((x0, y0, x1, y1))) for x0, x1 in cols]
    return poses


def render(poses: list[Image.Image], frame: int, height: int) -> list[Image.Image]:
    # One scale for the whole strip: per-pose scaling would make the body size pulse between frames.
    scale = height / max(p.height for p in poses)
    cells = []
    for p in poses:
        small = p.resize((max(1, round(p.width * scale)), max(1, round(p.height * scale))), NEAREST)
        small = outlined(quantize(small)[0])
        a = np.array(small)
        # Centre on the body mass, not the bbox, so spread legs do not shift the torso between frames.
        cx = int(round(np.nonzero(a[:, :, 3])[1].mean()))
        x = min(max(frame // 2 - cx, 0), frame - small.width)
        cell = Image.new("RGBA", (frame, frame))
        cell.paste(small, (x, frame - small.height))
        cells.append(cell)
    return cells


def report(cells: list[Image.Image]) -> None:
    a = [np.array(c) for c in cells]
    colors = len({tuple(px) for arr in a for px in arr[arr[:, :, 3] > 0][:, :3]})
    feet = [int(np.nonzero(x[:, :, 3].any(axis=1))[0].max()) for x in a]
    heights = [int(np.ptp(np.nonzero(x[:, :, 3].any(axis=1))[0]) + 1) for x in a]
    widths = [int(np.ptp(np.nonzero(x[:, :, 3].any(axis=0))[0]) + 1) for x in a]
    alpha = sorted({int(v) for x in a for v in np.unique(x[:, :, 3])})
    dup = [i + 1 for i in range(1, len(a)) if np.array_equal(a[i], a[i - 1])]
    print(f"frames {len(cells)}  colors {colors}  alpha {alpha}")
    aspect = float(np.mean([w / h for w, h in zip(widths, heights)]))
    print(f"aspect w/h {aspect:.2f}  (original Khabib: idle 1.03, walk 0.78; a much lower value means a mismatched, slimmer body)")
    print(f"feet rows {feet}  heights {heights}  widths {widths}  duplicate-of-previous {dup or 'none'}")


def main() -> None:
    ap = argparse.ArgumentParser(description=__doc__, formatter_class=argparse.RawDescriptionHelpFormatter)
    ap.add_argument("raw", type=pathlib.Path)
    ap.add_argument("--count", type=int, required=True, help="total number of poses")
    ap.add_argument("--rows", type=int, default=1, help="rows of poses in the source image (count must divide)")
    ap.add_argument("--out", type=pathlib.Path, required=True, help="output path without extension")
    ap.add_argument("--frame", type=int, default=160)
    ap.add_argument("--height", type=int, default=120, help="body height of the tallest pose, in output pixels")
    ap.add_argument("--fps", type=float, default=10)
    a = ap.parse_args()

    cells = render(isolate(Image.open(a.raw), a.count, a.rows), a.frame, a.height)
    a.out.parent.mkdir(parents=True, exist_ok=True)
    strip(cells).save(a.out.with_suffix(".png"))
    report(cells)

    # Preview on the game's navy at 3x, and a looping GIF at the target fps to judge the motion.
    bg = (20, 35, 74, 255)
    frames = []
    for c in cells:
        f = Image.new("RGBA", c.size, bg)
        f.alpha_composite(c)
        frames.append(f.resize((c.width * 3, c.height * 3), NEAREST).convert("P"))
    prev = Image.new("RGBA", (a.frame * len(cells), a.frame), bg)
    prev.alpha_composite(strip(cells))
    prev.resize((prev.width * 2, prev.height * 2), NEAREST).save(a.out.parent / f"{a.out.name}-preview.png")
    frames[0].save(a.out.with_suffix(".gif"), save_all=True, append_images=frames[1:], loop=0,
                   duration=round(1000 / a.fps), disposal=2)
    print(f"wrote {a.out.with_suffix('.png')}, {a.out.name}-preview.png, {a.out.with_suffix('.gif')}")


if __name__ == "__main__":
    main()
