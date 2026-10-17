# Prompt Gemini: sprite boss v2

Bốn prompt dùng để Gemini vẽ lại sprite di chuyển của boss theo bộ số frame và fps trong `../sprite-v2.md`. Giữ nguyên giọng văn, bảng màu và quy tắc của các prompt cũ trong `assets/source/boss-*-prompt.txt`; chỉ phần layout và quy tắc chuyển động là mới.

| File | Boss | Ảnh tham chiếu | Chuyển động thứ ba |
| --- | --- | --- | --- |
| `boss-khabib-v2-prompt.txt` | Khabib | `assets/concepts/bosses/boss-khabib-v1.png` | `press` (PRESSING RUSH) |
| `boss-vought-v2-prompt.txt` | SA Vought | `assets/concepts/bosses/boss-sa-vought-v1.png` | `deflect` (DEFLECT SWAT) |
| `boss-yasuka-v2-prompt.txt` | Yasuka | `assets/concepts/bosses/boss-yasuka-v1.png` | `dodge` (EVASIVE DODGE SLIDE) |
| `boss-doc-cong-v2-prompt.txt` | Đốc Công | `assets/source/boss-doc-cong.png` (chưa có concept riêng) | `charge` (CHARGE RUN) |

Mansa chỉ có concept, chưa có sprite hay chuyển động nào trong sandbox, nên chưa có prompt.

## Cách dùng

1. Mở Gemini, đính kèm ảnh tham chiếu ở bảng trên, rồi dán nội dung prompt.
2. Yêu cầu ảnh ngang tỉ lệ 3:2 (ví dụ 2400×1600) để mỗi ô là hình vuông 400×400.
3. Kết quả có nền magenta `#FF00FF`, giống các sheet cũ, để tách nền bằng hàm `remove_magenta_bg` như trong `scripts/build_boss_khabib.py`.

## Layout atlas v2

Lưới 6 cột × 4 hàng, 19 pose, 5 ô cuối để trống.

| Ô | Nội dung | Frame | fps trong game |
| --- | --- | --- | --- |
| 1–6 | `idle` | 6 | 6 |
| 7–14 | `walk` | 8 | 15 (hoặc 7.5 nếu boss chậm 120 px/s) |
| 15–19 | chuyển động thứ ba | 5 | 30 |
| 20–24 | trống | | |

Nên kiểm tra sau khi sinh ảnh:
- Chân đứng đúng một hàng pixel ở mọi ô.
- `idle` pose 6 nối mượt về pose 1, `walk` pose 8 nối mượt về pose 1.
- Tỉ lệ thân, đầu, đạo cụ giống nhau ở mọi ô, nhân vật không trôi ngang trong ô.
- Không có ô nào trùng nhau và không có hiệu ứng vẽ lên nhân vật.

Gemini thường sai số lượng pose hoặc làm lệch lưới. Nếu vậy, sinh lại hoặc tách `idle`, `walk`, chuyển động thứ ba thành ba lần sinh riêng.

## Chưa làm

Chưa có script cắt sheet v2 thành strip và `regions.json` cho layout 6×4 này. Các sheet boss cũ dùng lưới 6×6 hoặc 4×4 (xem `docs/asset-pipeline.md`).
