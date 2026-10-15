"""Runtime sprite contract transcribed from docs/sprite-spec.md (never modifies it)."""
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]
BASE = ['0B1020','0B1530','14234A','22345F','2F4A7F','3FE0FF','F2E3C2','C9B48C','FF5A5A','B8303A','F6C9A0','E0A274','A86B4A']
COLORS = {
    'teal': ['0E5E66','1A9AA0','4FD1C5'],
    'violet': ['4B2A7A','7B4BC4','B08CF0'],
    'amber': ['8A5A00','E0A21B','FFD25E'],
    'mint': ['1F6B45','3FBF7F','8EF0B5'],
    'orange': ['8A3A10','E06A1F','FFA25E'],
}
PLAYER = {'idle':(4,6,True),'run':(8,12,True),'jump':(2,10,False),'fall':(2,10,True),'attack':(5,18,False),'dash':(4,16,False),'hurt':(2,10,False),'down':(4,4,True),'revive':(4,10,False)}
SKILLS = {
    'teal': {'id':'one-line-fix', 'title':'Em sửa có một dòng', 'height':54,
             'effects':[('keyboard-cleave',80,80,16,False),('keyboard-charge',64,64,10,True)]},
    'violet': {'id':'out-of-scope', 'title':'Không thuộc scope em', 'height':42,
               'effects':[('forward-shield',48,64,12,True),('returned-task',32,24,16,True)]},
    'amber': {'id':'meeting-minutes', 'title':'Theo biên bản đã chốt', 'height':50,
              'effects':[('stamp-projectile',32,32,14,True),('boss-stamp-mark',48,48,8,True)]},
    'mint': {'id':'works-on-my-machine', 'title':'Máy em vẫn chạy', 'height':44,
             'effects':[('hotfix-burst',128,48,16,False),('hotfix-zone',128,48,8,True)]},
    'orange': {'id':'clock-out', 'title':'Em về đúng giờ', 'height':48,
               'effects':[('timecard-trail',96,32,16,True),('timecard-impact',48,48,18,False)]},
}
BOSS = {'idle':(4,6,True),'volley_windup':(4,8,False),'volley':(4,12,False),'wave_windup':(4,8,False),'wave':(4,12,False),'wind_clock':(6,8,True),'hurt':(2,10,False),'defeat':(6,8,False)}
BOSS_MOVEMENT = {'walk':(4,8,True),'charge_windup':(4,5,False),'charge':(4,12,True),'charge_recover':(4,4,False)}
BOSS_MOVEMENT_VFX = {'boss-foot-dust':(64,32,4,12,False),'charge-warning':(64,16,4,8,True)}
BOSS_DROP = {'drop_windup':(4,8,False),'drop':(4,12,False),'drop_jam':(6,8,True)}
BOSS_DROP_VFX = {'file-stack':(64,48,4,12,True),'file-burst':(96,48,5,18,False),
                 'drop-warning':(64,32,2,6,True),'alert-mark':(16,24,2,6,True)}
BOSS_KHABIB = {'idle':(4,6,True),'walk':(4,8,True),'press_windup':(4,8,False),'press':(4,12,False),
               'spam_windup':(4,8,False),'spam':(4,12,True),'teleport_out':(2,8,False),
               'teleport_in':(2,10,False),'recover':(2,6,True),'hurt':(2,10,False),'defeat':(6,8,False)}
BOSS_KHABIB_VFX = {'khabib-press-wave':(48,48,4,14,True),'khabib-nag-projectile':(32,24,4,12,True),
                   'khabib-teleport-warning':(64,32,4,8,True),'khabib-portal':(48,48,4,12,False)}
BOSS_VOUGHT = {'idle':(4,6,True),'walk':(4,8,True),'deflect_windup':(4,8,False),'deflect':(4,12,False),
               'stamp_throw':(4,12,False),'shoulder_pat':(2,8,False),'reflect_shield':(4,10,True),
               'recover':(2,6,True),'hurt':(2,10,False),'defeat':(6,8,False)}
BOSS_VOUGHT_VFX = {'vought-deflect-ribbon':(48,48,4,12,True),'vought-bug-stamp':(32,32,4,12,True),
                   'vought-speech-shield':(48,48,4,10,True),'vought-pat-mark':(32,32,4,8,True)}
BOSS_YASUKA = {'idle':(4,6,True),'walk':(4,8,True),'dodge_windup':(4,8,False),'dodge':(4,12,False),
               'procedure':(4,12,False),'silence':(4,8,True),'recover':(2,6,True),
               'hurt':(2,10,False),'defeat':(6,8,False)}
BOSS_YASUKA_VFX = {'yasuka-silence-dome':(64,48,4,10,True),'yasuka-dodge-trail':(48,48,4,12,False),
                   'yasuka-stamp-trap':(32,32,4,8,True),'yasuka-paper-flurry':(48,48,4,12,False)}
VFX = {'task-card':(16,12,4,12,True),'deadline-wave':(32,32,4,12,True),'floor-warning':(32,8,2,6,True),'swing-arc':(48,48,4,18,False),'code-pulse':(32,16,4,16,True),'land-dust':(32,16,4,16,False),'hit-spark':(24,24,3,18,False)}
PROPS = {'desk':(96,48),'chair':(32,48),'monitor':(32,32),'plant':(32,48),'filing-cabinet':(32,48),'water-cooler':(32,48),'ceiling-light':(64,16),'poster':(32,48)}

def rgb(value):
    return tuple(bytes.fromhex(value))

def entries():
    out = []
    def add(path, w, h, n=1, fps=0, loop=False, kind='prop'):
        out.append(dict(path=path, frameWidth=w, frameHeight=h, frames=n, fps=fps, loop=loop, kind=kind))
    for color in COLORS:
        for name,(n,fps,loop) in PLAYER.items():
            add(f'assets/sprites/player/{color}/{name}.png',64,64,n,fps,loop,'player')
    for color, skill in SKILLS.items():
        add(f'assets/sprites/player/{color}/skill-windup.png',64,64,4,8,False,'player-skill')
        add(f'assets/sprites/player/{color}/skill-cast.png',64,64,4,16,False,'player-skill')
        for name,w,h,fps,loop in skill['effects']:
            add(f'assets/sprites/vfx/skills/{name}.png',w,h,4,fps,loop,'vfx')
        add(f'assets/ui/skills/{skill["id"]}.png',24,24,kind='ui')
    for name,(n,fps,loop) in BOSS.items():
        add(f'assets/sprites/boss/doc-cong/{name}.png',160,160,n,fps,loop,'boss')
    for name,(n,fps,loop) in BOSS_MOVEMENT.items():
        add(f'assets/sprites/boss/doc-cong/{name}.png',160,160,n,fps,loop,'boss')
    for name,(w,h,n,fps,loop) in BOSS_MOVEMENT_VFX.items():
        add(f'assets/sprites/vfx/{name}.png',w,h,n,fps,loop,'vfx')
    for name,(n,fps,loop) in BOSS_DROP.items():
        add(f'assets/sprites/boss/doc-cong/{name}.png',160,160,n,fps,loop,'boss')
    for name,(w,h,n,fps,loop) in BOSS_DROP_VFX.items():
        add(f'assets/sprites/vfx/{name}.png',w,h,n,fps,loop,'vfx')
    for name,(n,fps,loop) in BOSS_KHABIB.items():
        add(f'assets/sprites/boss/khabib/{name}.png',160,160,n,fps,loop,'boss')
    for name,(w,h,n,fps,loop) in BOSS_KHABIB_VFX.items():
        add(f'assets/sprites/vfx/{name}.png',w,h,n,fps,loop,'vfx')
    for name,(n,fps,loop) in BOSS_VOUGHT.items():
        add(f'assets/sprites/boss/vought/{name}.png',160,160,n,fps,loop,'boss')
    for name,(w,h,n,fps,loop) in BOSS_VOUGHT_VFX.items():
        add(f'assets/sprites/vfx/{name}.png',w,h,n,fps,loop,'vfx')
    for name,(n,fps,loop) in BOSS_YASUKA.items():
        add(f'assets/sprites/boss/yasuka/{name}.png',160,160,n,fps,loop,'boss')
    for name,(w,h,n,fps,loop) in BOSS_YASUKA_VFX.items():
        add(f'assets/sprites/vfx/{name}.png',w,h,n,fps,loop,'vfx')
    for name,(w,h,n,fps,loop) in VFX.items():
        add(f'assets/sprites/vfx/{name}.png',w,h,n,fps,loop,'vfx')
    add('assets/sprites/items/coffee.png',16,16,4,6,True,'item')
    add('assets/sprites/items/document.png',16,16,kind='item')
    add('assets/tiles/office.png',256,64,kind='tiles')
    for name,(w,h) in PROPS.items():
        add(f'assets/sprites/props/{name}.png',w,h)
    add('assets/sprites/props/elevator.png',64,96,4,8,False)
    for name in ['bg-0-sky','bg-1-windows','bg-2-office']:
        add(f'assets/backgrounds/{name}.png',1920,1080,kind='background')
    add('assets/ui/heart.png',9,8,2,kind='ui')
    add('assets/ui/portrait.png',24,24,kind='portrait')
    for color in list(COLORS)[1:]:
        add(f'assets/ui/portrait-{color}.png',24,24,kind='portrait')
    add('assets/ui/deadline-icon.png',16,16,kind='ui')
    return out
