#!/usr/bin/env python3
"""Generate images headless through Antigravity's `agy` CLI (its image-generator subagent).

    tools/agy_gen.py PROMPT.txt --ref assets/concepts/bosses/boss-khabib-v1.png --out /tmp/khabib

agy saves generated images under ~/.gemini/antigravity-cli/brain/<conversation_id>/ and the parent reply only says
"dispatched", so this script runs agy, waits for the images to appear and copies them to --out.
`--collect ID` skips the agy run and only copies the images of an earlier conversation.
"""
import argparse
import json
import pathlib
import shutil
import subprocess
import sys
import time

BRAIN = pathlib.Path.home() / ".gemini/antigravity-cli/brain"
EXTS = {".png", ".jpg", ".jpeg", ".webp"}


def images(conv: str) -> list[pathlib.Path]:
    # Top level only: .system_generated and .user_uploaded hold agy's own copies.
    return sorted((p for p in (BRAIN / conv).glob("*") if p.suffix.lower() in EXTS), key=lambda p: p.stat().st_mtime)


def subagent_report(conv: str) -> str:
    """Why nothing was generated (quota, refusal...): the image-generator subagent messages its parent."""
    msgs = sorted((BRAIN / conv / ".system_generated/messages").glob("*.json"))
    return "".join(f"\nimage-generator said: {json.loads(m.read_text()).get('content', '')[:600]}" for m in msgs)


def run_agy(prompt: pathlib.Path, refs: list[pathlib.Path], model: str, timeout: int) -> str:
    listed = "\n".join(f"- {r}" for r in refs) or "- (none)"
    task = (
        "Use your image-generator subagent. Give it the brief below verbatim and attach these reference images "
        f"by absolute path:\n{listed}\n\nBrief:\n{prompt.read_text()}\n\nReply with the saved image paths only."
    )
    cmd = ["agy", "-p", task, "--model", model, "--output-format", "json", "--print-timeout", f"{timeout}s"]
    for d in sorted({str(r.parent) for r in refs}):
        cmd += ["--add-dir", d]  # read access to the references, no blanket permission skip
    out = subprocess.run(cmd, capture_output=True, text=True, cwd=prompt.parent).stdout
    line = next((l for l in out.splitlines() if l.startswith("{")), "")
    if not line:
        sys.exit(f"agy printed no JSON:\n{out}")
    data = json.loads(line)
    if data.get("status") != "SUCCESS":
        sys.exit(f"agy failed: {data}")
    return data["conversation_id"]


def wait_for_images(conv: str, settle: int = 20, limit: int = 300) -> list[pathlib.Path]:
    """The subagent can still be writing after the parent returns: wait until the file list stops growing."""
    end, last, stable_since = time.time() + limit, [], time.time()
    while time.time() < end:
        now = images(conv)
        if now != last:
            last, stable_since = now, time.time()
        elif now and time.time() - stable_since >= settle:
            break
        time.sleep(2)
    return last


def main() -> None:
    ap = argparse.ArgumentParser(description=__doc__, formatter_class=argparse.RawDescriptionHelpFormatter)
    ap.add_argument("prompt", type=pathlib.Path, nargs="?", help="text file with the image brief")
    ap.add_argument("--ref", type=pathlib.Path, action="append", default=[], help="reference image (repeatable)")
    ap.add_argument("--out", type=pathlib.Path, required=True, help="directory for the copied images")
    ap.add_argument("--model", default="gemini-3.8-flash-medium")
    ap.add_argument("--timeout", type=int, default=600, help="agy print timeout, seconds")
    ap.add_argument("--collect", metavar="ID", help="copy images of an existing conversation, do not run agy")
    a = ap.parse_args()

    if a.collect:
        conv = a.collect
    elif a.prompt:
        refs = [r.resolve() for r in a.ref]
        missing = [str(r) for r in refs if not r.is_file()]
        if missing:
            sys.exit(f"missing reference images: {missing}")
        conv = run_agy(a.prompt.resolve(), refs, a.model, a.timeout)
    else:
        ap.error("give a prompt file or --collect ID")

    found = images(conv) if a.collect else wait_for_images(conv)  # collect: the run already finished
    if not found:
        sys.exit(f"no images in {BRAIN / conv}{subagent_report(conv)}")
    a.out.mkdir(parents=True, exist_ok=True)
    print(f"conversation {conv}")
    for p in found:
        shutil.copy2(p, a.out / p.name)
        print(a.out / p.name)


if __name__ == "__main__":
    main()
