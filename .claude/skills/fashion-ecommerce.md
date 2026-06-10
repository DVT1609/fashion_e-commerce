# Fashion High-Scale Engine — CLAUDE.md

Tài liệu hướng dẫn đầy đủ để Claude Code hiểu kiến trúc, luồng xử lý, pattern và convention của project này trước khi thực hiện bất kỳ thay đổi nào.

---

## 1. Mục tiêu & Bối cảnh

Project là một **backend thương mại điện tử thời trang** được thiết kế để chịu tải cực cao — mục tiêu chính là xử lý **Flash Sale** với 1 triệu request đồng thời.

**Hai nhóm tính năng:**

| Nhóm | Mô tả | Độ ưu tiên |
|------|-------|-----------|
| **Group 1 — Standard eCommerce** | User & Auth, Catalog, Cart, Checkout, Invoicing | Nền tảng |
| **Group 2 — High-Concurrency Flash Sale** | Rate limiting, Aerospike inventory, Kafka queue, async worker, invoice | Core focus |

**Success metrics:**
- 95% request phải response trong **< 100ms**
- Xử lý **50,000 – 100,000 RPS** trên máy local
- **Toàn vẹn dữ liệu tuyệt đối** với 10,000 đơn hàng đồng thời

---

## 2. Technology Stack

| Layer | Công nghệ | Phiên bản | Vai trò |
|-------|-----------|-----------|---------|
| **Web framework** | [Fiber](https://gofiber.io/) v3 | v3 | HTTP server dựa trên `fasthttp`, zero-allocation |
| **Database** | MySQL | 9.x | Lưu trữ persistent — users, orders, products |
| **Cache / Inventory** | Aerospike | v7 (Go client) | RAM-based storage, TTL 300s cho đăng ký tạm thời |
| **Message queue** | Kafka | latest | Async pipeline — tách API và Worker |
| **ORM** | GORM | v2 | MySQL abstraction, AutoMigrate, Raw query |
| **Validation** | go-playground/validator | v10 | Struct tag validation + custom error messages |
| **Password hashing** | argon2id (`golang.org/x/crypto`) | — | Heavy hashing chỉ trong Worker |
| **Light hashing** | SHA-256 (`crypto/sha256`) | stdlib | Bảo vệ đường truyền tới Kafka, không dùng cho lưu trữ |
| **Containerization** | Docker + Docker Compose | — | Môi trường dev & deploy |
| **Load testing** | k6 | — | Kiểm thử hiệu năng |
| **Live reload** | Air | — | Dev workflow, cấu hình `.air.toml` |
| **Go version** | Go | 1.25.0 | |

---

## 3. Kiến trúc tổng thể

### 3.1 Layered Architecture

```
[HTTP Request]
     ↓
[Handler]      — Bind, validate input, map sang Request struct
     ↓
[Service]      — Business logic, orchestration
     ↓
[Repository]   — Data access (MySQL / Aerospike / Kafka Producer)
     ↓
[Database/Queue] — MySQL, Aerospike, Kafka
```

**Worker** là một pipeline song song, không thuộc request path:

```
[Kafka Topic] → [Worker goroutines × 4] → [Batch buffer] → [MySQL Bulk Insert]
```

### 3.2 Dependency Wiring (main.go)

Toàn bộ dependency injection được wire **thủ công** trong `cmd/api/main.go` — không dùng DI framework:

```
ConnectMysql() → db
ConnectAerospike() → aerospikeClient
kafka.CreateProducer() → writer
kafka.CreateConsumer() → kafkaReader

UserAerospikeRepository{client, ctxBase, ctxWrite}
UserMysqlRepository{db}
UserKafkaProducerRepository{writer}

NewUserService(kafkaRepo, mysqlRepo, aerospikeRepo)
NewUserHandler(userService, validate)

NewRegisterWorker(kafkaReader, db, aerospikeClient, ctxWrite)
→ 4 goroutines: go registerWorker.WorkerRegister(ctx)

app.Listen(":3004")
```

---

## 4. Cấu trúc thư mục

```
Fashion E-Commerce/
├── cmd/api/
│   └── main.go                          # Entry point, wire toàn bộ dependency
├── config/
│   ├── config_mysql.go                  # Đọc MySQL config từ env
│   └── config_aerospike.go              # Đọc Aerospike config từ env
├── internal/
│   ├── database/
│   │   ├── mysql/connect_mysql.go       # Kết nối MySQL + pool config
│   │   └── aerospike/connect_aerospike.go
│   ├── handler/
│   │   └── user_handler.go              # HTTP handler cho /Register
│   ├── models/
│   │   ├── user.go                      # User, InputRegister, RegisterRequest
│   │   ├── product.go / product_model.go
│   │   ├── cart.go / cart_item.go
│   │   ├── order.go / order_item.go
│   │   └── review.go
│   ├── repository/
│   │   ├── repositoryMysql/
│   │   │   └── user_mysql_repository.go
│   │   ├── repositoryAerospike/
│   │   │   └── user_aerospike_repository.go
│   │   ├── repositoryKafkaProducer/
│   │   │   └── user_kafka_producer_repository.go
│   │   └── repositoryKafkaConsumer/
│   │       └── user_kafka_consumer_repository.go
│   ├── service/
│   │   └── user_service.go
│   ├── message_queue/kafka/
│   │   └── connect_kafka.go             # CreateProducer(), CreateConsumer()
│   ├── utils/security/
│   │   └── password.go                  # HashPasswordArgon2()
│   └── worker/
│       └── user_register_worker.go      # Async batch worker
├── pkg/vldtr/
│   └── validator.go                     # Error messages tiếng Việt
├── tests/                               # Chưa có test, cần bổ sung
├── .env                                 # DB credentials, Kafka config
├── .air.toml                            # Live reload config
├── docker-compose.yaml
├── dockerfile
└── aerospike.conf
```

---

## 5. Register Feature — Luồng đầy đủ

### 5.1 Flow diagram

```
POST /Register
    │
    ▼
[UserHandler.Register()]
  ├─ Bind JSON → InputRegister
  ├─ Validate struct tags (required, min, email, len, eqfield, startswith, numeric)
  ├─ Check special char trong Password (regex: [^a-zA-Z0-9])
  └─ Map → RegisterRequest (bỏ PasswordConfirm)
    │
    ▼
[UserService.Register()]
  ├─ (1) CheckUserExistsAerospike(email, username)
  │       key: "pending:email:<email>", "pending:username:<username>"
  │       namespace: "fashion_e-commerce", set: "usersRegister"
  │       → Nếu exists: return error ngay (fast fail)
  │
  ├─ (2) Nếu không có trong Aerospike → CheckUserExistsMysql(email, username)
  │       UNION ALL query, dừng ngay khi tìm thấy 1 record
  │       → Nếu exists: return "Email hoặc username đã tồn tại"
  │
  ├─ (3) SHA-256 hash password (bảo vệ đường truyền tới Kafka)
  │       KHÔNG dùng để lưu vào DB — chỉ là placeholder tạm
  │
  ├─ (4) CreateRecordRegister(userModel) → Aerospike (TTL 300s)
  │       Ghi 2 key: email + username, rollback email nếu username lỗi
  │
  ├─ (5) ProduceRegisterMessage(userModel) → Kafka topic "user-registered"
  │       key: "email-username", value: JSON(User)
  │       → Return 200 OK ngay cho client
  │
  └─ (ERR) Nếu Kafka lỗi → DeleteRecordRegister() xóa Aerospike reservation
    │
    ▼
[Kafka topic: "user-registered"]
  BatchSize: 10,000 | BatchTimeout: 50ms | Async: true | Balancer: LeastBytes
    │
    ▼ (4 goroutines song song)
[RegisterWorker.WorkerRegister()]
  ├─ goroutine đọc Kafka → parse JSON → Argon2id hash password nặng
  │   (params: memory=8192KB, t=1, p=1, saltLen=16, keyLen=32)
  │   → push vào userChan (capacity 1000)
  │
  └─ main loop select:
      ├─ case user ← userChan → append vào batch[]
      │   └─ len(batch) >= 1000 → flushBatch() + reset ticker
      └─ case ←ticker.C (mỗi 10s) → flushBatch() nếu batch > 0
    │
    ▼
[flushBatch()]
  ├─ db.Create(&users) — GORM Bulk Insert
  ├─ Thành công → log "Đã Bulk Insert thành công N tài khoản"
  └─ Lỗi → handleBatchError() — xử lý từng record lẻ
              └─ Error 1062 / "Duplicate entry" → compensateAerospikeCache()
                  Cập nhật status: "confirmed_in_db" cho cả 2 key
```

### 5.2 Lý do thiết kế quan trọng

| Quyết định | Lý do |
|-----------|-------|
| Aerospike check trước MySQL | Aerospike là RAM-based, response < 1ms, tránh hit DB cho trường hợp spam |
| SHA-256 tại API, Argon2id tại Worker | SHA-256 không block CPU API (mục tiêu < 100ms response). Argon2id nặng (~50-200ms) chạy async ở Worker |
| Aerospike reservation (TTL 300s) | Kafka là async — nếu chỉ check MySQL, 2 request cùng email trong < 300ms sẽ cả 2 pass check nhưng Worker sẽ chỉ insert 1 thành công (gây mất data hoặc lỗi). Aerospike làm "pessimistic lock" tạm thời |
| Bulk Insert theo batch 1000 + 10s timer | Giảm số lượng round-trip tới MySQL, tối ưu throughput |
| 4 Worker goroutines | Khớp với số logical CPU cores của máy dev (4 cores) |
| Kafka key = "email-username" | Đảm bảo message của cùng 1 user luôn vào cùng 1 partition → tránh race condition giữa các Worker |

---

## 6. Models

### User (`internal/models/user.go`)

```go
type User struct {
    UserID       uint      `gorm:"primaryKey"`
    Username     string    `gorm:"uniqueIndex;type:varchar(50);not null"`
    Email        string    `gorm:"uniqueIndex;type:varchar(100);not null"`
    PasswordHash string    `gorm:"type:varchar(255);not null"`
    Role         string    `gorm:"type:varchar(20);default:'customer';index"`
    FullName     string    `gorm:"type:varchar(100)"`
    Phone        string    `gorm:"type:varchar(10)"`
    Address      string    `gorm:"type:text"`
    CreatedAt    time.Time `gorm:"column:create_at"`
    UpdatedAt    time.Time `gorm:"column:update_at"`
    Orders  []Order  // HasMany
    Cart    Cart     // HasOne
    Reviews []Review // HasMany
}
```

**Lưu ý:** `CreatedAt` → column `create_at`, `UpdatedAt` → column `update_at` (không phải snake_case GORM mặc định).

### InputRegister — Validation rules

| Field | Rules |
|-------|-------|
| `username` | required, min=6 |
| `full_name` | required, min=6 |
| `email` | required, email format |
| `phone` | required, len=10, startswith=0, numeric |
| `address` | required |
| `password` | required, min=10, **phải có ký tự đặc biệt** (check riêng bằng regex) |
| `password_confirm` | required, eqfield=Password |

`RegisterRequest` = `InputRegister` trừ `password_confirm` — struct này được truyền vào Service.

---

## 7. Repository Patterns

### MySQL Repository

- Dùng `Raw().Scan()` thay vì `First()`/`Where()` khi cần query phức tạp hoặc optimize index
- Khi `Raw().Scan()` không tìm thấy record, **không** trả về `gorm.ErrRecordNotFound` — phải check `user.Email == ""` thủ công
- Connection pool: `MaxOpenConns=9`, `MaxIdleConns=9`, `ConnMaxLifetime=30m` (công thức: CPU×2+1)

### Aerospike Repository

- Namespace: `"fashion_e-commerce"`, Set: `"usersRegister"`
- Key pattern: `"pending:email:<email>"`, `"pending:username:<username>"`
- `WritePolicy` được tạo với TTL 300s: `as.NewWritePolicy(0, 300)`
- Khi ghi 2 key (email + username), rollback key đầu tiên nếu key thứ hai lỗi
- `compensateAerospikeCache()` cập nhật `status: "confirmed_in_db"` — không xóa record, giúp chặn đăng ký lại với email đã tồn tại trong DB

### Kafka Producer

- Topic: `"user-registered"` | Broker: `kafka:9092` (Docker service name)
- `Async: true` — không block goroutine của request
- `BatchSize: 10000`, `BatchTimeout: 50ms`, `Balancer: LeastBytes`
- `AllowAutoTopicCreation: true` — tự tạo topic khi chưa có

### Kafka Consumer (Worker)

- `GroupID: "register-worker-group"` — Kafka tự phân phối partition giữa các consumer trong group
- `MaxBytes: 10MB`

---

## 8. Password Hashing

### Argon2id params (`internal/utils/security/password.go`)

```go
Argon2Memory      = 8 * 1024  // 8192 KB = 8 MB
Argon2Iterations  = 1
Argon2Parallelism = 1
Argon2SaltLength  = 16 bytes
Argon2KeyLength   = 32 bytes
```

Output format: `$argon2id$v=19$m=8192,t=1,p=1$<base64salt>$<base64hash>`

**Quan trọng:** Hàm `HashPasswordArgon2()` nhận input là **SHA-256 hash** (không phải plaintext). Worker nhận `user.PasswordHash` (SHA-256) từ Kafka và hash lại bằng Argon2id → lưu vào MySQL.

---

## 9. Validation & Error Messages

`pkg/vldtr/GetErrorMessageRegister()` map validator tag → tiếng Việt:

| Tag | Message |
|-----|---------|
| `required` | "Trường X không được để trống" |
| `email` | "Định dạng email không hợp lệ" |
| `min` | "Trường X phải có ít nhất N ký tự" |
| `max` | "Trường X không được vượt quá N ký tự" |
| `len` | "Trường X phải có đúng N ký tự" |
| `eqfield` | "Trường X phải khớp với trường Y" |
| `startswith` | "Trường X phải bắt đầu bằng Y" |
| `numeric` | "Trường X chỉ được chứa các chữ số" |

Handler trả về **mảng lỗi** (`"errors": [...]`) khi validation fail, hoặc **single error** (`"error": "..."`) cho các trường hợp khác.

---

## 10. Infrastructure

### Docker Compose services

Xem `docker-compose.yaml`. Các service chính:
- `mysql` — expose port 3306
- `aerospike` — expose port 3000 (client), 3001 (fabric), 3002 (mesh), 3003 (HTTP)
- `kafka` — expose port 9092
- App chạy trên port **3004** (`app.Listen(":3004")`)

### Environment Variables (`.env`)

Config được load qua `config/config_mysql.go` và `config/config_aerospike.go`. Các biến cần có:
- `MYSQL_USER`, `MYSQL_PASSWORD`, `MYSQL_HOST`, `MYSQL_DATABASE`
- Aerospike host config (xem `config_aerospike.go`)

### Aerospike config

File `aerospike.conf` ở root — mount vào container Aerospike.

---

## 11. Conventions & Patterns

### Naming

- Package theo chức năng: `handler`, `service`, `repositoryMysql`, `repositoryAerospike`, `repositoryKafkaProducer`, `worker`
- Struct tên: `UserHandler`, `UserService`, `UserMysqlRepository`, `UserAerospikeRepository`, `UserKafkaProducerRepository`, `RegisterWorker`
- Constructor: `New<StructName>(deps...) *StructName`

### Error handling

- Service layer return `error` — Handler xử lý HTTP status code
- Khi Kafka lỗi: compensation logic xóa Aerospike reservation trước khi return error
- Worker: log lỗi và `continue` — không crash goroutine
- Batch lỗi: fallback sang xử lý từng record lẻ (`handleBatchError`)

### Code comments

- Comment giải thích **lý do** (trade-off performance, lý do dùng thuật toán X) — không comment "what"
- Comments tiếng Việt là bình thường trong project này

### HTTP responses

```json
// Success
{"message": "Dữ liệu hợp lệ, tiến hành đăng ký"}

// Validation errors (nhiều lỗi)
{"errors": ["Trường X không được để trống", ...]}

// Single error
{"error": "Mô tả lỗi"}
```

---

## 12. Cách chạy môi trường dev

```bash
# Khởi động toàn bộ infrastructure (MySQL, Aerospike, Kafka)
docker-compose up -d

# Chạy app với live reload (Air)
air

# Hoặc chạy trực tiếp
go run cmd/api/main.go
```

App listen tại `http://localhost:3004`

**Test endpoint đăng ký:**
```bash
curl -X POST http://localhost:3004/Register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "testuser01",
    "full_name": "Nguyen Van A",
    "email": "test@example.com",
    "phone": "0123456789",
    "address": "123 Nguyen Hue, HCM",
    "password": "MyPass@2024",
    "password_confirm": "MyPass@2024"
  }'
```

---

## 13. Trạng thái hiện tại & TODO

### Đã implement (branch: `feature-new-register`)

- [x] POST `/Register` — full async pipeline
- [x] Handler validation với error messages tiếng Việt
- [x] Two-stage password hashing (SHA-256 API → Argon2id Worker)
- [x] Aerospike reservation pattern (chống race condition)
- [x] Kafka producer (async, batched)
- [x] Worker với batch insert 1000 records / 10s flush
- [x] Compensation logic khi Kafka lỗi
- [x] MySQL connection pooling

### Chưa implement

- [ ] Login / JWT authentication
- [ ] Product catalog (Group 1)
- [ ] Cart & Checkout (Group 1)
- [ ] Flash Sale feature (Group 2) — Aerospike inventory, rate limiting
- [ ] Unit tests & integration tests (`tests/` còn trống)
- [ ] k6 load test scripts

### Bug đã biết

- `user_register_worker.go` line 9: `user_register_worker.go` đang là file modified (M) trong git — có thể đang có thay đổi chưa commit.

---

## 14. Lưu ý khi thêm tính năng mới

1. **Thêm entity mới** → Tạo file model trong `internal/models/`, thêm vào `AutoMigrate()` trong `main.go`
2. **Thêm endpoint** → Handler mới trong `internal/handler/`, wire vào `main.go`, tạo route `app.Method("/path", handler.Method)`
3. **Thêm Kafka topic** → Tạo producer/consumer mới trong `internal/message_queue/kafka/`, tạo worker mới trong `internal/worker/`
4. **Validation error message** → Thêm case vào `pkg/vldtr/validator.go`
5. **Không dùng Argon2id trong API request path** — chỉ dùng trong Worker để tránh CPU bottleneck
6. **Aerospike namespace phải là** `"fashion_e-commerce"` — trùng với cấu hình trong `aerospike.conf`
