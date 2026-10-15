# Sandbox Khabib

Màn thử boss Khabib, không nằm trong chuỗi màn chính. Concept: `assets/concepts/boss-khabib-v1.png`. Chọn "Sandbox Khabib" ở phòng chờ (chủ phòng). Bấm Esc (chủ phòng) để về phòng chờ.

## Luật

- Dùng lại map màn 1 (`levels/sprint.json`). Không có DEADLINE, không thua.
- Người chơi gục thì tự đứng dậy sau 2 giây, đầy máu.
- Khabib hạ xuống 0 máu thì đầy máu lại ngay.
- Máu boss tính như màn 1 (theo số người). Pha 2 (dưới 50% máu) nhanh hơn như màn 1.

## Các chiêu (lặp theo thứ tự)

| Chiêu | Trạng thái | Cách hoạt động | Cách né |
| --- | --- | --- | --- |
| Pressing | `press_windup`, `press` | Báo 0,75 giây, rồi tiến về người gần nhất 2,5 giây, vừa đi vừa tung sóng áp lực dọc sàn. Thân chạm người thì mất 1 tim (một lần mỗi chiêu). | Nhảy qua sóng, lướt (dash) qua thân. |
| Talk Spam | `spam_windup`, `spam` | Bắn lời giục "XONG CHUA?", "BAO CAO!", "UPDATE!" mỗi 0,25 giây trong 2,5 giây, mỗi lời ở một làn ngẫu nhiên (sàn hoặc mặt bàn). | Đứng khác làn, nhảy, lướt. P2 phản được. |
| Teleport | `teleport_out`, `teleport_warn`, `teleport_in` | Mờ dần 0,5 giây, biến mất 0,6 giây (không đánh trúng được) và đánh dấu điểm hạ cánh màu coral ngay sau lưng người gần nhất. Người đứng trên bàn thì Khabib hạ cánh lên bàn (người đang nhảy thì lấy mặt đất bên dưới), và ở lại đó. Khabib đi bộ trên bàn như người chơi: Pressing có thể đi quá mép bàn và rơi xuống (bàn hoặc sàn bên dưới), sóng áp lực luôn chạy dọc sàn (chấn động mặt đất) kể cả khi boss đứng trên bàn, và thêm một sóng dọc mặt bàn, dừng ở mép bàn. Hạ cánh đè lên ai đứng yên thì mất 1 tim. | Rời điểm đánh dấu trước khi hạ cánh. |
| Teleport liên hoàn | `chain_windup`, rồi `teleport_out`/`warn`/`in` nhiều lần | Báo 0,75 giây, rồi nhảy liên tục vài lần, mỗi lần nhanh hơn Teleport thường (dissolve 0,2 giây, đánh dấu 0,4 giây). Từ 2 người trở lên: mỗi người bị một cú, theo thứ tự ngẫu nhiên, hạ cánh sau lưng người đó (đúng tầng của họ). Một người: 2 đến 4 cú, mỗi cú rơi vào một chỗ ngẫu nhiên (sàn hoặc mặt bàn bất kỳ), nhìn về phía người chơi. | Rời điểm đánh dấu ngay mỗi lần, đừng đứng yên. |

Vòng lặp: Pressing, Talk Spam, Teleport, Pressing, Talk Spam, Teleport liên hoàn.

Sau Pressing, Teleport và cả chuỗi Teleport liên hoàn là `recover` 1 giây: sát thương nhận ×2.

## Trạng thái code

- Server: `infinitefighters-backend/khabib.go`, test `khabib_test.go`. Chế độ vào bằng `{"type":"start","level":"sandbox"}`, thoát bằng `{"type":"end"}`.
- Client: dùng sprite riêng `sprites/boss/khabib/` khi đã có (trạng thái thiếu anim mượn anim gần nhất cùng bộ); chưa có thì mượn tạm sprite Đốc Công, nhuộm xanh tím.
- Hitbox 80×110 theo `docs/sprite-spec.md`. VFX riêng đã nối vào client: `khabib-press-wave` (sóng áp lực), `khabib-teleport-warning` (điểm hạ cánh), `khabib-portal` (khi biến mất và hạ cánh). Lời giục của Talk Spam vẫn vẽ bằng chữ, chưa dùng `khabib-nag-projectile`.
