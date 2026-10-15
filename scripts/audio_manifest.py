"""MVP audio contract, matching the audio tables in docs/game-scenario.md."""
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]
SAMPLE_RATE = 48000
BPM = 112
BEAT = 60 / BPM

# duration, recommended gain, priority (higher wins), event
SFX = {
    'footstep-1': (.075,.24,1,'player.footstep.1'),
    'footstep-2': (.082,.24,1,'player.footstep.2'),
    'jump': (.16,.52,4,'player.jump'),
    'land': (.17,.40,3,'player.land'),
    'dash': (.18,.50,4,'player.dash'),
    'keyboard-swing': (.15,.45,3,'player.attack.swing'),
    'keyboard-hit': (.12,.70,5,'combat.keyboard.hit'),
    'code-pulse': (.21,.55,4,'player.code-pulse'),
    'player-hurt': (.24,.65,6,'player.hurt'),
    'player-down': (.48,.72,9,'player.down'),
    'revive': (.62,.67,8,'player.revived'),
    'coffee': (.28,.48,5,'pickup.coffee'),
    'document': (.25,.48,5,'pickup.document'),
    'checkpoint': (.66,.60,7,'checkpoint.activated'),
    'ui-confirm': (.10,.38,4,'ui.confirm'),
    'ui-back': (.10,.38,4,'ui.back'),
    'ping': (.18,.45,7,'team.ping'),
    'deadline-tick': (.065,.55,9,'deadline.tick'),
    'victory': (3.2,.65,10,'match.victory'),
    'defeat': (3.0,.65,10,'match.defeat'),
    'boss-volley-windup': (.48,.80,10,'boss.volley.windup'),
    'boss-volley': (.30,.65,7,'boss.volley.release'),
    'boss-wave-windup': (.49,.85,10,'boss.wave.windup'),
    'boss-wave-slam': (.52,.80,9,'boss.wave.impact'),
    'boss-wind-clock': (.75,.35,5,'boss.wind-clock.start'),
    'boss-hurt': (.23,.42,5,'boss.hurt'),
    'boss-defeat': (1.6,.75,10,'boss.defeated'),
}

BOSS_MOVEMENT_SFX = {
    'boss-footstep': (.18,.30,1,'boss.walk.footstep'),
    'boss-charge-windup': (.90,.80,10,'boss.charge.windup'),
    'boss-charge': (.60,.60,7,'boss.charge.start'),
}
SFX.update(BOSS_MOVEMENT_SFX)

BOSS_DROP_SFX = {
    'boss-drop-windup': (.90,.80,10,'boss.drop.windup'),
    'boss-drop-impact': (.25,.65,7,'boss.drop.impact'),
    'boss-paper-jam': (1.20,.40,5,'boss.drop.jam'),
}
SFX.update(BOSS_DROP_SFX)

ROLES = {
    'teal': dict(slot='P1',name='Người gánh task',skill='Em sửa có một dòng',cooldownSeconds=6,
        chargeMaxSeconds=1,chargeCancelOnDash=True,identity='Arpeggio phím cơ, tích lực lên cao, cú bổ trầm.'),
    'violet': dict(slot='P2',name='Chuyên viên chuyển tiếp',skill='Không thuộc scope em',cooldownSeconds=8,
        shieldSeconds=.6,maxReflections=3,identity='Bass tròn, âm khiên trong và tiếng thư bật ngược.'),
    'amber': dict(slot='P3',name='Người chốt biên bản',skill='Theo biên bản đã chốt',cooldownSeconds=10,
        markSeconds=4,damageMultiplier=1.15,markStacks=False,identity='Nhịp chắc, tiếng dấu đóng và chốt khóa.'),
    'mint': dict(slot='P4',name='Người cứu production',skill='Máy em vẫn chạy',cooldownSeconds=14,
        zoneSeconds=3,healHeartsPerPlayerPerCast=1,revives=False,identity='Chuông ấm, hợp âm sáng, vùng hotfix êm.'),
    'orange': dict(slot='P5',name='Người đòi tan làm',skill='Em về đúng giờ',cooldownSeconds=7,
        maxDashTiles=3,invulnerabilitySeconds=.15,sharesDashCooldown=True,
        identity='Bass bước nhanh, quẹt thẻ và gió lao ngang.'),
}

# duration, gain, priority, role, exact event hook. Shared cues have no role.
SKILL_SFX = {
    'teal-charge': (1.,.42,4,'teal','skill.teal.charge.start'),
    'teal-smash': (.30,.58,5,'teal','skill.teal.release'),
    'teal-smash-hit': (.19,.65,6,'teal','skill.teal.hit'),
    'teal-charge-cancel': (.12,.25,3,'teal','skill.teal.charge.cancel'),
    'violet-shield': (.20,.48,5,'violet','skill.violet.shield.start'),
    'violet-reflect': (.14,.50,6,'violet','skill.violet.projectile.reflected'),
    'violet-shield-end': (.13,.22,3,'violet','skill.violet.shield.end'),
    'amber-stamp-cast': (.18,.45,4,'amber','skill.amber.cast'),
    'amber-mark-hit': (.28,.60,6,'amber','skill.amber.mark.applied'),
    'amber-mark-expire': (.17,.20,2,'amber','skill.amber.mark.expired'),
    'mint-hotfix-deploy': (.38,.52,5,'mint','skill.mint.zone.created'),
    'mint-hotfix-zone': (1.,.15,2,'mint','skill.mint.zone.active'),
    'mint-hotfix-heal': (.34,.45,6,'mint','skill.mint.player.healed'),
    'mint-hotfix-end': (.22,.20,2,'mint','skill.mint.zone.expired'),
    'orange-timecard-dash': (.27,.55,5,'orange','skill.orange.dash.start'),
    'orange-timecard-hit': (.17,.58,6,'orange','skill.orange.target.hit'),
    'orange-timecard-stop': (.16,.24,3,'orange','skill.orange.dash.end'),
    'skill-ready': (.16,.20,2,None,'skill.cooldown.ready.local'),
    'skill-unavailable': (.095,.20,2,None,'skill.unavailable.local'),
}

SKILL_BINDINGS = {
    'teal': dict(charge='teal-charge',release='teal-smash',hit='teal-smash-hit',cancel='teal-charge-cancel'),
    'violet': dict(start='violet-shield',reflect='violet-reflect',end='violet-shield-end'),
    'amber': dict(cast='amber-stamp-cast',markApplied='amber-mark-hit',markExpired='amber-mark-expire'),
    'mint': dict(deploy='mint-hotfix-deploy',activeLoop='mint-hotfix-zone',heal='mint-hotfix-heal',end='mint-hotfix-end'),
    'orange': dict(start='orange-timecard-dash',hit='orange-timecard-hit',end='orange-timecard-stop'),
}

def entries():
    result=[]
    for name,bars,gain in [('lobby',16,.35),('sprint',24,.35),('boss',32,.35),('urgency',8,.30)]:
        result.append(dict(key=name,category='music',path=f'assets/audio/music/{name}.ogg',
            seconds=bars*4*BEAT,channels=2,loop=True,bpm=BPM,bars=bars,gain=gain,
            priority=2,event=f'music.{name}',optional=name=='urgency'))
    for color,role in ROLES.items():
        result.append(dict(key=f'player-{color}',category='music',path=f'assets/audio/music/player-{color}.ogg',
            seconds=16*4*BEAT,channels=2,loop=True,bpm=BPM,bars=16,gain=.35,priority=2,
            event=f'music.player.{color}.intro',optional=True,pack='player-skills',player=color,
            title=f"{role['slot']} · {role['name']}",description=role['identity']))
    result.append(dict(key='office',category='ambience',path='assets/audio/ambience/office.ogg',
        seconds=24.,channels=2,loop=True,gain=.18,priority=0,event='ambience.office',optional=False))
    for name,(seconds,gain,priority,event) in SFX.items():
        result.append(dict(key=name,category='sfx',path=f'assets/audio/sfx/{name}.ogg',
            seconds=seconds,channels=1,loop=name=='boss-wind-clock',gain=gain,
            priority=priority,event=event,optional=False))
        if name in BOSS_MOVEMENT_SFX:
            result[-1]['pack']='boss-movement'
        if name in BOSS_DROP_SFX:
            result[-1]['pack']='boss-drop'
    for name,(seconds,gain,priority,color,event) in SKILL_SFX.items():
        result.append(dict(key=name,category='sfx',path=f'assets/audio/sfx/{name}.ogg',seconds=seconds,
            channels=1,loop=name=='mint-hotfix-zone',gain=gain,priority=priority,event=event,
            optional=False,pack='player-skills',player=color,
            title=(f"{ROLES[color]['slot']} · {ROLES[color]['skill']}" if color else 'Tín hiệu kỹ năng chung')))
    for e in result:
        relative=Path(e['category'])/f"{e['key']}.wav"
        e['source']=str(Path('assets/audio/source')/relative)
        e['fallback']=e['path'].removesuffix('.ogg')+'.mp3'
        e['frames']=round(e['seconds']*SAMPLE_RATE)
    return result
