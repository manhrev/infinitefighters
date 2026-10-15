#!/usr/bin/env python3
"""Normalize imagegen sources for the falling-file attack; asset directories only."""
import json
import html
import numpy as np
from PIL import Image, ImageDraw
from build_sprites import ROOT, SOURCE, save, strip, segments, outlined
from build_skill_sprites import fit_frames, palette_image
from sprite_manifest import BASE, BOSS_DROP, BOSS_DROP_VFX, entries

SOURCES = {
    'drop_windup':'boss-drop-windup', 'drop':'boss-drop-release', 'drop_jam':'boss-drop-jam',
    'file-stack':'falling-file-stack', 'file-burst':'falling-file-burst',
    'drop-warning':'falling-drop-warning', 'alert-mark':'falling-alert-mark',
}

def extract(key,count,character=False):
    source=Image.open(SOURCE/f'{key}.png').convert('RGBA')
    a=np.array(source)
    opaque=a[:,:,3]>=128
    columns=segments(opaque.sum(axis=0)>3,12)
    if character and len(columns)!=count:
        raise ValueError(f'{key}: expected {count} isolated character columns, got {columns}; inspect source')
    raw=[];boxes=[]
    for i in range(count):
        x0,x1=columns[i] if character else (round(i*source.width/count),round((i+1)*source.width/count))
        cell=source.crop((x0,0,x1,source.height))
        ca=np.array(cell);ca[ca[:,:,3]<(16 if key=='falling-drop-warning' else 128)]=0
        cell=Image.fromarray(ca)
        bbox=cell.getbbox()
        if not bbox:
            raise ValueError(f'{key}: empty frame {i}')
        raw.append(cell.crop(bbox))
        boxes.append([x0+bbox[0],bbox[1],x0+bbox[2],bbox[3]])
    (SOURCE/f'{key}-regions.json').write_text(json.dumps(boxes,indent=2)+'\n')
    return raw

def warning_frames(raw):
    frames=[]
    for image in raw:
        # Keep the generated warning's diagonal pattern; lower interior opacity
        # so desk surfaces remain readable. No caps across the vertical repeat.
        resized=image.resize((64,32),Image.Resampling.NEAREST)
        alpha=np.array(resized)[:,:,3]
        resized.putalpha(255) # Quantize stripe RGB before restoring its faint alpha.
        art=palette_image(resized,['FF5A5A','B8303A'])
        a=np.array(art)
        a[:,:,3]=np.where(alpha>=16,np.minimum(alpha,48),0)
        a[:,:2,3]=255;a[:,-2:,3]=255
        # Same two boundary rows and derivative on every repeated module.
        a[-1]=a[0];a[-2]=a[1]
        a[a[:,:,3]==0]=0
        frames.append(Image.fromarray(a))
    return frames

def main():
    safe=[c for c in BASE if c not in ('FF5A5A','B8303A')]
    groups={}
    for name,(count,fps,loop) in BOSS_DROP.items():
        groups[name]=extract(SOURCES[name],count,True)
    # Separate generated sheets have different source pixel scales. Calibrate
    # by anatomy rather than making the overhead watch and bent jam pose equal
    # heights. Keep a fixed scale within each animation so its limbs don't jitter.
    calibration=json.loads((SOURCE/'boss-drop-calibration.json').read_text())
    rendered={}
    for name,(count,fps,loop) in BOSS_DROP.items():
        raw=groups[name]
        height=raw[0].height*calibration['targetHairToSoleHeight']/calibration['sourceBodyHeights'][name]
        frames=fit_frames(raw,(160,160),safe,character=True,height=height)
        save(strip(frames),f'assets/sprites/boss/doc-cong/{name}.png')
        rendered[name]=frames
    for name,(w,h,count,fps,loop) in BOSS_DROP_VFX.items():
        raw=extract(SOURCES[name],count)
        if name=='drop-warning':
            frames=warning_frames(raw)
        elif name=='file-stack':
            # The requested danger stack occupies the complete two-tile width.
            frames=[]
            for f in raw:
                alpha=np.array(f)[:,:,3]>=128
                occupancy=alpha.sum(axis=1)
                bands=segments(occupancy>occupancy.max()*.5,10)
                top=max(bands,key=lambda band:band[1]-band[0])[0]
                canvas=Image.new('RGBA',(64,48))
                body=palette_image(f.crop((0,top,f.width,f.height)).resize((64,32),Image.Resampling.NEAREST),BASE)
                bounds=body.getbbox()
                body=body.crop((bounds[0],0,bounds[2],32)).resize((64,32),Image.Resampling.NEAREST)
                canvas.paste(body,(0,16))
                if top:
                    sheets=f.crop((0,0,f.width,top)).resize((64,16),Image.Resampling.NEAREST)
                    canvas.paste(sheets,(0,0))
                frames.append(outlined(palette_image(canvas,BASE)))
        elif name=='file-burst':
            frames=fit_frames(raw,(w,h),BASE)
            aligned=[]
            for f in frames:
                box=f.getbbox();crop=f.crop(box)
                canvas=Image.new('RGBA',(w,h))
                canvas.paste(crop,((w-crop.width)//2,h-crop.height))
                aligned.append(canvas)
            frames=aligned
        else:
            frames=[]
            for f in raw:
                scale=min((w-2)/f.width,(h-2)/f.height)
                art=palette_image(f.resize((max(1,round(f.width*scale)),max(1,round(f.height*scale))),Image.Resampling.NEAREST),BASE)
                frame=Image.new('RGBA',(w,h))
                frame.paste(art,((w-art.width)//2,(h-art.height)//2))
                frames.append(frame)
            # The symbol's hard one-pixel outline remains readable at 16x24.
            frames=[outlined(f) for f in frames]
        save(strip(frames),f'assets/sprites/vfx/{name}.png')
        rendered[name]=frames
    width=2040;row_height=365
    preview=Image.new('RGB',(width,len(rendered)*row_height),'#0B1530')
    draw=ImageDraw.Draw(preview)
    for row,(name,frames) in enumerate(rendered.items()):
        y=row*row_height
        draw.text((12,y+10),f'{name}: {frames[0].width}x{frames[0].height}, {len(frames)} frames',fill='#F2E3C2')
        scale=2 if frames[0].width==160 else 4
        for i,frame in enumerate(frames):
            art=frame.resize((frame.width*scale,frame.height*scale),Image.Resampling.NEAREST)
            px=12+i*336
            if frames[0].width==96: px=12+i*400
            preview.paste(art,(px,y+35),art)
        preview.crop((0,y,width,y+row_height)).save(ROOT/f'assets/preview/boss-drop-{name}.png')
    save(preview,'assets/preview/boss-drop-contact-sheet.png')
    # Show a vertical repeat over existing desk artwork, to audit transparency
    # and both boundary rows without altering the runtime warning sprite.
    repeat=Image.new('RGBA',(64,224))
    desk=Image.open(ROOT/'assets/sprites/props/desk.png').convert('RGBA')
    repeat.alpha_composite(desk.crop((16,0,80,48)),(0,160))
    for y in range(0,192,32):repeat.alpha_composite(rendered['drop-warning'][0],(0,y))
    repeat.alpha_composite(rendered['alert-mark'][0],(24,0))
    save(repeat.resize((256,896),Image.Resampling.NEAREST),'assets/preview/boss-drop-warning-repeat.png')
    overview=Image.new('RGB',(1120,370),'#0B1530')
    od=ImageDraw.Draw(overview)
    od.text((20,8),'drop_windup',fill='#F2E3C2')
    od.text((420,8),'falling files / warning / desk',fill='#F2E3C2')
    od.text((790,8),'drop_jam',fill='#F2E3C2')
    for image,pos in [(rendered['drop_windup'][-1],(20,35)),(rendered['drop_jam'][2],(790,35)),
                      (desk,(420,245))]:
        art=image.resize((image.width*2,image.height*2),Image.Resampling.NEAREST)
        overview.paste(art,pos,art)
    tile=rendered['drop-warning'][0].resize((128,64),Image.Resampling.NEAREST)
    for y in range(53,245,64):overview.paste(tile,(452,y),tile)
    stack=rendered['file-stack'][0].resize((128,96),Image.Resampling.NEAREST)
    overview.paste(stack,(452,145),stack)
    mark=rendered['alert-mark'][0].resize((32,48),Image.Resampling.NEAREST)
    overview.paste(mark,(500,28),mark)
    save(overview,'assets/preview/boss-drop-overview.png')
    cards=[]
    for name,frames in rendered.items():
        boss=name in BOSS_DROP
        meta=BOSS_DROP[name] if boss else BOSS_DROP_VFX[name][2:]
        count,fps,loop=meta
        w,h=frames[0].size
        url=f'../sprites/boss/doc-cong/{name}.png' if boss else f'../sprites/vfx/{name}.png'
        style=f'width:{w}px;height:{h}px;background-image:url({url});--end:-{w*count}px;--duration:{count/fps}s;--frames:{count}'
        cards.append(f'<article><h2>{html.escape(name)}</h2><p>{w}×{h} · {count} frames · {fps} FPS · loop: {str(loop).lower()}</p><div class="stage" style="width:{w*2}px;height:{h*2}px"><div class="sprite" style="{style}"></div></div><a href="{url}">PNG strip</a></article>')
    audio_cards=[]
    for key in ['boss-drop-windup','boss-drop-impact','boss-paper-jam']:
        audio_cards.append(f'<article><h2>{key}</h2><audio controls preload="none"><source src="../audio/sfx/{key}.ogg" type="audio/ogg"><source src="../audio/sfx/{key}.mp3" type="audio/mpeg"></audio></article>')
    page='''<!doctype html><html lang="vi"><meta charset="utf-8"><meta name="viewport" content="width=device-width,initial-scale=1"><title>Task rơi từ trên cao — Asset preview</title><style>
    body{margin:0;background:#0B1530;color:#F2E3C2;font:16px/1.5 system-ui}main{max-width:1200px;margin:auto;padding:24px}h1{color:#3FE0FF}h2{font-size:18px}p{color:#bdcbe3}.grid{display:grid;grid-template-columns:repeat(auto-fit,minmax(340px,1fr));gap:16px}article{padding:16px;background:#14234A;border:1px solid #2F4A7F;border-radius:8px;overflow:auto}.stage{position:relative;margin:12px auto;background-color:#0B1530;background-image:linear-gradient(45deg,#14234A 25%,transparent 25%),linear-gradient(-45deg,#14234A 25%,transparent 25%);background-size:16px 16px}.sprite{transform:scale(2);transform-origin:top left;image-rendering:pixelated;background-repeat:no-repeat;animation:frames var(--duration) steps(var(--frames)) infinite}@keyframes frames{to{background-position:var(--end) 0}}.paused .sprite{animation-play-state:paused}a{color:#3FE0FF}button{padding:10px 18px;margin-bottom:16px;background:#3FE0FF;color:#0B1020;border:0;border-radius:6px;font-weight:bold}audio{width:100%}</style><main><h1>Task rơi từ trên cao</h1><p>Preview asset ở 2× nearest-neighbor. Các strip lặp ở đây để dễ duyệt; loop thực tế theo manifest. Thời gian báo đòn, rơi và kẹt giấy do server quyết định.</p><button id="pause">Dừng / chạy animation</button><div class="grid">'''+''.join(cards)+'''</div><h2>Âm thanh</h2><div class="grid">'''+''.join(audio_cards)+'''</div><p><a href="boss-drop-contact-sheet.png">Tất cả frame</a> · <a href="boss-drop-warning-repeat.png">Kiểm tra cột nối dọc trên bàn</a></p></main><script>document.getElementById('pause').onclick=()=>document.body.classList.toggle('paused');const players=[...document.querySelectorAll('audio')];players.forEach(a=>a.addEventListener('play',()=>players.forEach(b=>{if(b!==a)b.pause()})));</script></html>'''
    (ROOT/'assets/preview/boss-drop.html').write_text(page)
    manifest_path=ROOT/'assets/sprites/manifest.json'
    manifest=json.loads(manifest_path.read_text());manifest['entries']=entries()
    manifest_path.write_text(json.dumps(manifest,indent=2)+'\n')
    print('Boss drop: 3 boss strips, 4 VFX strips, manifest and previews written.')

if __name__=='__main__':main()
