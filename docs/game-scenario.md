# TASK VÔ TẬN

Trạng thái: phần MVP đã triển khai; số liệu thực tế và các điểm lệch so với thiết kế nằm ở mục **Trạng thái triển khai**. Các phần còn lại là định hướng.
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

Đồng bộ với code ngày 2026-10-15 (`infinitefighters-backend/`, `infinitefighters-frontend/`). Thông số dưới đây là giá trị đang chạy, vẫn là mốc khởi đầu để cân bằng khi chơi thử.

**Đã có:** phòng chờ, vào lại bằng token, di chuyển (chạy, nhảy, xuyên bục một chiều, rơi xuống bục), đánh bàn phím, xung code, lướt, năm chiêu riêng, boss Đốc Công Hôm Qua đầy đủ (đi, đạn task, sóng deadline, lao, lên dây đồng hồ), máu và trạng thái “Nghỉ phép”, cứu đồng đội, cà phê hồi máu, ping, thanh DEADLINE, thắng/thua, âm thanh và HUD.

**Điều khiển:** A/D hoặc ←/→ chạy · Space/W/↑ nhảy · S/↓ + nhảy để xuống bục · J đánh · K xung code · L/Shift lướt (theo hướng đang nhìn hoặc hướng đang giữ phím, bất tử suốt 0,25 giây lướt) · **R** chiêu riêng · giữ **E** cạnh đồng đội để cứu · 1/2/3 ping.

| Thông số co-op | Giá trị |
| --- | --- |
| Máu người chơi | 5 tim; trúng đòn mất 1 tim và bất tử 0,6 giây. Cứu xong có 3 tim. Giữ E 3 giây trong tầm 48 px. |
| Máu boss | 120 + 60 cho mỗi người thêm, đếm mọi slot trong phòng lúc bắt đầu (kể cả người đang rớt mạng); chơi lại sau khi thua giữ nguyên. |
| Đòn chung | Đánh bàn phím 1, xung code 2 (hồi 1,5 giây), lướt 0,25 giây bất tử (hồi 0,8 giây). Lúc boss lên dây đồng hồ, mọi sát thương ×2. |
| P1 Em sửa có một dòng | Giữ tối đa 1 giây, giữ đủ thì tự bổ; sát thương 3 (thả ngay) đến 10 (đầy lực). Lướt hoặc bị trúng đòn đều hủy tích lực. Hồi 6 giây từ lúc thả/hủy. |
| P2 Không thuộc scope em | Khiên 0,6 giây, bật lại tối đa 3 task bay vào mặt khiên; task bật lại bay về phía boss, gây 3. Hồi 8 giây. |
| P3 Theo biên bản đã chốt | Dấu gây 2, đánh dấu boss 4 giây: đòn bàn phím, cú bổ P1 và xung code của cả đội +15%. Đóng dấu lại chỉ làm mới thời gian. Hồi 10 giây. |
| P4 Máy em vẫn chạy | Vùng 128 × 48 px tồn tại 3 giây, gây 3 lúc đặt, mỗi người trong vùng hồi tối đa 1 tim mỗi lần dùng, không cứu người gục. Hồi 14 giây. |
| P5 Em về đúng giờ | Lao 3 tile trong 0,2 giây, bất tử 0,15 giây đầu, gây 4 một lần, dừng ở tường, dùng được trên không. Dùng chiêu khóa lướt 7 giây; lướt thường khóa chiêu 0,8 giây. |
| Boss | Vòng đòn: đạn task → sóng → lao → đạn task → sóng → lao → lên dây đồng hồ (1,5 giây). Không đủ chỗ lao về phía người chơi thì lao ngược lại. Dưới 50% máu boss cuồng nộ: đạn task và sóng luôn có thêm một đợt, không nghỉ giữa các lượt. Nhịp nhanh (chỉnh 2026-10-15 để khó hơn): nghỉ 0,33 giây giữa các lượt, đi tối đa 1 giây, báo đòn 0,75 giây, hồi sức sau lao 1 giây; đạn task 600 px/s, sóng 540 px/s. Từ 2 người, đạn task bay cả hai phía; từ 3 người, đạn và sóng có thêm một đợt. Khoảng 23 lượt đòn mỗi phút. |
| Đòn rơi hồ sơ | “Em đang ở trên thì tiện nhận luôn nhé!”. Khi boss dưới 75% máu, tối đa một lượt mỗi 8 giây, thay cho lượt ném task nếu có người đứng trên bàn (không ai trên bàn thì vẫn ném task). Khóa tối đa 3 cột rộng 2 tile vào người trên bàn, mỗi bàn một cột, cột nằm gọn trên mặt bàn và chừa ít nhất 2 tile trống. Báo đòn 1 giây (cột cảnh báo và dấu “!”), hồ sơ rơi cách nhau 0,25 giây từ ngay dưới bàn phía trên (hoặc trần), vỡ khi chạm bàn đích nên không trúng tầng trên hay tầng dưới. Trúng mất 1 tim, mỗi người tối đa một lần mỗi lượt. Sau đó boss kẹt giấy 1,2 giây, nhận sát thương ×2. Khiên P2 không bật lại hồ sơ. |
| Mode khó (`-hard`) | Đòn rơi có ngay từ đầu trận và thả 6 chồng cùng lúc: nhắm người trên bàn trước, phần còn lại ngẫu nhiên trên mọi bàn và sàn; mỗi bàn một cột, trên cùng một tầng hai cột cách nhau ít nhất 2 tile trống. Cuồng nộ ngay từ đầu (luôn thêm một đợt đạn/sóng, không nghỉ giữa lượt). Không có bất tử sau khi trúng đòn; mỗi sóng vẫn chỉ trừ 1 tim mỗi người. Vẫn giữ bất tử khi lướt, 0,15 giây đầu cú lao P5 và ngay sau khi được cứu. |
| Boss nửa máu (pha 2) | Khi boss còn dưới 50% máu, ở mọi chế độ: tốc độ đi và lao gấp đôi, mọi bộ đếm thời gian của boss còn 2/3 (báo đòn 0,5 giây), đạn task và sóng nhanh gấp rưỡi. Đo được khoảng 37 lượt đòn mỗi phút, gấp rưỡi pha 1 (24). Vẫn né được: nhảy qua cú lao có cửa sổ 0,3 giây. |
| Cà phê | 3 cốc trên bàn, chạm vào hồi 1 tim (người đầy máu đi qua không lấy), 30 giây sau có lại. |
| Kết thúc | Thua: hiện lý do 4 giây rồi server tự chơi lại màn. Thắng: hiện “APPROVED!” 6 giây rồi về phòng chờ; người đang rớt mạng mất chỗ. |

**Khác với thiết kế:**

- Chiêu riêng dùng phím **R**, không phải E như gợi ý, vì E đã là giữ để cứu đồng đội.
- Boss không gây sát thương khi chạm; chỉ các đòn có báo trước (đạn, sóng, thân boss đang lao) mới gây sát thương.
- Một camera chung đóng khung cả đấu trường; chưa có đoạn vượt màn nên chưa có camera riêng bám từng người.
- Khung nhìn 1280 × 720 và đấu trường 40 × 22 tile (xem mục Không gian multiplayer).
- Thua thì server tự chơi lại sau 4 giây, không chờ người chơi bấm; nhạc đổi đúng lúc chơi lại.
- Boss nhanh hơn thông số thiết kế để tăng độ khó: đi tối đa 1 giây (thiết kế 2 giây), báo lao 0,75 giây (thiết kế 0,9), hồi sức sau lao 1 giây (thiết kế 1,5). Thiếu chỗ lao thì lao ngược hướng; chỉ khi cả hai phía đều thiếu chỗ mới đổi sang lên dây đồng hồ.

**Màn “Anh em cây khế” — extension loạn đấu (đã có, trừ phần NPC):** hạ boss xong thì những người còn kết nối chuyển sang tiệc tổng kết trên map `party.json` (người đang rớt mạng không vào tiệc). Chủ phòng cũng có thể chọn thẳng màn này ở phòng chờ, không cần đánh boss trước. Có đủ các luật đã chốt: đếm ngược 5 giây kèm luật chơi, 100 máu, đòn và năm chiêu theo bảng PvP, két phun 50 × N xu mỗi 15 giây và 150 × N xu ở giây 150 (băng rôn HR), tiền rơi 10%/50% làm tròn lên, chủ cũ chờ 1 giây, ATM giữ E 3 giây, hồi sinh 3 giây ở điểm xa đối thủ nhất kèm 2 giây bảo vệ, rớt mạng 5 giây, bản gom thưởng solo, bảng kết quả “Ai ôm tiền” rồi về phòng chờ sau 10 giây. Cách hiểu đã chọn ở chỗ tài liệu chưa nói rõ:
- Bị trúng đòn khi đang tích lực P1 thì hủy tích lực (giống co-op).
- Mỗi lần két phun chia thành 5 bó (đợt cuối 15 bó), mỗi bó 10 × N xu.
- Vùng hotfix P4 hồi dần cho chính P4 khi đứng trong vùng, tối đa 20 máu trong khoảng 1 giây.
- Tiền rơi văng ra rồi rơi xuống bề mặt bên dưới; bó tiền không tự biến mất.
- Phòng chờ hiện hai thẻ màn kèm mô tả: “Đường Đua Sprint” (đánh boss) và “Anh em cây khế” (loạn đấu giành tiền); chủ phòng bấm thẻ nào thì vào màn đó. Cờ `-party` vẫn có, chỉ để chọn sẵn màn tiệc khi chủ phòng không gửi lựa chọn.

Chưa có: NPC khách mời, tặng quà, “Ai có hẹn” (phần “Tiền + gái” chưa có luật chốt); hình và âm riêng của tiệc (ATM, két, bó tiền, nhạc disco) đang dùng hình/âm tạm cho tới khi có asset.

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

### Năm chiêu riêng cho P1–P5 — đánh boss bằng nghiệp vụ công sở

Bổ sung thiết kế cho màn co-op MVP: mỗi player có một chiêu riêng theo slot, dùng chung nền tảng đánh bàn phím, xung code và lướt né. Đã triển khai; số liệu đang chạy nằm ở mục Trạng thái triển khai.

| Player | Chiêu riêng | Cách đánh và tác dụng | Câu châm biếm |
| --- | --- | --- | --- |
| **P1 — Teal: người gánh task** | **Em sửa có một dòng** | Giữ chiêu tối đa 1 giây để dồn lực vào bàn phím, rồi bổ mạnh trước mặt. Đánh trúng gây sát thương lớn lên một mục tiêu; có thể thả sớm để đánh nhẹ. Trong lúc giữ vẫn đi chậm được nhưng dễ trúng đòn; lướt sẽ hủy tích lực. Hồi chiêu 6 giây từ lúc thả hoặc hủy. | “Một dòng thôi, sao production tối rồi?” |
| **P2 — Tím: chuyên viên chuyển tiếp** | **Không thuộc scope em** | Dựng một tấm khiên thư chuyển tiếp phía trước trong 0,6 giây, bật lại tối đa 3 viên đạn task về phía kẻ bắn. Đạn trả lại gây sát thương lên boss; sóng deadline và đòn cận chiến vẫn xuyên qua. Hồi chiêu 8 giây. | “Em xin phép loop anh vào xử lý.” |
| **P3 — Vàng: người chốt biên bản** | **Theo biên bản đã chốt** | Phóng một con dấu tầm ngắn, gây sát thương và đánh dấu boss trong 4 giây. Các đòn bàn phím và xung code của cả đội lên mục tiêu được tăng 15% sát thương; dấu không cộng dồn. Hồi chiêu 10 giây. | “Anh đã đồng ý trong cuộc họp không có anh.” |
| **P4 — Mint: người cứu production** | **Máy em vẫn chạy** | Đặt một vùng hotfix tại chân trong 3 giây. Khi đặt, xung điện gây sát thương một lần lên kẻ địch trong vùng; bản thân và đồng đội đứng trong vùng được hồi tối đa 1 tim mỗi lần dùng. Không cứu người đã gục; vẫn phải né đòn và tương tác cứu như thường. Hồi chiêu 14 giây. | “Để em zip nguyên máy gửi anh.” |
| **P5 — Cam: người đòi tan làm** | **Em về đúng giờ** | Lao ngang theo hướng đang nhìn tối đa 3 tile, để lại vệt thẻ chấm công gây sát thương một lần lên mỗi kẻ địch đi qua. Có thể dùng trên không, dừng khi gặp tường; chỉ có bất tử trong 0,15 giây đầu. Hồi chiêu 7 giây, dùng chung hồi chiêu với lướt thường để tránh né liên tục. | “Task mai tính, cửa đang đóng!” |

- **Phối hợp:** P3 đóng dấu lúc boss lên dây đồng hồ → P1 bổ bàn phím và P5 lao qua; P2 canh đợt task tiếp theo để trả đạn, P4 đặt hotfix ở điểm tập hợp. Đây là cơ hội phối hợp, không phải chuỗi bắt buộc.
- **Chơi ít người:** chỉ xuất hiện các chiêu của slot đang có người; mọi nhân vật vẫn tự gây sát thương và né được các đòn boss bằng bộ chiêu chung. Không có cơ chế buộc đủ năm vai trò mới thắng; solo mặc định P1.
- **Giới hạn:** không chiêu nào dừng thanh DEADLINE, hủy báo đòn boss hoặc làm boss choáng liên tục. Chiêu hụt vẫn tiêu hồi chiêu; không gây sát thương hay đẩy đồng đội. Chiêu P2 là kỹ năng riêng trong co-op; phản đòn “Có biên bản” của PvP vẫn thuộc bản mở rộng.
- **Điều khiển và hình ảnh:** dùng chung một nút kỹ năng riêng, phím **R** (E đã dành cho giữ để cứu đồng đội); HUD hiện biểu tượng, tên chiêu và hồi chiêu theo P1–P5. Phân biệt bằng biểu tượng và chuyển động: bàn phím bổ, thư bật lại, dấu đóng, vòng hotfix, vệt thẻ; chữ châm biếm hiện ngắn trên UI, không che vùng báo đòn. VFX và âm thanh của từng chiêu đã có.

### Boss MVP biết di chuyển — Đốc Công Hôm Qua

Bổ sung thiết kế: boss đi lại trên sàn chính và có đòn lao tới, buộc cả đội đổi vị trí thay vì đứng một chỗ đánh. Các thông số là mức khởi đầu để thử nghiệm; đã triển khai cùng hoạt ảnh và âm thanh.

| Hành vi | Cách hoạt động | Cách đối phó |
| --- | --- | --- |
| **Đi kiểm tra tiến độ** | Giữa các lượt đánh, boss đi về phía vị trí ngang của một người chơi còn hoạt động, tốc độ bằng 40% tốc độ chạy thường của người chơi. Chọn mục tiêu gần nhất lúc bắt đầu đi, giữ tối đa 2 giây rồi dừng khi còn cách 3 tile hoặc đi đủ 6 tile. Nếu mục tiêu gục/rớt mạng thì dừng và chọn lại ở lượt sau. | Chạy vòng, đổi bục hoặc tiếp cận từ phía sau. Boss chỉ đi trên sàn chính, không nhảy lên bục và không dịch chuyển tức thời. |
| **Họp nhanh 5 phút** | Boss đứng lại, giơ đồng hồ và báo một hành lang nguy hiểm trên sàn trong 0,9 giây. Khóa hướng và điểm kết thúc ngay khi bắt đầu báo, rồi lao thẳng tối đa 8 tile trong 0,6 giây; dừng sớm nếu gặp tường hoặc giới hạn đấu trường. Chỉ thân boss đang lao gây sát thương, tối đa một lần lên mỗi người trong lượt đó. | Nhảy qua thân boss, lên bục hoặc ra sau vị trí xuất phát trước khi boss lao. Boss không quay đầu hay bám theo sau khi đã báo hướng. Thoại: “Có 5 phút thôi, xuống đây anh trao đổi!” |
| **Task rơi từ trên cao** | Khi boss đã yếu, giơ đồng hồ chỉ lên trên và thả chồng hồ sơ xuống những người đang đứng trên bàn (thông số ở mục Trạng thái triển khai). | Chạy khỏi cột, nhảy sang bàn khác hoặc chủ động thả xuống. Đứng lì trên bàn không còn an toàn. |
| **Chờ mọi người vào call** | Sau cú lao, boss đứng thở và chỉnh đồng hồ trong 1,5 giây, không gây sát thương tiếp xúc. Sau đạn task hoặc sóng deadline vẫn có khoảng lên dây đồng hồ như trước. | Tận dụng khoảng đứng yên để đánh cận chiến, đóng dấu hoặc đặt hotfix gần đội. |

- **Nhịp trận:** đổi vị trí → báo đòn → ra đòn → hồi sức. Luân phiên đạn task và sóng deadline; sau mỗi hai lượt này, lượt tiếp theo là “Họp nhanh 5 phút”. Chỉ bắt đầu đổi vị trí/lượt mới khi đạn và sóng của lượt cũ đã biến mất; không vừa lao vừa phóng task hoặc tạo sóng.
- **Lối né:** phạm vi đi và lao phải giữ khoảng trống ít nhất 3 tile giữa mép thân boss và mỗi tường biên. Nếu không đủ đường lao tối thiểu 3 tile hoặc không có đường nhảy/bục an toàn, thay lượt lao bằng lên dây đồng hồ. Cơ thể boss không chặn đường người chơi; đi thường, báo đòn và hồi sức không gây sát thương tiếp xúc, không đẩy người chơi vào tường.
- **Đánh từ vị trí mới:** boss đứng yên và khóa hướng khi bắt đầu báo đạn task hoặc sóng. Đạn task vẫn chừa sàn an toàn; sóng vẫn có thể né bằng bục. Tín hiệu nguy hiểm lấy theo vị trí thực tế của boss, không theo điểm đứng ban đầu.
- **Multiplayer:** bỏ qua người gục hoặc rớt mạng khi chọn mục tiêu. Nếu mục tiêu biến mất sau khi đã báo lao, boss vẫn đi theo hành lang đã khóa để giữ tín hiệu nhất quán. Khi toàn bộ người chơi rớt mạng, cả chuyển động và bộ đếm đòn đều tạm dừng; chơi lại đưa boss về vị trí xuất phát.
- **Hình ảnh và âm thanh (đã có):** `walk`, `charge_windup`, `charge`, `charge_recover`; bụi chân và hành lang báo lao không che mép bục. Tiếng bước nặng khi đi, chuông mời họp khi báo lao và gió quét khi lao. Âm báo cùng lúc với hình ảnh; tắt tiếng vẫn đọc được hướng và điểm dừng.

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

Thiết kế chi tiết Hầm Tri Thức đã chốt cho bản chơi thử tại [knowledge-vault-spec.md](knowledge-vault-spec.md) (chưa triển khai): hai boss, năm chiêu, tài liệu mở cầu và công tắc mở khiên; bao gồm luật solo, tương tác người chơi và điều kiện nghiệm thu. Với màn này, tương tác công tắc dùng giữ E theo spec chi tiết.

| Khu vực | Boss | Đòn đánh và cách vượt |
| --- | --- | --- |
| Sảnh Onboarding — Vacancy | Ngài Định Hướng | Đi qua biển chỉ dẫn đổi hướng và bàn họp. Boss tạo vòng họp đứng yên gây sát thương, phóng mũi tên chỉ việc mâu thuẫn; nhảy sang bục trống rồi đánh khi boss đang đổi slide. |
| Tháp Kiến Trúc — Vacancy | Kiến Trúc Sư Ngoài Phạm Vi | Khiên blueprint chặn đòn phía trước. Boss ném bom yêu cầu xuống sàn; dẫn bom về khiên để phá giáp, sau đó tiếp cận tấn công. Thoại: “Phần này dev tự xử lý nhé.” |
| Đường Đua Sprint — NA | Đốc Công Hôm Qua | Đi kiểm tra tiến độ trên sàn, phóng đạn task, tạo sóng deadline và lao tới bằng “Họp nhanh 5 phút”. Nhảy lên bục để né sóng/cú lao, xuống sàn né loạt đạn cao, đánh lúc boss lên dây đồng hồ hoặc hồi sức sau lao. Thoại: “Task nhỏ thôi!” |
| Hầm Tri Thức — NA | Song Thần Giữ Bài FE & BE | Hai boss đứng ở hai cụm bục, khóa cầu nối bằng code bí truyền. Nhặt trang tài liệu để mở đường; đánh công tắc hai bên để vô hiệu khiên. Solo cho phép kích hoạt tuần tự trong cửa sổ dài hơn. Thoại: “Không hiểu thì cứ hỏi.” — trạng thái: Đang bận. |
| Phòng Điều Hành | QUY TRÌNH | Máy in khổng lồ có tay máy, kết hợp task, khiên và sóng deadline. Những đường giấy nối sang các phòng ban chuyển sát thương đi; cắt đường giấy bằng đòn đánh, rồi đánh lõi máy in lúc nó kẹt giấy. |

## Trận cuối và kết thúc

Khi QUY TRÌNH còn 10% máu, nút TÁI CẤU TRÚC xuất hiện: “Hy sinh một đồng đội để hồi đầy tài nguyên.” Đây là tình tiết lựa chọn trong màn co-op, cần cả đội đồng ý để tránh một người tự loại người khác.

- Chọn tái cấu trúc: kết thúc nhanh, nhân vật bị hy sinh xuất hiện như boss hư cấu trong đoạn kết; vòng lặp công sở tiếp tục.
- Từ chối: sống sót thêm một phase và kích hoạt bốn công tắc Ưu tiên rõ / Chia sẻ kiến thức / Quyền quyết định / Trách nhiệm chung. Có thể bật lần lượt, không đòi đủ bốn người. Giáp máy in vỡ, cả đội đánh lõi và mở cửa thoát.

Đội bước ra ngoài lúc trời sáng. Điện thoại rung: “Mai retro nhẹ để xem tại sao hôm nay về trễ nhé.”

## Extension cuối game — Anh em hết mình, tiền tình hết hồn

Định hướng mở rộng sau chiến dịch, ngoài phạm vi MVP. Dù chọn kết thúc nào, những người chơi còn lại được mời đến tiệc tổng kết. Nhánh tái cấu trúc giữ hậu quả đã chọn: nhân vật bị hy sinh vẫn chỉ xuất hiện như boss trong đoạn kết.

HR công bố quỹ thưởng chỉ đủ cho một người, kèm danh hiệu “Người gánh team” và một buổi hẹn do các cô nàng NPC khách mời tự chọn người tham gia. Màn hình hiện bảng công trạng: ai cũng được ghi là người đóng góp chính, trừ phần trách nhiệm. Một người nói: “Tiền chia sao cũng được, anh em là chính.” Ngay sau đó, người ấy nhặt chiếc vali tiền và chạy.

### Từ co-op sang loạn đấu — màn “Anh em cây khế”, luật đã chốt

Chốt thiết kế ngày 2026-10-15; vẫn thuộc bản mở rộng sau chiến dịch, cần triển khai và thử cân bằng. Các thông số dưới đây là mốc thử nghiệm đầu tiên.

- **Mở phase:** cả đội nâng ly: “Anh em là chính.” HR đặt vali xuống: “Ngân sách chỉ đủ một suất.” Nhãn ĐỒNG ĐỘI chuyển thành ĐỐI TƯỢNG CẠNH TRANH, nhạc chuyển disco chiptune. Hiển thị luật “Nhặt tiền → đánh rơi tiền đối thủ → gửi ATM” và đếm ngược 5 giây; lúc này chưa đánh, nhặt hay gửi được. Vali trong cảnh mở đầu chỉ là đạo cụ, không cho ai lợi thế tiền trước giờ bắt đầu.
- **Xuất phát:** giữ nhân vật và chiêu riêng P1–P5, mọi người bắt đầu với 100 máu, mọi chiêu sẵn sàng, ví và số dư ATM bằng 0. Xóa buff, đạn, vật phẩm và tài nguyên còn lại từ boss. Đặt người chơi ở các điểm xuất phát cách đều đường tới két. Trận kéo dài 180 giây từ khi đếm ngược kết thúc.
- **Đấu trường:** sảnh tiệc gọn, nhiều bàn ghế và 5–7 bục ở 2–3 cao độ; trang trí không che đòn hoặc chặn đường. Hai ATM nằm ở hai phía, mỗi máy có ít nhất hai đường tiếp cận. Camera đóng khung toàn sảnh; người chơi không va chạm cứng với nhau để tránh chặn ATM bằng thân người.
- **Thắng:** chỉ tính tiền đã gửi thành công. Hết giờ, người có số dư cao nhất thắng; hòa thì đồng hạng, kể cả cùng 0. Tiền còn trong ví hoặc trên sàn không cộng điểm. Số lần hạ người chỉ là thống kê vui, không phá hòa.

### Giao tranh và năm chiêu riêng trong PvP

Đánh thường bằng bàn phím gây 10 sát thương, xung code gây 10; hồi chiêu lần lượt 0,4 và 1,2 giây. Lướt thường kéo dài 0,2 giây, bất tử trong lúc lướt, hồi chiêu 2 giây. Mỗi lần nhận sát thương có 0,6 giây bất tử để tránh bị nhiều người khóa combo; đòn bị chặn bởi bất tử không làm rơi tiền hay ngắt ATM. Mỗi lượt vung hoặc lần dùng chiêu chỉ đánh trúng mỗi mục tiêu một lần. Không có choáng nối tiếp; sát thương vượt máu còn lại không tạo thêm tiền rơi.

| Player | Chiêu giữ lại | Điều chỉnh cho loạn đấu |
| --- | --- | --- |
| P1 | **Em sửa có một dòng** | Giữ tối đa 1 giây, sát thương tăng tuyến tính từ 15 lên 30. Đánh một đối thủ gần nhất trước mặt; tích lực đi chậm, lướt hủy chiêu. Hồi chiêu 6 giây từ lúc thả/hủy. |
| P2 | **Không thuộc scope em** | Khiên trước mặt 0,6 giây, phản tối đa 3 xung code hoặc con dấu P3 về hướng kẻ bắn tại thời điểm phản. Đạn phản gây 10 sát thương, thuộc về P2 và mất hiệu ứng đánh dấu; không phản lần nữa. Không chặn bàn phím, P1 bổ, vùng hotfix hoặc cú lao P5. Hồi chiêu 8 giây. |
| P3 | **Theo biên bản đã chốt** | Con dấu gây 10 sát thương và đánh dấu một đối thủ trong 4 giây. Chỉ đòn bàn phím/xung code của chính P3 được tăng 15% sát thương lên người đó; không tăng cho người khác. Mỗi P3 chỉ giữ một dấu, dấu không cộng dồn. Hồi chiêu 10 giây. |
| P4 | **Máy em vẫn chạy** | Vùng hotfix tồn tại 3 giây. Lúc đặt, gây 10 sát thương một lần lên từng đối thủ trong vùng; chỉ hồi cho chính P4 tối đa 20 máu trong mỗi lần dùng, không vượt 100. Không hồi máu đối thủ. Hồi chiêu 14 giây. |
| P5 | **Em về đúng giờ** | Lao ngang tối đa 3 tile, gây 15 sát thương một lần lên mỗi đối thủ trên đường; bất tử 0,15 giây đầu, dừng ở tường. Hồi chiêu riêng 7 giây và dùng chung cổng hồi với lướt: dùng chiêu khóa lướt 7 giây; dùng lướt thường khóa chiêu riêng 2 giây. |

Sát thương có phần lẻ được giữ nguyên khi tính máu, UI làm tròn để hiển thị. Chiêu riêng hụt vẫn hồi chiêu. Dấu P3 không tăng tiền rơi. Các giới hạn không đánh đồng đội của co-op được thay bằng luật PvP trong phase này; “Có biên bản” của chế độ tay đôi không tự thêm vào bộ chiêu loạn đấu.

### Nhặt tiền, cướp tiền và gửi ATM

- **Nguồn tiền:** đơn vị tiền nguyên, gọi là xu thưởng. Với N người tham gia lúc bắt đầu phase, két giữa sảnh phun tổng cộng 50 × N xu mỗi 15 giây, từ giây 0 đến 135. Tiền chia thành các bó nhỏ rải quanh két để không một người nhặt toàn bộ bằng một lần chạm. N giữ cố định khi có người rớt mạng.
- **Tiền rơi khi đánh:** đòn gây sát thương nhưng chưa hạ đối thủ làm rơi 10% ví hiện tại, làm tròn lên và không vượt số tiền đang có. Đòn kết liễu chỉ làm rơi 50% ví trước đòn, làm tròn lên; không cộng thêm khoản 10%. Phần còn lại giữ trong ví sau hồi sinh. Tiền đã gửi không bao giờ bị cướp.
- **Nhặt lại:** tiền rơi nằm tại điểm trúng đòn, ai cũng nhặt được; chủ cũ phải chờ 1 giây mới nhặt lại phần tiền đó. Mỗi bó chỉ được nhận một lần; nếu chạm cùng lúc, xử lý theo thứ tự xác nhận của server. Bó tiền không tự biến mất trước khi hết phase; tiền lọt ra ngoài vùng chơi được đưa về gần két.
- **ATM:** giữ tương tác và đứng yên trong tầm 1 tile suốt 3 giây để gửi toàn bộ ví. Nhận sát thương, di chuyển, ra đòn, dùng chiêu/lướt hoặc thả tương tác thì hủy, không mất tiền. Không được bắt đầu gửi trong lúc đang tích lực, giữ khiên hoặc lao; vùng hotfix đã đặt vẫn tồn tại. Mỗi máy phục vụ một người, người khác phải tự bắt đầu lại khi máy trống, không có hàng đợi tự động. Nếu cùng yêu cầu, server nhận yêu cầu hợp lệ đầu tiên.
- **Khi đang gửi:** không nhặt thêm tiền hoặc tiêu tiền tặng quà; UI hiện số tiền đang gửi và thanh tiến độ để cả người gửi lẫn đối thủ đọc được. Gửi thành công mới chuyển nguyên số tiền từ ví sang số dư; giao dịch hủy không chuyển một phần. Một nút tương tác ưu tiên ATM khi trong tầm, còn tặng quà cần chọn rõ hành động riêng để tránh tiêu nhầm tiền.
- **Cao trào 30 giây cuối:** ở giây 150, HR hô “Giải ngân trước khi khóa sổ!”, két bung một đợt 150 × N xu. Hai ATM vẫn hoạt động, không tăng sát thương hay thêm luật đột ngột. Mười giây cuối có đếm ngược rõ trên UI.
- **Khóa sổ:** giao dịch hoàn thành tại hoặc trước mốc 180 giây được tính. Sau mốc đó đóng ATM, dừng sát thương, nhặt tiền và tương tác rồi hiện kết quả; giao dịch chưa xong bị hủy. UI luôn phân biệt VÍ / ĐÃ GỬI và hiển thị số dư của mọi người.

### Bị hạ, hồi sinh và rớt mạng

- **Bị hạ:** không có cứu đồng đội trong phase loạn đấu. Hồi sinh sau 3 giây với 100 máu, ở điểm có khoảng cách tới đối thủ gần nhất lớn nhất trong các điểm hồi sinh hợp lệ. Xóa dấu/buff và hiệu ứng do nhân vật tạo ra, gồm hotfix, đạn và vệt lao; hồi chiêu tiếp tục chạy trong lúc chờ, không đặt về đầy khi hồi sinh.
- **Bảo vệ hồi sinh:** bất tử 2 giây, có viền nhấp nháy và không nhặt/gửi tiền trong thời gian đó. Chủ động tấn công hoặc dùng chiêu riêng kết thúc bảo vệ ngay trước khi thực hiện hành động; đi, nhảy hoặc lướt thường không hủy sớm. Chết không mất lượt tham gia và không bị loại khỏi phase.
- **Rớt mạng:** hủy gửi tiền và thao tác đang giữ; nhân vật đứng yên, vẫn có thể bị đánh trong tối đa 5 giây. Nếu bị hạ thì xử lý mất 50% ví một lần như bình thường, không tự hồi sinh khi còn offline. Sau 5 giây mất kết nối, toàn bộ phần ví còn lại rơi tại vị trí nhân vật rồi nhân vật rời sân; số dư đã gửi được giữ. Không phạt hay thả cùng khoản tiền hai lần.
- **Vào lại:** nối lại trong 5 giây giữ đúng máu/vị trí hiện có, không cấp bất tử mới; nếu đã bị hạ thì tiếp tục thời gian chờ hồi sinh. Vào lại sau khi đã rời sân dùng quy trình hồi sinh, ví bằng 0 và giữ số dư ATM. Các timer hồi chiêu không được reset để lấy lợi thế.
- **Số người:** nếu chỉ một người đủ điều kiện tham gia từ đầu (kể cả do nhánh tái cấu trúc), chơi bản gom thưởng solo với cùng luật két/ATM và thời gian; không thêm bot đối thủ. Nếu trận đã bắt đầu với từ hai người, một người rớt mạng không chuyển thành solo và không đổi luật thắng. Người bị hy sinh ở phần kết có thể xem trận, không điều khiển nhân vật trong sảnh.
- **Tạm dừng:** còn ít nhất một người kết nối thì đồng hồ tiếp tục. Khi tất cả cùng mất kết nối, dừng đồng hồ trận, timer rớt mạng, đạn và hiệu ứng; có người vào lại thì tiếp tục. Việc mất kết nối tự nó không trao chiến thắng sớm.

### Tiền + gái

Các cô nàng NPC là khách mời có tên, cá tính và gu riêng: người thích tay chơi nhiều tiền, người thích người biết giúp đỡ, người chỉ muốn ăn tiệc yên ổn. Người chơi có thể tặng quà bằng tiền chưa gửi, mời nhảy hoặc làm yêu cầu nhỏ giữa trận để tăng thiện cảm. NPC tự nhận lời hoặc từ chối; có nhiều tiền không bảo đảm có hẹn, và buổi hẹn không phải điều kiện thắng PvP.

Đánh nhau cạnh bàn của một NPC có thể làm mất thiện cảm với cô ấy. Cảnh hài chính: hai ông đang tranh công lấy lòng thì bị cả hai từ chối vì làm đổ ly nước. Các quyết định và phản ứng đều có tín hiệu ngắn, rõ trên màn hình để người chơi hiểu mình vừa được cộng hay trừ điểm.

### Đoạn kết mở rộng

Kết quả hiện thành hai mục riêng: **Ai ôm tiền** và **Ai có hẹn**. Một người có thể thắng cả hai, chỉ được một, hoặc trắng tay. Người thắng tiền nâng vali lên thì nhận thông báo: “Khoản thưởng đang chờ Approve.” Người có hẹn vừa ra tới cửa thì điện thoại rung: “Em rảnh tối nay không? Tiện xem giúp chị cái lỗi production.”

Khung hình cuối: cả nhóm ngồi trước cửa công ty, người ôm vali khóa, người ôm laptop. Câu thoại: “Mai vẫn party nhé?” — “Ừ. Nhưng bật chia loot từ đầu.”

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
