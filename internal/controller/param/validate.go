package param

import (
	"github.com/full-finger/user-system/internal/apperror"
	"github.com/labstack/echo/v4"
)

func BindAndValidate(c echo.Context, req any) error {
	if err := c.Bind(req); err != nil {
		return apperror.BadRequest("参数错误")
	}
	if err := c.Validate(req); err != nil {
		return apperror.BadRequest(err.Error())
	}
	return nil
}
