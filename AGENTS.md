# Task Vô Tận (infinitefighters)

2D pixel-art co-op platformer, LAN multiplayer in the browser. Go server (authoritative) + Phaser 3 client.

## Layout

- `infinitefighters-backend/` — Go server; embeds `web/` (built frontend) and `levels/*.json` (Tiled maps).
- `infinitefighters-frontend/` — Phaser 3 + TypeScript + Vite; builds into `../infinitefighters-backend/web`.
- `assets/` — runtime art/audio imported by the frontend. Only `sprites/ tiles/ backgrounds/ ui/` and audio `manifest.json` + `music/ sfx/ ambience/` are in git; sources, concepts, previews stay local.
- `scripts/` — Python asset pipeline (sprites, audio, level generation, validators).
- `docs/` — design, in Vietnamese. `game-scenario.md` overview + audio; `levels/NN-*.md` one file per level; `progress.md` level tracker; `sprite-spec.md`, `tech-stack.md`.

## Commands

- Build all: `cd infinitefighters-backend && make build` (frontend build, then `go build`).
- Run server: `cd infinitefighters-backend && go run .` — flags `-addr :8080 -code -deadline 10m -boss-hp -hard -party`.
- Dev client: `cd infinitefighters-frontend && npm run dev` (proxies `/ws`, `/levels` to :8080). `?debug` logs match state.
- Typecheck: `npm run typecheck`. Go tests: `go test ./...` (needs `web/` built first, `//go:embed web`).
- Scripts: `.venv-sprites/bin/python scripts/<name>.py`; deps in `scripts/requirements-*.txt`. `validate_audio.py` reads the audio tables in `docs/game-scenario.md`.

## Rules

- Keep `docs/` in sync with code: level rules/numbers in `docs/levels/`, status in `docs/progress.md`.
- Code comments in English; docs in Vietnamese.
- Do not commit build output (`web/`, binaries) or asset sources.
