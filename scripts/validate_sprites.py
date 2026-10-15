#!/usr/bin/env python3
"""Checks runtime contract plus provenance masks, strip motion, tile seams and swaps.

Visual acceptance (absence of text, facing, identity, upper-left light) is
recorded separately in assets/preview/review.json; this script cannot prove those
semantic properties from a PNG and never claims to do OCR.
"""
import argparse
import hashlib
import json
import sys
import numpy as np
from PIL import Image
from sprite_manifest import ROOT, BASE, COLORS, PLAYER, PROPS, rgb, entries

def validate():
    failures=[]; checks=0
    distinct_roster=(ROOT/'assets/source/player-teal-v2.png').is_file()
    def check(condition,message):
        nonlocal checks
        checks+=1
        if not condition: failures.append(message)
    count=0
    manifest_path=ROOT/'assets/sprites/manifest.json'
    check(manifest_path.is_file(),'missing runtime sprite manifest')
    if manifest_path.is_file():
        manifest=json.loads(manifest_path.read_text())
        check(manifest.get('entries')==entries(),'runtime manifest does not match sprite contract')
    all_teal=np.array([rgb(c) for c in COLORS['teal']])
    danger=np.array([rgb('FF5A5A'),rgb('B8303A')])
    allowed_hazards={'task-card.png','deadline-wave.png','floor-warning.png','charge-warning.png','volley.png','volley_windup.png',
                     'file-stack.png','file-burst.png','drop-warning.png','alert-mark.png',
                     'khabib-press-wave.png','khabib-nag-projectile.png','khabib-teleport-warning.png',
                     'vought-bug-stamp.png','vought-pat-mark.png',
                     'yasuka-silence-dome.png','yasuka-stamp-trap.png'}
    for e in entries():
        path=ROOT/e['path']; tag=e['path']
        check(path.is_file(),f'{tag}: missing')
        if not path.is_file(): continue
        count+=1
        with Image.open(path) as im:
            check(im.format=='PNG',f'{tag}: not PNG')
            check(im.mode=='RGBA',f'{tag}: expected RGBA, got {im.mode}')
            check(im.size==(e['frameWidth']*e['frames'],e['frameHeight']),f'{tag}: incorrect size {im.size}')
            a=np.array(im.convert('RGBA'))
        if a.shape!=(e['frameHeight'],e['frameWidth']*e['frames'],4): continue
        visible=a[:,:,3]>0
        check(visible.any(),f'{tag}: blank asset')
        if e['kind']!='vfx':
            check(np.isin(a[:,:,3],[0,255]).all(),f'{tag}: partial alpha')
        # Every visible runtime color must be an exact palette entry.
        palette=BASE.copy()
        if e['kind'] in ['player','player-skill','portrait']:
            color=path.parent.name if e['kind'] in ['player','player-skill'] else ('teal' if path.stem=='portrait' else path.stem.removeprefix('portrait-'))
            palette+=COLORS[color]
        if path.name=='hotfix-zone.png':
            palette+=COLORS['mint']
        allowed={rgb(x) for x in palette}
        used={tuple(x) for x in a[:,:,:3][visible].tolist()}
        check(used.issubset(allowed),f'{tag}: outside palette {used-allowed}')
        if path.name not in allowed_hazards:
            check(not np.any(np.all(a[:,:,None,:3]==danger[None,None,:,:],axis=3)&visible[:,:,None]),f'{tag}: danger color on non-hazard')
        if e['kind'] in ['player','player-skill','boss','item','portrait','ui','prop']:
            check(not a[0,0,3] and not a[0,-1,3],f'{tag}: background or clipping at top corners')
        if e['kind'] in ['player','player-skill','boss']:
            hashes=[]
            for i in range(e['frames']):
                f=a[:,i*e['frameWidth']:(i+1)*e['frameWidth']]
                check(np.any(f[-1,:,3]==255),f'{tag} frame{i}: feet/body not on bottom row')
                if e['kind'] in ['player','player-skill']:
                    shoe=np.any(np.all(f[-4:,:,:3]==rgb('F2E3C2'),axis=2)|np.all(f[-4:,:,:3]==rgb('C9B48C'),axis=2))
                    check(shoe,f'{tag} frame{i}: no sneaker pixels at feet baseline')
                check(not np.any(f[:,0,3]) and not np.any(f[:,-1,3]),f'{tag} frame{i}: horizontal clipping')
                check(np.any(f[:,:,3]==0),f'{tag} frame{i}: missing transparency')
                hashes.append(f.tobytes())
            check(len(set(hashes))==e['frames'],f'{tag}: duplicate animation frames')
        if e['kind'] in ['player','portrait']:
            if e['kind']=='player':
                color=path.parent.name
                mask_path=ROOT/f'assets/sprite-masks/player-{path.stem}.png'
                if distinct_roster:
                    mask_path=ROOT/f'assets/sprite-masks/player-{color}-{path.stem}.png'
                original=ROOT/f'assets/sprites/player/teal/{path.name}'
            else:
                color='teal' if path.stem=='portrait' else path.stem.removeprefix('portrait-')
                mask_path=ROOT/'assets/sprite-masks/portrait.png'
                if distinct_roster:
                    mask_path=ROOT/f'assets/sprite-masks/portrait-{color}.png'
                original=ROOT/'assets/ui/portrait.png'
            check(mask_path.is_file(),f'{tag}: missing source hoodie mask')
            if mask_path.is_file():
                mask=np.array(Image.open(mask_path))>0
                check(mask.shape==visible.shape,f'{tag}: hoodie mask shape')
                if mask.shape==visible.shape:
                    hoodie=np.array([rgb(x) for x in COLORS[color]])
                    actual=np.any(np.all(a[:,:,None,:3]==hoodie[None,None,:,:],axis=3),axis=2)&visible
                    check(np.array_equal(actual,mask),f'{tag}: hoodie colors outside source clothing mask / missing inside')
                    if distinct_roster:
                        check(actual.any(),f'{tag}: missing player clothing colors')
                    else:
                        for c in COLORS[color]:
                            check(rgb(c) in used,f'{tag}: missing hoodie shade #{c}')
                    check(np.all(visible[mask]),f'{tag}: transparent hoodie pixels')
                    if not distinct_roster and color!='teal' and original.is_file():
                        teal=np.array(Image.open(original))
                        check(np.array_equal(teal[~mask],a[~mask]),f'{tag}: palette swap changed non-hoodie pixels')
                        check(np.array_equal(teal[:,:,3],a[:,:,3]),f'{tag}: swap changed alpha')
        if e['kind']=='background':
            check(np.array_equal(a[:,0],a[:,-1]),f'{tag}: horizontal edges mismatch')
            # Also require adjacent edge derivatives to match; equal columns alone
            # can hide a sharp discontinuity at the repeat boundary.
            check(np.array_equal(a[:,1],a[:,-2]),f'{tag}: near-edge mismatch')
            if 'sky' in path.name: check(np.all(a[:,:,3]==255),f'{tag}: sky must be opaque')
            else: check(np.any(a[:,:,3]==0),f'{tag}: missing transparent parallax regions')
        if path.name in ['deadline-wave.png','floor-warning.png','charge-warning.png']:
            for i in range(e['frames']):
                f=a[:,i*e['frameWidth']:(i+1)*e['frameWidth']]
                check(np.array_equal(f[:,0],f[:,-1]),f'{tag} frame{i}: effect does not tile')
        if path.parent.name=='skills' and e['kind']=='vfx':
            frames=[a[:,i*e['frameWidth']:(i+1)*e['frameWidth']] for i in range(e['frames'])]
            check(all(np.any(f[:,:,3]>0) for f in frames),f'{tag}: blank skill effect frame')
            check(len({f.tobytes() for f in frames})==e['frames'],f'{tag}: duplicate skill effect frames')
        if path.name in ['file-stack.png','file-burst.png','drop-warning.png','alert-mark.png']:
            frames=[a[:,i*e['frameWidth']:(i+1)*e['frameWidth']] for i in range(e['frames'])]
            check(all(np.any(f[:,:,3]>0) for f in frames),f'{tag}: blank drop effect frame')
            check(len({f.tobytes() for f in frames})==e['frames'],f'{tag}: duplicate drop effect frames')
            for i,f in enumerate(frames):
                if path.name=='drop-warning.png':
                    check(np.array_equal(f[0],f[-1]) and np.array_equal(f[1],f[-2]),f'{tag} frame{i}: vertical repeat mismatch')
                    check(np.mean(f[:,2:-2,3])<64,f'{tag} frame{i}: warning interior obscures desks')
                if path.name=='file-stack.png':
                    check(np.any(f[:,0,3]>0) and np.any(f[:,-1,3]>0),f'{tag} frame{i}: stack does not span requested 64px')
                if path.name=='file-burst.png':
                    check(np.any(f[-1,:,3]>0),f'{tag} frame{i}: burst not bottom-anchored')
    if distinct_roster:
        silhouettes=[]
        for color in COLORS:
            idle=ROOT/f'assets/sprites/player/{color}/idle.png'
            if idle.is_file():
                a=np.array(Image.open(idle))[:,:64,3]
                silhouettes.append(a.tobytes())
        check(len(silhouettes)==5 and len(set(silhouettes))==5,'player roster: silhouettes are not distinct')
    tile=ROOT/'assets/tiles/office.png'
    if tile.is_file():
        a=np.array(Image.open(tile));tiles=[a[(i//8)*32:(i//8+1)*32,(i%8)*32:(i%8+1)*32] for i in range(16)]
        for left,right in [(0,1),(1,1),(1,2),(4,5),(5,5),(5,6)]:
            check(np.array_equal(tiles[left][:,-1],tiles[right][:,0]),f'tiles {left}/{right}: seam mismatch')
        for i in [0,1,2,4,5,6]:
            top=np.array(rgb('F2E3C2'))
            check(np.all(tiles[i][1,:,:3]==top),f'tile{i}: cream top not horizontal')
        for i in [4,5,6]:
            check(np.all(tiles[i][12:,:,3]==0),f'tile{i}: platform taller than12px')
    contact=ROOT/'assets/preview/contact-sheet.png'
    check(contact.is_file(),'missing assets/preview/contact-sheet.png')
    review=ROOT/'assets/preview/review.json'
    check(review.is_file(),'missing visual review record')
    if review.is_file():
        record=json.loads(review.read_text())
        for key in ['no_text_watermark','side_view_faces_right','identity_consistent','upper_left_light','no_sprite_background']:
            check(record.get(key) is True,f'visual review incomplete: {key}')
        for e in entries():
            p=ROOT/e['path']
            if p.is_file():
                check(record.get('sha256',{}).get(e['path'])==hashlib.sha256(p.read_bytes()).hexdigest(),f"{e['path']}: visual review stale")
    result={'files':count,'checks':checks,'failures':failures}
    output=ROOT/'assets/preview/validation.json'
    output.parent.mkdir(parents=True,exist_ok=True)
    output.write_text(json.dumps(result,indent=2)+'\n')
    print(f"Validated {count} runtime PNGs; {checks} checks; {len(failures)} failures.")
    for error in failures: print('FAIL:',error)
    return len(failures)

if __name__=='__main__': sys.exit(bool(validate()))
