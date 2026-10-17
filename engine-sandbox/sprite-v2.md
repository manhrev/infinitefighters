# Sprite v2 cho engine-sandbox

Đề xuất số frame và fps để chuyển động mượt trong sandbox. Số liệu tính từ các hằng số trong `world.go` (`runSpeed` 260, `sprintSpeed` 400, `dashTicks` 10, `jumpSpeed` 880, `gravity` 2400, `tickRate` 60), chưa kiểm tra bằng mắt. Spec chính thức của game vẫn là `docs/sprite-spec.md`; tài liệu này chỉ áp dụng cho bộ sprite v2 của sandbox.

## Nguyên tắc chọn

1. **Hết trượt chân.** Quãng đường một chu kỳ chạy là `tốc độ × số frame / fps`. Quãng này nên bằng khoảng 1.2–2.2 lần chiều cao nhân vật. Ngoài khoảng đó, chân trượt trên sàn dù có nhiều frame.
2. **fps chia hết 60.** Giá trị an toàn: 6, 7.5, 10, 12, 15, 20, 30. Màn hình vẽ 60 lần mỗi giây, nên fps chia hết thì mọi frame hiện đúng nhau. Các giá trị như 8, 16, 18, 24 làm một số frame hiện lâu hơn số khác và gây khựng nhẹ.
3. **Chân luôn ở cùng một hàng pixel** ở mọi frame, nhân vật căn giữa đáy frame, quay mặt sang phải (quy tắc chung của `docs/sprite-spec.md`).

## Player (frame 64×64, hitbox 20×44)

| Chuyển động | Frame | fps | Chu kỳ | Ghi chú |
| --- | --- | --- | --- | --- |
| `idle` | 6 | 6 | 1 giây | thở nhẹ |
| `run` (260 px/s) | 8 | 20 | 0.4 giây, 104 px | bước dài khoảng 1.1 lần chiều cao |
| `sprint` (400 px/s) | 8 | 30 | 0.27 giây, 107 px | strip riêng, cùng quãng bước với `run` |
| `dash` (10 tick = 0.167 giây) | 5 | 30 | 0.167 giây | mỗi frame đúng 2 tick |
| nhảy lên | 2 | 30 | không lặp | 0.067 giây |
| `land` | 3 | 20 | không lặp | 0.15 giây |

## Boss (frame 160×160, hitbox cao 110)

| Chuyển động | Frame | fps | Chu kỳ | Ghi chú |
| --- | --- | --- | --- | --- |
| `idle` | 6 | 6 | 1 giây | |
| `walk` (260 px/s) | 8 | 15 | 0.53 giây, 139 px | khoảng 1.26 lần chiều cao |
| `walk` nếu boss chậm 120 px/s | 8 | 7.5 | 1.07 giây, 128 px | hợp dáng nặng hơn |
| dash-like (`press`, `charge`, `deflect`, `dodge`) | 5 | 30 | 0.167 giây | theo `dashTicks` |

Sandbox hiện cho mọi nhân vật chạy cùng tốc độ 260 px/s. Nếu chọn boss chậm hơn, cần thêm trường tốc độ vào bảng `chars` trong `world.go`.

## Nhảy và rơi

Chọn pose theo vận tốc dọc `vy`, không theo thời gian. Vẽ 5 pose: lên nhanh, lên chậm, đỉnh, rơi chậm, rơi nhanh, và ánh xạ `vy` từ -880 đến +900 (`maxFall`) sang chúng. Cách này luôn khớp với vật lý dù nhảy cao hay thấp. Hiện client chỉ có hai trạng thái `jump` và `fall`; cần sửa `web/index.html` để dùng `vy`.

## Khác với spec chính

Spec hiện tại dùng `run` 8 frame ở 12 fps và `walk` của boss 4 frame ở 8 fps. Bộ v2 đổi cả hai. Nếu đưa v2 vào game chính, cần cập nhật `assets/sprites/manifest.json` và `docs/sprite-spec.md`.
