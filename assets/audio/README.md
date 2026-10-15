# TASK VÔ TẬN — Audio MVP

62 âm thanh đã tạo, gồm bộ MVP ban đầu 32 âm, 24 âm bổ sung cho năm player (**5 nhạc chủ đề + 19 hiệu ứng chiêu riêng**), 3 âm cho bước đi/cú lao và 3 âm cho đòn thả hồ sơ của boss. Danh sách tương ứng nằm trong phần âm thanh của `docs/game-scenario.md`; lớp nhạc `urgency` bổ sung nhịp khi gần hết deadline. Mọi âm đều được tổng hợp bằng script; không dùng bản thu, giọng nói, nhạc hoặc sample tải từ bên ngoài.

## Nghe duyệt

Mở `preview/index.html` trong trình duyệt, bấm phát từng mẫu. Nhạc chủ đề và hiệu ứng của mỗi player được gom vào một mục P1–P5, bên cạnh nhạc và hiệu ứng chung. Trang chỉ phát một mẫu cùng lúc, có nút dừng tất cả; các loop tự lặp. Có thể mở file trực tiếp, hoặc từ thư mục gốc dự án chạy:

```sh
python3 -m http.server 8765
```

Sau đó mở `http://localhost:8765/assets/audio/preview/`. Server này chỉ phục vụ nghe duyệt, không phải game multiplayer.

## File bàn giao

| Nhóm | Số âm | Nội dung |
| --- | --- | --- |
| `music/` | 9 | Lobby34,286 giây; sprint51,429 giây; boss68,571 giây; urgency17,143 giây;5 theme player, mỗi bản34,286 giây. Tất cả112 BPM, loop theo số ô nhịp nguyên. |
| `ambience/` | 1 | Văn phòng24 giây: điều hòa, quạt server và tiếng điện nhỏ. Stereo, loop. |
| `sfx/` | 52 | 33 hiệu ứng chung/boss; 17 hiệu ứng theo năm chiêu; 2 cue chung báo chiêu sẵn sàng/chưa dùng được. |

Mỗi âm có ba bản: **WAV PCM16 48 kHz** trong `source/<nhóm>/`, **OGG Vorbis** ở đúng đường dẫn runtime đã mô tả, và **MP3** cùng tên làm dự phòng. Tổng 186 file audio. Nhạc/ambience stereo; SFX mono. `boss-wind-clock` là loop0,75 giây và `mint-hotfix-zone` là loop1 giây; các SFX còn lại phát một lần.

`manifest.json` ghi đường dẫn, bản dự phòng, thời lượng, số kênh, số sample, BPM, ô nhịp, loop, sự kiện, gain gợi ý và độ ưu tiên. `roles` ánh xạ P1–P5 tới theme, các bước chiêu và thông số theo kịch bản; `skillPlayback` mô tả lúc dừng, hủy hoặc gộp âm. Gain trong manifest là mức gợi ý khi trộn; không phải mức phát đã kiểm chứng trong gameplay.

## Bộ âm thanh cho năm player

| Player / chiêu | Nhạc chủ đề | Hiệu ứng |
| --- | --- | --- |
| P1 Teal — Em sửa có một dòng | `player-teal`: arpeggio nhanh và lead sắc, tích lực lên cao | `teal-charge`, `teal-smash`, `teal-smash-hit`, `teal-charge-cancel` |
| P2 Tím — Không thuộc scope em | `player-violet`: bass tròn, giai điệu đáp ngược, echo ngắn | `violet-shield`, `violet-reflect`, `violet-shield-end` |
| P3 Vàng — Theo biên bản đã chốt | `player-amber`: bass chắc và nhịp đóng dấu | `amber-stamp-cast`, `amber-mark-hit`, `amber-mark-expire` |
| P4 Mint — Máy em vẫn chạy | `player-mint`: chuông ấm, hòa âm sáng, thưa nhịp | `mint-hotfix-deploy`, `mint-hotfix-zone`, `mint-hotfix-heal`, `mint-hotfix-end` |
| P5 Cam — Em về đúng giờ | `player-orange`: bass bước nhanh và nốt ngắn gấp gáp | `orange-timecard-dash`, `orange-timecard-hit`, `orange-timecard-stop` |

Tên bảng bỏ đuôi file; theme ở `music/`, hiệu ứng ở `sfx/`. Hai cue thêm là `skill-ready` và `skill-unavailable`, chỉ phát cục bộ. Theme có giai điệu, tiết tấu và hòa âm riêng, không phải năm bản đổi cao độ của cùng một bài.

Theme dùng cho nghe duyệt hoặc giới thiệu nhân vật khi tích hợp sau; không tự thêm màn chọn nhân vật hay bật theme theo slot trong trận. Multiplayer dùng một BGM chung; phát chiêu chỉ kích hoạt SFX, không thay nhạc. Âm chiêu có ưu tiên thấp hơn tín hiệu boss và đồng đội gục.

Charge của P1 là one-shot tối đa1 giây, phải dừng khi thả/hủy; không giữ loop khi đầy lực. Khiên P2 chỉ phát phản đạn khi có phản đạn thật, tối đa3 theo lần dùng. Dấu P3 chỉ có âm áp dấu/hết dấu, không kêu theo mọi đòn tăng sát thương. Hotfix P4 gắn loop với ID vùng, dừng theo hết hạn3 giây; mỗi người chỉ phát heal khi thực sự được hồi và tối đa một lần mỗi lượt dùng, gộp các lượt hồi trong100 ms. P5 dừng âm lao nếu chạm tường sớm; dùng chung hồi chiêu với dash thường. Các thời gian và tác dụng này là quy tắc tích hợp, chưa phải gameplay đã được triển khai.

## Bước đi và cú lao của boss

| File trong `sfx/` | Thời lượng | Sự kiện và chất âm |
| --- | --- | --- |
| `boss-footstep.ogg` | 0,18 giây | `boss.walk.footstep`: bước giày nặng, tiếng kim loại ngắn; gain gợi ý 0,30, ưu tiên thấp |
| `boss-charge-windup.ogg` | 0,90 giây | `boss.charge.windup`: ba nốt chuông mời họp gấp, khác tiếng xòe giấy và bánh răng; gain 0,80, ưu tiên 10 |
| `boss-charge.ogg` | 0,60 giây | `boss.charge.start`: gió lao ngang và đồng hồ rung; gain 0,60, ưu tiên 7 |

Cả ba là mono, one-shot, có WAV/OGG/MP3. Bước chân phát khi chân chạm sàn; chuông bắt đầu cùng báo hành lang và tiếng lao phát một lần khi bắt đầu lao. Không phát lại theo snapshot hoặc từng tile; chiều dài audio không điều khiển thời điểm sát thương. Các file và metadata sẵn sàng để tích hợp, không có thay đổi code game trong đợt bàn giao này.

## Task rơi từ trên cao — bộ âm mới

Trang nghe thử có mục đòn này ở đầu trang. Cả ba file là mono, one-shot, có WAV gốc PCM16 48 kHz và OGG/MP3; tên file dùng trực tiếp làm key trong manifest.

| File trong `sfx/` | Thời lượng | Gain / ưu tiên | Event |
| --- | --- | --- | --- |
| `boss-drop-windup.ogg` | 0,90 giây | 0,80 / 10 | `boss.drop.windup` |
| `boss-drop-impact.ogg` | 0,25 giây | 0,65 / 7 | `boss.drop.impact` |
| `boss-paper-jam.ogg` | 1,20 giây | 0,40 / 5 | `boss.drop.jam` |

Windup là hai nốt thông báo điện tử đi lên và một tiếng tick cuối, khác tiếng xòe giấy của volley, bánh răng của wave và chuông mời họp ba nốt của charge. Impact gồm tiếng hồ sơ dày đập xuống gỗ và giấy rơi văng. Jam là ba nhịp motor chậm dần, khựng lại rồi tiếng giật giấy; không có giọng người.

Phát windup ngay khi bắt đầu báo đòn 1 giây. Phát impact một lần theo ID va chạm của từng chồng hồ sơ, tối đa3 chồng cách nhau0,25 giây; không phát theo từng tờ giấy hay snapshot. Phát paper-jam một lần khi bắt đầu khoảng yếu1,2 giây, dừng nếu trạng thái bị ngắt/rời màn, không phát lại lịch sử khi reconnect. Các quy tắc phát được ghi trong `dropPlayback` của manifest.

Phạm vi đợt này là âm thanh: không thay đổi sprite, backend hoặc frontend. Điều kiện dưới60% HP, hồi đòn12 giây, mục tiêu trên bàn, nhân đôi sát thương khi kẹt giấy và P2 không phản được hồ sơ vẫn do server xử lý. Manifest cung cấp asset và metadata; việc tải tự động chưa đồng nghĩa các hook phát âm đã được kiểm thử trong gameplay.

## Hướng âm thanh

Nhạc là các đoạn chiptune có giai điệu, hòa âm, bass, bộ gõ và biến thể theo đoạn: lobby mang chất nhạc thang máy vui vẻ; sprint nhanh và nảy; boss căng hơn với hòa âm thứ, nhịp đồng hồ và tiếng máy in tổng hợp. Tiếng phím cơ/giấy/kim loại được thiết kế bằng lớp noise có lọc, transient và các thành phần cộng hưởng; đây là âm thanh cách điệu, không phải bản thu đồ vật thật.

Bốn âm báo đòn được làm khác nhau: volley là tiếng xòe giấy cao, tăng dần; wave là hai tiếng bánh răng/kim loại trầm; charge là chuông mời họp ba nốt; drop là thông báo hai nốt đi lên và tiếng tick. Các âm báo bắt đầu ở sự kiện windup; thời điểm gây sát thương vẫn do logic game quyết định.

`urgency.ogg` là lớp nhịp bổ sung, không thay nhạc boss. Khi còn dưới `min(60 giây,10% thời gian ban đầu)`, có thể bật loop này một lần tại ranh giới nhịp tiếp theo và hạ nhạc nền. Mười giây cuối dùng `deadline-tick.ogg` tối đa một lần mỗi giây. Chọn mức âm khi nghe thử để tránh lớp urgency lấn tín hiệu boss.

## Tạo lại và kiểm tra

Chạy từ thư mục gốc; dùng lại môi trường Python của bộ sprite hoặc tạo một môi trường riêng:

```sh
uv pip install --python .venv-sprites/bin/python -r scripts/requirements-audio.txt
.venv-sprites/bin/python scripts/generate_audio.py
.venv-sprites/bin/python scripts/validate_audio.py
```

Tạo lại riêng một âm, ví dụ:

```sh
.venv-sprites/bin/python scripts/generate_audio.py --only boss-wave-windup
```

Chỉ tạo lại24 âm của gói bổ sung:

```sh
.venv-sprites/bin/python scripts/generate_audio.py --pack player-skills
```

Chỉ tạo lại 3 âm bước đi/cú lao boss:

```sh
.venv-sprites/bin/python scripts/generate_audio.py --pack boss-movement
```

Chỉ tạo lại 3 âm thả hồ sơ:

```sh
.venv-sprites/bin/python scripts/generate_audio.py --pack boss-drop
```

Mỗi âm có seed riêng, nên tạo riêng vẫn khớp bản tạo toàn bộ. Script tự xuất lại manifest và trang nghe duyệt. FFmpeg được cung cấp bởi gói `imageio-ffmpeg`, không cần cài binary hệ thống hoặc API key.

## Kiểm tra và giới hạn

`validation.json` chứa kết quả giải mã từng WAV/OGG/MP3, định dạng,48 kHz, số kênh, thời lượng, peak, RMS, DC offset, khoảng lặng đầu, độ chênh sample tại điểm nối loop, sai số nén Vorbis và tính đầy đủ so với danh sách trong kịch bản. Kiểm tra thêm ánh xạ năm role/chiêu, PCM gốc không trùng nhau, và mọi link trên trang nghe duyệt. Âm boss cũng được kiểm tra độ khác biệt phổ tần; phép đo này không thay việc nghe thực tế.

Các file WAV và OGG loop giữ đúng số sample. Điểm nối được làm mượt5 ms mà không thêm khoảng im; MP3 có thể xử lý padding khác nhau giữa các bộ giải mã, nên ưu tiên OGG khi chạy vòng lặp. Cần thử trình duyệt mục tiêu trước khi quyết định dùng MP3 cho loop.

Đợt bổ sung này tạo asset và cập nhật manifest/trang nghe duyệt, không thay code game. Chưa kiểm thử hook phát ba âm thả hồ sơ trong gameplay, phối âm với5 người chơi hoặc nghe trên tai nghe/loa laptop; các bước này không được tính là đã đạt bởi kiểm tra tự động. Các boss/khu vực khác, PvP và extension tiệc cuối game nằm ngoài audio MVP.
