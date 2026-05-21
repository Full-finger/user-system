package validator

import (
	"strings"

	"github.com/go-playground/validator/v10"
)

type CustomValidator struct {
	validator *validator.Validate
}

func New() *CustomValidator {
	return &CustomValidator{
		validator: validator.New(),
	}
}

func (cv *CustomValidator) Validate(i interface{}) error {
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
			default:
				errMessages = append(errMessages, e.Field()+"不合法")
			}
		}
		return &ValidationError{Messages: errMessages}
	}
	return nil
}

type ValidationError struct {
	Messages []string
}

func (v *ValidationError) Error() string {
	return strings.Join(v.Messages, "；")
}
