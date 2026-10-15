# Hầm Tri Thức — Song Thần Giữ Bài FE & BE

Trạng thái: chốt thiết kế cho bản chơi thử ngày 2026-10-15; chưa triển khai. Các số liệu là cấu hình khởi đầu, cần cân bằng bằng chơi thử. Tài liệu này là nguồn thiết kế chi tiết của màn Hầm Tri Thức; không thay đổi luật màn Sprint.

Màn 4 · cánh NA. Tổng quan: [game-scenario.md](../game-scenario.md).

Cốt truyện: trở về cánh NA, cả đội phải lấy lại kiến thức bàn giao bị giữ kín để mở đường đi tiếp.

Tóm tắt: Hai boss đứng ở hai cụm bục, khóa cầu nối bằng code bí truyền. Nhặt trang tài liệu để mở đường; đánh công tắc hai bên để vô hiệu khiên. Solo cho phép kích hoạt tuần tự trong cửa sổ dài hơn. Thoại: “Không hiểu thì cứ hỏi.” — trạng thái: Đang bận.

## Mục tiêu và phạm vi

Màn co-op 1–5 người: học đường đi, mở cầu bằng tài liệu, dùng công tắc mở khiên rồi hạ hai boss. Độ khó đến từ thay đổi vị trí và phối hợp, không tăng sát thương mỗi đòn. Mọi slot đều chơi solo được; không yêu cầu P2 phản đạn hoặc P4 hồi máu để qua màn.

Bản đầu dùng camera chung 1280 × 720, map 40 × 22 tile, tile 32 px. Giữ nguyên điều khiển, 5 tim, cứu đồng đội, bất tử sau trúng đòn và năm bộ chiêu của Sprint. Chưa thêm biến thể hard cho màn này; cờ hard chỉ áp dụng Sprint cho đến khi có thiết kế riêng.

## Bản đồ và mở cầu

Tọa độ dưới đây dùng tile từ 0; điểm đứng là vị trí chân. Sàn chính có mặt trên tại y=21, liên tục từ trái sang phải, không có hố chết. Tường hai biên. FE đứng tại (8,21), BE tại (32,21), hitbox mỗi boss 70 × 110 px; không gây sát thương tiếp xúc và không chặn người đi qua. Boss đứng tại chỗ, quay về mục tiêu khi bắt đầu báo đòn.

- Bục trái: x=3..10 tại y=17 và x=9..15 tại y=13; bục phải đối xứng x=29..36 tại y=17 và x=24..30 tại y=13. Tất cả là bục một chiều. Xác minh các bước nhảy bằng vật lý thực tế trước khi duyệt map.
- Cầu một chiều x=16..23 tại y=13 ban đầu chưa có va chạm. Khóa là hologram trên cầu, không phải tường ngăn toàn chiều cao. Sàn dưới cầu luôn là đường vòng để không khóa khả năng tiếp cận hai bên.
- Ba tài liệu đặt tại (6,17), (13,13), (27,13). Chạm để nhặt; tiến độ 0/3 dùng chung đội, mỗi trang chỉ tính một lần, không rơi khi gục/rớt mạng. Không cần giữ E.
- Đủ 3 trang: cầu mở vĩnh viễn đến hết lượt, bật va chạm một chiều và animation giải mã. Không đóng cầu khi boss đổi pha, chết hoặc có người rớt mạng. Người đang dưới cầu không bị đẩy hoặc kẹt khi cầu xuất hiện.
- Hai công tắc trên sàn: trái tại (12,21), điều khiển khiên BE; phải tại (28,21), điều khiển khiên FE. Có dây/biểu tượng nối rõ tới boss tương ứng.
- Ba cà phê tại (5,17), (20,21), (35,17), hồi 1 tim, hồi lại sau 30 giây như Sprint. Spawn P1–P5 tại x=17,18.5,20,21.5,23, y=21.

Lúc bắt đầu có hướng dẫn 3 giây, chưa trừ deadline. Sau đó deadline 10 phút chạy ngay, người chơi tìm tài liệu trong khi boss ngủ và có khiên. Nhặt đủ trang thì mở cầu, kích hoạt công tắc, báo thức tỉnh 2 giây rồi bắt đầu vòng đánh. Boss không nhận sát thương trước khi thức tỉnh. Không có checkpoint trong bản này.

## Công tắc và khiên

Giữ E trong tầm 48 px, đứng trên mặt sàn, liên tục 1 giây để kích hoạt. Di chuyển, nhảy, nhận sát thương, ra đòn, dùng chiêu/lướt hoặc thả E thì hủy tiến độ. Mỗi công tắc phục vụ một người; yêu cầu hợp lệ đầu tiên theo thứ tự xử lý server được nhận. Giữ E cạnh đồng đội gục ưu tiên cứu; chốt mục tiêu tương tác khi bắt đầu giữ, phải thả E mới chọn mục tiêu khác. Hiện prompt trước thao tác và thanh tiến độ khi giữ.

- Công tắc trái mở khiên BE; công tắc phải mở khiên FE. Hai công tắc hoạt động độc lập, không yêu cầu kích hoạt đồng thời.
- Khiên mở 12 giây với 2–5 slot; 18 giây với 1 slot. Cấu hình này cố định từ đầu lượt, không đổi theo số người online. Không gia hạn bằng giữ E; công tắc khóa trong lúc khiên đang mở và 2 giây sau khi đóng.
- Khi có từ hai slot nhưng chỉ còn một người kết nối và chưa gục, người đó giữ công tắc trong 0,5 giây; không đổi máu boss hoặc thời lượng mở khiên. Quy tắc xét tại lúc bắt đầu thao tác, không rút ngắn một thao tác đang giữ.
- Khiên chặn mọi sát thương và dấu P3, kể cả phản đạn; đòn vẫn tiêu hao bình thường. Có VFX và âm chặn. Khiên không ảnh hưởng lịch tấn công, không nhân sát thương khi mở.
- Khi boss bị hạ, công tắc điều khiển boss đó tắt và không nhận tương tác nữa.

## Năm chiêu boss

Mỗi đòn gây mất 1 tim; áp dụng bất tử sau trúng đòn 0,6 giây như Sprint. Mục tiêu là người đang kết nối, chưa gục, gần boss thi triển nhất; hòa chọn ID nhỏ hơn. Khóa mục tiêu/vùng tại lúc bắt đầu báo đòn, không đuổi theo trong thời gian báo. Nếu không có mục tiêu hợp lệ, chờ và chọn lại; không nhắm người gục/offline.

| Chiêu | Báo trước và hiệu lực | Cách né / đặc tính |
| --- | --- | --- |
| FE — Máy anh hiển thị đúng | Chọn bục cố định có tâm gần mục tiêu nhất, hòa chọn ID bục nhỏ hơn; viền sọc báo 1 giây rồi điện phủ toàn bục 1,2 giây, cao 16 px trên mặt bục. Nếu mục tiêu ở sàn, vẫn chọn bục gần nhất. | Rời bục hoặc nhảy lên khỏi điện; sàn chính luôn an toàn với đòn này. Mỗi người tối đa một lần trúng mỗi lượt. Cầu không thuộc tập bục có thể bị chọn. |
| FE — Responsive rồi mà | Khóa x mục tiêu, đặt cột rộng 2 tile, kẹp gọn trong map. Báo 1 giây rồi quét từ trần đến sàn 0,35 giây. | Đi ngang khỏi cột; không đẩy hoặc xóa bục. Mỗi người tối đa một lần trúng mỗi lượt. |
| BE — API trả 200 rồi | Báo 0,9 giây; bắn 3 viên cách nhau 0,2 giây từ ngực tới điểm giữa hitbox mục tiêu đã khóa, tốc độ 360 px/s; sống tối đa 2 giây hoặc đến khi chạm tường solid. | Né sang cao độ khác; đạn xuyên bục một chiều. P2 phản được. |
| BE — Retry vô hạn | Báo 1 giây; tạo 3 vòng tròn tâm tại x mục tiêu và x ± 3 tile, y=21 tile, bán kính 24 px; kẹp tâm trong [1,39] tile, gộp vòng trùng. Vòng nổ cao 48 px trên sàn trong 0,3 giây. | Nhảy hoặc lên bục; không phản được. Một lượt gây tối đa một lần trúng mỗi người, kể cả vùng giao nhau. |
| Kết hợp — Lỗi bên kia | Cả hai boss cùng báo, FE khóa cột rộng 2 tile vào mục tiêu trong 1,2 giây. BE gửi một gói dữ liệu rơi dọc cột với tốc độ 400 px/s, từ y=1 tile tới sàn; không va chạm bục, hủy khi tới sàn. | Rời cột; chỉ một cột với mọi số người. Không phản được. Không có điện bục hoặc chiêu khác trong lượt kết hợp. |

Đạn BE khi phản: bay 480 px/s về tâm hitbox boss đã bắn tại thời điểm phản, gây 3 sát thương, sống tối đa 2 giây; không phản lần hai. Không tự tìm boss khác. Boss nguồn đã chết thì đạn vẫn bay theo hướng đã khóa và hủy khi hết tuổi/chạm tường; có thể va chạm boss sống đầu tiên trên đường. Va chạm khiên tiêu hủy đạn.

## Lịch đánh, pha 2 và một boss gục

Vòng cơ bản: FE điện bục → BE loạt đạn → FE cột quét → BE retry → kết hợp. Mỗi lượt chỉ bắt đầu sau khi toàn bộ nguy hiểm lượt trước biến mất và nghỉ 0,8 giây. Hai boss không tự chạy AI độc lập; một bộ điều phối quyết định lượt. Đòn kết hợp là một lượt chung, có hai hoạt ảnh nhưng một nguồn sát thương. Cấu trúc này cho năm mẫu đòn mà không tạo tổ hợp phủ kín sân ngoài kiểm soát.

Pha 2 kích hoạt một lần khi tổng máu còn lại của cả hai ≤50% tổng máu tối đa ban đầu. Báo chuyển pha 1 giây sau khi nguy hiểm hiện tại kết thúc; sau đó thời gian báo của mọi đòn ×0,85, nghỉ giữa lượt còn 0,5 giây, tốc độ đạn/rơi ×1,15. Giữ nguyên kích thước vùng, số đạn, thời gian mở khiên và sát thương. Không đổi thông số đòn đang diễn ra.

Một boss gục: hủy phần đòn đang báo hoặc nguy hiểm đang tồn tại do boss đó sở hữu; nếu là lượt kết hợp thì hủy cả lượt. Xóa dấu P3 của boss đó. Boss còn lại không hồi máu và không hồi sinh đồng đội. Khiên boss sống mở vĩnh viễn, công tắc còn lại tắt. Bỏ lượt boss chết và đòn kết hợp; lặp hai chiêu của boss sống. Áp dụng pha 2 khi đạt ngưỡng tổng máu, không thêm hệ số cuồng nộ nữa. Như vậy dồn hạ một boss là chiến thuật hợp lệ và trận bớt kéo dài cuối màn.

## Người chơi và sát thương nhiều mục tiêu

Giữ số liệu chiêu đang chạy của Sprint. Chỉ mở rộng chọn mục tiêu cho màn này:

- Đòn bàn phím, P1 bổ và P5 lao có thể trúng mỗi boss một lần mỗi lần dùng. Dùng tập ID mục tiêu đã trúng, thay cho một cờ hit chung; va chạm khiên cũng tính là đã thử trúng boss đó cho lần dùng này.
- Xung code và dấu P3 trúng boss đầu tiên trên đường rồi biến mất, kể cả khi bị khiên chặn; hòa chọn ID boss. Dấu P3 tồn tại riêng từng boss, 4 giây, +15% cho các loại đòn giống Sprint; không cộng dồn trên cùng boss.
- P4 gây 3 sát thương lúc đặt lên từng boss trong vùng một lần; hồi tối đa 1 tim cho mỗi người mỗi vùng, không cứu người gục. P2 phản tối đa 3 viên mỗi khiên, chỉ áp dụng chiêu loạt đạn BE.
- P5 đi xuyên thân boss như hiện tại; không xuyên tường, không tự mở cầu. Không bắt buộc bất kỳ chiêu riêng nào cho tài liệu hoặc công tắc.
- Đòn đã khóa sát thương không bị hủy bởi trúng đánh, đóng dấu, mở khiên hoặc bị phản đạn. Boss chỉ hủy đòn khi gục hoặc lượt kết thúc.

## Thông số trận và vòng đời

N là tổng slot lúc bắt đầu, gồm người offline, giữ cố định qua lần thua. Máu **mỗi boss = 80 + 40 × (N−1)**, tổng 160–480. Người chơi 5 tim, cứu 3 giây được 3 tim, cà phê và hồi chiêu giữ như Sprint. Không tự thêm sát thương theo N.

Thứ tự mỗi tick: xử lý thao tác người chơi và sát thương lên boss → ghi nhận boss gục, hủy đòn tương ứng → cập nhật nguy hiểm còn lại và sát thương người chơi → cứu/nhặt đồ → xét kết quả. Hai boss cùng gục thì thắng, kể cả deadline vừa về 0 trong tick đó; nếu chưa thắng và hết giờ hoặc tất cả người online gục thì thua. Giữ luật người offline miễn sát thương và không tính vào all-down như Sprint.

Thua hiện lý do 4 giây rồi reset toàn bộ lượt: cả hai boss, tài liệu, cầu, công tắc, cà phê, người chơi, đạn và deadline. Thắng hiện 6 giây rồi sang tiệc nếu được cấu hình, nếu không về lobby. Reconnect khôi phục toàn bộ trạng thái hiện tại, không đặt lại timer hoặc phát lại hiệu ứng cũ. Tất cả rớt mạng thì dừng mọi timer gameplay.

Bản đầu cho chủ phòng chọn Sprint hoặc Hầm Tri Thức trong lobby; mặc định Sprint. Server gửi ID/URL màn đã chọn khi start và reconnect. Chưa tự nối Sprint → Hầm Tri Thức; nhánh tiệc chỉ nối sau Sprint (extension màn 1); thắng Hầm Tri Thức thì về phòng chờ. Thay đổi lựa chọn chỉ hợp lệ trong lobby.

## Hiển thị, asset và hợp đồng kỹ thuật

- HUD: hai thanh máu FE/BE, biểu tượng khiên và giây còn mở; tài liệu 0/3; deadline; prompt công tắc có tên boss đích. Lúc chỉ còn một boss gọn thành một thanh. Dùng cả chữ, hình và màu để đọc nguy hiểm khi tắt tiếng.
- Asset mới: hai bộ idle, telegraph cho mỗi chiêu, cast, hurt, shield, defeat; animation kết hợp dùng trạng thái của hai boss. VFX điện bục, cột quét, đạn dữ liệu/phản, retry, gói dữ liệu rơi, cầu giải mã, công tắc tắt/giữ/sáng. Tái sử dụng tài liệu, cà phê, người chơi, hit spark và tile văn phòng cho prototype; bổ sung nền server/tủ tài liệu khi hoàn thiện hình ảnh.
- Âm mới: báo riêng cho năm chiêu, phát chiêu, khiên chặn/mở/đóng, tài liệu, mở cầu, công tắc thành công, mỗi boss gục và nhạc Hầm Tri Thức. Dùng ID sự kiện/counter để không phát lặp theo snapshot, giữ ưu tiên âm báo đòn như Sprint.
- Server là nguồn sự thật cho boss, vật phẩm, tương tác, va chạm cầu và kết quả. Snapshot màn này có levelId, stage (collect/combat/result), bosses[] với ID/kind/HP/state/shieldMs/markMs/counters, document IDs đã lấy, bridgeOpen, switches[] với targetBossId/ownerId/progress/cooldown, hazards[] có ownerBossId và vùng đã khóa.
- Tiled thêm boss FE/BE có ID/kind, document, switch có targetBossId, bridge và platform có ID ổn định. Parser phải giữ properties và nhiều boss; không ghi đè boss cuối vào một trường Boss như hiện tại. Client/server dùng chung ID hình học, không suy ra va chạm cầu chỉ từ animation.
- Giữ nhánh Sprint hoạt động trong quá trình chuyển đổi: định tuyến luật theo levelId; renderer, damage và snapshot hỗ trợ nhiều boss bằng ID. Chuẩn hóa target/hit tracking trước khi thêm AI mới.

## Điều kiện nghiệm thu

1. Solo với từng slot đều nhặt đủ tài liệu, mở cả hai khiên và hạ boss được; không đòi đứng đồng thời hai phía. Chơi thử xác nhận đủ thời gian đi từ công tắc tới boss và tung ít nhất một đòn P1 đầy lực trước khi đóng khiên.
2. Kiểm tra đường nhảy tới ba trang, hai công tắc và hai boss; mở cầu không kẹt người; mọi người đều tới cứu đồng đội được trước và sau khi mở cầu.
3. Kiểm tra từng chiêu có đường né bằng di chuyển/nhảy thường, không cần chiêu riêng; pha 2 vẫn có cửa sổ né. Không chiêu nguy hiểm nào lọt qua bộ điều phối để chồng ngoài thiết kế.
4. Test ý nghĩa trên server: khiên chặn sát thương/dấu; multi-hit đúng một lần/boss; phản đúng nguồn; bật công tắc đồng thời; boss gục trong lượt kết hợp; thứ tự thắng/timeout; reset và pause không để sót trạng thái.
5. Chơi 2 và 5 client; rớt/vào lại lúc giữ E, khiên mở, cầu vừa mở và boss vừa gục. UI/va chạm phải khớp server, âm không lặp. Có người rớt không gây khóa tiến trình; solo sống sót vẫn mở được công tắc.
6. Chạy lại kiểm tra Sprint và nhánh tiệc hiện có sau khi đổi protocol/renderer; bản build production chứa map và manifest mới.

Thứ tự thực hiện: nhiều boss và chọn màn → map/thu tài liệu/cầu → công tắc/khiên → năm chiêu và bộ điều phối → HUD/asset/âm → kiểm thử và cân bằng. Bản đầu dùng hình tạm để kiểm chứng vòng chơi trước khi hoàn thiện art.
