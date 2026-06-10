package handler

import (
	"regexp"

	"github.com/DVT1609/fashion_e-commerce.git/internal/models"
	"github.com/DVT1609/fashion_e-commerce.git/internal/service"
	"github.com/DVT1609/fashion_e-commerce.git/pkg/vldtr"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v3"
)

type UserHandler struct {
	UserService *service.UserService
	validate    *validator.Validate // Thêm trường này vào struct
}

func NewUserHandler(service *service.UserService, validate *validator.Validate) *UserHandler {

	return &UserHandler{
		UserService: service,
		validate:    validate,
	}
}

func (handler *UserHandler) Register(ctx fiber.Ctx) error {

	// 1. Bind dữ liệu từ Body
	var inputRegister models.InputRegister
	if err := ctx.Bind().Body(&inputRegister); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Đầu vào không hợp lệ: " + err.Error(),
		})
	}

	// 2. Kiểm tra các ràng buộc cơ bản qua tag validate
	if err := handler.validate.Struct(inputRegister); err != nil {
		var errors []string
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			for _, e := range validationErrors {
				errors = append(errors, vldtr.GetErrorMessageRegister(e)) // Sử dụng hàm chuyển đổi lỗi của bạn
			}
		}
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"errors": errors, // Trả về mảng các thông báo lỗi
		})
	}

	// 3. Kiểm tra ký tự đặc biệt cho Password (Validator mặc định không có tag này)
	// Regex: tìm ít nhất một ký tự không phải chữ cái hoặc số
	specialChar := regexp.MustCompile(`[^a-zA-Z0-9]`)
	if !specialChar.MatchString(inputRegister.Password) {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Mật khẩu phải chứa ít nhất 1 ký tự đặc biệt",
		})
	}

	var registerRequest = models.RegisterRequest{
		Username: inputRegister.Username,
		FullName: inputRegister.FullName,
		Email:    inputRegister.Email,
		Phone:    inputRegister.Phone,
		Address:  inputRegister.Address,
		Password: inputRegister.Password,
	}

	// 4. Nếu mọi thứ ổn, gọi tiếp tầng Service
	err := handler.UserService.Register(&registerRequest)

	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Đã xảy ra lỗi khi đăng ký người dùng: " + err.Error(),
		})
	}

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Dữ liệu hợp lệ, tiến hành đăng ký",
	})

}
