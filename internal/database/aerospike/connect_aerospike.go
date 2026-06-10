package aerospike

import (
	"log"
	"time"

	"github.com/DVT1609/fashion_e-commerce.git/config"
	as "github.com/aerospike/aerospike-client-go/v7"
)

func ConnectAerospike() (*as.Client, error) {
	policy := as.NewClientPolicy()
	// 1. Cấu hình Connection Pool
	policy.ConnectionQueueSize = 1024 // Tăng ConnectionQueueSize để hỗ trợ nhiều kết nối hơn
	policy.LimitConnectionsToQueueSize = true

	// 2. Cấu hình Timeout chuẩn v6 (Đổi từ policy.Timeout)
	policy.Timeout = 500 * time.Millisecond

	// 3. QUAN TRỌNG: Cấu hình cho Docker
	// Giúp tránh lỗi "FAIL_FORBIDDEN" khi Cluster trả về IP nội bộ
	policy.UseServicesAlternate = true

	config := config.LoadConfigAerospike()
	client, err := as.NewClientWithPolicy(policy, config.Aerospike_Host, 3000)
	if err != nil {
		log.Fatalf("Lỗi kết nối aerospike 1: %v", err)
		return nil, err
	}
	log.Printf("Kết nối aerospike thành công")
	return client, nil
}
