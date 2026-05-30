package security

import (
	"crypto/rand"
	// "crypto/subtle"
	"encoding/base64"
	"fmt"

	// "strings"
	"golang.org/x/crypto/argon2"
)

// Cấu hình Argon2id (Bạn có thể điều chỉnh tùy theo tài nguyên Server)
const (
	Argon2Memory      = 8 * 1024 // 16MB
	Argon2Iterations  = 1
	Argon2Parallelism = 1
	Argon2SaltLength  = 16
	Argon2KeyLength   = 32
)

// HashPasswordArgon2 tạo ra chuỗi hash bao gồm cả Salt theo định dạng chuẩn
func HashPasswordArgon2(password string) (string, error) {
	// 1. Tạo Salt ngẫu nhiên
	salt := make([]byte, Argon2SaltLength)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}

	// 2. Băm mật khẩu
	hash := argon2.IDKey([]byte(password), salt, Argon2Iterations, Argon2Memory, Argon2Parallelism, Argon2KeyLength)

	// 3. Đóng gói thành chuỗi: $argon2id$v=19$m=65536,t=1,p=4$salt$hash
	b64Salt := base64.RawStdEncoding.EncodeToString(salt)
	b64Hash := base64.RawStdEncoding.EncodeToString(hash)

	encodedHash := fmt.Sprintf("$argon2id$v=19$m=%d,t=%d,p=%d$%s$%s",
		Argon2Memory, Argon2Iterations, Argon2Parallelism, b64Salt, b64Hash)

	return encodedHash, nil
}
