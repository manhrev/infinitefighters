# Đường Đua Sprint — Đốc Công Hôm Qua

Màn 1 · cánh NA. Màn co-op của bản MVP, đã triển khai. Tổng quan game, bối cảnh, âm thanh và không gian multiplayer: [game-scenario.md](../game-scenario.md).

Cốt truyện: đêm chốt sprint, Đốc Công Hôm Qua dí cả đội ngay tại cánh NA. Hạ boss là sprint “done”, cả đội được mời đi tiệc tổng kết: extension loạn đấu [01b-party.md](01b-party.md).

Đi kiểm tra tiến độ trên sàn, phóng đạn task, tạo sóng deadline và lao tới bằng “Họp nhanh 5 phút”. Nhảy lên bục để né sóng/cú lao, xuống sàn né loạt đạn cao, đánh lúc boss lên dây đồng hồ hoặc hồi sức sau lao. Thoại: “Task nhỏ thôi!”

## Trạng thái triển khai

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
- Khung nhìn 1280 × 720 và đấu trường 40 × 22 tile (xem mục Không gian multiplayer trong [game-scenario.md](../game-scenario.md)).
- Thua thì server tự chơi lại sau 4 giây, không chờ người chơi bấm; nhạc đổi đúng lúc chơi lại.
- Boss nhanh hơn thông số thiết kế để tăng độ khó: đi tối đa 1 giây (thiết kế 2 giây), báo lao 0,75 giây (thiết kế 0,9), hồi sức sau lao 1 giây (thiết kế 1,5). Thiếu chỗ lao thì lao ngược hướng; chỉ khi cả hai phía đều thiếu chỗ mới đổi sang lên dây đồng hồ.

## Năm chiêu riêng cho P1–P5 — đánh boss bằng nghiệp vụ công sở

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

## Boss MVP biết di chuyển — Đốc Công Hôm Qua

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
