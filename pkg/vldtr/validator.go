package vldtr // Package vldtr chứa các hàm tiện ích để chuyển đổi lỗi từ validator thành thông báo tiếng Việt

import (
	"fmt"
	"github.com/go-playground/validator/v10"
)

// GetErrorMessageRegister chuyển đổi lỗi từ validator thành thông báo tiếng Việt cho trường hợp đăng ký
func GetErrorMessageRegister(err validator.FieldError) string {
	switch err.Tag() {
	case "required":
		return fmt.Sprintf("Trường %s không được để trống", err.Field())
	case "email":
		return "Định dạng email không hợp lệ"
	case "min":
		return fmt.Sprintf("Trường %s phải có ít nhất %s ký tự", err.Field(), err.Param())
	case "max":
		return fmt.Sprintf("Trường %s không được vượt quá %s ký tự", err.Field(), err.Param())
	case "len":
		return fmt.Sprintf("Trường %s phải có đúng %s ký tự", err.Field(), err.Param())
	case "eqfield":
		return fmt.Sprintf("Trường %s phải khớp với trường %s", err.Field(), err.Param())
	case "startswith":
		return fmt.Sprintf("Trường %s phải bắt đầu bằng %s", err.Field(), err.Param())
	case "numeric":
		return fmt.Sprintf("Trường %s chỉ được chứa các chữ số", err.Field())
	default:
		return fmt.Sprintf("Trường %s không hợp lệ", err.Field())
	}
}