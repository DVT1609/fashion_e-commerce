package repositoryAerospike

import (
	"github.com/DVT1609/fashion_e-commerce.git/internal/models"
	as "github.com/aerospike/aerospike-client-go/v7"
)

type UserAerospikeRepository struct {
	Client   *as.Client
	CtxBase  *as.BasePolicy
	CtxWrite *as.WritePolicy
}

func (repo *UserAerospikeRepository) CheckUserExistsAerospike(email string, username string) (bool, error) {
	// Kiểm tra trong Aerospike nếu có email hoặc username trùng lặp
	// Sử dụng email và username làm key để tra cứu nhanh
	keyEmail, _ := as.NewKey("fashion_e-commerce", "usersRegister", "pending:email:"+email)
	keyUsername, _ := as.NewKey("fashion_e-commerce", "usersRegister", "pending:username:"+username)
	existsEmail, err := repo.Client.Exists(repo.CtxBase, keyEmail)
	if err != nil {
		return false, err
	}
	existsUsername, err := repo.Client.Exists(repo.CtxBase, keyUsername)
	if err != nil {
		return false, err
	}

	return existsEmail || existsUsername, nil
}

func (repo *UserAerospikeRepository) CreateRecordRegister(userModel models.User) error {
	// 1. Tạo Key cho cả Email và Username để chặn trùng lặp tuyệt đối
	keyEmail, _ := as.NewKey("fashion_e-commerce", "usersRegister", "pending:email:"+userModel.Email)
	keyUsername, _ := as.NewKey("fashion_e-commerce", "usersRegister", "pending:username:"+userModel.Username)

	// Gom chung một bộ Bins đầy đủ cho cả 2 Key, để sau này có thể dễ dàng
	//truy xuất email từ username hoặc ngược lại nếu cần thiết
	bins := as.BinMap{
		"email"         :userModel.Email,
        "username"      :userModel.Username,
        "password_hash" :userModel.PasswordHash,
        "full_name"     :userModel.FullName,
	}

	// 2. Sử dụng chính sách WritePolicy đã cấu hình (có TTL 300s từ main.go)
	// Ghi record cho Email
	err := repo.Client.Put(repo.CtxWrite, keyEmail, bins)
	if err != nil {
		return err
	}

	// 3. Ghi record cho Username
	err = repo.Client.Put(repo.CtxWrite, keyUsername, bins)
	if err != nil {
		// Rollback: Nếu ghi Username lỗi, hãy xóa Email đã ghi trước đó để tránh treo record
		_, err = repo.Client.Delete(repo.CtxWrite, keyEmail)
		return err
	}

	return nil
}

func (repo *UserAerospikeRepository) DeleteRecordRegister(email, username string) (bool, error) {
	// Xóa record tạm thời trong Aerospike nếu Kafka xử lý lỗi hoặc sau khi đã xử lý xong
	keyEmail, _ := as.NewKey("fashion_e-commerce", "usersRegister", "pending:email:"+email)
	keyUsername, _ := as.NewKey("fashion_e-commerce", "usersRegister", "pending:username:"+username)

	// Xóa record cho Email
	_, errEmail := repo.Client.Delete(repo.CtxWrite, keyEmail)
	if errEmail != nil {
		return false, errEmail
	}

	// Xóa record cho Username
	_, errUsername := repo.Client.Delete(repo.CtxWrite, keyUsername)
	if errUsername != nil {
		return false, errUsername
	}

	return true, nil
}
