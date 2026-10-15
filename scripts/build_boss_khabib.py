#!/usr/bin/env python3
"""Build Boss Khabib runtime sprite strips and attack VFX from source sheets."""
import json
import numpy as np
from PIL import Image, ImageDraw

from build_sprites import ROOT, SOURCE, rgb, save, strip, quantize, outlined, fit
from sprite_manifest import entries

NEAREST = Image.Resampling.NEAREST

BOSS_KHABIB = {
    'idle': (4, 6, True),
    'walk': (4, 8, True),
    'press_windup': (4, 8, False),
    'press': (4, 12, False),
    'spam_windup': (4, 8, False),
    'spam': (4, 12, True),
    'teleport_out': (2, 8, False),
    'teleport_in': (2, 10, False),
    'recover': (2, 6, True),
    'hurt': (2, 10, False),
    'defeat': (6, 8, False),
}

BOSS_KHABIB_VFX = {
    'khabib-press-wave': (48, 48, 4, 14, True),
    'khabib-nag-projectile': (32, 24, 4, 12, True),
    'khabib-teleport-warning': (64, 32, 4, 8, True),
    'khabib-portal': (48, 48, 4, 12, False),
}

def remove_magenta_bg(im):
    a = np.array(im.convert('RGBA'))
    r, g, b = a[:, :, :3].astype(int).transpose(2, 0, 1)
    magenta = (r > 130) & (b > 130) & (g < 110) & (r > g * 1.3) & (b > g * 1.3)
    a[:, :, 3] = np.where(magenta | (a[:, :, 3] < 128), 0, 255)
    a[a[:, :, 3] == 0, :3] = 0
    return Image.fromarray(a)

def extract_khabib_character_cells():
    source = remove_magenta_bg(Image.open(SOURCE / 'boss-khabib.png'))
    y_cuts = [0, 175, 350, 520, 695, 860, 1024]
    x_cuts = [
        [0, 168, 338, 508, 680, 835, 1024],
        [0, 165, 335, 510, 675, 850, 1024],
        [0, 185, 355, 525, 690, 855, 1024],
        [0, 172, 342, 510, 680, 852, 1024],
        [0, 175, 340, 510, 680, 850, 1024],
        [0, 175, 370, 580, 590, 795, 1024]
    ]
    cells = {}
    boxes = []
    for r in range(6):
        y0, y1 = y_cuts[r], y_cuts[r+1]
        xs = x_cuts[r]
        for c in range(len(xs)-1):
            x0, x1 = xs[c], xs[c+1]
            crop = source.crop((x0, y0, x1, y1))
            bbox = crop.getbbox()
            if bbox:
                actual_crop = crop.crop(bbox)
                cells[(r, c)] = actual_crop
                boxes.append([x0 + bbox[0], y0 + bbox[1], x0 + bbox[2], y0 + bbox[3]])
    (SOURCE / 'boss-khabib-regions.json').write_text(json.dumps(boxes, indent=2) + '\n')
    return cells

def extract_khabib_vfx_cells():
    source = remove_magenta_bg(Image.open(SOURCE / 'boss-khabib-vfx.png'))
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
                actual_crop = crop.crop(bbox)
                vfx_cells[(r, c)] = actual_crop
                boxes.append([x0 + bbox[0], y0 + bbox[1], x0 + bbox[2], y0 + bbox[3]])
    (SOURCE / 'boss-khabib-vfx-regions.json').write_text(json.dumps(boxes, indent=2) + '\n')
    return vfx_cells

def build_character_sprites(cells):
    mapping = {
        'idle': [(0, 0), (0, 1), (0, 2), (0, 3)],
        'walk': [(0, 4), (0, 5), (1, 0), (1, 1)],
        'press_windup': [(1, 2), (1, 3), (1, 4), (1, 5)],
        'press': [(2, 0), (2, 1), (2, 2), (2, 3)],
        'hurt': [(2, 4), (2, 5)],
        'spam_windup': [(3, 0), (3, 1), (3, 2), (3, 3)],
        'spam': [(3, 2), (3, 3), (3, 4), (3, 5)],
        'teleport_out': [(4, 0), (4, 1)],
        'teleport_in': [(4, 2), (4, 3)],
        'recover': [(4, 4), (4, 5)],
    }
    
    out_frames = {}
    for anim_name, keys in mapping.items():
        frames = []
        for key in keys:
            crop = cells[key]
            target_h = min(150, round(crop.height * 120 / 150))
            f, _ = fit(crop, 160, 160, target_h, danger=False, max_w=152)
            frames.append(f)
        out_frames[anim_name] = frames
        save(strip(frames), f'assets/sprites/boss/khabib/{anim_name}.png')

    # Defeat animation: 6 sequential distinct poses
    f0, _ = fit(cells[(5, 0)], 160, 160, 116, danger=False, max_w=152)
    f1, _ = fit(cells[(5, 1)], 160, 160, 114, danger=False, max_w=152)
    f2, _ = fit(cells[(5, 2)], 160, 160, 118, danger=False, max_w=152)
    c_fall = cells[(5, 2)].rotate(-10, resample=NEAREST, expand=True)
    f3, _ = fit(c_fall, 160, 160, 105, danger=False, max_w=152)
    f4, _ = fit(cells[(5, 4)], 160, 160, 80, danger=False, max_w=152)
    f5, _ = fit(cells[(5, 5)], 160, 160, 78, danger=False, max_w=152)

    defeat_frames = [f0, f1, f2, f3, f4, f5]
    out_frames['defeat'] = defeat_frames
    save(strip(defeat_frames), 'assets/sprites/boss/khabib/defeat.png')
    return out_frames

def build_vfx_sprites(vfx_cells):
    mapping = {
        'khabib-press-wave': [(0, 0), (0, 1), (0, 2), (0, 3)],
        'khabib-nag-projectile': [(1, 0), (1, 1), (1, 2), (1, 3)],
        'khabib-teleport-warning': [(2, 0), (2, 1), (2, 2), (2, 3)],
        'khabib-portal': [(3, 0), (3, 1), (3, 2), (3, 3)],
    }
    
    out_vfx = {}
    for name, (w, h, count, fps, loop) in BOSS_KHABIB_VFX.items():
        keys = mapping[name]
        frames = []
        is_hazard = (name in ['khabib-press-wave', 'khabib-nag-projectile', 'khabib-teleport-warning'])
        for key in keys:
            crop = vfx_cells[key]
            scale = min((w - 4) / crop.width, (h - 4) / crop.height)
            nw = max(1, round(crop.width * scale))
            nh = max(1, round(crop.height * scale))
            small = crop.resize((nw, nh), NEAREST)
            small, _ = quantize(small, danger=is_hazard)
            small = outlined(small)
            canvas = Image.new('RGBA', (w, h))
            if name == 'khabib-teleport-warning':
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
    
    save(preview, 'assets/preview/boss-khabib.png')

    vfx_sheet = Image.new('RGBA', (4 * 100 + 40, 4 * 80 + 40), rgb('0B1530') + (255,))
    vfx_draw = ImageDraw.Draw(vfx_sheet)
    for r, (name, frames) in enumerate(vfx_frames.items()):
        y = 10 + r * 80
        vfx_draw.text((10, y), name, fill=rgb('3FE0FF') + (255,))
        for i, f in enumerate(frames):
            enlarged = f.resize((f.width * 2, f.height * 2), NEAREST)
            vfx_sheet.alpha_composite(enlarged, (10 + i * 100, y + 16))
    save(vfx_sheet, 'assets/preview/boss-khabib-vfx.png')

def main():
    cells = extract_khabib_character_cells()
    vfx_cells = extract_khabib_vfx_cells()
    char_frames = build_character_sprites(cells)
    vfx_frames = build_vfx_sprites(vfx_cells)
    make_preview(char_frames, vfx_frames)
    
    manifest_path = ROOT / 'assets/sprites/manifest.json'
    manifest = json.loads(manifest_path.read_text())
    manifest['entries'] = entries()
    manifest_path.write_text(json.dumps(manifest, indent=2) + '\n')
    print('Boss Khabib: 11 animation strips, 4 VFX strips, previews and manifest updated successfully.')

if __name__ == '__main__':
    main()
