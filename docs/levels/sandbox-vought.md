# Sandbox SA Vought

Màn thử boss phụ SA Vought (Solution Architect Sang Tên Bug), không nằm trong chuỗi màn chính. Concept: `assets/concepts/boss-sa-vought-v1.png`. Chọn "Sandbox SA Vought" ở phòng chờ (chủ phòng). Bấm Esc (chủ phòng) để về phòng chờ.

## Luật

Giống [sandbox Khabib](sandbox-khabib.md): dùng map màn 1, không DEADLINE, không thua. Người gục tự đứng dậy sau 2 giây. Vought hạ xuống 0 máu thì đầy máu lại. Hitbox 36×95 theo `docs/sprite-spec.md`.

Vought đi bộ tới người gần nhất (như boss màn 1), rồi dùng chiêu kế tiếp trong vòng lặp. Mọi con số dưới đây là bản thử, chưa cân bằng.

## Bốn chiêu (lặp theo thứ tự)

| Chiêu | Trạng thái | Cách hoạt động | Cách né / phá |
| --- | --- | --- | --- |
| Múa mồm chuyển việc | `deflect_windup`, `deflect` | Báo 0,75 giây, rồi tung 5 task ("DEV LO NHE"…) bay vòng cung, mỗi 1/6 giây, về phía người chơi. Một người: ba điểm quanh người đó. Nhiều người: lần lượt từng người. Chạm là mất 1 tim. | Đứng chỗ khác, nhảy, lướt. P2 phản được. |
| Vỗ vai tâm sự | `pat_out`, `pat_warn`, `shoulder_pat` | Mờ dần, biến mất (không đánh trúng được), đánh dấu điểm hạ cánh sau lưng người gần nhất, rồi giữ người đó đứng im 1,5 giây. Giữ đủ thì người đó bị "gắn task lên lưng": 5 giây không lướt và không dùng chiêu riêng. | Rời điểm đánh dấu hoặc lướt. Đồng đội đánh trúng Vought thì ngắt cuộc tâm sự (Vought lảo đảo, `recover`). Khi chơi một mình: đập J hoặc L 6 lần để thoát. |
| Sang tên bug | `stamp_windup`, `stamp_throw` | Báo 0,75 giây, rồi đóng con dấu DEV lên người gần nhất. Sau 3,5 giây không phá được thì người đó mất 1 tim và bị gắn task (như trên). | Đánh con dấu (J, K, hoặc đòn bất kỳ trúng nó) 3 lần, người nào đánh cũng được. Số giây và số lần còn lại hiện trên đầu. |
| Phản đòn yêu cầu | `reflect_windup`, `reflect_shield` | Hít vào 0,6 giây, rồi mở khiên lời nói 2 giây. Đòn đánh vào khiên không gây sát thương và bắn trả một task đỏ ("EM LAM!") vào người gần nhất. Xung code bị khiên bắn ngược lại theo đúng làn bay. | Đừng đánh khi khiên mở; chờ khiên đóng. Sau khiên là `recover` 1 giây, ×2 sát thương. |

Sau Vỗ vai và Phản đòn là `recover` 1 giây: sát thương nhận ×2.

## Trạng thái code

- Server: `infinitefighters-backend/vought.go`, test `vought_test.go`. Vào bằng `{"type":"start","level":"sandbox-vought"}`.
- Client: dùng `sprites/boss/vought/` khi đã có, kèm `vought-speech-shield` và `vought-bug-stamp`. Chưa dùng `vought-deflect-ribbon`, `vought-pat-mark`.
- Chưa cân bằng và chưa có âm thanh riêng (đang mượn âm của boss màn 1).
