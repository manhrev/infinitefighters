# TASK VÔ TẬN

Trạng thái: phần MVP đã triển khai; số liệu thực tế và các điểm lệch so với thiết kế nằm ở mục **Trạng thái triển khai** trong file của từng màn (`levels/`). Các phần còn lại là định hướng.
Cập nhật: 2026-10-15.

## Ý tưởng

Game 2D platformer màn hình ngang, châm biếm công sở: chạy, nhảy qua bục, né đạn task và đánh boss bằng bàn phím. Hai chế độ: tay đôi và cả đội đánh boss. Chơi trên trình duyệt qua mạng LAN theo định hướng hiện có của dự án.

Nhân vật và công ty đều hư cấu. Sự hài hước đến từ cách làm việc: thiếu điều phối, né trách nhiệm, dí deadline, giữ kiến thức và đổ thừa.

## Phạm vi bản đầu (MVP)

Đã chốt 2026-10-15. Các phần khác trong tài liệu là định hướng dài hạn.

- **Một màn co-op:** Đường Đua Sprint, boss Đốc Công Hôm Qua. Chưa làm chế độ Tay đôi, các khu vực khác và trận cuối.
- **Số người:** 1–5 người cùng phòng.
- **Thời gian:** mỗi màn có thanh DEADLINE đếm ngược (mặc định 10 phút, chỉnh được). Hết giờ là thua.
- **Thua và chơi lại:** cả đội gục hết hoặc hết giờ thì chơi lại màn từ đầu.
- **Boss:** máu và số hướng tấn công tính một lần lúc bắt đầu trận theo số người trong phòng, không đổi giữa trận.
- **Rớt mạng:** giữ chỗ cho người chơi. Nhân vật đứng yên, không nhận sát thương và không tính vào điều kiện "cả đội gục". Người chơi mở lại trang là vào lại đúng nhân vật. Nếu mọi người cùng rớt, game tạm dừng.
- **Lobby:** server có một phòng duy nhất. Khi khởi động, server in ra địa chỉ và một mã 4 ký tự. Người chơi mở địa chỉ, nhập tên và mã, rồi vào phòng chờ. Người vào đầu tiên là chủ phòng và bấm Bắt đầu.

### Trạng thái triển khai

Số liệu đang chạy, điều khiển và điểm lệch so với thiết kế: [01-sprint.md](levels/01-sprint.md) (co-op) và [06-party.md](levels/06-party.md) (loạn đấu).

**Chưa làm:** trang tài liệu nhặt được, checkpoint thang máy, chế độ Tay đôi, các khu vực và boss khác, trận cuối, phần NPC của extension.

**Cờ server:** `-code` (mã phòng cố định), `-deadline 10m`, `-boss-hp 120` (máu boss khi có 1 người, để chơi thử), `-hard` (mode khó), `-party` (vào thẳng tiệc), `-addr :8080`. Mở trang với `?debug` để đọc trạng thái trận trong console khi test.

## Bối cảnh

Mọi tên gọi chỉ để gợi nhớ cho vui, không ám chỉ đội nhóm hay người thật.

Chi nhánh Việt Nam vừa thành lập của tập đoàn Trách Nhiệm Hữu Hạn Trách Nhiệm có hai cánh văn phòng: NA — Not Assigned và Vacancy Platform.

Đội Vacancy vừa biến mất sau một đợt tái cấu trúc; bàn ghế vẫn còn, ticket tiếp tục tự assign. Đội NA được chuyển từ V-Soul sang, mang theo hệ thống cũ và những lời hứa chưa được bàn giao.

Bạn là dev vừa nhận việc. HR đưa laptop, thẻ ra vào và lời chào: “Ở đây ai cũng có ownership. Quyền quyết định đang chờ cấp.”

Một đêm, hệ thống quản trị biến thành thực thể QUY TRÌNH, khóa cửa văn phòng: “Muốn tan làm, vui lòng hoàn thành các đầu việc còn tồn đọng.” Mỗi task xong lại sinh hai task mới. Cả đội phải vượt qua năm khu vực để tìm người có quyền bấm Approve.

## Cách chơi cốt lõi

- Di chuyển trái/phải, nhảy, nhảy xuyên bục một chiều từ dưới lên, chủ động rơi xuống bục dưới, đánh thường và lướt né có hồi chiêu.
- Vũ khí nhận diện: bàn phím cận chiến; kỹ năng phụ là phát xung code tầm ngắn.
- Boss báo trước đòn bằng hoạt ảnh và vùng nguy hiểm, để người chơi nhảy, chạy hoặc đổi bục né tránh.
- Mỗi khu vực có đoạn platforming ngắn, checkpoint thang máy và đấu trường boss.
- Vòng lặp: vượt chướng ngại → nhặt cà phê/tài liệu → đến checkpoint → học nhịp đánh boss → mở khu vực tiếp theo.
- Ưu tiên cảm giác di chuyển và chiến đấu; tình tiết công sở xuất hiện qua đòn đánh, vật thể và câu thoại ngắn.

## Âm thanh

Bộ âm thanh được lưu riêng trong `assets/audio/`, gồm nhạc nền, hiệu ứng chung và phần bổ sung cho năm chiêu riêng; cần tích hợp và nghe thử phối âm trong gameplay. Phạm vi MVP vẫn là một màn co-op 1–5 người, boss Đốc Công Hôm Qua. Các phần âm thanh cho khu vực khác và PvP thuộc bản mở rộng.

### Phong cách

Nhạc chiptune 16-bit pha tiếng nhạc cụ gõ làm từ đồ văn phòng: phím cơ thành hi-hat, máy in thành nhịp bass, tiếng quẹt thẻ và chuông thang máy thành điểm nhấn. Giai điệu có vẻ vui vẻ như video giới thiệu công ty, nhưng nhịp điệu dần gấp và hơi tréo ngoe khi deadline đến gần.

Hiệu ứng ngắn, sắc và dễ phân biệt. Bàn phím đánh trúng nghe “cạch–bốp”, task bay nghe giấy quất gió, boss lên dây đồng hồ nghe bánh cóc xoay khô khốc. Chất hài nằm ở âm sắc và thời điểm; tránh phát một câu thoại dài sau mỗi lần đánh. MVP không cần lồng tiếng; câu châm biếm hiển thị bằng chữ trong game.

### Nhạc nền và không gian — MVP

| File dự kiến | Khi phát | Hướng âm thanh |
| --- | --- | --- |
| `assets/audio/music/lobby.ogg` | Phòng chờ | Loop 30–45 giây, chiptune nhẹ pha nhạc thang máy, một tiếng thông báo lệch nhịp làm điểm hài. |
| `assets/audio/music/sprint.ogg` | Đoạn vượt màn | Loop 45–60 giây, khoảng 112 BPM, bass nảy và nhịp phím cơ; chừa khoảng âm cho tiếng nhảy, đòn đánh. |
| `assets/audio/music/boss.ogg` | Bắt đầu đấu boss | Loop 60–90 giây, cùng nhịp với sprint để chuyển mượt; motif đồng hồ, bass dồn và tiếng máy in có tiết tấu. |
| `assets/audio/ambience/office.ogg` | Trong màn chơi | Loop 20–30 giây rất nhỏ: điều hòa, server và tiếng điện đèn; không có tiếng gõ phím ngẫu nhiên dễ nhầm với đòn đánh. |

Khi thời gian còn dưới 10% ban đầu, tối đa 60 giây, thêm nhịp đồng hồ gấp và hạ nhẹ nhạc nền. Chỉ kích hoạt một lần khi vượt ngưỡng; không phát liên tục mỗi lần nhận trạng thái trận đấu. Mười giây cuối có tiếng tick rõ hơn mỗi giây và đồng hồ trên UI cùng nháy. Không tăng tốc nhạc theo FPS hoặc theo tốc độ máy người chơi.

Khi thắng, dừng nhạc boss và phát một câu nhạc chiến thắng ngắn kết bằng tiếng thang máy mở. Khi thua, nhạc hụt hơi, máy in kẹt giấy rồi dừng. Đợi người chơi bấm chơi lại mới chuyển nhạc về trạng thái đầu màn; không chồng nhạc cũ với nhạc mới.

### Hiệu ứng nhân vật, vật phẩm và UI — MVP

Các tên dưới đây là file dự kiến trong `assets/audio/sfx/`; mỗi sự kiện chỉ phát một lần tại thời điểm xảy ra.

| File | Sự kiện | Chất âm / độ dài mục tiêu |
| --- | --- | --- |
| `footstep-1.ogg`, `footstep-2.ogg` | Chân chạm sàn khi chạy | Hai tiếng giày trên sàn kim loại luân phiên, 50–100 ms; không phát khi đang ở trên không. |
| `jump.ogg`, `land.ogg` | Bật nhảy / tiếp đất | Một nốt bật lên ngắn / tiếng giày chạm sàn trầm, 100–200 ms. |
| `dash.ogg` | Bắt đầu lướt | Gió quét và tiếng kéo ghế rất ngắn, 120–200 ms. |
| `keyboard-swing.ogg` | Bàn phím vung qua điểm ra đòn | Gió quét kèm phím lạch cạch, 100–180 ms; có thể nghe cả khi đánh hụt. |
| `keyboard-hit.ogg` | Đòn được xác nhận trúng mục tiêu | “Cạch–bốp” rõ, 80–150 ms; tách khỏi swing để người chơi biết đã đánh trúng. |
| `code-pulse.ogg` | Phát xung code | Tiếng synth điện tử ngắn, khác âm task bay, 120–250 ms. |
| `player-hurt.ogg`, `player-down.ogg` | Nhận sát thương / gục | Nốt chát ngắn / tiếng hệ thống tắt lịm, 150–500 ms. Không phát hurt cho đòn bị chặn bởi bất tử. |
| `revive.ogg` | Cứu đồng đội thành công | Ba nốt khởi động lại đi lên, 400–700 ms. |
| `coffee.ogg`, `document.ogg` | Nhặt cà phê / tài liệu | Tiếng uống nhỏ kèm nốt hồi sức / giấy lật kèm tiếng lưu, 150–350 ms. |
| `checkpoint.ogg` | Kích hoạt checkpoint | Chuông thang máy hai nốt, 400–700 ms; không lặp khi đứng trong vùng kích hoạt. |
| `ui-confirm.ogg`, `ui-back.ogg` | Xác nhận / quay lại | Tiếng quẹt thẻ đi lên / đi xuống, 60–120 ms; chỉ phát trên máy thao tác. |
| `ping.ogg` | Đồng đội gửi ping | Tiếng thông báo mềm, 100–180 ms; giới hạn tần suất để tránh spam. |
| `deadline-tick.ogg` | Nhịp đếm ngược khẩn cấp | Tick ngắn, 40–80 ms; tối đa một tiếng mỗi giây, có nháy UI đi kèm. |
| `victory.ogg`, `defeat.ogg` | Kết thúc trận | Câu nhạc thắng / thua dài 2–4 giây, mỗi trận chỉ phát một kết quả. |

### Nhạc chủ đề cho năm player

Năm loop riêng dùng để nghe duyệt hoặc giới thiệu nhân vật khi có màn giới thiệu phù hợp, không tự thêm cơ chế chọn nhân vật vào lobby. Mỗi bản dài 16 ô nhịp, khoảng 34,286 giây, stereo, 112 BPM. Trong trận multiplayer vẫn dùng một nhạc nền chung; dùng chiêu không đổi nhạc và không chồng năm theme.

| File | Player và chất nhạc |
| --- | --- |
| `assets/audio/music/player-teal.ogg` | P1 — Người gánh task: arpeggio phím cơ gọn và sáng, giai điệu đi lên như đang tích lực cho một dòng sửa code. |
| `assets/audio/music/player-violet.ogg` | P2 — Chuyên viên chuyển tiếp: bass tròn, nhịp lệch nhẹ và giai điệu đáp lại như thư bị chuyển ngược. |
| `assets/audio/music/player-amber.ogg` | P3 — Người chốt biên bản: bass chắc, nhịp đóng dấu rõ, lead vuông và dứt khoát. |
| `assets/audio/music/player-mint.ogg` | P4 — Người cứu production: chuông ấm, hợp âm sáng, nhịp thưa hơn để gợi hồi phục. |
| `assets/audio/music/player-orange.ogg` | P5 — Người đòi tan làm: bass bước nhanh, nốt ngắn và chuyển nửa cung, cảm giác đang chạy kịp giờ quẹt thẻ. |

### Hiệu ứng cho năm chiêu riêng — MVP

Các file dưới đây nằm trong `assets/audio/sfx/`. Âm gắn với sự kiện chiêu, không phát lại theo frame animation, snapshot hay từng tick sát thương. Không lồng tiếng câu châm biếm.

| Player | File và mốc phát |
| --- | --- |
| P1 — Em sửa có một dòng | `teal-charge.ogg`: bắt đầu giữ, tiếng điện và phím tăng dần tối đa 1 giây, không loop. Thả sớm dừng charge và phát `teal-smash.ogg`; có thể giảm gain theo lực tích. Trúng mục tiêu phát thêm `teal-smash-hit.ogg`. Lướt/hủy tích lực dừng charge và phát `teal-charge-cancel.ogg`, không phát smash. |
| P2 — Không thuộc scope em | `violet-shield.ogg`: dựng khiên. `violet-reflect.ogg`: khi task thực sự bị bật lại, tối đa 3 lần theo giới hạn chiêu; gộp các phản đạn xảy ra cùng lúc để tránh chồng tiếng. `violet-shield-end.ogg`: khi khiên hết hiệu lực 0,6 giây hoặc bị kết thúc sớm. Sóng deadline không tạo tiếng phản đạn. |
| P3 — Theo biên bản đã chốt | `amber-stamp-cast.ogg`: phóng dấu. `amber-mark-hit.ogg`: dấu trúng và được áp lên mục tiêu. `amber-mark-expire.ogg`: dấu hết hạn sau 4 giây, phát nhỏ. Không tạo loop hoặc tiếng nhắc theo mỗi đòn được tăng sát thương; dấu không cộng dồn. |
| P4 — Máy em vẫn chạy | `mint-hotfix-deploy.ogg`: đặt vùng và phát xung đầu. `mint-hotfix-zone.ogg`: loop nền 1 giây rất nhỏ trong khi vùng tồn tại, dừng khi vùng hết hạn 3 giây hoặc rời màn. `mint-hotfix-heal.ogg`: khi một người thực sự nhận hồi máu, tối đa một lần mỗi người mỗi lần dùng; gộp các lượt hồi trong 100 ms. `mint-hotfix-end.ogg`: vùng kết thúc. Không dùng âm revive vì chiêu không cứu người đã gục. |
| P5 — Em về đúng giờ | `orange-timecard-dash.ogg`: tiếng quẹt thẻ và gió khi bắt đầu lao. `orange-timecard-hit.ogg`: mỗi mục tiêu thực sự bị đánh trúng một lần trong lượt lao. `orange-timecard-stop.ogg`: kết thúc hoặc gặp tường; dừng âm lao nếu va tường sớm. Âm không quyết định khoảng bất tử 0,15 giây hay khoảng cách tối đa 3 tile. |
| Dùng chung | `skill-ready.ogg`: báo hồi chiêu xong, nhỏ và chỉ trên máy người sở hữu. `skill-unavailable.ogg`: bấm khi chưa dùng được, chỉ cục bộ, tối đa một lần mỗi 300 ms. P5 dùng trạng thái hồi chiêu chung với lướt thường, không phát hai ready cue. |

Thời gian chiêu và hồi chiêu lấy từ trạng thái gameplay, không lấy từ độ dài file âm thanh. Loop hotfix gắn với ID vùng; charge và các one-shot gắn với ID lần dùng chiêu. Reconnect chỉ khôi phục loop của vùng còn tồn tại, không phát lại deploy, phản đạn, hit hay heal cũ. Âm chiêu riêng có ưu tiên dưới âm báo đòn boss và đồng đội gục.

### Tín hiệu âm thanh của boss — MVP

| File trong `assets/audio/sfx/` | Mốc kích hoạt | Ý nghĩa với người chơi |
| --- | --- | --- |
| `boss-volley-windup.ogg` | Bắt đầu `volley_windup` | Tiếng xòe giấy sắc, tăng nhẹ cao độ: sắp có loạt task bay. |
| `boss-volley.ogg` | Boss thực sự phóng loạt task | Tiếng quất giấy một lần cho cả loạt; không phát thêm một tiếng cho từng viên đạn. |
| `boss-wave-windup.ogg` | Bắt đầu `wave_windup` | Hai tiếng bánh răng nặng, khác hẳn tiếng xòe giấy: chuẩn bị né sóng ở sàn. |
| `boss-wave-slam.ogg` | Đồng hồ chạm sàn, sóng xuất hiện | Tiếng kim loại nện kèm luồng điện lướt, 300–600 ms; không lặp theo từng tile sóng. |
| `boss-footstep.ogg` (đã có asset) | Chân chạm sàn trong `walk` | Bước nặng ngắn, nhỏ hơn âm báo đòn; không phát theo mỗi snapshot. |
| `boss-charge-windup.ogg` (đã có asset) | Bắt đầu `charge_windup` | Chuông mời họp gấp, phân biệt với báo task và sóng; bắt đầu cùng hành lang báo lao. |
| `boss-charge.ogg` (đã có asset) | Bắt đầu `charge` | Gió quét và đồng hồ rung, một lần cho mỗi cú lao. |
| `boss-drop-windup.ogg` (đã có asset) | Bắt đầu báo “Task rơi từ trên cao” | Thông báo văn phòng hai nốt đi lên và một tiếng tick, dài 0,9 giây, ưu tiên 10. Phát một lần cùng báo đòn 1 giây; khác chuông mời họp của cú lao. |
| `boss-drop-impact.ogg` (đã có asset) | Một chồng hồ sơ chạm bàn | Hồ sơ dày đập mặt gỗ và giấy văng, dài 0,25 giây, ưu tiên 7. Một âm cho mỗi chồng, tối đa 3 cột cách nhau 0,25 giây; không phát theo từng tờ giấy. |
| `boss-paper-jam.ogg` (đã có asset) | Bắt đầu trạng thái kẹt giấy sau lượt thả | Máy in nghiến, khựng rồi giật giấy, dài 1,2 giây, ưu tiên 5, không loop. Thời gian dễ bị đánh và nhân đôi sát thương do server quyết định, không lấy từ độ dài âm. |
| `boss-wind-clock.ogg` | Boss bắt đầu `wind_clock` | Chuỗi bánh cóc 0,5–1 giây, lặp nhỏ trong cửa sổ dễ bị đánh; dừng ngay khi boss đổi trạng thái. |
| `boss-hurt.ogg`, `boss-defeat.ogg` | Boss mất máu / bị hạ | Đồng hồ rung chát / lò xo bung và giấy rơi. Boss defeat phát trước câu nhạc chiến thắng. |

Âm báo đòn phải bắt đầu cùng tín hiệu hình ảnh, trước thời điểm gây sát thương theo nhịp windup đã chốt của boss. Âm thanh hỗ trợ đọc đòn; tắt tiếng vẫn nhìn được loại đòn và vùng nguy hiểm. Không lấy chiều dài file audio làm bộ đếm thời gian gây sát thương.

### Trộn âm và multiplayer

- Có Master, Music, SFX và nút tắt tiếng; nhớ lựa chọn trên từng trình duyệt. Mức khởi đầu gợi ý: Master 70%, Music 35%, SFX 70%, ambience thấp hơn nhạc rõ rệt; cần nghe thử rồi cân bằng.
- Ưu tiên nghe: báo đòn boss và đồng đội gục → nhân vật của mình → đòn trúng và vật phẩm → đồng đội khác → tiếng bước chân và ambience. Hạ nhạc ngắn khi có báo đòn quan trọng.
- Tối đa 16 hiệu ứng cùng lúc; giữ chỗ cho tối đa 4 âm quan trọng. Khi quá tải, bỏ tiếng bước chân hoặc tiếng đánh của đồng đội trước; không cắt âm báo đòn để phát tiếng trang trí.
- Tiếng đồng đội nhỏ hơn tiếng của mình, giảm theo khoảng cách; vị trí trái/phải theo khung nhìn, pan nhẹ để dùng loa laptop vẫn nghe được. Âm báo boss quan trọng và tín hiệu trận đấu luôn nghe rõ.
- Sự kiện chiến đấu có ID từ trạng thái trận đấu để client phát một lần, không phát lại trên mỗi snapshot hoặc khi reconnect. Reconnect chỉ khôi phục nhạc/ambience/trạng thái boss đang lặp, không phát lại lịch sử đòn đánh.
- Nhạc là nền trên từng client; nhịp sát thương, deadline và kết quả lấy từ trạng thái trận đấu. Khi trận tạm dừng vì mọi người rớt mạng, nhạc/âm lặp cũng tạm dừng; rời màn phải dừng và giải phóng các loop.
- Trình duyệt chỉ bắt đầu âm thanh sau thao tác đầu tiên như bấm Vào phòng. Không phát âm ở trang chờ chưa tương tác; đổi mute hoặc âm lượng không ảnh hưởng gameplay.

### Bàn giao và kiểm tra âm thanh

Giữ bản gốc WAV PCM 48 kHz, 16-bit trong `assets/audio/source/`; bản runtime OGG Vorbis ở các đường dẫn trên, kèm MP3 dự phòng nếu cần cho trình duyệt mục tiêu. SFX thường là mono, nhạc/ambience stereo. Không clipping; cắt khoảng lặng thừa ở đầu SFX và thêm fade rất ngắn để tránh tiếng bật mép sóng. Loop nhạc/ambience phải nối mượt, không có khoảng im hoặc tiếng click ở điểm lặp.

Kiểm tra bằng tai nghe và loa laptop: phân biệt được ba đòn boss (task, sóng, lao) khi 5 người cùng đánh; không nhân tiếng theo số đạn/tile/snapshot; đổi màn, chơi lại và reconnect không chồng loop; tắt tiếng vẫn đủ tín hiệu hình ảnh; mọi file runtime có bản ghi sự kiện tương ứng và tải được trên các trình duyệt mục tiêu. Âm thanh tự tạo hoặc có quyền dùng và bàn giao cùng game; lưu nguồn và giấy phép nếu dùng âm thanh bên ngoài.

### Hướng mở rộng sau MVP

Mỗi khu vực có một âm sắc riêng: Sảnh Onboarding dùng chuông họp, Tháp Kiến Trúc dùng tiếng giấy blueprint mở, Hầm Tri Thức dùng khóa và tiếng ổ cứng, QUY TRÌNH dùng máy in kẹt giấy. Chế độ tay đôi cần tín hiệu vào trận và kết quả riêng. Extension tiệc cuối game chuyển sang disco chiptune, tiếng tiền rơi, ATM chốt số dư và phản ứng NPC ngắn; giảm nhạc khi có đối thoại và không lặp câu thoại theo mỗi đòn đánh.

## Hai chế độ

### Tay đôi — Trao đổi nhanh 5 phút

- 1v1 PvP giữa hai người chơi; có thể tập solo đấu boss AI.
- Bản đồ gọn hơn co-op, bố trí đối xứng với sàn chính và 2–3 bục nhảy.
- Người chơi dùng cùng nền tảng di chuyển, khác bộ kỹ năng và ngoại hình.
- Đấu khoảng 3–5 phút. Thắng khi hạ đối thủ; hết giờ so máu còn lại.
- Phản đòn đúng lúc kích hoạt “Có biên bản”, bật lại đạn task.

### Co-op — Sprint cuối cùng

- Mục tiêu ban đầu: 1–5 người. Sample hình ảnh thể hiện 4 người.
- Cùng vượt màn và đánh boss; tắt sát thương và va chạm cứng giữa đồng đội để tránh chặn đường nhau.
- Năm màu nhận diện teal, tím, vàng, mint, cam; thêm nhãn P1–P5 và dấu chỉ vị trí để không chỉ phụ thuộc màu.
- Mỗi người có camera riêng bám theo nhân vật khi vượt màn. Khi đánh boss, camera mỗi máy đóng khung cùng đấu trường và chỉ nới trong giới hạn dễ đọc. MVP chỉ có đấu trường boss nên dùng một camera đóng khung cố định.
- Đồng đội gục ở trạng thái “Nghỉ phép”; người khác giữ tương tác gần họ để cứu trong lúc né đòn.
- Boss tăng sức bền và số hướng tấn công theo số người; luôn chừa lối né.
- Ping nhanh: Tôi nhận / Cần hỗ trợ / Chặn scope.

## Không gian multiplayer rộng hơn

Thông số dưới đây là giả định thiết kế ban đầu, cần thử bằng prototype.

- Khung nhìn tham chiếu 1280 × 720, tile cơ sở 32 × 32, nhân vật khoảng 32 × 48 pixel gốc. Phóng vừa cửa sổ bằng nearest-neighbor, chừa viền khi tỷ lệ màn hình khác 16:9; tỷ lệ phóng có thể lẻ. Đổi từ 1920 × 1080 ngày 2026-10-15 để đấu trường chật và gần như concept: người chơi chiếm khoảng 7% chiều cao khung, boss khoảng 17%.
- Đấu trường boss MVP 40 × 22 tile (1280 × 704 px), camera đóng khung cả đấu trường. Sàn chính ở đáy; 7 bàn một chiều trên 3 tầng cách nhau 4 tile (nhảy cao tối đa khoảng 5 tile): 2 bàn thấp, 1 bàn giữa và 2 gờ sát tường, 2 bàn cao, bố trí đối xứng. Văn phòng chật đồ: bàn có màn hình, ghế, tủ hồ sơ, cây, máy nước, poster, đèn trần; trang trí làm tối nhẹ. Đoạn vượt màn sau này có thể dài hơn vì camera bám theo từng người.
- Có sàn chính liên tục và 2–3 cao độ phụ, gồm 5–7 bục rộng được bố trí lệch nhau. Chừa hành lang ngang, khoảng đáp và khoảng nhảy giữa các bục.
- Điểm bắt đầu: bục rộng 6–10 tile, khoảng trống ngang 3–5 tile, chênh cao 3–4 tile; kiểm chứng lại sau khi chốt vận tốc và độ cao nhảy.
- Boss khoảng 2–3 lần chiều cao người chơi (đang dùng 2,5 lần). Nhân vật và boss vẫn nhỏ so với khung hình để thấy rõ đường chạy và toàn bộ tín hiệu đòn đánh.
- Chừa khoảng trống hai bên boss và nhiều đường tiếp cận; không để trang trí che chân nhân vật hoặc mép bục.
- Nền văn phòng và skyline có parallax, tương phản thấp hơn mặt sàn chơi được. Các cao độ đều thuộc một mặt phẳng 2D nhìn ngang.
- Nếu thử 6 người, mở rộng đấu trường/đường di chuyển trước; tránh thu nhỏ nhân vật quá mức bằng camera.

## Các khu vực và boss

Mỗi khu vực là một màn, mô tả trong file riêng ở `levels/`.

| Khu vực | Boss | Trạng thái | Mô tả |
| --- | --- | --- | --- |
| Đường Đua Sprint — NA | Đốc Công Hôm Qua | Đã triển khai (MVP) | [01-sprint.md](levels/01-sprint.md) |
| Sảnh Onboarding — Vacancy | Ngài Định Hướng | Định hướng | [02-onboarding.md](levels/02-onboarding.md) |
| Tháp Kiến Trúc — Vacancy | Kiến Trúc Sư Ngoài Phạm Vi | Định hướng | [03-architecture-tower.md](levels/03-architecture-tower.md) |
| Hầm Tri Thức — NA | Song Thần Giữ Bài FE & BE | Chốt thiết kế, chưa triển khai | [04-knowledge-vault.md](levels/04-knowledge-vault.md) |
| Phòng Điều Hành | QUY TRÌNH | Định hướng; gồm trận cuối và kết thúc | [05-control-room.md](levels/05-control-room.md) |

Extension sau chiến dịch: “Anh em cây khế” — loạn đấu giành tiền, đã triển khai trừ phần NPC: [06-party.md](levels/06-party.md).

## Hướng hình ảnh và asset mẫu

Pixel art 16-bit, silhouette rõ, nhân vật hài hước. Nền navy, màn hình cyan, mặt sàn màu kem ấm và tín hiệu nguy hiểm coral đỏ. Tất cả gameplay nhìn ngang; mép trên của bục phải rõ và nằm ngang.

### Bảng mẫu v02 để duyệt

- Phần lớn diện tích dành cho một đấu trường co-op rộng với bốn người chơi nhỏ, phân bố ở nhiều vị trí và cao độ; boss ở bên phải, có chỗ chạy vòng hai bên.
- Dải tham chiếu nhỏ bên dưới: dev đứng/chạy/nhảy/đánh, boss PM, tile sàn/bục/server, cà phê, tài liệu, chìa khóa và đạn task.
- Giữ hướng nhân vật và màu của sample 01. Giảm tỷ lệ nhân vật so với môi trường, tăng không gian di chuyển.
- Đây là concept để duyệt bố cục và phong cách, chưa phải atlas hoặc sprite sheet có thể cắt dùng trực tiếp.

### Asset dự kiến sau khi duyệt

| Nhóm | Nội dung |
| --- | --- |
| Người chơi | Idle, run, jump, fall, attack, dash, hurt, down, revive; 5 màu và nhãn người chơi |
| Boss | Idle, walk, charge_windup, charge, charge_recover, báo trước đạn/sóng, đánh, lên dây đồng hồ, trúng đòn, chuyển phase, bị hạ |
| Tile và props | Sàn nối được, bục một chiều, bàn, server, cửa thang máy, đèn, cây, biển chỉ dẫn |
| Nền | Skyline, tường kính, nội thất xa và gần dưới dạng các lớp parallax |
| Vật phẩm | Cà phê, trang tài liệu, chìa khóa server, checkpoint |
| VFX và UI | Đạn task, cung đánh bàn phím, bụi đáp đất, vùng báo đòn, máu boss, nhãn P1–P5, ping và dấu cứu đồng đội |

Mẫu ảnh và prompt tạo ảnh lưu trong `assets/concepts/`. Các file này phục vụ duyệt hình ảnh, không được hiểu là asset runtime đã hoàn thiện.
