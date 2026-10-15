# MVP runtime sprite delivery

## Falling-file attack — Task rơi từ trên cao

Seven additional runtime strips: `drop_windup` (4 × 160 × 160, 8 FPS, no loop), `drop` (4 × 160 × 160, 12 FPS, no loop), `drop_jam` (6 × 160 × 160, 8 FPS, loop); `file-stack` (4 × 64 × 48, 12 FPS, loop), `file-burst` (5 × 96 × 48, 18 FPS, no loop), `drop-warning` (2 × 64 × 32, 6 FPS, loop) and `alert-mark` (2 × 16 × 24, 6 FPS, loop). Every new entry is in `assets/sprites/manifest.json`.

Open `boss-drop.html` for animated previews of all seven strips and audio playback; the page loops strips for review while runtime loop flags stay as specified in the manifest. `boss-drop-overview.png` shows the attack's key silhouettes. `boss-drop-contact-sheet.png` and `boss-drop-<strip>.png` show every frame. `boss-drop-warning-repeat.png` repeats the column vertically over a desk to inspect seams and transparency. The warning has two thin coral rails and a faint striped interior; the explicit `!` pictogram is the only baked symbol, without words or labels.

All art was generated with built-in imagegen. Source sheets and exact prompts are retained in `assets/source/boss-drop-*` and `falling-*`. Original long-arm windup/release sources remain as `*-before-compact.png`; imagegen's compact-arm edits make room for the overhead watch without shrinking the body. `boss-drop-calibration.json` records each sheet's inspected body scale. The processing script keeps one scale per animation and fits the 64 px-wide stack's solid body separately from its fluttering sheets so the physical stack stays stable.

Rebuild: `.venv-sprites/bin/python scripts/build_sprites.py boss-drop`, then regenerate the full contact sheet and run sprite validation. Matching audio and its listening preview are in `assets/audio/`. This is an asset-only delivery: no files under either game code directory are edited by these asset scripts.

## Boss movement and charge

Six additional runtime PNGs: `walk`, `charge_windup`, `charge`, `charge_recover` under `assets/sprites/boss/doc-cong/`, plus `boss-foot-dust` and horizontally repeating `charge-warning` under `assets/sprites/vfx/`. The original boss strips are preserved. Source sheets and exact built-in imagegen prompts are retained in `assets/source/boss-movement*`; rebuild with `.venv-sprites/bin/python scripts/build_sprites.py boss-motion`.

`boss-movement.png` previews every new frame. These are asset deliverables only. The frontend's existing animation loader can use the new state names; wiring the locked charge corridor, mirrored chevrons and foot dust into gameplay is separate work.

## Five office skill sprite sets

Additional 25 runtime PNGs support the five abilities in `docs/levels/01-sprint.md`: 10 character animation strips (`skill-windup` and `skill-cast` for each player), 10 standalone VFX strips, and 5 skill icons. The full contract, names, sizes, FPS and loop flags are in `docs/sprite-spec.md` and `assets/sprites/manifest.json`.

Preview the five abilities and icons in `skills-overview.png`, every new frame in `skills-contact-sheet.png`, or inspect `skill-teal.png`, `skill-violet.png`, `skill-amber.png`, `skill-mint.png`, `skill-orange.png` separately. Source sheets and exact prompts are retained as `assets/source/skill-<color>*`; these were generated with the built-in imagegen tool. Teal's cast row receives a separate imagegen cleanup to remove the baked slash; the original source is retained as `skill-teal-before-cleanup.png`.

Rebuild: `.venv-sprites/bin/python scripts/build_sprites.py skills`, followed by `.venv-sprites/bin/python scripts/make_sprite_preview.py` and `.venv-sprites/bin/python scripts/validate_sprites.py`. Source alpha projections locate the actual row and character bounds before downsampling, rather than assuming imagegen produced mathematically equal cells. Runtime sprites have binary alpha and exact palettes. Effects use one scale per strip to retain growth between frames.

The existing manifest-based frontend loader picks up all new sprite and animation keys. Gameplay activation, hit timings, cooldowns, reflection, marks and healing are separate from this asset delivery; sprite dimensions do not define attack ranges or effect lifetimes. All existing player movement, attack sprites and portraits remain unchanged.

## Distinct player roster v2

The five players now have independently generated body designs: teal tall/slim, violet short/round, amber muscular, mint petite, orange stocky with a pot belly. All nine animation paths, frame counts, FPS and 64 × 64 cells are unchanged. Gameplay hitboxes and stats remain shared.

- `player-roster-v2.png`: five idle silhouettes at 4×, on a shared baseline.
- `player-roster-v2-all-frames.png`: every player animation frame at 4×.
- Original runtime players and portraits are retained in `assets/archive/player-roster-v1/`.
- Five built-in imagegen source sheets and exact prompts are in `assets/source/player-<color>-v2*`.
- `scripts/build_player_roster.py` locates connected silhouettes and records exact regions and animation selections in `*-v2-layout.json`. Source sheets have 7–8 poses in some requested 8-cell rows: a tucked airborne pose supplies a fall transition, a forward swing supplies a missing dash entry, and an airborne pose fills a missing run step. These selections retain the runtime frame counts without duplicating frames within a strip.
- Rebuild with `.venv-sprites/bin/python scripts/build_sprites.py player`, then `.venv-sprites/bin/python scripts/make_roster_preview.py` and `.venv-sprites/bin/python scripts/make_sprite_preview.py`.
- The full rebuild detects the v2 sources; legacy palette cloning is skipped. Hoodie masks are now per identity, and each portrait comes from its own processed idle pose. The validator checks each identity's mask instead of requiring identical anatomy across colors.
- `runtime-scene.png` remains the earlier v1 composition; use the v2 roster previews to inspect the current player art.

All 120 runtime PNGs listed in `assets/sprites/manifest.json` are RGBA at native game-pixel sizes. Load strips as Phaser spritesheets using the manifest's `frameWidth`, `frameHeight`, frame count and FPS. Use integer scaling, nearest-neighbor filtering and `pixelArt: true`. Character render origin is bottom-center; hitboxes are metadata, not painted pixels.

## Outputs

- 45 player strips: nine animations in each of teal, violet, amber, mint and orange.
- Eight boss strips for Đốc Công Hôm Qua.
- Seven VFX strips, coffee strip and document pickup.
- `assets/tiles/office.png`: 256 × 64, eight 32 × 32 tiles per row, two rows. Tile indices are in `assets/tiles/office.json`; indices11–15 are empty.
- Eight office props and a four-frame checkpoint elevator.
- Three horizontally repeating background layers, each1920 × 1080. Sky is opaque; glass and far-office layers have transparent regions.
- Heart strip (full/empty), stopwatch icon and five identity-specific portraits.
- `contact-sheet.png`: every runtime asset and every animation frame at exactly4× nearest-neighbor, including full-size background layers; filename and frame labels are outside artwork. Large image:8000 ×18632.
- `runtime-scene.png`: native-size composition for scale review, with an enlarged UI swatch section underneath. This is a preview only, not runtime art.

## Rebuild

Run from the repository root. Dependencies are pinned in `scripts/requirements-sprites.txt`.

```sh
uv venv .venv-sprites
uv pip install --python .venv-sprites/bin/python -r scripts/requirements-sprites.txt
.venv-sprites/bin/python scripts/build_sprites.py all
.venv-sprites/bin/python scripts/make_sprite_preview.py
.venv-sprites/bin/python scripts/validate_sprites.py
```

Individual processing stages: `player`, `boss`, `effects`, `tiles`, `props`, `backgrounds`, `ui`. `all` executes these in order and writes the Phaser loading manifest. With v2 sources present, player processing also produces portraits and palette swapping is skipped. Source PNGs and exact imagegen prompts are retained in `assets/source/`; rebuilding requires no API call.

`assets/sprite-masks/` contains source-color-derived hoodie masks before variant swapping. Validation checks palette membership, clothing mask isolation, unchanged non-clothing pixels, per-frame sneaker baseline, visible bottom alignment, distinct frames, exact dimensions, binary alpha, danger color restrictions, platform joins, VFX joins and both horizontal edge columns of backgrounds. Semantic visual checks are recorded in `review.json` with hashes; that record becomes stale if runtime artwork changes. PNG pixel checks cannot prove absence of text or artistic identity, so those properties were inspected visually.

## Spec conflicts and explicit choices

- The boss description requests a red tie, while the global palette reserves coral red for danger. The tie is cream; only dangerous cards/waves/warnings use coral.
- White stripes, highlights and dust use the supplied cream palette entries. No unlisted pure white was added.
- Plant foliage is office blue because the non-player palette does not contain green. Teal colors remain exclusive to player hoodies.
- Hearts are cyan, not coral, to respect the danger-only color rule.
- The down animation has no Zz lettering; the stricter no-text constraint wins. Eyes close and the character sits with the same keyboard.
- Code pulse is an abstract angular bolt; no readable code glyphs are baked in.
- Prop dimensions are implementation choices within the allowed multiples of16, recorded explicitly in the manifest. The elevator keeps its exact64 ×96 frame size.
- Portrait variant paths were unspecified: teal stays at `assets/ui/portrait.png`; others use `portrait-violet.png`, `portrait-amber.png`, `portrait-mint.png`, `portrait-orange.png`.
- Bottom anchoring applies even to airborne, seated and collapsed poses. The validator checks the bottom visible silhouette in all character frames and sneaker presence in the bottom four rows of player frames; it cannot semantically identify a boss shoe in a collapsed pose.
- Backgrounds are composed at native resolution from downsampled generated motifs. Pillars extend by repeating native segments, not by enlarging source pixels.

No requested MVP assets were skipped. The v2 roster updates the player section of `docs/sprite-spec.md`. Other bosses, duel mode, server keys and final-boss artwork are outside the sprite spec's MVP scope and were not produced.

Generation used the built-in imagegen tool. All original, v2 roster and skill prompts are stored next to their corresponding source PNGs. The processing scripts document source regions, frame selection and deterministic structural cleanup.
