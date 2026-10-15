#!/usr/bin/env python3
"""Validate every MVP WAV master, Vorbis runtime file and MP3 fallback.

Measures signal integrity, not artistic quality or real-device/browser playback.
"""
import hashlib
import json
import re
import sys
from html.parser import HTMLParser

import numpy as np
import soundfile as sf
from audio_manifest import ROOT, SAMPLE_RATE as SR, ROLES, SKILL_BINDINGS, entries

def main():
    failures=[];checks=0;metrics=[]
    def check(ok,message):
        nonlocal checks
        checks+=1
        if not ok:failures.append(message)
    contract=entries()
    by_key={e['key']:e for e in contract}
    for key,(duration,priority) in {'boss-drop-windup':(.9,10),'boss-drop-impact':(.25,7),'boss-paper-jam':(1.2,5)}.items():
        e=by_key.get(key,{})
        check(e.get('seconds')==duration and e.get('priority')==priority,f'{key}: drop audio timing/priority mismatch')
        check(e.get('channels')==1 and e.get('loop') is False and e.get('category')=='sfx',f'{key}: expected mono one-shot SFX')
    # Verify actual document coverage independently of the exported manifest.
    doc=(ROOT/'dev/game-scenario.md').read_text()
    section=doc.split('## Âm thanh',1)[1].split('## Hai chế độ',1)[0]
    # Asset-only extensions may be documented in the audio delivery contract
    # before the gameplay scenario is edited or integrated.
    section+='\n'+(ROOT/'assets/audio/README.md').read_text()
    names=set(re.findall(r'`(?:assets/audio/(?:music|ambience)/)?([a-z0-9-]+\.ogg)`',section))
    expected={e['key']+'.ogg' for e in contract}
    check(names.issubset(expected),f'Uncovered documented files: {names-expected}')
    required={e['key']+'.ogg' for e in contract if not e['optional']}
    check(required.issubset(names),f'Required audio missing from scenario: {required-names}')
    manifest_path=ROOT/'assets/audio/manifest.json'
    check(manifest_path.is_file(),'Missing audio manifest')
    if manifest_path.is_file():
        manifest=json.loads(manifest_path.read_text())
        check(manifest.get('assets')==contract,'Manifest differs from audio contract')
        for color,role in ROLES.items():
            binding=manifest.get('roles',{}).get(color,{})
            check(binding.get('skill')==role['skill'],f'{color}: skill missing from manifest')
            check(binding.get('sounds')==SKILL_BINDINGS[color],f'{color}: invalid sound bindings')
            check(binding.get('theme')==f'player-{color}',f'{color}: missing player theme')
            check(role['skill'] in doc,f'{color}: role no longer matches scenario')
            for key in SKILL_BINDINGS[color].values():
                check(key+'.ogg' in expected,f'{color}: binding references missing asset{key}')
    hashes=[];master_hashes=[]
    for e in contract:
        decoded={}
        for field,ext in [('source','wav'),('path','ogg'),('fallback','mp3')]:
            path=ROOT/e[field];tag=e[field]
            check(path.is_file(),f'{tag}: missing')
            if not path.is_file():continue
            try:
                info=sf.info(path);x,sr=sf.read(path,always_2d=True,dtype='float32')
            except Exception as exc:
                check(False,f'{tag}: failed to decode: {exc}');continue
            check(sr==SR,f'{tag}: expected48000Hz, got{sr}')
            check(info.channels==e['channels'],f'{tag}: wrong channels')
            check(np.isfinite(x).all(),f'{tag}: invalid samples')
            delta=abs(len(x)-e['frames'])
            check(delta<=(SR*.04 if ext=='mp3' else 1),f'{tag}: incorrect decoded duration ({len(x)} frames)')
            peak=float(np.max(np.abs(x)));rms=float(np.sqrt(np.mean(x*x)))
            check(.001<peak<.97,f'{tag}: silent or clipping (peak{peak:.5f})')
            check(rms>.0001,f'{tag}: near-silent asset')
            check(float(np.max(np.abs(np.mean(x,axis=0))))<.015,f'{tag}: excessive DC offset')
            if ext=='wav':
                check(info.format=='WAV' and info.subtype=='PCM_16',f'{tag}: expected PCM16 WAV')
                master_hashes.append(hashlib.sha256(x.tobytes()).hexdigest())
            if ext=='ogg':
                check(info.format=='OGG' and info.subtype=='VORBIS',f'{tag}: expected OGG Vorbis')
                hashes.append(hashlib.sha256(path.read_bytes()).hexdigest())
            if e['loop']:
                jump=float(np.max(np.abs(x[0]-x[-1])))
                check(jump<(.035 if ext!='mp3' else .05),f'{tag}: loop boundary jump{jump:.5f}')
                # A loop may have intentional rhythmic rests, but not an encoder's
                # appended blank tail. WAV/OGG must retain exact sample lengths.
                if ext!='mp3':check(len(x)==e['frames'],f'{tag}: changed loop length')
            else:
                active=np.flatnonzero(np.max(np.abs(x),axis=1)>max(.0005,peak*.01))
                onset=(active[0]/SR) if len(active) else 999
                check(onset<.012,f'{tag}: leading silence{onset:.4f}s')
            decoded[ext]=x
            metrics.append(dict(path=tag,seconds=len(x)/SR,channels=info.channels,format=info.format,
                subtype=info.subtype,peak=round(peak,6),peakDb=round(20*np.log10(peak),2),
                rms=round(rms,6),loop=e['loop'],boundaryJump=round(float(np.max(np.abs(x[0]-x[-1]))),6)))
        if 'wav' in decoded and 'ogg' in decoded:
            wav=decoded['wav'];ogg=decoded['ogg'];n=min(len(wav),len(ogg))
            err=float(np.sqrt(np.mean((wav[:n]-ogg[:n])**2)))
            check(err<.07,f"{e['key']}: excessive Vorbis encode error{err:.5f}")
    check(len(hashes)==len(set(hashes)),'Duplicate runtime files / placeholder copies')
    check(len(master_hashes)==len(set(master_hashes)),'Duplicate PCM masters / reused audio')
    for path in ['assets/audio/preview/index.html','assets/audio/README.md']:
        check((ROOT/path).is_file(),f'Missing{path}')
    preview=ROOT/'assets/audio/preview/index.html'
    if preview.is_file():
        class PreviewSources(HTMLParser):
            def __init__(self):super().__init__();self.sources=[];self.players=0
            def handle_starttag(self,tag,attrs):
                if tag=='source':self.sources.append(dict(attrs).get('src',''))
                if tag=='audio':self.players+=1
        parser=PreviewSources();parser.feed(preview.read_text())
        check(parser.players==len(contract),'Preview does not expose every audio asset')
        for url in parser.sources:
            check((preview.parent/url).resolve().is_file(),f'Preview broken source: {url}')
    # Keep the important warning cues physically different: low two-clank wave
    # warning vs rising paper fan. Compare spectral centroids, not just file names.
    def centroid(key):
        p=ROOT/f'assets/audio/source/sfx/{key}.wav'
        if not p.exists():return None
        x,_=sf.read(p);power=np.abs(np.fft.rfft(x))**2;f=np.fft.rfftfreq(len(x),1/SR)
        return float((power*f).sum()/max(power.sum(),1e-12))
    volley=centroid('boss-volley-windup');wave=centroid('boss-wave-windup')
    if volley is not None and wave is not None:
        check(volley>wave*1.5,f'Boss warning cues insufficiently spectrally distinct ({volley:.1f}/{wave:.1f}Hz)')
    def warning_spectrum(key):
        p=ROOT/f'assets/audio/source/sfx/{key}.wav'
        if not p.is_file():return None
        x,_=sf.read(p);power=np.abs(np.fft.rfft(x))**2;freq=np.fft.rfftfreq(len(x),1/SR)
        edges=np.geomspace(80,12000,49)
        bands=np.array([power[(freq>=lo)&(freq<hi)].sum() for lo,hi in zip(edges[:-1],edges[1:])])
        return bands/max(np.linalg.norm(bands),1e-12)
    drop=warning_spectrum('boss-drop-windup');drop_differences={}
    if drop is not None:
        for other in ['boss-volley-windup','boss-wave-windup','boss-charge-windup']:
            reference=warning_spectrum(other)
            check(reference is not None,f'{other}: missing comparison warning')
            if reference is not None:
                distance=float(np.linalg.norm(drop-reference));drop_differences[other]=distance
                check(distance>.15,f'Drop warning spectrum too similar to {other}: {distance:.4f}')
    result=dict(assets=len(contract),filesDecoded=len(metrics),checks=checks,failures=failures,
        bossCueCentroidsHz=dict(volley=volley,wave=wave),dropWarningSpectralDistances=drop_differences,measurements=metrics,
        limits='Automated file and signal checks only. Listening on headphones/laptop, in-game mixing and browser playback still require audition/integration.')
    report=ROOT/'assets/audio/validation.json';report.parent.mkdir(parents=True,exist_ok=True)
    report.write_text(json.dumps(result,indent=2)+'\n')
    summary=f'Validated {len(contract)} audio assets; {len(metrics)} files decoded; {checks} checks; {len(failures)} failures.'
    print(summary)
    for message in failures:print('FAIL:',message)
    return bool(failures)

if __name__=='__main__':sys.exit(main())
