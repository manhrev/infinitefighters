# Sprite Spec — MVP

Status: ready for generation (2026-10-15). Covers only the MVP level: "Đường Đua Sprint", boss "Đốc Công Hôm Qua", 1–5 players. See `game-scenario.md` for the game and `tech-stack.md` for the engine (Phaser 3, `pixelArt: true`).

Style reference: `assets/concepts/multiplayer-arena-sample-02.png`. That image is concept art only. Match its look, not its pixel sizes.

## Global rules

- **Format:** PNG, RGBA.
- **Native size:** every asset is drawn at its final in-game pixel size, 1 art pixel = 1 game pixel. The game scales up by integer factors only. Never deliver upscaled art (for example a 32×32 tile drawn as 1024×1024).
- **Alpha:** characters, tiles, props and items use only fully transparent (alpha 0) or fully opaque (alpha 255) pixels. No anti-aliasing, no soft shadows, no blur. VFX may use partial alpha.
- **Outline:** 1 px dark outline `#0B1020` around characters, items and props.
- **Light:** from upper left.
- **View:** flat orthographic side view. No perspective, no isometric.
- **Animation strips:** one PNG per animation. Frames laid left to right in a single row, all frames the same size, no padding or gaps between frames. Strip width = frame width × frame count.
- **Anchor:** characters stand on the bottom edge of the frame, horizontally centered. The feet must be at the same pixel row in every frame, otherwise the sprite jitters.
- **Facing:** all characters face right. The game mirrors them to face left.
- **Text:** no text baked into sprites. Labels, numbers and dialogue are rendered by the game.

## Palette

| Use | Hex |
| --- | --- |
| Outline | `#0B1020` |
| Navy background, dark / mid | `#0B1530`, `#14234A` |
| Office blue (furniture, walls) | `#22345F`, `#2F4A7F` |
| Cyan screens and glow | `#3FE0FF` |
| Cream platform top | `#F2E3C2`, shadow `#C9B48C` |
| Coral danger (projectiles, warnings) | `#FF5A5A`, dark `#B8303A` |
| Skin (light / mid / shadow) | `#F6C9A0`, `#E0A274`, `#A86B4A` |

Coral red is reserved for danger. Do not use it on players or pickups.

## Players

Five distinct developer designs with keyboard weapons, matching sample 02's style. Updated 2026-10-15: body shape, face and hair distinguish players in addition to hoodie color.

- **Frame:** 64 × 64. Upright heights vary from about 42–54 pixels according to the roster below, centered, feet on the bottom row. Leave room for keyboard and limbs during attack and dash.
- **Hitbox (for reference, not drawn):** 20 × 44, bottom-center.

| File | Frames | FPS | Loop | Notes |
| --- | --- | --- | --- | --- |
| `idle.png` | 4 | 6 | yes | Breathing, keyboard held low |
| `run.png` | 8 | 12 | yes | |
| `jump.png` | 2 | 10 | no | Rising |
| `fall.png` | 2 | 10 | yes | Falling |
| `attack.png` | 5 | 18 | no | Keyboard swing. Also used for the code-pulse skill |
| `dash.png` | 4 | 16 | no | Low forward lunge |
| `hurt.png` | 2 | 10 | no | Recoil |
| `down.png` | 4 | 4 | yes | "Nghỉ phép" state: sitting asleep, small Zz bubble allowed |
| `revive.png` | 4 | 10 | no | Standing back up |

### Five character identities

Generate each character's complete animation set independently. Do not create the roster by palette swapping one body.

| Player | Body | Identifying features | Approx. upright height |
| --- | --- | --- | --- |
| P1 teal | Tall and skinny | Long limbs, messy hair, rectangular glasses | 54 px |
| P2 violet | Short and fat | Round belly and face, short legs, bowl cut, round glasses | 42 px |
| P3 amber | Muscular | Broad shoulders, big arms, square jaw, crew cut, sleeveless hoodie | 50 px |
| P4 mint | Petite and slim | Compact build, small hair bun, large round glasses | 44 px |
| P5 orange | Average height, stocky, pot belly | Receding hair, moustache, zip hoodie, relaxed slouch | 48 px |

All are adult fictional developers. Each identity must remain consistent across all nine animations and its portrait. Body differences are cosmetic: all players retain the same movement, damage, reach and 20 × 44 bottom-centered hitbox. Character colors use the corresponding ramp below; avoid those colors on skin, hair and equipment.

| Player | Shadow | Base | Highlight |
| --- | --- | --- | --- |
| P1 teal | `#0E5E66` | `#1A9AA0` | `#4FD1C5` |
| P2 violet | `#4B2A7A` | `#7B4BC4` | `#B08CF0` |
| P3 amber | `#8A5A00` | `#E0A21B` | `#FFD25E` |
| P4 mint | `#1F6B45` | `#3FBF7F` | `#8EF0B5` |
| P5 orange | `#8A3A10` | `#E06A1F` | `#FFA25E` |

Paths: `assets/sprites/player/<color>/<file>.png`. Preserve frame counts, FPS and runtime keys. Imagegen source sheets and prompts use `assets/source/player-<color>-v2.*`; the original teal sheet is retained.

## Player skill sprites — five office abilities

Additional assets for the five skills in `levels/sprint.md`. Existing nine player animations and portraits stay intact. This delivery supplies art and loader entries; combat logic, cooldowns and healing/reflection rules are implemented separately.

For each color, add `skill-windup.png` and `skill-cast.png` under `assets/sprites/player/<color>/`. Each is a 64 × 64 RGBA strip with 4 frames, facing right, bottom-centered. Windup: 8 FPS, no loop; cast: 16 FPS, no loop. Preserve the v2 identity and body proportions. Hold the final windup pose for a sustained charge or guard; skill state determines duration, never the animation's loop count. P1's ability to walk slowly while charging may use the windup pose as the movement presentation. P5's normal dash animation remains separate.

Standalone VFX live in `assets/sprites/vfx/skills/`; all have 4 frames and a center render origin. Their transparent bounds describe artwork, not attack hitboxes. Player-attached effects should follow the player; world effects stay at cast position. Mirror directional effects with facing. Keep friendly effects cyan/cream, plus mint for healing, with transparent interiors so boss telegraphs remain readable. Never recolor boss task-card sprites to indicate reflection: use the returned-task sprite below.

| Slot / skill | Character action | Primary VFX (frame, FPS, loop) | Secondary VFX (frame, FPS, loop) |
| --- | --- | --- | --- |
| P1 / Em sửa có một dòng | Charge keyboard overhead → heavy downward chop | `keyboard-cleave.png` (80 × 80, 16, no) | `keyboard-charge.png` (64 × 64, 10, yes) |
| P2 / Không thuộc scope em | Brace keyboard → block and push back | `forward-shield.png` (48 × 64, 12, yes) | `returned-task.png` (32 × 24, 16, yes) |
| P3 / Theo biên bản đã chốt | Wind up office stamp → throw | `stamp-projectile.png` (32 × 32, 14, yes) | `boss-stamp-mark.png` (48 × 48, 8, yes) |
| P4 / Máy em vẫn chạy | Crouch with keyboard → deploy patch at feet | `hotfix-burst.png` (128 × 48, 16, no) | `hotfix-zone.png` (128 × 48, 8, yes) |
| P5 / Em về đúng giờ | Lean forward → horizontal lunge | `timecard-trail.png` (96 × 32, 16, yes) | `timecard-impact.png` (48 × 48, 18, no) |

Skill icons are separate 24 × 24 RGBA images in `assets/ui/skills/`: `one-line-fix.png`, `out-of-scope.png`, `meeting-minutes.png`, `works-on-my-machine.png`, `clock-out.png`. Icons show the corresponding primary effect; names, input hints and cooldown overlays are drawn by UI, without baked lettering.

Source sheets and exact built-in imagegen prompts: `assets/source/skill-<color>.png` and `skill-<color>-prompt.txt`. Exact source bounds are recorded in `skill-<color>-regions.json`. Rebuild with `scripts/build_sprites.py skills`. Preview: `assets/preview/skills-contact-sheet.png` and `skill-<color>.png`. Total addition: 10 character strips, 10 VFX strips, 5 icons.

## Boss: Đốc Công Hôm Qua

Angry manager in a dark suit, red tie, holding a giant stopwatch. Comedic, readable silhouette.

- **Frame:** 160 × 160. Body about 120 tall (2.5× player), centered, feet on the bottom row.
- **Hitbox (reference):** 70 × 110, bottom-center.

| File | Frames | FPS | Loop | Notes |
| --- | --- | --- | --- | --- |
| `idle.png` | 4 | 6 | yes | Tapping foot, glaring |
| `volley_windup.png` | 4 | 8 | no | Telegraph: pulls back arm full of task cards. Must be clearly readable |
| `volley.png` | 4 | 12 | no | Throws task cards forward |
| `wave_windup.png` | 4 | 8 | no | Telegraph: lifts stopwatch overhead |
| `wave.png` | 4 | 12 | no | Slams stopwatch on the floor, starts deadline wave |
| `wind_clock.png` | 6 | 8 | yes | Vulnerable window: winding the stopwatch, distracted, sweating |
| `hurt.png` | 2 | 10 | no | Flinch |
| `defeat.png` | 6 | 8 | no | Collapses, stopwatch breaks, papers fly |

Path: `assets/sprites/boss/doc-cong/<file>.png`

### Boss movement extension

Keep the original eight boss strips. Add the following 160 × 160 RGBA strips with the same manager identity, stopwatch, right-facing art, shared scale and bottom-center render origin. The existing 70 × 110 hitbox remains unchanged; animation length does not control attack timing.

| File | Frames | FPS | Loop | Purpose |
| --- | --- | --- | --- | --- |
| `walk.png` | 4 | 8 | yes | Heavy steps while inspecting progress |
| `charge_windup.png` | 4 | 5 | no | Raises stopwatch, leans forward, locks charge direction; hold last pose until the server releases the charge |
| `charge.png` | 4 | 12 | yes | Fast horizontal lunge, stopwatch held forward |
| `charge_recover.png` | 4 | 4 | no | Brakes, catches breath and adjusts stopwatch; hold final pose for remaining recovery |

Additional VFX in `assets/sprites/vfx/`: `boss-foot-dust.png` (64 × 32, 4 frames, 12 FPS, no loop, cream/tan, bottom-center), and `charge-warning.png` (64 × 16, 4 frames, 8 FPS, loop, coral chevrons pointing right, horizontally tileable). Flip the warning for a leftward charge. Its width follows the server's locked `chargeFrom`/`chargeTo` positions plus the 70 px body width, and never tracks a player's later movement. Show it during windup/charge, on the floor without covering platform edges. Walking and recovery have no danger coloring.

Source sheets and exact imagegen prompts: `assets/source/boss-movement*`. Exact crop regions are retained alongside each sheet. Build with `.venv-sprites/bin/python scripts/build_sprites.py boss-motion`; inspect `assets/preview/boss-movement.png`. Integrated in the game on 2026-10-15.

### Falling-file attack asset extension — Task rơi từ trên cao

Asset contract only, based on the supplied request. Do not edit either game code directory. Preserve every existing boss strip. New boss strips face right, have 1 px navy outlines, upper-left light, binary alpha, a shared body scale and feet on the bottom row.

| File under `assets/sprites/boss/doc-cong/` | Frame | Frames | FPS | Loop | Silhouette |
| --- | --- | --- | --- | --- | --- |
| `drop_windup.png` | 160 × 160 | 4 | 8 | no | Stopwatch overhead in one hand, other index finger points straight up; hold final pose for the remaining 1 s telegraph |
| `drop.png` | 160 × 160 | 4 | 12 | no | Snaps the pointing hand down as an order; does not slam the watch |
| `drop_jam.png` | 160 × 160 | 6 | 8 | yes | Papers jammed in watch and stuck to suit, sweat and repeated sheet tug; loop only during the 1.2 s vulnerable state |

| File under `assets/sprites/vfx/` | Frame | Frames | FPS | Loop | Notes |
| --- | --- | --- | --- | --- | --- |
| `file-stack.png` | 64 × 48 | 4 | 12 | yes | Cream folders/binders, coral tag, sheets fluttering above; complete 64 px width |
| `file-burst.png` | 96 × 48 | 5 | 18 | no | Sideways paper scatter and low dust; bottom-center anchor at desk surface |
| `drop-warning.png` | 64 × 32 | 2 | 6 | yes | Two thin coral rails, translucent striped interior, exact top-to-bottom repeat |
| `alert-mark.png` | 16 × 24 | 2 | 6 | yes | Coral warning pictogram `!`, navy outline and cream glint; explicit pictogram exception to the no-baked-text rule, no words or labels |

Integration contract supplied by the user: target up to 3 desks occupied by players, at most one column per desk; column begins below any desk above. Telegraph lasts 1 s; falling stacks arrive 0.25 s apart and break on the target desk. Below 75% HP, replace a volley turn when at least 8 s have elapsed since the previous drop (tuned 2026-10-15; hard mode drops 6 stacks at once from full HP, see `levels/sprint.md`). In phase 2 (below 50% HP) every boss timer runs at 2/3. Paper jam lasts 1.2 s and doubles incoming damage. P2 cannot reflect falling files. These values describe server behavior, not sprite timing. Integrated in the game on 2026-10-15.

Imagegen source sheets and exact prompts are retained under `assets/source/boss-drop-*` and `assets/source/falling-*`; native runtime strips are generated by `scripts/build_boss_drop.py`. Build with `.venv-sprites/bin/python scripts/build_sprites.py boss-drop`. Review `assets/preview/boss-drop-contact-sheet.png`, per-strip previews and `boss-drop-warning-repeat.png`. Audio contract and listening preview are under `assets/audio/`.

## Projectiles and VFX

VFX may use partial alpha and glow.

| File | Frame | Frames | FPS | Loop | Notes |
| --- | --- | --- | --- | --- | --- |
| `task-card.png` | 16 × 12 | 4 | 12 | yes | Coral red card labeled with a white blank stripe (no text), spinning |
| `deadline-wave.png` | 32 × 32 | 4 | 12 | yes | Coral flame-like floor wave. Tiles horizontally; game repeats it along the floor |
| `floor-warning.png` | 32 × 8 | 2 | 6 | yes | Coral striped warning strip on the floor before an attack. Tiles horizontally |
| `swing-arc.png` | 48 × 48 | 4 | 18 | no | Cyan crescent of the keyboard swing |
| `code-pulse.png` | 32 × 16 | 4 | 16 | yes | Short-range cyan `</>`-style energy bolt (shape only, no readable text) |
| `land-dust.png` | 32 × 16 | 4 | 16 | no | |
| `hit-spark.png` | 24 × 24 | 3 | 18 | no | White/cyan impact star |

Path: `assets/sprites/vfx/<file>.png`

## Items

| File | Size | Frames | Notes |
| --- | --- | --- | --- |
| `coffee.png` | 16 × 16 | 4 (steam loop, 6 FPS) | Heals |
| `document.png` | 16 × 16 | 1 | Collectible |

Path: `assets/sprites/items/<file>.png`

## Tiles (for Tiled)

One tileset image, `assets/tiles/office.png`, 32 × 32 tiles, no spacing or margin, 8 tiles per row.

Required tiles (row 1–2 is enough):

1. Floor top-left edge, floor top-middle, floor top-right edge (cream top, dark base)
2. Floor fill (below the top row)
3. One-way desk platform: left end, middle, right end (cream top clearly flat and horizontal; thin, about 12 px tall, at the top of the tile)
4. Platform support leg (thin, non-solid decoration)
5. Server block (solid), top and bottom halves
6. Wall fill (dark navy)

Platform top edges must be perfectly horizontal and must line up across adjacent tiles.

## Props (decoration, not solid)

Separate PNGs in `assets/sprites/props/`, any size that is a multiple of 16:

`desk.png`, `chair.png`, `monitor.png` (cyan screen), `plant.png`, `filing-cabinet.png`, `water-cooler.png`, `ceiling-light.png`, `poster.png` (blank, no text).

Checkpoint elevator: `elevator.png`, frame 64 × 96, 4 frames (closed → open), 8 FPS, cyan outline glow.

## Background parallax

Three layers, each **1920 × 1080**, seamless when tiled horizontally (left and right edges must match). Lower contrast than the playable floor and platforms.

| File | Content |
| --- | --- |
| `assets/backgrounds/bg-0-sky.png` | Night sky, moon, far skyline with lit windows |
| `assets/backgrounds/bg-1-windows.png` | Tall glass window frames and pillars, transparent glass areas |
| `assets/backgrounds/bg-2-office.png` | Faint far-office furniture silhouettes, transparent above |

## UI

| File | Size | Notes |
| --- | --- | --- |
| `assets/ui/heart.png` | 9 × 8, 2 frames | Full, empty |
| `assets/ui/portrait.png` | 24 × 24 | Teal player head; other identities use `portrait-<color>.png`, each cropped from its own idle pose |
| `assets/ui/deadline-icon.png` | 16 × 16 | Stopwatch icon for the DEADLINE bar |

Bars, labels (P1–P5), pings and text are drawn by the game.

## Acceptance checks

Each delivered PNG must pass:

1. Size matches this spec exactly (strip width = frame width × frame count).
2. Characters, tiles, props and items: every pixel alpha is 0 or 255.
3. Each player uses its own clothing palette and has a distinct silhouette; every animation and portrait preserves that player's identity.
4. Feet on the same bottom row in every frame of a strip.
5. No text, no watermark, no background behind characters or items.
6. Background layers tile seamlessly left-to-right.

## Not in MVP

Other bosses and zones, Tay đôi mode, server key item, final-boss assets, boss phase-change animation.
