import http from 'k6/http';
import { check, sleep } from 'k6';

export const options = {
  scenarios: {
    stress_test: {
      executor: 'ramping-arrival-rate',
      startRate: 100,
      timeUnit: '1s',
      preAllocatedVUs: 500, // Số lượng nhân viên ảo sẵn sàng
      maxVUs: 2000,        // Tối đa 2000 nhân viên ảo để xử lý Bcrypt chậm
      stages: [
        { duration: '2m', target: 500 },   // Trong 2 phút đầu nâng tốc độ lên 500 req/s
        { duration: '11m', target: 1200 }, // Duy trì và nâng lên 1200 req/s (để đạt ~1M request)
        { duration: '2m', target: 0 },
      ],
    },
  },
};
// Giúp k6 không bị ngắt quãng giữa chừng, sau 1 giờ K6 sẽ dừng không bắn request nữa
export default function () {
  const url = 'http://localhost:3004/Register'; // Cổng bạn đã cấu hình trong main.go
  
  // Tạo dữ liệu ngẫu nhiên để không bị trùng lặp khi check Aerospike/MySQL
  const id = Math.floor(Math.random() * 100000);
  const payload = JSON.stringify({
    username: `user_${id}`,
    full_name: `User Test ${id}`,
    email: `test_${id}@example.com`,
    phone: "0912345678",
    address: "Hanoi, Vietnam",
    password: "Password@123",
    password_confirm: "Password@123"
  });

  const params = {
    headers: {
      'Content-Type': 'application/json',
    },
  };

  const res = http.post(url, payload, params);

  // Kiểm tra xem API có trả về 200 OK không
  check(res, {
    'is status 200': (r) => r.status === 200,
  });

  // Nghỉ 0.1 giây giữa các request để tránh làm treo máy cục bộ quá nhanh
  // sleep(0.1);
}