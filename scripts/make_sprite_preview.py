#!/usr/bin/env python3
"""Every runtime image at exactly4x nearest-neighbor; full-size backgrounds included."""
from PIL import Image, ImageDraw, ImageFont
from sprite_manifest import ROOT, entries

def main():
    width=8000; margin=20; gap=24
    placed=[]; x=margin; y=margin; row_h=0
    for e in entries():
        w=e['frameWidth']*4; h=e['frameHeight']*4
        block_w=e['frames']*(w+8)-8; block_h=h+48
        if x+block_w>width-margin:
            x=margin; y+=row_h+gap; row_h=0
        placed.append((e,x,y,block_w,block_h))
        x+=block_w+gap; row_h=max(row_h,block_h)
    height=y+row_h+margin
    sheet=Image.new('RGB',(width,height),'#0B1530')
    draw=ImageDraw.Draw(sheet)
    try: font=ImageFont.truetype('/System/Library/Fonts/Menlo.ttc',16)
    except OSError: font=ImageFont.load_default(size=16)
    for e,x,y,bw,bh in placed:
        draw.text((x,y),e['path'],fill='#F2E3C2',font=font)
        im=Image.open(ROOT/e['path']).convert('RGBA')
        for i in range(e['frames']):
            fw=e['frameWidth']; fh=e['frameHeight']
            f=im.crop((i*fw,0,(i+1)*fw,fh)).resize((fw*4,fh*4),Image.Resampling.NEAREST)
            px=x+i*(fw*4+8)
            draw.text((px,y+22),f'{i}',fill='#3FE0FF',font=font)
            sheet.paste(f,(px,y+48),f)
    destination=ROOT/'assets/preview/contact-sheet.png'
    destination.parent.mkdir(parents=True,exist_ok=True)
    sheet.save(destination)
    print(f'Contact sheet: {width}x{height}, all {len(placed)} assets at4x nearest-neighbor.')

if __name__=='__main__': main()
