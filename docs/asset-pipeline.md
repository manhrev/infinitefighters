# Asset pipeline

Cách tạo lại, kiểm tra và thêm asset. Rule hình ảnh: [sprite-spec.md](sprite-spec.md). Rule âm thanh: mục **Âm thanh** trong [game-scenario.md](game-scenario.md).

Mọi thứ trong `assets/` đều nằm trong git, kể cả nguồn, nên clone về là tạo lại được. Không cần API key: nguồn ảnh và prompt đã lưu sẵn, âm thanh được tổng hợp bằng script.

## Thư mục

| Thư mục | Vai trò | Ai ghi |
| --- | --- | --- |
| `assets/sprites/`, `tiles/`, `backgrounds/`, `ui/` | Ảnh runtime, frontend import qua `sprites/manifest.json` | `build_sprites.py` |
| `assets/audio/{music,sfx,ambience}/`, `audio/manifest.json` | Âm runtime OGG + MP3 dự phòng | `generate_audio.py` |
| `assets/source/` | Ảnh nguồn từ imagegen, kèm prompt (`*-prompt.txt`) và vùng cắt (`*-regions.json`, `*-layout.json`) | Thủ công |
| `assets/sprite-masks/` | Mask màu áo dùng khi đổi màu nhân vật và khi kiểm tra | `build_sprites.py` |
| `assets/audio/source/` | Bản gốc WAV 48 kHz 16-bit | `generate_audio.py` |
| `assets/preview/`, `assets/audio/preview/` | Ảnh xem duyệt, trang nghe duyệt, kết quả kiểm tra | Script preview/validate |
| `assets/concepts/` | Concept duyệt phong cách, không phải asset runtime | Thủ công |
| `assets/archive/` | Bản cũ giữ lại để tham khảo | Thủ công |

Ghi chú chi tiết: [assets/preview/README.md](../assets/preview/README.md) (sprite, các lựa chọn khi spec mâu thuẫn), [assets/audio/README.md](../assets/audio/README.md) (âm thanh).

## Môi trường

Chạy từ thư mục gốc repo.

```sh
uv venv .venv-sprites
uv pip install --python .venv-sprites/bin/python -r scripts/requirements-sprites.txt -r scripts/requirements-audio.txt
```

## Script

| Script | Việc | Chạy |
| --- | --- | --- |
| `build_sprites.py` | Cắt ảnh nguồn thành sprite runtime; `all` chạy mọi stage rồi ghi `assets/sprites/manifest.json` | `build_sprites.py all` hoặc một stage: `player skills boss boss-motion boss-drop effects tiles props backgrounds ui` |
| `build_player_roster.py`, `build_skill_sprites.py`, `build_boss_movement.py`, `build_boss_drop.py` | Stage con của `build_sprites.py` | Gọi qua `build_sprites.py <stage>` |
| `palette_swap.py` | Đổi màu áo nhân vật, giữ nguyên hình | Tự chạy trong `build_sprites.py all` |
| `sprite_manifest.py` | Hợp đồng sprite (tên, kích thước, FPS, loop) chép từ `sprite-spec.md` | Thư viện, không chạy riêng |
| `make_sprite_preview.py`, `make_roster_preview.py` | Ảnh xem duyệt phóng 4× vào `assets/preview/` | Chạy sau khi build |
| `validate_sprites.py` | Kiểm tra sprite theo hợp đồng, mask, mép tile | Chạy sau khi build; exit khác 0 là lỗi |
| `generate_audio.py` | Tổng hợp nhạc và hiệu ứng, xuất WAV/OGG/MP3, manifest và trang nghe duyệt | `generate_audio.py`, `--only <key>` hoặc `--pack <tên>` |
| `audio_manifest.py` | Hợp đồng âm thanh khớp bảng âm trong `game-scenario.md` | Thư viện, không chạy riêng |
| `validate_audio.py` | Giải mã và đo mọi file âm, đối chiếu với bảng âm trong `game-scenario.md` | Chạy sau khi tạo âm |
| `gen_levels.py` | Sinh map Tiled `infinitefighters-backend/levels/*.json` | `python3 scripts/gen_levels.py` |

Lệnh đầy đủ có dạng `.venv-sprites/bin/python scripts/<script>.py ...`.

## Thứ tự

```sh
.venv-sprites/bin/python scripts/build_sprites.py all
.venv-sprites/bin/python scripts/make_sprite_preview.py
.venv-sprites/bin/python scripts/validate_sprites.py
.venv-sprites/bin/python scripts/generate_audio.py
.venv-sprites/bin/python scripts/validate_audio.py
```

Sau đó build game (`cd infinitefighters-backend && make build`) để frontend nhận asset mới.

## Thêm asset cho màn mới

1. Ghi asset cần có vào file màn ở `docs/levels/`. Màn có góc nhìn khác platformer ghi luôn quy ước hình ảnh ở đó.
2. Thêm rule kích thước, palette, hoạt ảnh vào `sprite-spec.md`; thêm âm vào bảng âm trong `game-scenario.md`.
3. Lưu ảnh nguồn và prompt vào `assets/source/`, viết stage cắt ảnh, thêm entry vào `sprite_manifest.py` hoặc `audio_manifest.py`.
4. Chạy lại thứ tự ở trên; hai validator phải sạch lỗi.
5. Commit cả nguồn lẫn runtime.
