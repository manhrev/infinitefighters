#!/usr/bin/env python3
"""Normalize generated skill sheets, preserving v2 player sprites unchanged."""
import json
import numpy as np
from PIL import Image, ImageDraw
from build_sprites import ROOT, SOURCE, save, strip, outlined, segments
from sprite_manifest import BASE, COLORS, SKILLS, rgb, entries

NEAREST = Image.Resampling.NEAREST

def palette_image(image, colors):
    a = np.array(image.convert('RGBA'))
    p = np.array([rgb(c) for c in colors], dtype=int)
    src = a[:,:,:3].astype(int)
    distance = ((src[:,:,None,:] - p[None,None,:,:]) ** 2).sum(axis=3)
    a[:,:,:3] = p[distance.argmin(axis=2)]
    a[:,:,3] = np.where(a[:,:,3] >= 128, 255, 0)
    a[a[:,:,3] == 0] = 0
    return Image.fromarray(a)

def fit_frames(raw, size, colors, character=False, height=None):
    if character:
        scale = height / raw[0].height
    else:
        scale = min((size[0]-4)/max(f.width for f in raw), (size[1]-4)/max(f.height for f in raw))
    scale = min(scale, (size[0]-4)/max(f.width for f in raw), (size[1]-4)/max(f.height for f in raw))
    result = []
    for image in raw:
        width, h = max(1, round(image.width*scale)), max(1, round(image.height*scale))
        image = palette_image(image.resize((width,h), NEAREST), colors)
        if character:
            image = outlined(image)
        frame = Image.new('RGBA', size)
        frame.paste(image, ((size[0]-width)//2, size[1]-h if character else (size[1]-h)//2))
        result.append(frame)
    return result

def main():
    safe_base = [c for c in BASE if c not in ('FF5A5A','B8303A')]
    preview = Image.new('RGB', (1600, 5*620), '#0B1530')
    draw = ImageDraw.Draw(preview)
    for row, (color, skill) in enumerate(SKILLS.items()):
        source = Image.open(SOURCE/f'skill-{color}.png').convert('RGBA')
        opaque = np.array(source)[:,:,3] >= 128
        rows = segments(opaque.sum(axis=1) > 10, 6)
        merged = []
        for start, end in rows:
            if merged and start - merged[-1][1] <= 2:
                merged[-1] = (merged[-1][0], end)
            else:
                merged.append((start, end))
        rows = merged
        if len(rows) != 4:
            raise ValueError(f'{color}: expected four source rows, got {rows}; inspect source')
        raw, boxes = [], []
        for i in range(16):
            x,y = i%4, i//4
            box = (round(x*source.width/4), rows[y][0], round((x+1)*source.width/4), rows[y][1])
            if y < 2:
                columns = segments(opaque[rows[y][0]:rows[y][1]].sum(axis=0) > 3, 10)
                if len(columns) != 4:
                    raise ValueError(f'{color}: expected four isolated character silhouettes in row {y}, got {columns}')
                box = (columns[x][0], rows[y][0], columns[x][1], rows[y][1])
            cell = source.crop(box)
            a = np.array(cell)
            a[a[:,:,3] < 128] = 0
            cell = Image.fromarray(a)
            bbox = cell.getbbox()
            if not bbox:
                raise ValueError(f'{color}: empty skill cell {i}; inspect source layout')
            raw.append(cell.crop(bbox))
            boxes.append([box[0]+bbox[0],box[1]+bbox[1],box[0]+bbox[2],box[1]+bbox[3]])
        (SOURCE/f'skill-{color}-regions.json').write_text(json.dumps(boxes,indent=2)+'\n')
        poses = fit_frames(raw[:8], (64,64), safe_base+COLORS[color], character=True, height=skill['height'])
        save(strip(poses[:4]),f'assets/sprites/player/{color}/skill-windup.png')
        save(strip(poses[4:]),f'assets/sprites/player/{color}/skill-cast.png')
        effects=[]
        for effect_row, (name,w,h,fps,loop) in enumerate(skill['effects']):
            colors = safe_base + (COLORS['mint'] if name=='hotfix-zone' else [])
            frames = fit_frames(raw[8+effect_row*4:12+effect_row*4],(w,h),colors)
            save(strip(frames),f'assets/sprites/vfx/skills/{name}.png')
            effects.append(frames)
        icon = fit_frames([raw[8]],(24,24),safe_base)[0]
        save(icon,f'assets/ui/skills/{skill["id"]}.png')
        y = row*620
        draw.text((12,y+8),f'{color}: {skill["id"]} / windup, cast, primary VFX, secondary VFX',fill='#F2E3C2')
        for j,frames in enumerate([poses[:4],poses[4:]]+effects):
            for i,frame in enumerate(frames):
                enlarged = frame.resize((frame.width*2,frame.height*2),NEAREST)
                preview.paste(enlarged,(20+i*390,y+35+j*145),enlarged)
        preview.crop((0,y,1600,y+620)).save(ROOT/f'assets/preview/skill-{color}.png')
        print(f'{color}: windup, cast, two VFX strips, icon')
    save(preview,'assets/preview/skills-contact-sheet.png')
    overview = Image.new('RGB',(1500,340),'#0B1530')
    labels = ImageDraw.Draw(overview)
    for i,(color,skill) in enumerate(SKILLS.items()):
        x = i*300
        labels.text((x+12,12),f'{color}: {skill["id"]}',fill='#F2E3C2')
        cast = Image.open(ROOT/f'assets/sprites/player/{color}/skill-cast.png').crop((64,0,128,64))
        cast = cast.resize((192,192),NEAREST)
        overview.paste(cast,(x+12,30),cast)
        name,w,h,_,_ = skill['effects'][0]
        effect = Image.open(ROOT/f'assets/sprites/vfx/skills/{name}.png').crop((w,0,2*w,h))
        ratio = max(1,min(3,270//effect.width,90//effect.height))
        effect = effect.resize((round(effect.width*ratio),round(effect.height*ratio)),NEAREST)
        overview.paste(effect,(x+12,235),effect)
        icon = Image.open(ROOT/f'assets/ui/skills/{skill["id"]}.png').resize((48,48),NEAREST)
        overview.paste(icon,(x+238,35),icon)
    save(overview,'assets/preview/skills-overview.png')
    manifest_path = ROOT/'assets/sprites/manifest.json'
    manifest = json.loads(manifest_path.read_text())
    manifest['entries'] = entries()
    manifest_path.write_text(json.dumps(manifest,indent=2)+'\n')

if __name__ == '__main__':
    main()
