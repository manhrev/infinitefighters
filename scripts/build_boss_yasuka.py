#!/usr/bin/env python3
"""Build Boss Yasuka runtime sprite strips and attack VFX from source sheets."""
import json
import numpy as np
from PIL import Image, ImageDraw

from build_sprites import ROOT, SOURCE, rgb, save, strip, quantize, outlined, fit
from sprite_manifest import entries

NEAREST = Image.Resampling.NEAREST

BOSS_YASUKA = {
    'idle': (4, 6, True),
    'walk': (4, 8, True),
    'dodge_windup': (4, 8, False),
    'dodge': (4, 12, False),
    'procedure': (4, 12, False),
    'silence': (4, 8, True),
    'recover': (2, 6, True),
    'hurt': (2, 10, False),
    'defeat': (6, 8, False),
}

BOSS_YASUKA_VFX = {
    'yasuka-silence-dome': (64, 48, 4, 10, True),
    'yasuka-dodge-trail': (48, 48, 4, 12, False),
    'yasuka-stamp-trap': (32, 32, 4, 8, True),
    'yasuka-paper-flurry': (48, 48, 4, 12, False),
}

def remove_magenta_bg(im):
    arr = np.array(im.convert('RGBA'))
    r, g, b = arr[:, :, 0].astype(int), arr[:, :, 1].astype(int), arr[:, :, 2].astype(int)
    is_m = (r > 160) & (b > 160) & (r > g * 1.3) & (b > g * 1.3)
    arr[is_m, 3] = 0
    arr[~is_m, 3] = 255
    arr[arr[:, :, 3] == 0, :3] = 0
    return Image.fromarray(arr)

def extract_yasuka_character_cells():
    source = remove_magenta_bg(Image.open(SOURCE / 'boss-yasuka.png'))
    w, h = source.size
    row_h = h // 6
    col_w = w // 6
    cells = {}
    boxes = []
    for r in range(6):
        for c in range(6):
            cell = source.crop((c * col_w, r * row_h, (c + 1) * col_w, (r + 1) * row_h))
            bbox = cell.getbbox()
            if bbox:
                actual = cell.crop(bbox)
                cells[(r, c)] = actual
                boxes.append([c * col_w + bbox[0], r * row_h + bbox[1], c * col_w + bbox[2], r * row_h + bbox[3]])
    (SOURCE / 'boss-yasuka-regions.json').write_text(json.dumps(boxes, indent=2) + '\n')
    return cells

def extract_yasuka_vfx_cells():
    source = remove_magenta_bg(Image.open(SOURCE / 'boss-yasuka-vfx.png'))
    w, h = source.size
    row_h = h // 4
    col_w = w // 4
    vfx_cells = {}
    boxes = []
    for r in range(4):
        for c in range(4):
            x0, y0, x1, y1 = c * col_w, r * row_h, (c + 1) * col_w, (r + 1) * row_h
            crop = source.crop((x0, y0, x1, y1))
            bbox = crop.getbbox()
            if bbox:
                actual = crop.crop(bbox)
                vfx_cells[(r, c)] = actual
                boxes.append([x0 + bbox[0], y0 + bbox[1], x0 + bbox[2], y0 + bbox[3]])
    (SOURCE / 'boss-yasuka-vfx-regions.json').write_text(json.dumps(boxes, indent=2) + '\n')
    return vfx_cells

def build_character_sprites(cells):
    mapping = {
        'idle': [(0, 0), (0, 1), (0, 2), (0, 3)],
        'walk': [(0, 4), (0, 5), (1, 0), (1, 1)],
        'dodge_windup': [(1, 2), (1, 3), (1, 4), (1, 5)],
        'dodge': [(2, 0), (2, 1), (2, 2), (2, 3)],
        'hurt': [(2, 4), (2, 5)],
        'procedure': [(3, 0), (3, 1), (3, 2), (3, 3)],
        'silence': [(4, 0), (4, 1), (4, 2), (4, 3)],
        'recover': [(4, 4), (4, 5)],
        'defeat': [(5, 0), (5, 1), (5, 2), (5, 3), (5, 4), (5, 5)],
    }

    out_frames = {}
    for anim_name, keys in mapping.items():
        frames = []
        for key in keys:
            crop = cells[key]
            # Yasuka upright height ~ 100 px, max width 152
            target_h = min(150, round(crop.height * 100 / 150))
            if anim_name == 'defeat' and key[1] >= 3:
                # Collapsed poses on ground
                target_h = min(150, round(crop.height * 80 / 100))
            f, _ = fit(crop, 160, 160, target_h, danger=False, max_w=152)
            frames.append(f)
        out_frames[anim_name] = frames
        save(strip(frames), f'assets/sprites/boss/yasuka/{anim_name}.png')

    return out_frames

def build_vfx_sprites(vfx_cells):
    mapping = {
        'yasuka-silence-dome': [(0, 0), (0, 1), (0, 2), (0, 3)],
        'yasuka-dodge-trail': [(1, 0), (1, 1), (1, 2), (1, 3)],
        'yasuka-stamp-trap': [(2, 0), (2, 1), (2, 2), (2, 3)],
        'yasuka-paper-flurry': [(3, 0), (3, 1), (3, 2), (3, 3)],
    }

    out_vfx = {}
    for name, (w, h, count, fps, loop) in BOSS_YASUKA_VFX.items():
        keys = mapping[name]
        frames = []
        is_hazard = (name in ['yasuka-silence-dome', 'yasuka-stamp-trap'])
        for key in keys:
            crop = vfx_cells[key]
            scale = min((w - 4) / crop.width, (h - 4) / crop.height)
            nw = max(1, round(crop.width * scale))
            nh = max(1, round(crop.height * scale))
            small = crop.resize((nw, nh), NEAREST)
            small, _ = quantize(small, danger=is_hazard)
            small = outlined(small)
            canvas = Image.new('RGBA', (w, h))
            if name == 'yasuka-silence-dome' or name == 'yasuka-stamp-trap':
                canvas.paste(small, ((w - nw) // 2, h - nh))
            else:
                canvas.paste(small, ((w - nw) // 2, (h - nh) // 2))
            frames.append(canvas)
        out_vfx[name] = frames
        save(strip(frames), f'assets/sprites/vfx/{name}.png')
    return out_vfx

def make_preview(char_frames, vfx_frames):
    rows = len(char_frames)
    row_h = 180
    sheet_w = 6 * 170 + 20
    sheet_h = rows * row_h + 40
    preview = Image.new('RGBA', (sheet_w, sheet_h), rgb('0B1530') + (255,))
    draw = ImageDraw.Draw(preview)

    for r, (anim_name, frames) in enumerate(char_frames.items()):
        y = 20 + r * row_h
        draw.text((10, y), f'{anim_name} ({len(frames)} frames)', fill=rgb('F2E3C2') + (255,))
        for i, f in enumerate(frames):
            preview.alpha_composite(f, (10 + i * 170, y + 16))

    save(preview, 'assets/preview/boss-yasuka.png')

    vfx_sheet = Image.new('RGBA', (4 * 100 + 40, 4 * 80 + 40), rgb('0B1530') + (255,))
    vfx_draw = ImageDraw.Draw(vfx_sheet)
    for r, (name, frames) in enumerate(vfx_frames.items()):
        y = 10 + r * 80
        vfx_draw.text((10, y), name, fill=rgb('3FE0FF') + (255,))
        for i, f in enumerate(frames):
            enlarged = f.resize((f.width * 2, f.height * 2), NEAREST)
            vfx_sheet.alpha_composite(enlarged, (10 + i * 100, y + 16))
    save(vfx_sheet, 'assets/preview/boss-yasuka-vfx.png')

def main():
    cells = extract_yasuka_character_cells()
    vfx_cells = extract_yasuka_vfx_cells()
    char_frames = build_character_sprites(cells)
    vfx_frames = build_vfx_sprites(vfx_cells)
    make_preview(char_frames, vfx_frames)

    manifest_path = ROOT / 'assets/sprites/manifest.json'
    manifest = json.loads(manifest_path.read_text())
    manifest['entries'] = entries()
    manifest_path.write_text(json.dumps(manifest, indent=2) + '\n')
    print('Boss Yasuka: 9 animation strips, 4 VFX strips, previews and manifest updated successfully.')

if __name__ == '__main__':
    main()
