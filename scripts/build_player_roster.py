#!/usr/bin/env python3
"""Normalize imagegen's five distinct character sheets to the runtime contract."""
import json
import numpy as np
from PIL import Image
from build_sprites import ROOT, SOURCE, MASKS, save, strip, mask_strip, quantize, outlined, largest_component
from sprite_manifest import COLORS, PLAYER, rgb

PROFILES = {
    'teal': ('tall-slim', 54),
    'violet': ('short-round', 42),
    'amber': ('muscular', 50),
    'mint': ('petite', 44),
    'orange': ('stocky', 48),
}
MAPPING = {
    'idle': [0, 1, 2, 3], 'jump': [4, 5], 'fall': [6, 7],
    'run': list(range(8, 16)), 'attack': list(range(16, 21)),
    'hurt': [21, 22], 'dash': list(range(24, 28)),
    'down': list(range(28, 32)), 'revive': list(range(32, 36)),
}

def source_layout(source):
    """Use connected silhouettes, since generated spacing is not an exact grid."""
    a = np.array(source)[:, :, 3] > 127
    seen = np.zeros_like(a)
    h, w = a.shape
    rows = [[] for _ in range(5)]
    for sy, sx in zip(*np.where(a)):
        if seen[sy, sx]:
            continue
        stack = [(int(sy), int(sx))]
        seen[sy, sx] = True
        xs, ys = [], []
        while stack:
            y, x = stack.pop()
            xs.append(x); ys.append(y)
            for yy, xx in [(y-1,x), (y+1,x), (y,x-1), (y,x+1)]:
                if 0 <= yy < h and 0 <= xx < w and a[yy, xx] and not seen[yy, xx]:
                    seen[yy, xx] = True
                    stack.append((yy, xx))
        if len(xs) > 1000:
            box = [min(xs), min(ys), max(xs)+1, max(ys)+1]
            row = min(4, int((box[1]+box[3]) / 2 / h * 5))
            rows[row].append(box)
    rows = [sorted(row, key=lambda b: b[0]) for row in rows]
    counts = list(map(len, rows))
    if counts[0] not in (7, 8) or counts[1] not in (7, 8) or counts[2] != 7 or counts[3] not in (7, 8) or counts[4] != 4:
        raise ValueError(f'Unexpected source silhouettes per row: {counts}; inspect source before mapping')
    boxes = {row * 8 + col: box for row, cells in enumerate(rows) for col, box in enumerate(cells)}
    mapping = {name: indices.copy() for name, indices in MAPPING.items()}
    if counts[0] == 7:
        # The second tucked airborne pose also serves as the fall transition.
        mapping['fall'] = [5, 6]
    if counts[1] == 7:
        mapping['run'] = [8, 9, 10, 11, 4, 12, 13, 14]
    if counts[3] == 7:
        # A generated forward swing supplies the dash entry, then three lunges.
        mapping['dash'] = [19, 24, 25, 26]
        mapping['down'] = [27, 28, 29, 30]
    return {'boxes': boxes, 'animations': mapping, 'sourceRowCounts': counts}

def main():
    for color, (body, height) in PROFILES.items():
        source = Image.open(SOURCE / f'player-{color}-v2.png').convert('RGBA')
        raw = {}
        boxes = {}
        regions = source_layout(source)
        (SOURCE / f'player-{color}-v2-layout.json').write_text(json.dumps(regions, indent=2) + '\n')
        for key, box in regions['boxes'].items():
            i = int(key)
            cell = source.crop(box)
            a = np.array(cell)
            a[a[:, :, 3] < 128] = 0
            a[a[:, :, 3] >= 128, 3] = 255
            cell = Image.fromarray(a)
            if cell.getbbox():
                raw[i] = largest_component(cell)
                boxes[i] = box
        mapping = regions.get('animations', MAPPING)
        required = {i for indices in mapping.values() for i in indices}
        if required - raw.keys():
            raise ValueError(f'{color}: missing cells {required - raw.keys()}')
        scale = height / np.median([raw[i].height for i in range(4)])
        for name, indices in mapping.items():
            frames, masks = [], []
            for i in indices:
                crop = raw[i]
                # Convert only colored clothing to this player's exact ramp.
                a = np.array(crop)
                hi = a[:, :, :3].max(axis=2).astype(int)
                lo = a[:, :, :3].min(axis=2).astype(int)
                r, g, b = a[:, :, :3].astype(int).transpose(2, 0, 1)
                tests = {
                    'teal': (g > r * 1.38) & (b > r * 1.1) & (g > b * .88) & (g > 45),
                    'violet': (b > g * 1.25) & (r > g * 1.15),
                    'amber': (r > b * 1.8) & (g > b * 1.8) & (g > r * .65),
                    'mint': (g > r * 1.15) & (g > b * 1.05),
                    'orange': (r > g * 1.4) & (g > b * 1.3),
                }
                # Skin is common to every identity and must retain the base palette.
                skin = (r > g) & (g > b) & (r - g < 85) & (g - b < 75) & (b > 65)
                cloth = tests[color] & (hi - lo > 25) & (a[:, :, 3] > 0)
                if color in ('orange', 'amber'):
                    skin_palette = np.array([rgb(x) for x in ['F6C9A0','E0A274','A86B4A']])
                    distance = ((a[:, :, None, :3].astype(int) - skin_palette) ** 2).sum(axis=3).min(axis=2)
                    cloth &= ~(skin & (distance < 900))
                nw, nh = max(1, round(crop.width * scale)), max(1, round(crop.height * scale))
                ratio = min(1, 60 / nw, 60 / nh)
                size = (max(1, round(nw * ratio)), max(1, round(nh * ratio)))
                small = crop.resize(size, Image.Resampling.NEAREST)
                mask = Image.fromarray((cloth * 255).astype('uint8')).resize(size, Image.Resampling.NEAREST)
                art, _ = quantize(small)
                ar = np.array(art)
                visible_cloth = np.array(mask) > 0
                ramp = np.array([rgb(c) for c in COLORS[color]])
                # Source luminance determines shadow/base/highlight, independent of hue.
                lum = np.array(small)[:, :, :3] @ np.array([.2126, .7152, .0722])
                levels = np.digitize(lum, [85, 155])
                ar[visible_cloth, :3] = ramp[levels[visible_cloth]]
                art = outlined(Image.fromarray(ar))
                canvas = Image.new('RGBA', (64, 64))
                canvas.paste(art, ((64 - size[0]) // 2, 64 - size[1]))
                frames.append(canvas)
                ca = np.array(canvas)
                actual = np.any(np.all(ca[:, :, None, :3] == ramp, axis=3), axis=2) & (ca[:, :, 3] > 0)
                masks.append(Image.fromarray((actual * 255).astype('uint8')))
            save(strip(frames), f'assets/sprites/player/{color}/{name}.png')
            MASKS.mkdir(exist_ok=True, parents=True)
            mask_strip(masks).save(MASKS / f'player-{color}-{name}.png')
        # Portrait uses the same identity as its idle animation.
        idle = Image.open(ROOT / f'assets/sprites/player/{color}/idle.png').crop((0, 0, 64, 64))
        crop = idle.crop(idle.getbbox())
        face = crop.crop((0, 0, crop.width, round(crop.height * .52)))
        ratio = min(22 / face.width, 23 / face.height)
        face = face.resize((max(1, round(face.width * ratio)), max(1, round(face.height * ratio))), Image.Resampling.NEAREST)
        portrait = Image.new('RGBA', (24, 24))
        portrait.paste(face, ((24 - face.width) // 2, 24 - face.height))
        suffix = '' if color == 'teal' else f'-{color}'
        save(portrait, f'assets/ui/portrait{suffix}.png')
        pa = np.array(portrait)
        ramp = np.array([rgb(c) for c in COLORS[color]])
        mask = np.any(np.all(pa[:, :, None, :3] == ramp, axis=3), axis=2) & (pa[:, :, 3] > 0)
        Image.fromarray((mask * 255).astype('uint8')).save(MASKS / f'portrait-{color}.png')
        (SOURCE / f'player-{color}-v2-regions.json').write_text(json.dumps(boxes, indent=2) + '\n')
        print(f'{color}: {body}, nine animation strips + portrait')

if __name__ == '__main__':
    main()
