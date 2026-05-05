1. Đề bài Project

Tên dự án: Fashion High-Scale Engine

Mô tả: Xây dựng hệ thống Backend cho nền tảng thương mại điện tử chuyên về thời trang. Hệ thống phải đảm bảo vận hành mượt mà các tính năng mua sắm thông thường và có khả năng chịu tải "khổng lồ" (đột biến lên đến 1 triệu request) trong các chiến dịch Flash Sale mà không gây sai lệch dữ liệu tài chính hay tồn kho.

2. Chi tiết các nhóm chức năng & Công nghệ

Nhóm 1: Quản lý thương mại điện tử cơ bản (Standard eCommerce)

Chức năngCông nghệCách hoạt độngUser & AuthGolang, MySQL, JWTĐăng ký, đăng nhập, quản lý profile.CatalogGolang, MySQLHiển thị danh sách quần áo, lọc theo size, màu, chất liệu.CartGolang, MySQLLưu trữ giỏ hàng tạm thời của người dùng.CheckoutGolang, MySQLQuy trình thanh toán, chọn địa chỉ và phương thức giao hàng.InvoicingGolang, MySQLXuất hóa đơn (PDF/Email) sau khi thanh toán thành công.

Nhóm 2: Hệ thống High-Concurrency Flash Sale (Trọng tâm)

Dưới đây là nơi bạn áp dụng cơ chế Bất đồng bộ (Asynchronous) để xử lý 1 triệu request:

Chặn tải & Kiểm soát (Rate Limiting):

Công nghệ: Golang Middleware.

Hoạt động: Dùng thuật toán Token Bucket hoặc Leaky Bucket để giới hạn số lượng request đi vào hệ thống, tránh bị DDOS hoặc sập server ngay lập tức.

Trừ tồn kho siêu tốc (Real-time Inventory):

Công nghệ: Aerospike.

Hoạt động: Tồn kho của các mẫu quần áo "Hot" được nạp sẵn vào RAM của Aerospike. Khi 1 triệu người nhấn mua, hệ thống trừ kho trên Aerospike bằng lệnh Operate (Atomic). Phản hồi cực nhanh (dưới 1ms).

Hàng đợi đơn hàng (Order Queuing):

Đây là điểm dùng bất đồng bộ: Thay vì bắt người dùng đợi ghi vào MySQL (mất 200-500ms), hệ thống chỉ đẩy thông tin đơn hàng vào Kafka.

Hoạt động: Trả kết quả "Đang xử lý" ngay lập tức cho khách hàng.

Xử lý đơn hàng hậu kỳ (Worker Service):

Công nghệ: Golang, MySQL.

Hoạt động: Các Consumer Group (như đã bàn ở trên) sẽ lấy đơn hàng từ Kafka ra để ghi vào MySQL, cập nhật hóa đơn và giảm tồn kho vật lý.

3. Các bước triển khai (Step-by-step)

Giai đoạn 1: Thiết lập nền móng (Infrastucture)

Dùng Docker dựng môi trường: MySQL, Kafka, Aerospike.

Thiết kế Schema MySQL (Users, Products, Orders, Invoices).

Giai đoạn 2: Phát triển Core API

Viết CRUD sản phẩm quần áo bằng Golang.

Xây dựng luồng mua hàng truyền thống (Order trực tiếp vào MySQL).

Giai đoạn 3: Phát triển Flash Sale Engine

Viết logic "Pre-warm": Lấy tồn kho từ MySQL đẩy vào Aerospike.

Viết API /flash-buy: Check Aerospike -> Push Kafka.

Giai đoạn 4: Xử lý bất đồng bộ (Consumers)

Viết Consumer Group 1: Lưu đơn hàng và xuất hóa đơn vào MySQL.

Viết Consumer Group 2: Giả lập gửi mail/thông báo thành công.

Giai đoạn 5: Tối ưu và Test

4. Tiêu chí đạt được (Success Metrics)

Tính toàn vẹn (Data Integrity): Chạy test 10.000 đơn hàng đồng thời, sau khi kết thúc, số lượng tồn kho trong MySQL + Aerospike phải khớp tuyệt đối. Không có hóa đơn nào bị mất.

Hiệu năng (Performance): 95% request mua hàng phải có thời gian phản hồi (Latency) dưới 100ms.

Khả năng chịu tải: Hệ thống xử lý được ít nhất 50.000 - 100.000 Request Per Second (RPS) trên máy local (tùy cấu hình máy).

5. Quy trình và Công nghệ Test

Để test mức độ "khủng" của hệ thống này, Postman là không đủ. Bạn cần dùng k6 (được viết bằng Go):

Công nghệ dùng thêm: k6 (Load Testing tool).

Cách test:

Test chức năng: Dùng Postman để đảm bảo logic mua hàng, trừ tiền, xuất hóa đơn đúng.

Test chịu tải (Stress Test): Dùng k6 viết script giả lập 1.000 người dùng tăng dần lên 10.000 rồi 50.000 người dùng liên tục nhấn mua.

Test giới hạn (Breakpoint Test): Tăng lượng request đến khi hệ thống bắt đầu báo lỗi để tìm ra điểm yếu nhất (thường là MySQL hoặc băng thông Kafka).

Mẹo nhỏ cho bạn: Trong CV, hãy nhấn mạnh việc bạn sử dụng Aerospike làm Primary High-speed Storage thay vì Redis. Điều này sẽ khiến bạn nổi bật vì Aerospike thường được dùng trong các hệ thống AdTech hoặc Fintech xử lý dữ liệu cực lớn, cho thấy bạn có tư duy chọn công nghệ rất "xịn".

Bạn có muốn tôi viết mẫu script k6 để bạn thử nghiệm tải ngay sau khi hoàn thành phần API mua hàng không?