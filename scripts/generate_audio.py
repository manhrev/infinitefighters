#!/usr/bin/env python3
"""Original deterministic chiptune score + synthesized office effects.

No downloaded samples, external audio service or API key. Outputs PCM16 WAV
masters, OGG Vorbis runtime files, MP3 fallbacks and listening preview.
"""
import argparse
import html
import json
import math
import subprocess
from pathlib import Path

import imageio_ffmpeg
import numpy as np
import soundfile as sf
from audio_manifest import ROOT, SAMPLE_RATE as SR, BPM, BEAT, ROLES, SKILL_SFX, SKILL_BINDINGS, entries

RNG=np.random.default_rng(15102026)

def seconds(n): return np.arange(round(n*SR),dtype=np.float64)/SR
def hz(midi): return 440*2**((midi-69)/12)

def envelope(x,attack=.002,release=.012):
    x=x.copy();n=len(x)
    a=min(round(attack*SR),n//2);r=min(round(release*SR),n//2)
    if a: x[:a]*=np.linspace(0,1,a)
    if r: x[-r:]*=np.linspace(1,0,r)
    return x

def oscillator(freq,duration,voice='sine'):
    t=seconds(duration)
    if np.isscalar(freq): phase=t*freq
    else: phase=np.cumsum(freq)/SR
    if voice=='sine': return np.sin(2*np.pi*phase)
    if voice=='triangle':
        out=np.zeros(len(t))
        for k in [1,3,5,7]: out+=(-1)**((k-1)//2)*np.sin(2*np.pi*k*phase)/(k*k)
        return out*.81
    if voice=='pulse':
        # Limited harmonics: bright console lead without an unbounded square-wave spectrum.
        out=np.zeros(len(t))
        maxfreq=float(np.max(freq)) if not np.isscalar(freq) else freq
        for k in [1,3,5,7]:
            if k*maxfreq<SR*.45: out+=np.sin(2*np.pi*k*phase)/k
        return out*.70
    if voice=='bell':
        return (np.sin(2*np.pi*phase)+.32*np.sin(2*np.pi*phase*2.01)+.12*np.sin(2*np.pi*phase*3.98))*.65
    raise ValueError(voice)

def tone(midi,duration,voice='triangle',decay=5):
    t=seconds(duration)
    x=oscillator(hz(midi),duration,voice)*np.exp(-t*decay)
    return envelope(x,attack=.003,release=min(.04,duration/4))

def glide(f0,f1,duration,voice='sine',decay=6):
    t=seconds(duration)
    f=f0*(f1/f0)**(t/max(duration,1e-8))
    return envelope(oscillator(f,duration,voice)*np.exp(-decay*t),.002,.012)

def noise(duration,kind='paper'):
    x=RNG.normal(0,1,round(duration*SR))
    size=9 if kind=='paper' else 45
    low=np.convolve(x,np.ones(size)/size,mode='same')
    out=(x-low)*.36 if kind=='paper' else low*2.0
    return out

def click(duration=.045,pitch=1750):
    t=seconds(duration)
    return envelope((noise(duration)*.55+oscillator(pitch,duration)*.28)*np.exp(-t*75),.001,.004)

def impact(duration=.16,weight=90):
    t=seconds(duration)
    return envelope(glide(weight*2,weight*.7,duration,decay=24)*.65+noise(duration,'low')*.40*np.exp(-t*35),.001,.015)

def paper(duration=.15,rise=False):
    t=seconds(duration)
    env=np.sin(np.pi*t/duration)**(.8 if rise else .35)
    # A narrow comb-like flutter turns filtered noise into a papery rather than explosive texture.
    flutter=.45+.55*np.sin(2*np.pi*(31*t+75*t*t))**2
    return envelope(noise(duration)*env*flutter,.002,.012)

def add(dest,x,start,gain=1.,pan=0.,wrap=False):
    index=round(start*SR)
    if dest.ndim==2:
        if x.ndim==1:
            angle=(pan+1)*np.pi/4
            x=np.column_stack((x*np.cos(angle),x*np.sin(angle)))
    elif x.ndim==2: x=x.mean(axis=1)
    x=x*gain
    if wrap:
        index%=len(dest)
        first=min(len(x),len(dest)-index)
        dest[index:index+first]+=x[:first]
        remaining=x[first:]
        if len(remaining): dest[:len(remaining)]+=remaining
    else:
        if index<0:
            x=x[-index:];index=0
        n=min(len(x),len(dest)-index)
        if n>0: dest[index:index+n]+=x[:n]

def seamless(x):
    """Five-ms smooth periodic join without adding silence or changing loop length."""
    x=x.copy();n=min(240,len(x)//8)
    u=np.linspace(0,1,n);s=u*u*(3-2*u)
    if x.ndim==2:s=s[:,None]
    midpoint=(x[0]+x[-1])/2
    x[:n]=midpoint*(1-s)+x[:n]*s
    x[-n:]=x[-n:]*(1-s)+midpoint*s
    x[0]=midpoint;x[1]=midpoint;x[-1]=midpoint;x[-2]=midpoint
    return x

def finish(x,peak=.55,loop=False):
    if loop: x=seamless(x)
    else: x=envelope(x,.0015,.012)
    x-=np.mean(x,axis=0)
    if not loop: x=envelope(x,.0015,.012)
    maximum=np.max(np.abs(x))
    if maximum>0: x*=peak/maximum
    return x.astype(np.float32)

def make_sfx(name,duration):
    if name in SKILL_SFX: return make_skill_sfx(name,duration)
    x=np.zeros(round(duration*SR))
    if name.startswith('footstep'):
        add(x,impact(duration,90 if name.endswith('1') else 110),0,.75)
        add(x,click(.025,1550 if name.endswith('1') else 1870),.004,.25)
    elif name=='jump': add(x,glide(260,900,duration,'triangle',4),0,.8)
    elif name=='land':
        add(x,impact(duration,80),0,.8);add(x,click(.045,1100),.003,.22)
    elif name=='dash':
        add(x,paper(duration),0,.7);add(x,glide(460,150,duration,'triangle',10),0,.3)
    elif name=='keyboard-swing':
        add(x,paper(duration),0,.7)
        for t in [0,.021,.047]:add(x,click(.04,1300+t*9000),t,.22)
    elif name=='keyboard-hit':
        add(x,impact(duration,125),0,.9)
        for t,p in [(0,1900),(.015,2400),(.03,1350)]:add(x,click(.055,p),t,.4)
    elif name=='code-pulse':
        add(x,glide(1100,330,duration,'pulse',12),0,.5)
        add(x,tone(86,.10,'bell',25),.02,.4)
    elif name=='player-hurt':
        add(x,glide(600,180,duration,'pulse',13),0,.7);add(x,impact(.09,120),0,.3)
    elif name=='player-down':
        for i,m in enumerate([69,64,57,45]):add(x,tone(m,.17,'triangle',14),i*.095,.65)
    elif name=='revive':
        for i,m in enumerate([62,69,74]):add(x,tone(m,.30,'bell',8),i*.14,.65)
    elif name=='coffee':
        add(x,glide(190,105,.11,'sine',10),0,.65)
        for i,m in enumerate([76,81]):add(x,tone(m,.15,'bell',18),.085+i*.07,.4)
    elif name=='document':
        add(x,paper(.11),0,.55);add(x,tone(86,.14,'triangle',18),.08,.4)
    elif name=='checkpoint':
        add(x,tone(81,.42,'bell',7),0,.65);add(x,tone(76,.40,'bell',8),.22,.58)
    elif name.startswith('ui-'):
        add(x,glide(600,1250 if name=='ui-confirm' else 330,duration,'triangle',14),0,.7)
    elif name=='ping': add(x,tone(88,duration,'bell',15),0,.65)
    elif name=='deadline-tick':
        add(x,click(duration,2450),0,.8);add(x,tone(91,.035,'triangle',70),0,.25)
    elif name=='boss-footstep':
        # Heavy shoe on office metal: low thud followed by a short sole/casing click.
        add(x,impact(duration,54),0,.85)
        add(x,click(.035,740),.004,.22)
        add(x,tone(40,.12,'triangle',30),0,.15)
    elif name=='boss-charge-windup':
        # Urgent meeting invitation: three bright office-bell notes, distinct
        # from the papery volley and the wave's low two-clank warning.
        for i,midi in enumerate([84,91,88]):
            add(x,tone(midi,.44,'bell',8),i*.18,.55)
            add(x,click(.025,2100+i*180),i*.18,.06)
        add(x,tone(72,.30,'triangle',16),0,.12)
    elif name=='boss-charge':
        t=seconds(duration)
        wind=np.sin(np.pi*t/duration)**.55
        add(x,envelope(noise(duration,'low')*wind,.003,.025),0,.65)
        add(x,paper(duration),0,.30)
        add(x,glide(650,110,duration,'triangle',5),0,.24)
        # A shaking stopwatch rides over the horizontal whoosh, without an impact.
        rattle=envelope(oscillator(1120,duration)*(1+.45*np.sin(2*np.pi*28*t))*np.exp(-t*7),.002,.025)
        add(x,rattle,0,.10)
    elif name=='boss-drop-windup':
        # Two rising, soft digital notification notes; the charge cue is a
        # bright three-note meeting bell. A separate tick finishes this cue.
        for start,midi in [(0,76),(.32,83)]:
            d=.48;t=seconds(d)
            body=(oscillator(hz(midi),d,'sine')*.78+
                oscillator(hz(midi)*2,d,'sine')*.16)*np.exp(-t*7)
            add(x,envelope(body,.004,.045),start,.58)
        add(x,click(.055,2700),.75,.18)
        add(x,tone(95,.055,'triangle',55),.75,.06)
    elif name=='boss-drop-impact':
        # Dry wooden desk/body thud, a broad folder slap and small paper bits.
        add(x,impact(.18,115),0,.78)
        for f in [220,470,730]:
            d=.17;t=seconds(d)
            add(x,envelope(oscillator(f,d)*np.exp(-t*32),.001,.015),0,.12)
        add(x,paper(.15),.012,.50)
        for start in [.09,.145,.18]:add(x,paper(.06),start,.16)
    elif name=='boss-paper-jam':
        # Three slowing motor stalls, followed by one irritated paper tug.
        for i,start in enumerate([0,.235,.47]):
            d=.21;t=seconds(d);fundamental=95-i*12
            motor=(oscillator(fundamental,d,'pulse')*.48+
                noise(d,'low')*.32)*(0.22+.78*np.sin(2*np.pi*(18-i*3)*t)**2)
            add(x,envelope(motor,.002,.018),start,.65)
            add(x,click(.035,700-i*90),start,.12)
        add(x,paper(.26,True),.83,.58)
        add(x,glide(460,160,.22,'triangle',13),.91,.20)
    elif name=='boss-volley-windup':
        for i in range(3):
            add(x,paper(.17,True),i*.14,.4+i*.16)
            add(x,glide(1000+i*260,1500+i*300,.12,'triangle',20),i*.14,.12)
    elif name=='boss-volley':
        add(x,paper(duration),0,.85);add(x,glide(1700,400,.2,'triangle',17),0,.15)
    elif name=='boss-wave-windup':
        for i in range(2):
            add(x,impact(.18,140-i*30),i*.23,.7)
            for f in [420,670]:add(x,envelope(oscillator(f,.16)*np.exp(-seconds(.16)*22)),i*.23,.16)
    elif name=='boss-wave-slam':
        add(x,impact(.45,62),0,.9)
        for f in [510,783,1103]:add(x,envelope(oscillator(f,.48)*np.exp(-seconds(.48)*14)),0,.15)
        add(x,glide(1600,110,.40,'triangle',12),.06,.28)
        add(x,paper(.30),.09,.28)
    elif name=='boss-wind-clock':
        for i in range(6):
            add(x,click(.055,1100+(i%2)*430),i*.125,.6)
            add(x,tone(55+(i%2)*3,.08,'triangle',35),i*.125,.35)
    elif name=='boss-hurt':
        for f in [920,1420,2190]:add(x,envelope(oscillator(f,duration)*np.exp(-seconds(duration)*25)),0,.22)
        add(x,impact(.10,125),0,.4)
    elif name=='boss-defeat':
        add(x,glide(1400,120,.80,'triangle',4),0,.45)
        for i in range(7):add(x,click(.07,1900-i*180),.12+i*.095,.23)
        for i in range(4):add(x,paper(.21),.55+i*.20,.3)
        add(x,impact(.40,70),1.0,.55)
    elif name=='victory':
        for i,m in enumerate([65,69,72,77]):add(x,tone(m,.65,'pulse',5),i*.18,.30)
        for m in [53,65,69,72]:add(x,tone(m,1.7,'bell',3),.88,.25)
        add(x,tone(84,.7,'bell',4),2.1,.35);add(x,tone(79,.65,'bell',5),2.4,.3)
    elif name=='defeat':
        for i,m in enumerate([69,68,65,61,50]):add(x,tone(m,.6,'triangle',6),i*.23,.40)
        for i in range(6):add(x,click(.07,850-i*85),1.50+i*.075,.3)
        add(x,glide(180,45,.7,'triangle',6),2.0,.45)
    else: raise ValueError(name)
    return finish(x,.60,loop=name=='boss-wind-clock')

def make_skill_sfx(name,duration):
    x=np.zeros(round(duration*SR))
    if name=='teal-charge':
        t=seconds(duration)
        build=glide(120,620,duration,'triangle',.2)*(.10+.65*t/duration)
        add(x,build,0,.62);add(x,click(.045,1250),0,.18)
        for i in range(7):
            add(x,click(.04,1000+i*180),.12+i*.11,.04+i*.012)
        add(x,tone(86,.09,'bell',18),.90,.15)
    elif name=='teal-smash':
        add(x,glide(820,105,duration,'pulse',11),0,.35)
        add(x,paper(.21),0,.55);add(x,impact(.19,65),.08,.30)
    elif name=='teal-smash-hit':
        add(x,impact(duration,56),0,.85)
        for i,f in enumerate([780,1175,1590]):
            add(x,envelope(oscillator(f,.16)*np.exp(-seconds(.16)*25)),i*.006,.20)
        add(x,click(.055,1750),0,.35)
    elif name=='teal-charge-cancel':
        add(x,glide(370,90,duration,'triangle',22),0,.65)
    elif name=='violet-shield':
        for i,m in enumerate([69,76,81]):add(x,tone(m,.15,'bell',18),i*.025,.30)
        add(x,paper(.11),0,.25)
    elif name=='violet-reflect':
        add(x,glide(1450,270,duration,'bell',18),0,.48)
        add(x,glide(350,1800,.085,'triangle',18),.025,.30)
        add(x,paper(.07),0,.35)
    elif name=='violet-shield-end':
        for m in [69,76]:add(x,tone(m,duration,'bell',30),0,.30)
    elif name=='amber-stamp-cast':
        add(x,paper(duration),0,.45);add(x,glide(320,720,.13,'triangle',20),0,.27)
    elif name=='amber-mark-hit':
        add(x,impact(.16,145),0,.9)
        add(x,click(.07,980),.015,.4)
        add(x,tone(74,.18,'bell',18),.08,.35)
        add(x,tone(81,.14,'triangle',22),.13,.22)
    elif name=='amber-mark-expire':
        add(x,click(.045,740),0,.5);add(x,glide(700,350,duration,'triangle',23),0,.26)
    elif name=='mint-hotfix-deploy':
        add(x,glide(180,740,.24,'sine',8),0,.3)
        for i,m in enumerate([65,72,77]):add(x,tone(m,.22,'bell',13),i*.055,.30)
        add(x,noise(.16,'low')*np.exp(-seconds(.16)*18),0,.2)
    elif name=='mint-hotfix-zone':
        # An integer number of carrier/LFO cycles creates a gentle one-second loop.
        t=seconds(duration)
        x=(.18*np.sin(2*np.pi*220*t)+.09*np.sin(2*np.pi*330*t))*(.75+.25*np.cos(2*np.pi*2*t))
        x+=.025*np.sin(2*np.pi*660*t)
    elif name=='mint-hotfix-heal':
        for i,m in enumerate([77,81,84]):add(x,tone(m,.20,'bell',14),i*.055,.30)
    elif name=='mint-hotfix-end':
        add(x,glide(520,220,duration,'sine',16),0,.5)
        add(x,tone(65,.14,'triangle',20),.04,.17)
    elif name=='orange-timecard-dash':
        add(x,glide(480,1900,.12,'triangle',16),0,.32)
        add(x,paper(.23),.025,.6)
        for i,p in enumerate([1100,790,520]):add(x,click(.04,p),.045+i*.035,.18)
    elif name=='orange-timecard-hit':
        add(x,impact(.14,100),0,.62);add(x,paper(.09),0,.4)
        add(x,glide(1400,320,.13,'pulse',22),.01,.22)
    elif name=='orange-timecard-stop':
        add(x,glide(650,180,duration,'triangle',18),0,.5)
        add(x,click(.07,800),.025,.28)
    elif name=='skill-ready':
        add(x,tone(81,.10,'triangle',24),0,.4);add(x,tone(88,.10,'bell',25),.045,.3)
    elif name=='skill-unavailable':
        add(x,tone(50,duration,'triangle',28),0,.45)
        add(x,tone(51,duration,'sine',30),0,.20)
    else: raise ValueError(name)
    return finish(x,.36 if name=='mint-hotfix-zone' else .56,loop=name=='mint-hotfix-zone')

def make_player_theme(color,frames,bars):
    """Five separately composed themes, not pitch shifts of the same music."""
    settings={
        'teal':dict(chords=[(50,[62,65,69]),(46,[58,62,65]),(53,[60,65,69]),(48,[60,64,67])],
            motif=[74,77,81,79,76,77,74,69],voice='pulse',lead=.15,bass=[0,1.5,2,3.5]),
        'violet':dict(chords=[(48,[60,63,67]),(44,[56,60,63]),(46,[58,62,65]),(43,[59,62,67])],
            motif=[75,72,67,72,74,70,67,62],voice='triangle',lead=.20,bass=[0,.75,2,2.75]),
        'amber':dict(chords=[(38,[62,65,69]),(46,[58,62,65]),(43,[59,62,67]),(45,[61,64,69])],
            motif=[74,74,69,77,76,74,73,69],voice='pulse',lead=.14,bass=[0,1,2,3]),
        'mint':dict(chords=[(53,[65,69,72,76]),(48,[64,67,72]),(50,[62,65,69,72]),(46,[62,65,69])],
            motif=[77,81,84,88,86,84,81,79],voice='bell',lead=.18,bass=[0,2]),
        'orange':dict(chords=[(45,[61,64,69]),(48,[60,64,67]),(50,[62,65,69]),(52,[64,68,71])],
            motif=[81,80,81,76,79,78,76,73],voice='triangle',lead=.20,bass=[0,.5,1,1.5,2,2.5,3,3.5]),
    }
    s=settings[color];out=np.zeros((frames,2))
    for bar in range(bars):
        root,chord=s['chords'][bar%4];base=bar*4*BEAT
        for i,pos in enumerate(s['bass']):
            extra=12 if color=='orange' and i%4==3 else (7 if color=='violet' and i%2 else 0)
            add(out,tone(root+extra,.27,'triangle',11),base+pos*BEAT,.23,wrap=True)
        for beat in range(4):
            pos=base+beat*BEAT
            if beat%2==0 or color=='amber':add(out,impact(.12,70 if color!='amber' else 58),pos,.18,wrap=True)
            else:add(out,click(.05,1400),pos,.10,pan=-.2,wrap=True)
        for i in range(8):
            if color=='mint' and i%2==0:continue
            add(out,click(.028,2350 if color=='teal' else 1800),base+i*.5*BEAT,.04,pan=.35,wrap=True)
        chord_positions=[.5,2.5] if color!='amber' else [1,3]
        for pos in chord_positions:
            for j,midi in enumerate(chord):
                add(out,tone(midi,.6 if color=='mint' else .37,'bell' if color=='mint' else 'triangle',8),base+pos*BEAT,.06,pan=-.6+j*.35,wrap=True)
            if color=='amber':add(out,impact(.10,140),base+pos*BEAT,.13,wrap=True)
        motif=s['motif']
        # Different rhythmic signatures, with a responding phrase every other bar.
        positions={'teal':[0,.5,1,1.5,2,2.5,3,3.5],
            'violet':[.25,.75,1.5,1.75,2.25,2.75,3.5,3.75],
            'amber':[0,.75,1,1.75,2,2.75,3,3.5],
            'mint':[0,.75,1.5,2,2.5,3,3.5,3.75],
            'orange':[0,.25,.75,1.5,2,2.25,2.75,3.5]}[color]
        transpose=[0,-2,3,-1][bar%4]
        for i,(pos,midi) in enumerate(zip(positions,motif)):
            if color=='mint' and i in [3,7]:continue
            if bar%8<4 and i==7:continue
            if bar%8>=4 and i in [2,6]:midi+=12 if color in ['teal','mint'] else -5
            if bar%2: midi=motif[-1-i]
            note=tone(midi+transpose,.38 if color=='mint' else .22,s['voice'],8 if color=='mint' else 12)
            add(out,note,base+pos*BEAT,s['lead'],pan=-.15,wrap=True)
            if color=='violet':add(out,note,base+pos*BEAT+.11,.045,pan=.55,wrap=True)
        if bar%4==3:
            for i in range(3):add(out,click(.035,1000+i*300),base+(3+i*.25)*BEAT,.055,wrap=True)
    return finish(out,.44,loop=True)

def make_music(name,frames,bars):
    if name.startswith('player-'):return make_player_theme(name.removeprefix('player-'),frames,bars)
    out=np.zeros((frames,2),dtype=np.float64)
    if name=='urgency':
        for i in range(bars*8):
            start=i*BEAT/2
            add(out,click(.05,1700+(i%2)*900),start,.30,pan=(-.1 if i%2 else .1),wrap=True)
            if i%4==0:add(out,tone(38,.20,'triangle',18),start,.35,wrap=True)
        return finish(out,.30,loop=True)
    progressions={
        'lobby':[(53,[65,69,72,76]),(50,[62,65,69,72]),(46,[58,62,65,69]),(48,[60,64,67,70])],
        'sprint':[(50,[62,65,69]),(46,[58,62,65]),(53,[60,65,69]),(48,[60,64,67])],
        'boss':[(38,[62,65,69]),(39,[63,67,70]),(46,[58,62,65]),(45,[61,64,69])],
    }
    melodies={
        'lobby':[[77,81,79,76,74,72],[74,77,81,79,77,76],[74,77,79,77,74,72],[76,79,82,79,76,72]],
        'sprint':[[74,77,81,79,77,76,74,69],[77,74,70,74,77,79,77,74],[77,81,84,81,79,77,76,72],[76,79,83,84,83,79,76,73]],
        'boss':[[74,69,74,77,76,74,69,65],[75,70,75,79,78,75,70,67],[77,74,70,74,77,79,77,74],[73,76,81,82,81,76,73,69]],
    }
    for bar in range(bars):
        root,chord=progressions[name][bar%4];base=bar*4*BEAT
        section=bar//8
        # Bass notes and a short printer-motor transient reinforce the backbeat.
        bass_positions=[0,1.5,2,3.5] if name!='boss' else [0,.5,1,1.5,2,2.5,3,3.5]
        for j,pos in enumerate(bass_positions):
            m=root+(12 if j%4==3 else 0)
            add(out,tone(m,.30 if name!='lobby' else .42,'triangle',10),base+pos*BEAT,.25,wrap=True)
            if name=='boss':add(out,noise(.05,'low')*np.exp(-seconds(.05)*50),base+pos*BEAT,.045,wrap=True)
        # Keyboard hats, metal-panel kick and snare: designed separately from player SFX.
        for eighth in range(8):
            start=base+eighth*.5*BEAT
            if name!='lobby' or eighth%2:
                add(out,click(.035,2600 if eighth%2 else 2100),start,.035 if name=='lobby' else .065,pan=.35,wrap=True)
        for beat in range(4):
            start=base+beat*BEAT
            if beat%2==0:add(out,impact(.14,68),start,.15 if name=='lobby' else .23,wrap=True)
            else:
                drum=envelope(noise(.10)*np.exp(-seconds(.10)*35),.001,.012)
                add(out,drum,start,.12 if name=='lobby' else .23,pan=-.12,wrap=True)
            if name=='boss':add(out,click(.05,1300+(beat%2)*700),start,.08,pan=-.3,wrap=True)
        # Short stereo chord stabs leave room for warnings and hit sounds.
        for pos in [.5,2.5]:
            for j,m in enumerate(chord):
                add(out,tone(m,.52,'triangle',7),base+pos*BEAT,.07,pan=(-.65+j*.4),wrap=True)
        motif=melodies[name][bar%4]
        positions=[.5,1,1.75,2.25,3,3.5] if name=='lobby' else [0,.5,1,1.75,2,2.5,3,3.5]
        for i,(pos,midi) in enumerate(zip(positions,motif)):
            # Section variation is composed, not a repeated single-bar placeholder.
            if name=='sprint' and section==0 and i in [3,7]:continue
            if name=='boss' and section%2 and i in [1,5]:midi+=12
            if name=='lobby' and bar%8>=4 and i==5:midi+=12
            duration=(.23 if name=='boss' else .31)
            voice='bell' if name=='lobby' else 'pulse'
            note=tone(midi,duration,voice,8 if name=='lobby' else 11)
            add(out,note,base+pos*BEAT,.14 if name=='lobby' else .17,pan=-.12,wrap=True)
            add(out,note,base+pos*BEAT+.16,.025,pan=.55,wrap=True)
        if bar%8==7:
            for i,m in enumerate([chord[0]+12,chord[1]+12,chord[2]+12]):
                add(out,tone(m,.22,'bell',12),base+(3+i*.25)*BEAT,.10,pan=.3,wrap=True)
        if name=='lobby' and bar in [5,13]:
            add(out,tone(88,.35,'bell',10),base+1.75*BEAT,.12,pan=.5,wrap=True)
    return finish(out,.44 if name=='lobby' else .48,loop=True)

def make_ambience(frames):
    # Integer-bin noise and fan frequencies are periodic across the24-second loop.
    frequencies=np.fft.rfftfreq(frames,1/SR)
    spectrum=(RNG.normal(size=len(frequencies))+1j*RNG.normal(size=len(frequencies)))
    shape=np.exp(-frequencies/470)/np.sqrt(np.maximum(frequencies,20))
    shape[frequencies<28]=0
    air=np.fft.irfft(spectrum*shape,n=frames);air/=np.max(np.abs(air))
    t=np.arange(frames)/SR
    hum=.16*np.sin(2*np.pi*50*t)+.06*np.sin(2*np.pi*150*t)+.035*np.sin(2*np.pi*360*t)
    base=air*.45+hum
    right=np.roll(air,31)*.45+hum
    return finish(np.column_stack((base,right)),.075,loop=True)

def make_preview(manifest):
    groups=[]
    sections=[('boss-drop','Boss · Task rơi từ trên cao'),('music','Nhạc nền chung'),('ambience','Không gian')]
    sections += [(color,f"{role['slot']} · {role['name']} — {role['skill']}") for color,role in ROLES.items()]
    sections += [('shared-sfx','Hiệu ứng chung và boss')]
    for category,label in sections:
        cards=[]
        for e in manifest['assets']:
            if category=='boss-drop':
                if e.get('pack')!='boss-drop':continue
            elif category in ROLES:
                if e.get('player')!=category:continue
            elif category=='shared-sfx':
                if e['category']!='sfx' or e.get('player') or e.get('pack')=='boss-drop':continue
            elif e['category']!=category or e.get('player'):continue
            ogg=html.escape('../'+str(Path(e['category'])/(e['key']+'.ogg')))
            mp3=ogg.removesuffix('.ogg')+'.mp3'
            loop=' loop' if e['loop'] else ''
            optional=' · lớp nhịp khẩn cấp bổ sung' if e['key']=='urgency' else ''
            description=html.escape(e.get('description',''))
            cards.append(f'<article><b>{html.escape(e["key"])}</b><small>{description}</small><small>{e["seconds"]:.3f}s · {e["channels"]} kênh · {"loop" if e["loop"] else "one-shot"}{optional}</small><audio controls preload="none"{loop}><source src="{ogg}" type="audio/ogg"><source src="{mp3}" type="audio/mpeg"></audio><small>{html.escape(e["event"])}</small></article>')
        groups.append(f'<section><h2>{label}</h2><div class="grid">{"".join(cards)}</div></section>')
    page='''<!doctype html><html lang="vi"><meta charset="utf-8"><meta name="viewport" content="width=device-width,initial-scale=1"><title>Task Vô Tận — Nghe duyệt âm thanh</title><style>
    *{box-sizing:border-box}body{margin:0;background:#0B1530;color:#F2E3C2;font:16px/1.5 system-ui}main{max-width:1200px;margin:auto;padding:32px 20px}h1{margin-bottom:8px;color:#3FE0FF}p{max-width:850px;color:#c4cfe2}.grid{display:grid;grid-template-columns:repeat(auto-fit,minmax(290px,1fr));gap:14px}article{border:1px solid #2F4A7F;background:#14234A;padding:16px;border-radius:10px}small{display:block;color:#bdcbe3;font-size:12px;margin:6px 0;overflow-wrap:anywhere}audio{width:100%;margin:10px 0}button{border:0;background:#3FE0FF;color:#0B1020;padding:12px 18px;border-radius:6px;cursor:pointer;font-weight:700}h2{margin-top:32px}</style><main><h1>TASK VÔ TẬN · Âm thanh MVP + 5 player</h1><p>Nhạc chiptune và hiệu ứng văn phòng tự tổng hợp. Mỗi player có nhạc chủ đề và bộ âm chiêu riêng trong cùng một mục. Nhạc chủ đề dùng nghe duyệt/giới thiệu nhân vật; trong trận dùng nhạc nền chung. Bấm phát để nghe từng mẫu, các loop tự lặp. Trang này chưa tích hợp gameplay.</p><button id="stop">Dừng tất cả</button>'''+''.join(groups)+'''</main><script>const players=[...document.querySelectorAll('audio')];players.forEach(a=>{a.volume=.5;a.addEventListener('play',()=>players.forEach(b=>{if(b!==a){b.pause();b.currentTime=0}}))});document.getElementById('stop').onclick=()=>players.forEach(a=>{a.pause();a.currentTime=0});</script></html>'''
    destination=ROOT/'assets/audio/preview/index.html'
    destination.parent.mkdir(parents=True,exist_ok=True);destination.write_text(page)

def main():
    parser=argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--only',help='Rebuild one asset key; existing manifest entries remain complete.')
    parser.add_argument('--pack',choices=['player-skills','boss-movement','boss-drop'],help='Build only the selected supplementary audio pack.')
    args=parser.parse_args()
    contract=entries();ffmpeg=imageio_ffmpeg.get_ffmpeg_exe()
    keys={e['key'] for e in contract}
    if args.only and args.only not in keys:parser.error(f'Unknown key: {args.only}')
    for e in contract:
        # Each asset gets an independent stable seed: --only matches a full rebuild.
        global RNG
        RNG=np.random.default_rng(15102026+sum((i+1)*ord(c) for i,c in enumerate(e['key'])))
        if args.only and args.only!=e['key']:continue
        if args.pack and e.get('pack')!=args.pack:continue
        if e['category']=='music':x=make_music(e['key'],e['frames'],e['bars'])
        elif e['category']=='ambience':x=make_ambience(e['frames'])
        else:x=make_sfx(e['key'],e['seconds'])
        source=ROOT/e['source'];runtime=ROOT/e['path'];fallback=ROOT/e['fallback']
        source.parent.mkdir(parents=True,exist_ok=True);runtime.parent.mkdir(parents=True,exist_ok=True)
        sf.write(source,x,SR,format='WAV',subtype='PCM_16')
        # Use isolated ffmpeg encoders: libsndfile's bundled long-file Vorbis
        # writer aborts on this macOS build. PCM masters remain the shared source.
        subprocess.run([ffmpeg,'-v','error','-y','-i',str(source),'-map_metadata','-1','-codec:a','libvorbis','-q:a','5',str(runtime)],check=True)
        subprocess.run([ffmpeg,'-v','error','-y','-i',str(source),'-map_metadata','-1','-codec:a','libmp3lame','-q:a','3',str(fallback)],check=True)
        print(f"{e['path']}: {e['frames']/SR:.3f}s, {e['channels']}ch",flush=True)
    manifest=dict(sampleRate=SR,sourceFormat='WAV PCM16',runtimeFormat='OGG Vorbis',fallbackFormat='MP3',
        bpm=BPM,provenance='Original deterministic synthesis; no external recordings or samples.',
        assets=contract,roles={color:dict(**role,theme=f'player-{color}',sounds=SKILL_BINDINGS[color]) for color,role in ROLES.items()},
        skillPlayback=dict(charge='One shot at charge start; stop on release/cancel. Never loop after the1s cap.',
            shield='At most3 reflection events per cast; coalesce simultaneous reflections, no per-frame playback.',
            mark='One applied sound; expiry is quiet and only plays on authoritative expiry, never stacks.',
            hotfix='Zone loop is owned by zone ID; stop on zone expiry or scene exit. Heal cue once per healed player per cast; coalesce within100ms.',
            dash='Stop start sound on early wall collision/end; hit once per target per cast. Audio never grants invulnerability.',
            cooldown='Ready/unavailable cues are local only; throttle unavailable to once every300ms.',
            music='Player themes are optional introduction/listening tracks. Do not layer five themes or switch BGM on skill use.'),
        dropPlayback=dict(windup='boss-drop-windup once at the start of the1s telegraph; the0.9s sound never triggers damage.',
            impact='boss-drop-impact once per stack impact ID, up to3 desks,0.25s apart; not once per scattered sheet or snapshot.',
            jam='boss-paper-jam once at the start of the1.2s vulnerable jam; no loop; stop on interruption or scene exit.',
            priority='Windup10, impact7, jam5; let the notification tail finish before the first impact.',
            reconnect='Never replay past windup/impact/jam events. Timing, targeting, vulnerability and reflection remain server decisions.'),
        mixing=dict(maxSfxVoices=16,reservedCriticalVoices=4,master=.70,music=.35,sfx=.70,
        musicGainOnWarning=.55,warningDuckSeconds=.5,remotePlayerGain=.45,pingCooldownSeconds=1,
        urgencyThreshold='min(60 seconds, 10% initial deadline)',
        triggers='Combat events play once per unique event ID; UI events are local; loops stop on scene exit.'))
    (ROOT/'assets/audio/manifest.json').write_text(json.dumps(manifest,indent=2,ensure_ascii=False)+'\n')
    make_preview(manifest)
    print(f'Audio ready: {len(contract)} entries, WAV + OGG + MP3, preview/index.html')

if __name__=='__main__':main()
