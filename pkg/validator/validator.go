// Package validator 提供 Echo validator 接口的中文错误翻译实现。
package validator

import (
	"regexp"
	"strings"
	"unicode"

	"github.com/go-playground/validator/v10"
)

var (
	usernameRe = regexp.MustCompile(`^[a-zA-Z0-9_]{3,30}$`)
)

// CustomValidator 实现 echo.Validator，将验证错误转为中文提示。
type CustomValidator struct {
	validator *validator.Validate
}

// New 创建 CustomValidator 实例，注册自定义验证规则。
func New() *CustomValidator {
	v := validator.New()
	_ = v.RegisterValidation("username", validateUsername)
	_ = v.RegisterValidation("nickname", validateNickname)
	_ = v.RegisterValidation("password", validatePassword)
	return &CustomValidator{validator: v}
}

// validateUsername 用户名：仅限字母、数字、下划线，3-30 位。
func validateUsername(fl validator.FieldLevel) bool {
	return usernameRe.MatchString(fl.Field().String())
}

// validateNickname 昵称：1-50 字符，禁止控制字符，禁止纯空白。
func validateNickname(fl validator.FieldLevel) bool {
	s := fl.Field().String()
	if len(s) == 0 || len(s) > 50 {
		return false
	}
	for _, r := range s {
		if unicode.IsControl(r) {
			return false
		}
	}
	return strings.TrimSpace(s) == s && len(strings.TrimSpace(s)) > 0
}

// validatePassword 密码：至少 8 位，必须包含至少一个字母和一个数字。
func validatePassword(fl validator.FieldLevel) bool {
	s := fl.Field().String()
	if len(s) < 8 {
		return false
	}
	var hasLetter, hasDigit bool
	for _, r := range s {
		if unicode.IsLetter(r) {
			hasLetter = true
		}
		if unicode.IsDigit(r) {
			hasDigit = true
		}
		if hasLetter && hasDigit {
			return true
		}
	}
	return false
}

func (cv *CustomValidator) Validate(i any) error {
	if err := cv.validator.Struct(i); err != nil {
		// 将验证错误转换为友好的中文提示
		var errMessages []string
		for _, e := range err.(validator.ValidationErrors) {
			switch e.Tag() {
			case "required":
				errMessages = append(errMessages, e.Field()+"不能为空")
			case "min":
				errMessages = append(errMessages, e.Field()+"长度不能少于"+e.Param())
			case "max":
				errMessages = append(errMessages, e.Field()+"长度不能超过"+e.Param())
			case "oneof":
				errMessages = append(errMessages, e.Field()+"值必须是"+e.Param()+"之一")
			case "username":
				errMessages = append(errMessages, "用户名仅限字母、数字和下划线，3-30 位")
			case "nickname":
				errMessages = append(errMessages, "昵称长度 1-50 字符，不能包含特殊控制字符")
			case "password":
				errMessages = append(errMessages, "密码至少 8 位，须包含字母和数字")
			default:
				errMessages = append(errMessages, e.Field()+"不合法")
			}
		}
		return &ValidationError{Messages: errMessages}
	}
	return nil
}

// ValidationError 验证错误集合。
type ValidationError struct {
	Messages []string
}

func (v *ValidationError) Error() string {
	return strings.Join(v.Messages, "；")
}
