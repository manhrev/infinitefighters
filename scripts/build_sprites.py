#!/usr/bin/env python3
"""Deterministic generated-source -> native RGBA sprite processing. Requires Pillow/numpy."""
import argparse
import json
import math
from pathlib import Path
import numpy as np
from PIL import Image, ImageDraw, ImageFilter
from sprite_manifest import ROOT, BASE, COLORS, PLAYER, BOSS, VFX, PROPS, rgb, entries

NEAREST = Image.Resampling.NEAREST
OUTLINE = (*rgb('0B1020'),255)
SOURCE = ROOT / 'assets/source'
MASKS = ROOT / 'assets/sprite-masks'

def save(im, path):
    path = ROOT / path
    path.parent.mkdir(parents=True, exist_ok=True)
    im.convert('RGBA').save(path)

def segments(values, minimum=2):
    a = np.flatnonzero(values)
    if not len(a): return []
    groups = np.split(a, np.where(np.diff(a)>1)[0]+1)
    return [(int(g[0]),int(g[-1])+1) for g in groups if len(g)>=minimum]

def remove_background(im):
    a = np.array(im.convert('RGBA'))
    r,g,b = a[:,:,:3].astype(int).transpose(2,0,1)
    magenta = (r>105)&(b>85)&(r>g*1.35)&(b>g*1.3)
    a[:,:,3] = np.where(magenta | (a[:,:,3]<128),0,255)
    a[a[:,:,3]==0,:3] = 0
    return Image.fromarray(a)

def extract_rows(name):
    im=remove_background(Image.open(SOURCE / f'{name}.png'))
    a=np.array(im)[:,:,3]>0
    rows=segments(a.sum(axis=1)>5, 8)
    result=[]
    boxes=[]
    for y0,y1 in rows:
        cols=segments(a[y0:y1].sum(axis=0)>2, 6)
        row=[]
        for x0,x1 in cols:
            crop=im.crop((x0,y0,x1,y1))
            bbox=crop.getbbox()
            if bbox and bbox[2]*bbox[3]>150:
                row.append(crop.crop(bbox)); boxes.append([x0+bbox[0],y0+bbox[1],x0+bbox[2],y0+bbox[3]])
        result.append(row)
    if name=='player-teal':
        # Three dash silhouettes overlap in their X projections, but not in the image.
        merged=boxes[24]
        if merged[2]-merged[0]>500:
            split_boxes=[[633,697,858,826],[858,697,1085,826],[1085,697,1312,826]]
            boxes[24:25]=split_boxes
            flat=[im.crop(tuple(box)).crop(im.crop(tuple(box)).getbbox()) for box in boxes]
            result=[flat]
    print(name, 'source rows:', [len(r) for r in result])
    (SOURCE / f'{name}-regions.json').write_text(json.dumps(boxes,indent=2)+'\n')
    return [im for row in result for im in row]

def quantize(im, hoodie=False, danger=False, background=False):
    a=np.array(im.convert('RGBA'))
    opaque=a[:,:,3]>127
    colors=[x for x in BASE if danger or x not in ['FF5A5A','B8303A']]
    if background: colors=['0B1020','0B1530','14234A','22345F','2F4A7F']
    p=np.array([rgb(x) for x in colors],dtype=np.int32)
    src=a[:,:,:3].astype(np.int32)
    if hoodie:
        r,g,b=src.transpose(2,0,1)
        cloth=opaque&(g>r*1.38)&(b>r*1.1)&(g>b*.88)&(g>45)
    else: cloth=np.zeros(opaque.shape,dtype=bool)
    out=np.zeros_like(a)
    # Bounded memory when processing full backgrounds.
    for y in range(a.shape[0]):
        distance=((src[y,:,None,:]-p[None,:,:])**2).sum(axis=2)
        out[y,:,:3]=p[distance.argmin(axis=1)]
    if hoodie:
        teal=np.array([rgb(x) for x in COLORS['teal']],dtype=np.int32)
        dist=((src[:,:,None,:]-teal[None,None,:,:])**2).sum(axis=3)
        out[cloth,:3]=teal[dist.argmin(axis=2)][cloth]
    out[:,:,3]=np.where(opaque,255,0)
    out[~opaque,:3]=0
    return Image.fromarray(out), Image.fromarray((cloth*255).astype('uint8'))

def outlined(im):
    alpha=im.getchannel('A')
    # Boundary is a single native pixel. Interior colors never become anti-aliased.
    eroded=alpha.filter(ImageFilter.MinFilter(3))
    a=np.array(im); edge=(np.array(alpha)>0)&(np.array(eroded)==0)
    # ImageFilter extends edge pixels; explicitly mark the occupied canvas edge.
    edge[0]|=np.array(alpha)[0]>0; edge[-1]|=np.array(alpha)[-1]>0
    edge[:,0]|=np.array(alpha)[:,0]>0; edge[:,-1]|=np.array(alpha)[:,-1]>0
    a[edge]=OUTLINE
    return Image.fromarray(a)

def largest_component(im):
    """Remove detached generated missiles; game emits its own projectile sprites."""
    a=np.array(im);mask=a[:,:,3]>0;seen=np.zeros(mask.shape,dtype=bool);parts=[]
    h,w=mask.shape
    for sy,sx in zip(*np.where(mask)):
        if seen[sy,sx]: continue
        stack=[(int(sy),int(sx))]; seen[sy,sx]=True; points=[]
        while stack:
            y,x=stack.pop();points.append((y,x))
            for yy,xx in [(y-1,x),(y+1,x),(y,x-1),(y,x+1)]:
                if 0<=yy<h and 0<=xx<w and mask[yy,xx] and not seen[yy,xx]:
                    seen[yy,xx]=True;stack.append((yy,xx))
        parts.append(points)
    keep=np.zeros_like(mask)
    for y,x in max(parts,key=len): keep[y,x]=True
    a[~keep]=0
    result=Image.fromarray(a)
    return result.crop(result.getbbox())

def fit(crop,w,h,target_h,hoodie=False,danger=False,max_w=None,baseline=True):
    crop=crop.crop(crop.getbbox())
    limit=max_w or w-4
    scale=min(target_h/crop.height,limit/crop.width)
    nw=max(1,round(crop.width*scale)); nh=max(1,round(crop.height*scale))
    small=crop.resize((nw,nh),NEAREST)
    small,mask=quantize(small,hoodie,danger)
    small=outlined(small)
    # Remove mask where boundary was converted to outline.
    ma=np.array(mask); ar=np.array(small)
    ma[np.all(ar[:,:,:3]==rgb('0B1020'),axis=2)]=0
    mask=Image.fromarray(ma)
    canvas=Image.new('RGBA',(w,h)); m=Image.new('L',(w,h))
    x=(w-nw)//2; y=h-nh if baseline else (h-nh)//2
    canvas.paste(small,(x,y)); m.paste(mask,(x,y))
    return canvas,m

def strip(frames):
    out=Image.new('RGBA',(frames[0].width*len(frames),frames[0].height))
    for i,f in enumerate(frames): out.paste(f,(i*f.width,0))
    return out

def mask_strip(frames):
    out=Image.new('L',(frames[0].width*len(frames),frames[0].height))
    for i,f in enumerate(frames): out.paste(f,(i*f.width,0))
    return out

def players():
    if (SOURCE/'player-teal-v2.png').is_file():
        from build_player_roster import main as build_roster
        build_roster()
        return
    raw=extract_rows('player-teal')
    # Generated sheet has an extra run pose and extra seated pose; select by inspected semantic order.
    mapping={'idle':[0,1,2,3], 'run':[4,5,6,7,8,9,10,11],
        'jump':[14,15], 'fall':[16,17], 'attack':[18,19,20,21,22],
        'dash':[23,24,25,26], 'hurt':[27,28], 'down':[29,30,31,32],
        'revive':[32,33,34,35]}
    # These indices are verified against source/player-teal-regions.json and source preview.
    for name,indices in mapping.items():
        frames=[]; masks=[]
        for j,i in enumerate(indices):
            # Preserve source head/limb scale instead of making every pose equally tall.
            height=min(60,round(raw[i].height*48/186))
            f,m=fit(raw[i],64,64,height,hoodie=True,max_w=60)
            frames.append(f); masks.append(m)
        save(strip(frames),f'assets/sprites/player/teal/{name}.png')
        MASKS.mkdir(parents=True,exist_ok=True)
        mask_strip(masks).save(MASKS/f'player-{name}.png')

def boss():
    source=remove_background(Image.open(SOURCE/'boss-doc-cong.png'))
    xs=[0,220,425,635,825,1050,source.width]
    ys=[0,210,410,650,850,1070,source.height]
    raw=[]; boxes=[]
    for row in range(6):
        for col in range(6 if row<5 else 4):
            box=(xs[col],ys[row],xs[col+1],ys[row+1])
            crop=source.crop(box); bbox=crop.getbbox()
            raw.append(crop.crop(bbox)); boxes.append(box)
    (SOURCE/'boss-doc-cong-regions.json').write_text(json.dumps(boxes,indent=2)+'\n')
    mapping={'idle':[0,1,2,3], 'volley_windup':[4,5,6,7], 'volley':[8,9,10,11],
        'wave_windup':[12,13,14,15], 'wave':[15,16,17,18],
        'wind_clock':[18,19,20,21,22,23], 'hurt':[24,25], 'defeat':[26,27,28,29,30,33]}
    for name,(n,_,_) in BOSS.items():
        frames=[]
        for j in range(n):
            crop=raw[mapping[name][j]]
            if name=='volley': crop=largest_component(crop)
            height=min(156,round(crop.height*120/165))
            f,_=fit(crop,160,160,height,danger=name in ['volley_windup','volley'],max_w=152)
            frames.append(f)
        save(strip(frames),f'assets/sprites/boss/doc-cong/{name}.png')

def effects():
    source=remove_background(Image.open(SOURCE/'effects-items.png'))
    # Source was generated as4 columns x9 rows. Generous margins make regular
    # cells robust; use the visible row projections for exact vertical bounds.
    a=np.array(source)[:,:,3]>0
    rows=segments(a.sum(axis=1)>2,10)
    if len(rows)!=9: raise ValueError(f'Expected9 VFX/item rows, got{rows}')
    groups=[]
    for y0,y1 in rows:
        group=[]
        for col in range(4):
            x0=round(col*source.width/4);x1=round((col+1)*source.width/4)
            c=source.crop((x0,y0,x1,y1))
            group.append(c.crop(c.getbbox()) if c.getbbox() else None)
        groups.append(group)
    for row,(name,(w,h,n,_,_)) in enumerate(VFX.items()):
        frames=[]
        for i in range(n):
            crop=groups[row][i]
            # Fit image-generated silhouettes, with effects free of physical outlines.
            if name in ['deadline-wave','floor-warning']:
                small=crop.resize((w,h),NEAREST)
                f,_=quantize(small,danger=True)
                aa=np.array(f)
                # Mirror two boundary columns for an exact loop; interiors retain motion.
                aa[:,-1]=aa[:,0];aa[:,-2]=aa[:,1]
                f=Image.fromarray(aa)
            else:
                ratio=min((w-2)/crop.width,(h-2)/crop.height)
                sz=(max(1,round(crop.width*ratio)),max(1,round(crop.height*ratio)))
                small,_=quantize(crop.resize(sz,NEAREST),danger=name=='task-card')
                f=Image.new('RGBA',(w,h));f.paste(small,((w-sz[0])//2,(h-sz[1])//2))
            frames.append(f)
        save(strip(frames),f'assets/sprites/vfx/{name}.png')
    # Same mug base in every frame; the generation may drift, so isolate one
    # canonical physical mug and animate only the generated steam silhouettes.
    mug=groups[7][0]
    mug_small,_=quantize(mug.resize((11,15),NEAREST))
    coffee=[]
    for i in range(4):
        f=Image.new('RGBA',(16,16));f.paste(mug_small,(2,1))
        # Preserve fixed cup/handle and replace top4 pixels with moving wisps.
        d=ImageDraw.Draw(f);d.rectangle((0,0,15,4),fill=(0,0,0,0))
        steam=rgb('F2E3C2')+(255,)
        x=5+(i%2)
        d.line([(x,4),(x-1,3),(x-1,2),(x,1 if i<2 else 0)],fill=steam)
        d.point((9+(i//2),3-(i%2)),fill=steam)
        coffee.append(f)
    save(strip(coffee),'assets/sprites/items/coffee.png')
    doc,_=fit(groups[8][0],16,16,14,max_w=12,baseline=False)
    save(doc,'assets/sprites/items/document.png')

def tiles():
    source=remove_background(Image.open(SOURCE/'office-tiles.png'))
    xs=[0,314,627,942,1254];ys=[0,275,574,891,1254]
    images=[]
    for i in range(16):
        x=i%4;y=i//4
        crop=source.crop((xs[x],ys[y],xs[x+1],ys[y+1]))
        tile,_=quantize(crop.resize((32,32),NEAREST))
        if i>=11: tile=Image.new('RGBA',(32,32))
        images.append(tile)
    # Enforce structural collision surfaces after nearest-neighbor sampling.
    # Generated texture is retained below these exactly aligned caps.
    for i in [0,1,2,4,5,6]:
        d=ImageDraw.Draw(images[i])
        d.line((0,0,31,0),fill=OUTLINE)
        d.rectangle((0,1,31,3),fill=rgb('F2E3C2')+(255,))
        d.rectangle((0,4,31,5),fill=rgb('C9B48C')+(255,))
        if i in [4,5,6]:
            d.rectangle((0,12,31,31),fill=(0,0,0,0))
            d.line((0,11,31,11),fill=OUTLINE)
    for triplet in [(0,1,2),(4,5,6)]:
        mid=np.array(images[triplet[1]])
        canonical=mid[:,1].copy()
        for i in triplet:
            ar=np.array(images[i]);ar[:,0]=canonical;ar[:,-1]=canonical
            images[i]=Image.fromarray(ar)
    atlas=Image.new('RGBA',(256,64))
    for i,tile in enumerate(images): atlas.paste(tile,((i%8)*32,(i//8)*32))
    save(atlas,'assets/tiles/office.png')
    mapping={'tileSize':32,'columns':8,'rows':2,'spacing':0,'margin':0,
        'tiles':{name:i for i,name in enumerate(['floor-left','floor-middle','floor-right','floor-fill','platform-left','platform-middle','platform-right','support-leg','server-top','server-bottom','wall-fill','empty-11','empty-12','empty-13','empty-14','empty-15'])}}
    (ROOT/'assets/tiles/office.json').write_text(json.dumps(mapping,indent=2)+'\n')

def props():
    source=remove_background(Image.open(SOURCE/'office-props.png'))
    boxes=[(35,105,450,325),(510,35,710,330),(795,100,1040,325),(1125,20,1405,340),
        (110,370,290,665),(470,355,605,670),(705,405,1095,585),(1160,380,1370,650)]
    for (name,(w,h)),box in zip(PROPS.items(),boxes):
        crop=source.crop(box);crop=crop.crop(crop.getbbox())
        f,_=fit(crop,w,h,h-2,max_w=w-2,baseline=False)
        if name=='plant':
            ar=np.array(f);ar[np.all(ar[:,:,:3]==rgb('3FE0FF'),axis=2),:3]=rgb('2F4A7F');f=Image.fromarray(ar)
        save(f,f'assets/sprites/props/{name}.png')
    elevators=[]
    boxes=[(40,682,335,1035),(395,682,690,1035),(755,682,1050,1035),(1110,682,1410,1035)]
    for box in boxes:
        crop=source.crop(box);crop=crop.crop(crop.getbbox())
        f,_=quantize(crop.resize((60,92),NEAREST))
        f=outlined(f);canvas=Image.new('RGBA',(64,96));canvas.paste(f,(2,4))
        d=ImageDraw.Draw(canvas)
        d.line((9,7,9,93),fill=rgb('3FE0FF')+(255,))
        d.line((55,7,55,93),fill=rgb('3FE0FF')+(255,))
        elevators.append(canvas)
    # One fixed casing; only the generated door aperture changes.
    for i in range(1,4):
        base=elevators[0].copy()
        base.paste(elevators[i].crop((12,18,52,93)),(12,18))
        elevators[i]=base
    save(strip(elevators),'assets/sprites/props/elevator.png')

def backgrounds():
    source=remove_background(Image.open(SOURCE/'background-toolkit.png'))
    def piece(box,limit):
        im=source.crop(box);im=im.crop(im.getbbox())
        scale=min(1,limit[0]/im.width,limit[1]/im.height)
        im=im.resize((max(1,round(im.width*scale)),max(1,round(im.height*scale))),NEAREST)
        return quantize(im,background=True)[0]
    sky=Image.new('RGBA',(1920,1080),rgb('0B1530')+(255,))
    d=ImageDraw.Draw(sky)
    # Native sky details, deterministic and limited to the approved dark palette.
    for i in range(70):
        x=(i*137+83)%1920; y=(i*79+47)%660
        d.point((x,y),fill=rgb('14234A')+(255,))
    buildings=[piece((x0,30,x1,380),(310,340)) for x0,x1 in [(40,345),(405,710),(765,1050),(1125,1410)]]
    moon=piece((80,450,345,715),(96,96));sky.alpha_composite(moon,(1530,205))
    def wrap_composite(canvas,im,x,y):
        for delta in [-1920,0,1920]:
            xx=x+delta
            if xx<1920 and xx+im.width>0: canvas.alpha_composite(im,(xx,y))
    for i,x in enumerate([-100,140,380,620,860,1100,1340,1580]):
        b=buildings[i%4]
        # Two rows of skyline depth; all source art is downsampled, never enlarged.
        back=b.copy(); aa=np.array(back)
        for c in ['2F4A7F','22345F']: aa[np.all(aa[:,:,:3]==rgb(c),axis=2),:3]=rgb('14234A')
        wrap_composite(sky,Image.fromarray(aa),x+55,1080-b.height-100)
        wrap_composite(sky,b,x,1080-b.height)
    # Fill skyline bases behind the bottom composition so roofs stay freestanding.
    windows=Image.new('RGBA',(1920,1080))
    pillar=piece((515,410,600,725),(32,256))
    beam=piece((735,535,1125,610),(320,16))
    # Extract repeatable interior modules: extending architectural beams uses
    # tiling, not upscaling raster art.
    shaft=pillar.crop((0,16,pillar.width,pillar.height-16))
    for x in range(-16,1920,320):
        for y in range(0,1080,shaft.height):
            windows.alpha_composite(shaft,(x,y))
    for y in [0,350,710,1055]:
        for x in range(0,1920,beam.width): windows.alpha_composite(beam,(x,y))
    office=Image.new('RGBA',(1920,1080))
    furniture=[piece(box,(300,160)) for box in [(35,795,385,1005),(410,825,755,1005),(785,750,1105,1005),(1125,780,1420,1005)]]
    for i,x in enumerate([50,390,800,1160,1530]):
        f=furniture[i%4]
        # Faint silhouettes remain less contrasty than playable furniture.
        a=np.array(f)
        a[np.all(a[:,:,:3]==rgb('2F4A7F'),axis=2),:3]=rgb('22345F')
        office.alpha_composite(Image.fromarray(a),(x,1080-f.height))
    for name,im in [('bg-0-sky',sky),('bg-1-windows',windows),('bg-2-office',office)]:
        a=np.array(im)
        # Border comes from one continuous source neighborhood. Match both edge
        # columns and adjacent columns for stable horizontal repeats.
        a[:,-1]=a[:,0];a[:,-2]=a[:,1]
        a[a[:,:,3]==0,:3]=0
        save(Image.fromarray(a),f'assets/backgrounds/{name}.png')

def ui():
    source=remove_background(Image.open(SOURCE/'ui-icons.png'))
    icons=[]
    for i in range(3):
        c=source.crop((round(i*source.width/3),0,round((i+1)*source.width/3),source.height))
        icons.append(c.crop(c.getbbox()))
    hearts=[]
    for c in icons[:2]:
        f,_=quantize(c.resize((9,8),NEAREST));hearts.append(outlined(f))
    save(strip(hearts),'assets/ui/heart.png')
    watch,_=fit(icons[2],16,16,16,max_w=16,baseline=False)
    save(watch,'assets/ui/deadline-icon.png')
    # Portrait is cropped from the SAME source pose as player idle, never a
    # newly generated face. Includes hood/shoulders for the palette swap.
    if (SOURCE/'player-teal-v2.png').is_file():
        # Distinct portraits are generated alongside their own player sheets.
        return
    raw=extract_rows('player-teal')[0]
    crop=raw.crop((0,0,raw.width,110))
    portrait,mask=fit(crop,24,24,24,max_w=22,hoodie=True,baseline=True)
    save(portrait,'assets/ui/portrait.png')
    mask.save(MASKS/'portrait.png')

def main():
    parser=argparse.ArgumentParser()
    parser.add_argument('stage',choices=['player','skills','boss','boss-motion','boss-drop','boss-khabib','boss-vought','boss-yasuka','effects','tiles','props','backgrounds','ui','all'])
    args=parser.parse_args()
    from build_skill_sprites import main as skills
    from build_boss_movement import main as boss_motion
    from build_boss_drop import main as boss_drop
    from build_boss_khabib import main as boss_khabib
    from build_boss_vought import main as boss_vought
    from build_boss_yasuka import main as boss_yasuka
    stages={'player':players,'skills':skills,'boss':boss,'boss-motion':boss_motion,'boss-drop':boss_drop,'boss-khabib':boss_khabib,'boss-vought':boss_vought,'boss-yasuka':boss_yasuka,'effects':effects,'tiles':tiles,'props':props,'backgrounds':backgrounds,'ui':ui}
    for name,fn in stages.items():
        if args.stage in [name,'all']: fn()
    if args.stage=='all':
        import palette_swap
        palette_swap.main()
        manifest={'pixelArt':True,'scaleMode':'NEAREST','entries':entries(),
            'playerHitbox':{'width':20,'height':44,'anchor':'bottom-center'},
            'bossHitbox':{'width':70,'height':110,'anchor':'bottom-center'}}
        (ROOT/'assets/sprites/manifest.json').write_text(json.dumps(manifest,indent=2)+'\n')

if __name__=='__main__': main()
