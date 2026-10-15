#!/usr/bin/env python3
"""Show all five runtime silhouettes and every player animation frame at 4x."""
from PIL import Image, ImageDraw
from sprite_manifest import ROOT, COLORS, PLAYER

def main():
    sheet = Image.new('RGB', (2180, 5 * 9 * 286 + 20), '#0B1530')
    draw = ImageDraw.Draw(sheet)
    lineup = Image.new('RGB', (1300, 320), '#0B1530')
    ld = ImageDraw.Draw(lineup)
    row = 0
    for col, color in enumerate(COLORS):
        for name, (count, _, _) in PLAYER.items():
            y = 20 + row * 286
            draw.text((12, y), f'{color} / {name}', fill='#F2E3C2')
            strip = Image.open(ROOT / f'assets/sprites/player/{color}/{name}.png')
            for i in range(count):
                frame = strip.crop((i*64, 0, (i+1)*64, 64)).resize((256, 256), Image.Resampling.NEAREST)
                sheet.paste(frame, (12+i*268, y+20), frame)
                if name == 'idle' and i == 0:
                    lineup.paste(frame, (col*260, 40), frame)
                    ld.text((col*260+20, 20), color, fill='#F2E3C2')
            row += 1
        sheet.crop((0, 20+col*9*286, 2180, (col+1)*9*286+20)).save(ROOT / f'assets/preview/player-{color}-v2-frames.png')
    sheet.save(ROOT / 'assets/preview/player-roster-v2-all-frames.png')
    lineup.save(ROOT / 'assets/preview/player-roster-v2.png')

if __name__ == '__main__':
    main()
