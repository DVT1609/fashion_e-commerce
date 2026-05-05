FROM golang:1.26-alpine

# 1. Cài đặt git và các thư viện cần thiết cho build
RUN apk add --no-cache git

# 2. Cài đặt Air (phiên bản mới nhất)
RUN go install github.com/air-verse/air@latest

# 3. Thiết lập thư mục làm việc
WORKDIR /app

# 4. Tối ưu hóa Cache: Copy file mod trước
COPY go.mod go.sum ./
RUN go mod download

# 5. Copy toàn bộ code vào Container
COPY . .

# 6. Mở cổng 3004
EXPOSE 3004

# 7. Chạy Air với đường dẫn tuyệt đối để tránh lỗi "command not found"
CMD ["/go/bin/air", "-c", ".air.toml"]