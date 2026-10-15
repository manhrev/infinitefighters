# Extension cuối game — Anh em hết mình, tiền tình hết hồn

Định hướng mở rộng sau chiến dịch, ngoài phạm vi MVP. Dù chọn kết thúc nào, những người chơi còn lại được mời đến tiệc tổng kết. Nhánh tái cấu trúc giữ hậu quả đã chọn: nhân vật bị hy sinh vẫn chỉ xuất hiện như boss trong đoạn kết.

HR công bố quỹ thưởng chỉ đủ cho một người, kèm danh hiệu “Người gánh team” và một buổi hẹn do các cô nàng NPC khách mời tự chọn người tham gia. Màn hình hiện bảng công trạng: ai cũng được ghi là người đóng góp chính, trừ phần trách nhiệm. Một người nói: “Tiền chia sao cũng được, anh em là chính.” Ngay sau đó, người ấy nhặt chiếc vali tiền và chạy.

## Trạng thái triển khai

**Màn “Anh em cây khế” — extension loạn đấu (đã có, trừ phần NPC):** hạ boss xong thì những người còn kết nối chuyển sang tiệc tổng kết trên map `party.json` (người đang rớt mạng không vào tiệc). Chủ phòng cũng có thể chọn thẳng màn này ở phòng chờ, không cần đánh boss trước. Có đủ các luật đã chốt: đếm ngược 5 giây kèm luật chơi, 100 máu, đòn và năm chiêu theo bảng PvP, két phun 50 × N xu mỗi 15 giây và 150 × N xu ở giây 150 (băng rôn HR), tiền rơi 10%/50% làm tròn lên, chủ cũ chờ 1 giây, ATM giữ E 3 giây, hồi sinh 3 giây ở điểm xa đối thủ nhất kèm 2 giây bảo vệ, rớt mạng 5 giây, bản gom thưởng solo, bảng kết quả “Ai ôm tiền” rồi về phòng chờ sau 10 giây. Cách hiểu đã chọn ở chỗ tài liệu chưa nói rõ:
- Bị trúng đòn khi đang tích lực P1 thì hủy tích lực (giống co-op).
- Mỗi lần két phun chia thành 5 bó (đợt cuối 15 bó), mỗi bó 10 × N xu.
- Vùng hotfix P4 hồi dần cho chính P4 khi đứng trong vùng, tối đa 20 máu trong khoảng 1 giây.
- Tiền rơi văng ra rồi rơi xuống bề mặt bên dưới; bó tiền không tự biến mất.
- Phòng chờ hiện hai thẻ màn kèm mô tả: “Đường Đua Sprint” (đánh boss) và “Anh em cây khế” (loạn đấu giành tiền); chủ phòng bấm thẻ nào thì vào màn đó. Cờ `-party` vẫn có, chỉ để chọn sẵn màn tiệc khi chủ phòng không gửi lựa chọn.

Chưa có: NPC khách mời, tặng quà, “Ai có hẹn” (phần “Tiền + gái” chưa có luật chốt); hình và âm riêng của tiệc (ATM, két, bó tiền, nhạc disco) đang dùng hình/âm tạm cho tới khi có asset.

## Từ co-op sang loạn đấu — màn “Anh em cây khế”, luật đã chốt

Chốt thiết kế ngày 2026-10-15; vẫn thuộc bản mở rộng sau chiến dịch, cần triển khai và thử cân bằng. Các thông số dưới đây là mốc thử nghiệm đầu tiên.

- **Mở phase:** cả đội nâng ly: “Anh em là chính.” HR đặt vali xuống: “Ngân sách chỉ đủ một suất.” Nhãn ĐỒNG ĐỘI chuyển thành ĐỐI TƯỢNG CẠNH TRANH, nhạc chuyển disco chiptune. Hiển thị luật “Nhặt tiền → đánh rơi tiền đối thủ → gửi ATM” và đếm ngược 5 giây; lúc này chưa đánh, nhặt hay gửi được. Vali trong cảnh mở đầu chỉ là đạo cụ, không cho ai lợi thế tiền trước giờ bắt đầu.
- **Xuất phát:** giữ nhân vật và chiêu riêng P1–P5, mọi người bắt đầu với 100 máu, mọi chiêu sẵn sàng, ví và số dư ATM bằng 0. Xóa buff, đạn, vật phẩm và tài nguyên còn lại từ boss. Đặt người chơi ở các điểm xuất phát cách đều đường tới két. Trận kéo dài 180 giây từ khi đếm ngược kết thúc.
- **Đấu trường:** sảnh tiệc gọn, nhiều bàn ghế và 5–7 bục ở 2–3 cao độ; trang trí không che đòn hoặc chặn đường. Hai ATM nằm ở hai phía, mỗi máy có ít nhất hai đường tiếp cận. Camera đóng khung toàn sảnh; người chơi không va chạm cứng với nhau để tránh chặn ATM bằng thân người.
- **Thắng:** chỉ tính tiền đã gửi thành công. Hết giờ, người có số dư cao nhất thắng; hòa thì đồng hạng, kể cả cùng 0. Tiền còn trong ví hoặc trên sàn không cộng điểm. Số lần hạ người chỉ là thống kê vui, không phá hòa.

## Giao tranh và năm chiêu riêng trong PvP

Đánh thường bằng bàn phím gây 10 sát thương, xung code gây 10; hồi chiêu lần lượt 0,4 và 1,2 giây. Lướt thường kéo dài 0,2 giây, bất tử trong lúc lướt, hồi chiêu 2 giây. Mỗi lần nhận sát thương có 0,6 giây bất tử để tránh bị nhiều người khóa combo; đòn bị chặn bởi bất tử không làm rơi tiền hay ngắt ATM. Mỗi lượt vung hoặc lần dùng chiêu chỉ đánh trúng mỗi mục tiêu một lần. Không có choáng nối tiếp; sát thương vượt máu còn lại không tạo thêm tiền rơi.

| Player | Chiêu giữ lại | Điều chỉnh cho loạn đấu |
| --- | --- | --- |
| P1 | **Em sửa có một dòng** | Giữ tối đa 1 giây, sát thương tăng tuyến tính từ 15 lên 30. Đánh một đối thủ gần nhất trước mặt; tích lực đi chậm, lướt hủy chiêu. Hồi chiêu 6 giây từ lúc thả/hủy. |
| P2 | **Không thuộc scope em** | Khiên trước mặt 0,6 giây, phản tối đa 3 xung code hoặc con dấu P3 về hướng kẻ bắn tại thời điểm phản. Đạn phản gây 10 sát thương, thuộc về P2 và mất hiệu ứng đánh dấu; không phản lần nữa. Không chặn bàn phím, P1 bổ, vùng hotfix hoặc cú lao P5. Hồi chiêu 8 giây. |
| P3 | **Theo biên bản đã chốt** | Con dấu gây 10 sát thương và đánh dấu một đối thủ trong 4 giây. Chỉ đòn bàn phím/xung code của chính P3 được tăng 15% sát thương lên người đó; không tăng cho người khác. Mỗi P3 chỉ giữ một dấu, dấu không cộng dồn. Hồi chiêu 10 giây. |
| P4 | **Máy em vẫn chạy** | Vùng hotfix tồn tại 3 giây. Lúc đặt, gây 10 sát thương một lần lên từng đối thủ trong vùng; chỉ hồi cho chính P4 tối đa 20 máu trong mỗi lần dùng, không vượt 100. Không hồi máu đối thủ. Hồi chiêu 14 giây. |
| P5 | **Em về đúng giờ** | Lao ngang tối đa 3 tile, gây 15 sát thương một lần lên mỗi đối thủ trên đường; bất tử 0,15 giây đầu, dừng ở tường. Hồi chiêu riêng 7 giây và dùng chung cổng hồi với lướt: dùng chiêu khóa lướt 7 giây; dùng lướt thường khóa chiêu riêng 2 giây. |

Sát thương có phần lẻ được giữ nguyên khi tính máu, UI làm tròn để hiển thị. Chiêu riêng hụt vẫn hồi chiêu. Dấu P3 không tăng tiền rơi. Các giới hạn không đánh đồng đội của co-op được thay bằng luật PvP trong phase này; “Có biên bản” của chế độ tay đôi không tự thêm vào bộ chiêu loạn đấu.

## Nhặt tiền, cướp tiền và gửi ATM

- **Nguồn tiền:** đơn vị tiền nguyên, gọi là xu thưởng. Với N người tham gia lúc bắt đầu phase, két giữa sảnh phun tổng cộng 50 × N xu mỗi 15 giây, từ giây 0 đến 135. Tiền chia thành các bó nhỏ rải quanh két để không một người nhặt toàn bộ bằng một lần chạm. N giữ cố định khi có người rớt mạng.
- **Tiền rơi khi đánh:** đòn gây sát thương nhưng chưa hạ đối thủ làm rơi 10% ví hiện tại, làm tròn lên và không vượt số tiền đang có. Đòn kết liễu chỉ làm rơi 50% ví trước đòn, làm tròn lên; không cộng thêm khoản 10%. Phần còn lại giữ trong ví sau hồi sinh. Tiền đã gửi không bao giờ bị cướp.
- **Nhặt lại:** tiền rơi nằm tại điểm trúng đòn, ai cũng nhặt được; chủ cũ phải chờ 1 giây mới nhặt lại phần tiền đó. Mỗi bó chỉ được nhận một lần; nếu chạm cùng lúc, xử lý theo thứ tự xác nhận của server. Bó tiền không tự biến mất trước khi hết phase; tiền lọt ra ngoài vùng chơi được đưa về gần két.
- **ATM:** giữ tương tác và đứng yên trong tầm 1 tile suốt 3 giây để gửi toàn bộ ví. Nhận sát thương, di chuyển, ra đòn, dùng chiêu/lướt hoặc thả tương tác thì hủy, không mất tiền. Không được bắt đầu gửi trong lúc đang tích lực, giữ khiên hoặc lao; vùng hotfix đã đặt vẫn tồn tại. Mỗi máy phục vụ một người, người khác phải tự bắt đầu lại khi máy trống, không có hàng đợi tự động. Nếu cùng yêu cầu, server nhận yêu cầu hợp lệ đầu tiên.
- **Khi đang gửi:** không nhặt thêm tiền hoặc tiêu tiền tặng quà; UI hiện số tiền đang gửi và thanh tiến độ để cả người gửi lẫn đối thủ đọc được. Gửi thành công mới chuyển nguyên số tiền từ ví sang số dư; giao dịch hủy không chuyển một phần. Một nút tương tác ưu tiên ATM khi trong tầm, còn tặng quà cần chọn rõ hành động riêng để tránh tiêu nhầm tiền.
- **Cao trào 30 giây cuối:** ở giây 150, HR hô “Giải ngân trước khi khóa sổ!”, két bung một đợt 150 × N xu. Hai ATM vẫn hoạt động, không tăng sát thương hay thêm luật đột ngột. Mười giây cuối có đếm ngược rõ trên UI.
- **Khóa sổ:** giao dịch hoàn thành tại hoặc trước mốc 180 giây được tính. Sau mốc đó đóng ATM, dừng sát thương, nhặt tiền và tương tác rồi hiện kết quả; giao dịch chưa xong bị hủy. UI luôn phân biệt VÍ / ĐÃ GỬI và hiển thị số dư của mọi người.

## Bị hạ, hồi sinh và rớt mạng

- **Bị hạ:** không có cứu đồng đội trong phase loạn đấu. Hồi sinh sau 3 giây với 100 máu, ở điểm có khoảng cách tới đối thủ gần nhất lớn nhất trong các điểm hồi sinh hợp lệ. Xóa dấu/buff và hiệu ứng do nhân vật tạo ra, gồm hotfix, đạn và vệt lao; hồi chiêu tiếp tục chạy trong lúc chờ, không đặt về đầy khi hồi sinh.
- **Bảo vệ hồi sinh:** bất tử 2 giây, có viền nhấp nháy và không nhặt/gửi tiền trong thời gian đó. Chủ động tấn công hoặc dùng chiêu riêng kết thúc bảo vệ ngay trước khi thực hiện hành động; đi, nhảy hoặc lướt thường không hủy sớm. Chết không mất lượt tham gia và không bị loại khỏi phase.
- **Rớt mạng:** hủy gửi tiền và thao tác đang giữ; nhân vật đứng yên, vẫn có thể bị đánh trong tối đa 5 giây. Nếu bị hạ thì xử lý mất 50% ví một lần như bình thường, không tự hồi sinh khi còn offline. Sau 5 giây mất kết nối, toàn bộ phần ví còn lại rơi tại vị trí nhân vật rồi nhân vật rời sân; số dư đã gửi được giữ. Không phạt hay thả cùng khoản tiền hai lần.
- **Vào lại:** nối lại trong 5 giây giữ đúng máu/vị trí hiện có, không cấp bất tử mới; nếu đã bị hạ thì tiếp tục thời gian chờ hồi sinh. Vào lại sau khi đã rời sân dùng quy trình hồi sinh, ví bằng 0 và giữ số dư ATM. Các timer hồi chiêu không được reset để lấy lợi thế.
- **Số người:** nếu chỉ một người đủ điều kiện tham gia từ đầu (kể cả do nhánh tái cấu trúc), chơi bản gom thưởng solo với cùng luật két/ATM và thời gian; không thêm bot đối thủ. Nếu trận đã bắt đầu với từ hai người, một người rớt mạng không chuyển thành solo và không đổi luật thắng. Người bị hy sinh ở phần kết có thể xem trận, không điều khiển nhân vật trong sảnh.
- **Tạm dừng:** còn ít nhất một người kết nối thì đồng hồ tiếp tục. Khi tất cả cùng mất kết nối, dừng đồng hồ trận, timer rớt mạng, đạn và hiệu ứng; có người vào lại thì tiếp tục. Việc mất kết nối tự nó không trao chiến thắng sớm.

## Tiền + gái

Các cô nàng NPC là khách mời có tên, cá tính và gu riêng: người thích tay chơi nhiều tiền, người thích người biết giúp đỡ, người chỉ muốn ăn tiệc yên ổn. Người chơi có thể tặng quà bằng tiền chưa gửi, mời nhảy hoặc làm yêu cầu nhỏ giữa trận để tăng thiện cảm. NPC tự nhận lời hoặc từ chối; có nhiều tiền không bảo đảm có hẹn, và buổi hẹn không phải điều kiện thắng PvP.

Đánh nhau cạnh bàn của một NPC có thể làm mất thiện cảm với cô ấy. Cảnh hài chính: hai ông đang tranh công lấy lòng thì bị cả hai từ chối vì làm đổ ly nước. Các quyết định và phản ứng đều có tín hiệu ngắn, rõ trên màn hình để người chơi hiểu mình vừa được cộng hay trừ điểm.

## Đoạn kết mở rộng

Kết quả hiện thành hai mục riêng: **Ai ôm tiền** và **Ai có hẹn**. Một người có thể thắng cả hai, chỉ được một, hoặc trắng tay. Người thắng tiền nâng vali lên thì nhận thông báo: “Khoản thưởng đang chờ Approve.” Người có hẹn vừa ra tới cửa thì điện thoại rung: “Em rảnh tối nay không? Tiện xem giúp chị cái lỗi production.”

Khung hình cuối: cả nhóm ngồi trước cửa công ty, người ôm vali khóa, người ôm laptop. Câu thoại: “Mai vẫn party nhé?” — “Ừ. Nhưng bật chia loot từ đầu.”
