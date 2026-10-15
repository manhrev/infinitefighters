#!/usr/bin/env python3
"""Build Boss SA Vought runtime sprite strips and attack VFX from source sheets."""
import json
import numpy as np
from PIL import Image, ImageDraw

from build_sprites import ROOT, SOURCE, rgb, save, strip, quantize, outlined, fit
from sprite_manifest import entries

NEAREST = Image.Resampling.NEAREST

BOSS_VOUGHT = {
    'idle': (4, 6, True),
    'walk': (4, 8, True),
    'deflect_windup': (4, 8, False),
    'deflect': (4, 12, False),
    'stamp_throw': (4, 12, False),
    'shoulder_pat': (2, 8, False),
    'reflect_shield': (4, 10, True),
    'recover': (2, 6, True),
    'hurt': (2, 10, False),
    'defeat': (6, 8, False),
}

BOSS_VOUGHT_VFX = {
    'vought-deflect-ribbon': (48, 48, 4, 12, True),
    'vought-bug-stamp': (32, 32, 4, 12, True),
    'vought-speech-shield': (48, 48, 4, 10, True),
    'vought-pat-mark': (32, 32, 4, 8, True),
}

def remove_magenta_bg(im):
    arr = np.array(im.convert('RGB'))
    r, g, b = arr[:, :, 0].astype(int), arr[:, :, 1].astype(int), arr[:, :, 2].astype(int)
    is_magenta = (r > 180) & (b > 200) & (g > 70) & (g < 150) & (b > g * 1.5) & (r > g * 1.5)
    out = np.dstack([arr, np.where(is_magenta, 0, 255).astype(np.uint8)])
    out[out[:, :, 3] == 0, :3] = 0
    return Image.fromarray(out)

def remove_vfx_magenta_bg(im):
    arr = np.array(im.convert('RGBA'))
    r, g, b = arr[:, :, 0], arr[:, :, 1], arr[:, :, 2]
    is_m = (r > 180) & (b > 180) & (g < 60)
    arr[is_m, 3] = 0
    arr[~is_m, 3] = 255
    arr[arr[:, :, 3] == 0, :3] = 0
    return Image.fromarray(arr)

def extract_vought_character_cells():
    source = remove_magenta_bg(Image.open(SOURCE / 'boss-vought.png'))
    row_spans = [(15, 171), (186, 342), (357, 512), (526, 682), (694, 851), (858, 1021)]
    
    col_cuts = [
        # Row 0: 8 cols (idle 0-3, walk 4-7)
        [(33, 98), (158, 246), (279, 402), (446, 536), (568, 618), (665, 739), (794, 867), (922, 998)],
        # Row 1: 7 cols (walk 0-2, deflect_windup 3-6)
        [(29, 103), (166, 222), (284, 388), (438, 534), (586, 685), (718, 849), (888, 1000)],
        # Row 2: 6 cols (deflect 0-3, hurt 4-5)
        [(12, 132), (168, 308), (333, 477), (507, 654), (725, 824), (894, 987)],
        # Row 3: 7 cols (stamp_throw 0-3, stamp_projectile 4, shoulder_pat 5-6)
        [(21, 113), (165, 276), (332, 463), (497, 630), (649, 674), (712, 840), (876, 1001)],
        # Row 4: 6 cols (reflect_shield 0-3, recover 4-5)
        [(22, 140), (182, 299), (342, 459), (502, 619), (726, 812), (894, 988)],
        # Row 5: 6 cols (defeat 0-5)
        [(29, 142), (194, 283), (330, 452), (494, 622), (653, 806), (829, 1013)],
    ]

    cells = {}
    boxes = []
    for r, (y0, y1) in enumerate(row_spans):
        for c, (x0, x1) in enumerate(col_cuts[r]):
            crop = source.crop((x0, y0, x1, y1))
            bbox = crop.getbbox()
            if bbox:
                actual = crop.crop(bbox)
                cells[(r, c)] = actual
                boxes.append([x0 + bbox[0], y0 + bbox[1], x0 + bbox[2], y0 + bbox[3]])

    (SOURCE / 'boss-vought-regions.json').write_text(json.dumps(boxes, indent=2) + '\n')
    return cells

def extract_vought_vfx_cells():
    source = remove_vfx_magenta_bg(Image.open(SOURCE / 'boss-vought-vfx.png'))
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
    (SOURCE / 'boss-vought-vfx-regions.json').write_text(json.dumps(boxes, indent=2) + '\n')
    return vfx_cells

def build_character_sprites(cells):
    mapping = {
        'idle': [(0, 0), (0, 1), (0, 2), (0, 3)],
        'walk': [(0, 4), (0, 5), (0, 6), (0, 7)],
        'deflect_windup': [(1, 3), (1, 4), (1, 5), (1, 6)],
        'deflect': [(2, 0), (2, 1), (2, 2), (2, 3)],
        'hurt': [(2, 4), (2, 5)],
        'stamp_throw': [(3, 0), (3, 1), (3, 2), (3, 3)],
        'shoulder_pat': [(3, 5), (3, 6)],
        'reflect_shield': [(4, 0), (4, 1), (4, 2), (4, 3)],
        'recover': [(4, 4), (4, 5)],
        'defeat': [(5, 0), (5, 1), (5, 2), (5, 3), (5, 4), (5, 5)],
    }

    out_frames = {}
    for anim_name, keys in mapping.items():
        frames = []
        for key in keys:
            crop = cells[key]
            # Height ~ 95 px (1.6x player, slender build), max width 152
            target_h = min(150, round(crop.height * 95 / 150))
            if anim_name == 'defeat' and key[1] >= 4:
                # Ground collapsed poses
                target_h = min(150, round(crop.height * 70 / 80))
            f, _ = fit(crop, 160, 160, target_h, danger=False, max_w=152)
            frames.append(f)
        out_frames[anim_name] = frames
        save(strip(frames), f'assets/sprites/boss/vought/{anim_name}.png')

    return out_frames

def build_vfx_sprites(vfx_cells):
    mapping = {
        'vought-deflect-ribbon': [(0, 0), (0, 1), (0, 2), (0, 3)],
        'vought-bug-stamp': [(1, 0), (1, 1), (1, 2), (1, 3)],
        'vought-speech-shield': [(2, 0), (2, 1), (2, 2), (2, 3)],
        'vought-pat-mark': [(3, 0), (3, 1), (3, 2), (3, 3)],
    }

    out_vfx = {}
    for name, (w, h, count, fps, loop) in BOSS_VOUGHT_VFX.items():
        keys = mapping[name]
        frames = []
        is_hazard = (name in ['vought-bug-stamp', 'vought-pat-mark'])
        for key in keys:
            crop = vfx_cells[key]
            scale = min((w - 4) / crop.width, (h - 4) / crop.height)
            nw = max(1, round(crop.width * scale))
            nh = max(1, round(crop.height * scale))
            small = crop.resize((nw, nh), NEAREST)
            small, _ = quantize(small, danger=is_hazard)
            small = outlined(small)
            canvas = Image.new('RGBA', (w, h))
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

    save(preview, 'assets/preview/boss-vought.png')

    vfx_sheet = Image.new('RGBA', (4 * 100 + 40, 4 * 80 + 40), rgb('0B1530') + (255,))
    vfx_draw = ImageDraw.Draw(vfx_sheet)
    for r, (name, frames) in enumerate(vfx_frames.items()):
        y = 10 + r * 80
        vfx_draw.text((10, y), name, fill=rgb('3FE0FF') + (255,))
        for i, f in enumerate(frames):
            enlarged = f.resize((f.width * 2, f.height * 2), NEAREST)
            vfx_sheet.alpha_composite(enlarged, (10 + i * 100, y + 16))
    save(vfx_sheet, 'assets/preview/boss-vought-vfx.png')

def main():
    cells = extract_vought_character_cells()
    vfx_cells = extract_vought_vfx_cells()
    char_frames = build_character_sprites(cells)
    vfx_frames = build_vfx_sprites(vfx_cells)
    make_preview(char_frames, vfx_frames)

    manifest_path = ROOT / 'assets/sprites/manifest.json'
    manifest = json.loads(manifest_path.read_text())
    manifest['entries'] = entries()
    manifest_path.write_text(json.dumps(manifest, indent=2) + '\n')
    print('Boss SA Vought: 10 animation strips, 4 VFX strips, previews and manifest updated successfully.')

if __name__ == '__main__':
    main()
