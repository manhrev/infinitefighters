# Tiến độ các màn

Cập nhật: 2026-10-15. Sửa file này mỗi khi một màn đổi trạng thái. Chi tiết thiết kế nằm trong file của từng màn ở [levels/](levels/).

Ký hiệu: ✅ xong · 🟡 một phần / tạm · ⬜ chưa làm · — không áp dụng.

## Tổng quan

| # | Màn | Thiết kế | Map | Server | Client | Sprite | Âm thanh | Cân bằng |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| 1 | [Đường Đua Sprint](levels/01-sprint.md) | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| 1b | [Anh em cây khế](levels/01b-party.md) (extension màn 1) | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| 2 | [Sảnh Onboarding](levels/02-onboarding.md) | ⬜ | ⬜ | ⬜ | ⬜ | ⬜ | ⬜ | ⬜ |
| 3 | [Tháp Kiến Trúc](levels/03-architecture-tower.md) | ⬜ | ⬜ | ⬜ | ⬜ | ⬜ | ⬜ | ⬜ |
| 4 | [Hầm Tri Thức](levels/04-knowledge-vault.md) | ✅ | ⬜ | ⬜ | ⬜ | ⬜ | ⬜ | ⬜ |
| 5 | [Phòng Điều Hành](levels/05-control-room.md) | ⬜ | ⬜ | ⬜ | ⬜ | ⬜ | ⬜ | ⬜ |

Cột: **Thiết kế** là luật đã chốt trong file màn · **Map** là file trong `infinitefighters-backend/levels/` · **Server** là logic Go · **Client** là hiển thị và điều khiển trong `infinitefighters-frontend/` · **Cân bằng** là chơi thử và chỉnh số liệu.

## Chi tiết từng màn

### 1 — Đường Đua Sprint (Đốc Công Hôm Qua)

Hoàn thành, 1–5 người. Map `levels/sprint.json`; code `game.go`, `skills.go`, `drop.go`.

Để sau (không chặn hoàn thành): sprite chuyển phase của boss (pha 2 hiện chỉ đổi nhịp đòn).

### 1b — Anh em cây khế (extension màn 1)

Hoàn thành. Vào sau khi hạ boss màn 1, hoặc chọn thẳng ở phòng chờ. Map `levels/party.json`; code `party.go`.

Để sau (không chặn hoàn thành): phần “Tiền + gái” (NPC khách mời, tặng quà, “Ai có hẹn”, chưa có luật chốt); hình và âm riêng cho ATM, két, bó tiền, nhạc disco (đang dùng tạm).

### 2 — Sảnh Onboarding (Ngài Định Hướng)

Mới có ý tưởng một đoạn. Cần: thiết kế chi tiết, đoạn platforming, asset biển chỉ dẫn.

### 3 — Tháp Kiến Trúc (Kiến Trúc Sư Ngoài Phạm Vi)

Mới có ý tưởng một đoạn. Cần: thiết kế chi tiết.

### 4 — Hầm Tri Thức (Song Thần Giữ Bài FE & BE)

Đã chốt thiết kế cho bản chơi thử, chưa bắt đầu code. Cần: map, hai boss, trang tài liệu mở cầu, công tắc mở khiên, asset và âm thanh.

### 5 — Phòng Điều Hành (QUY TRÌNH)

Mới có ý tưởng và phác thảo trận cuối, kết thúc. Cần: thiết kế chi tiết.

## Ngoài các màn

- ⬜ Trang tài liệu nhặt được (sprite `items/document.png` đã có, gameplay chưa).
- ⬜ Checkpoint thang máy và đoạn vượt màn có camera riêng.
- ⬜ Chế độ Tay đôi.
- ⬜ Asset: server, chìa khóa server, checkpoint.
