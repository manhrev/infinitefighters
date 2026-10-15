# Sandbox Yasuka

Màn thử boss Yasuka (sếp lớn nhân sự, "Đang chờ phản hồi"), không nằm trong chuỗi màn chính. Concept: `assets/concepts/boss-yasuka-v1.png`. Chọn "Sandbox Yasuka" ở phòng chờ (chủ phòng). Bấm Esc (chủ phòng) để về phòng chờ.

## Luật

Giống [sandbox Khabib](sandbox-khabib.md): dùng map màn 1, không DEADLINE, không thua. Người gục tự đứng dậy sau 2 giây. Yasuka hạ xuống 0 máu thì đầy máu lại. Hitbox 52×100 theo `docs/sprite-spec.md`. Yasuka đi bộ tới người gần nhất, rồi dùng chiêu kế tiếp trong vòng lặp. Mọi con số là bản thử, chưa cân bằng.

## Bốn chiêu (lặp theo thứ tự)

| Chiêu | Trạng thái | Cách hoạt động | Cách né / phá |
| --- | --- | --- | --- |
| Né yêu cầu | `dodge_windup`, `dodge` | Chắp tay sau lưng chờ 2,5 giây. Đòn đầu tiên trúng chị (đánh gần, xung code, chiêu riêng) bị lách đi, không gây sát thương. Chị trượt ra xa 200 px, rồi hở sườn 1 giây (×2 sát thương). Không ai đánh thì chị bỏ thế. | Đánh một nhát nhử, cả đội dồn đòn lúc chị hở sườn. |
| Đúng quy trình | `procedure_windup`, `procedure` | Đóng dấu từ chối: một dấu dưới chân mỗi người, và một hàng ba dấu trước mặt chị trên sàn. Dấu ngủ 0,8 giây (nhấp nháy), rồi thức 4 giây. Ai chạm dấu thức bị đẩy về chỗ xuất phát. | Rời chỗ trước khi dấu thức. Lướt (dash) đi qua được. |
| Im lặng là vàng | `silence_windup`, `silence`, `flurry_windup`, `flurry` | Quay lưng dựng mái vòm "…" 5 giây. Mái vòm hứng mọi đòn (độ bền 8, hiện bằng độ mờ). Vỡ thì hồ sơ tồn đọng bung ra: báo 0,5 giây, rồi giấy bay dọc mọi bục, cả hai phía (sàn an toàn). Sau đó `recover` 1 giây. Không ai phá nổi thì mái vòm tự hạ, chị `recover`. | Phá mái vòm rồi nhảy hoặc lướt tránh giấy, hoặc đứng ở sàn. |
| Gửi lại biểu mẫu | `forms_windup`, `forms` | Báo 0,75 giây, rồi ném cho **mỗi người chơi một tờ giấy** tự đuổi theo người đó (cùng tỏa ra theo hình quạt). Giấy bay chậm hơn tốc độ chạy (220 so với 260 px/s), quay đầu dần, sống 5 giây. Chạm bất kỳ người nào thì người đó mất 1 tim (đồng đội đứng chắn đỡ được). Mục tiêu gục thì giấy đổi sang người gần nhất. | Chạy khỏi, đánh (J) hoặc bắn xung code (K) để hạ giấy, lướt (dash) xuyên qua. |

## Chưa làm

Concept còn ba chiêu chưa làm: Gắn kết hình thức (chia P1–P5 lên các bục), Chị đã ghi nhận (bom hẹn giờ), Khảo sát ẩn danh. Tờ giấy đuổi theo là chiêu thêm theo yêu cầu, không có trong concept. Mê cung hồ sơ và công tắc giữ cùng lúc của "Đúng quy trình" được đơn giản thành dấu mộc đẩy về chỗ xuất phát.

## Trạng thái code

- Server: `infinitefighters-backend/yasuka.go`, test `yasuka_test.go`. Vào bằng `{"type":"start","level":"sandbox-yasuka"}`.
- Client: sprite `sprites/boss/yasuka/` và VFX `yasuka-silence-dome`, `yasuka-dodge-trail`, `yasuka-stamp-trap`, `yasuka-paper-flurry`. Giấy bay dọc bục (chiêu mái vòm vỡ) là chữ "HO SO", chưa có sprite riêng. Tờ giấy đuổi theo (Gửi lại biểu mẫu) là texture vẽ lúc khởi động trong `game.ts` (`makePaperTexture`), chưa có sprite trong pipeline. Âm thanh mượn của boss màn 1.
