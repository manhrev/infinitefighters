#!/usr/bin/env python3
"""Build additional boss movement sprites from retained imagegen source sheets."""
import json
import numpy as np
from PIL import Image, ImageDraw
from build_sprites import ROOT, SOURCE, segments, save, strip
from build_skill_sprites import fit_frames, palette_image
from sprite_manifest import BASE, BOSS_MOVEMENT, BOSS_MOVEMENT_VFX, entries

def extract(name, row_count, characters=False):
    source = Image.open(SOURCE/f'{name}.png').convert('RGBA')
    opaque = np.array(source)[:,:,3] >= 128
    rows = segments(opaque.sum(axis=1)>6,6)
    merged=[]
    for start,end in rows:
        if merged and start-merged[-1][1]<=2:
            merged[-1]=(merged[-1][0],end)
        else:
            merged.append((start,end))
    if len(merged)!=row_count:
        raise ValueError(f'{name}: expected {row_count} rows, got {merged}; inspect source')
    raw=[];boxes=[]
    for y0,y1 in merged:
        columns=segments(opaque[y0:y1].sum(axis=0)>3,20) if characters else []
        if characters and len(columns)!=4:
            raise ValueError(f'{name}: expected four characters, got {columns}; inspect source')
        for col in range(4):
            x0,x1=columns[col] if characters else (round(col*source.width/4),round((col+1)*source.width/4))
            crop=source.crop((x0,y0,x1,y1))
            a=np.array(crop);a[a[:,:,3]<128]=0
            crop=Image.fromarray(a)
            bbox=crop.getbbox()
            if not bbox:
                raise ValueError(f'{name}: empty source cell')
            raw.append(crop.crop(bbox))
            boxes.append([x0+bbox[0],y0+bbox[1],x0+bbox[2],y0+bbox[3]])
    (SOURCE/f'{name}-regions.json').write_text(json.dumps(boxes,indent=2)+'\n')
    return raw

def main():
    safe_base=[c for c in BASE if c not in ('FF5A5A','B8303A')]
    raw=extract('boss-movement',4,True)
    frames=fit_frames(raw,(160,160),safe_base,character=True,height=120)
    preview=Image.new('RGB',(1380,6*360),'#0B1530')
    draw=ImageDraw.Draw(preview)
    for row,(name,(count,fps,loop)) in enumerate(BOSS_MOVEMENT.items()):
        selected=frames[row*4:(row+1)*4]
        save(strip(selected),f'assets/sprites/boss/doc-cong/{name}.png')
        draw.text((12,row*360+10),f'{name}: {count} frames, {fps} FPS',fill='#F2E3C2')
        for i,frame in enumerate(selected):
            enlarged=frame.resize((320,320),Image.Resampling.NEAREST)
            preview.paste(enlarged,(12+i*340,row*360+35),enlarged)
    raw_fx=extract('boss-movement-vfx',2)
    for row,(name,(w,h,count,fps,loop)) in enumerate(BOSS_MOVEMENT_VFX.items()):
        selected=raw_fx[row*4:(row+1)*4]
        if name=='charge-warning':
            effects=[]
            for image in selected:
                art=palette_image(image.resize((w,h),Image.Resampling.NEAREST),BASE)
                a=np.array(art)
                a[:,-1]=a[:,0];a[:,-2]=a[:,1]
                effects.append(Image.fromarray(a))
        else:
            effects=fit_frames(selected,(w,h),safe_base)
        save(strip(effects),f'assets/sprites/vfx/{name}.png')
        y=(row+4)*360
        draw.text((12,y+10),f'{name}: {w} x {h}, {count} frames',fill='#F2E3C2')
        for i,frame in enumerate(effects):
            enlarged=frame.resize((w*4,h*4),Image.Resampling.NEAREST)
            preview.paste(enlarged,(12+i*340,y+60),enlarged)
    save(preview,'assets/preview/boss-movement.png')
    manifest_path=ROOT/'assets/sprites/manifest.json'
    manifest=json.loads(manifest_path.read_text())
    manifest['entries']=entries()
    manifest_path.write_text(json.dumps(manifest,indent=2)+'\n')
    print('Boss movement: four animation strips, two VFX strips, manifest and preview.')

if __name__=='__main__':
    main()
