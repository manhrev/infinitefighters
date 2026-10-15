#!/usr/bin/env python3
"""Exact RGB substitution; geometry, alpha, face and equipment remain byte-identical."""
import numpy as np
from PIL import Image
from sprite_manifest import ROOT, COLORS, PLAYER, rgb

def swap(source, destination, colors):
    original=np.array(Image.open(source).convert('RGBA'))
    result=original.copy()
    for old,new in zip(COLORS['teal'],colors):
        mask=np.all(original[:,:,:3]==rgb(old),axis=2)&(original[:,:,3]>0)
        result[mask,:3]=rgb(new)
    destination.parent.mkdir(parents=True,exist_ok=True)
    Image.fromarray(result).save(destination)

def main():
    if (ROOT/'assets/source/player-teal-v2.png').is_file():
        print('Distinct player roster active; palette cloning skipped.')
        return
    count=0
    for color,colors in COLORS.items():
        if color=='teal': continue
        for name in PLAYER:
            swap(ROOT/f'assets/sprites/player/teal/{name}.png',ROOT/f'assets/sprites/player/{color}/{name}.png',colors)
            count+=1
        swap(ROOT/'assets/ui/portrait.png', ROOT/f'assets/ui/portrait-{color}.png', colors)
        count+=1
    print(f'Palette swap: {count} PNGs written (4 variants, 9 strips + portrait each).')

if __name__=='__main__': main()
