# Đóng góp

Mọi đóng góp được nhận theo [MIT](LICENSE), áp dụng cho cả code, tài liệu và asset.

## Quy trình

1. Tạo branch từ `main`: `feat/<việc>`, `fix/<việc>`, `docs/<việc>`, `asset/<việc>`.
2. Commit theo Conventional Commits: `feat: ...`, `fix: ...`, `docs: ...`, `chore: ...`. Mỗi commit một việc.
3. Mở pull request vào `main`, mô tả thay đổi và cách đã kiểm tra. Đổi luật chơi hoặc cốt truyện cần chủ repo duyệt.

## Trước khi mở PR

Chạy những lệnh liên quan tới phần đã sửa:

| Sửa | Lệnh |
| --- | --- |
| Frontend | `cd infinitefighters-frontend && npm run typecheck` |
| Backend | `cd infinitefighters-backend && make build && go test ./...` |
| Sprite | `.venv-sprites/bin/python scripts/validate_sprites.py` |
| Âm thanh hoặc bảng âm trong kịch bản | `.venv-sprites/bin/python scripts/validate_audio.py` |

Không commit build output (`infinitefighters-backend/web/`, binary).

## Tài liệu

- Tài liệu viết bằng tiếng Việt; comment trong code viết bằng tiếng Anh.
- Sửa luật hoặc số liệu của một màn thì sửa luôn file màn trong `docs/levels/` và trạng thái trong `docs/progress.md` ở cùng PR.

## Thêm màn mới

- Cốt truyện mở: màn mới nối tiếp sau màn cuối hiện có, đặt tên `docs/levels/NN-<slug>.md` với NN là số kế tiếp. Không đánh số lại màn cũ.
- Mở đầu file màn bằng dòng `Màn NN · <khu vực>`, trạng thái, link về [game-scenario.md](docs/game-scenario.md), và đoạn **Cốt truyện** nối từ màn trước.
- Thêm dòng vào bảng khu vực trong `docs/game-scenario.md` và vào `docs/progress.md`; thêm link cốt truyện từ màn trước sang màn mới.
- Khuyến khích lối chơi hoặc góc nhìn mới (giải đố, lén lút, nhìn từ trên xuống, chạy đua, quản lý...). Giữ chất châm biếm công sở, nhân vật P1–P5 và chơi được 1–5 người qua LAN. Góc nhìn khác platformer thì ghi quy ước hình ảnh trong file màn và trong `docs/sprite-spec.md`.
- Endgame (`docs/levels/endgame.md`) đang tạm gác; chỉ thiết kế khi chủ repo yêu cầu.

## Asset

- Theo rule trong [docs/sprite-spec.md](docs/sprite-spec.md) và quy trình trong [docs/asset-pipeline.md](docs/asset-pipeline.md).
- Commit cả nguồn (ảnh nguồn kèm prompt, mask, WAV gốc) lẫn file runtime do script tạo ra.
- Chỉ dùng asset tự tạo hoặc có quyền phân phối theo MIT; ghi nguồn và giấy phép nếu dùng tài nguyên bên ngoài.
- Nhân vật, công ty và tình tiết đều hư cấu; không ám chỉ người hoặc đội nhóm có thật.
