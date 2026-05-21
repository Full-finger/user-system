package controller

import (
	"net/http"
	"strconv"

	"github.com/full-finger/user-system/internal/controller/param"
	"github.com/full-finger/user-system/internal/service"
	"github.com/labstack/echo/v4"
)

type UserController struct {
	svc        *service.UserService
	captchaSvc *service.CaptchaService
}

func NewUserController(svc *service.UserService, captchaSvc *service.CaptchaService) *UserController {
	return &UserController{svc: svc, captchaSvc: captchaSvc}
}

func success(c echo.Context, data any) error {
	return c.JSON(http.StatusOK, map[string]any{
		"code":    200,
		"message": "success",
		"data":    data,
	})
}

func fail(c echo.Context, code int, msg string) error {
	return c.JSON(code, map[string]any{
		"code":    code,
		"message": msg,
		"data":    nil,
	})
}

func (ctrl *UserController) Register(c echo.Context) error {
	var req param.RegisterRequest
	if err := c.Bind(&req); err != nil {
		return fail(c, http.StatusBadRequest, "参数错误")
	}
	if err := c.Validate(&req); err != nil {
		return fail(c, http.StatusBadRequest, err.Error())
	}
	user, err := ctrl.svc.Register(service.RegisterInput{
		Username: req.Username,
		Password: req.Password,
	})
	if err != nil {
		return fail(c, http.StatusBadRequest, err.Error())
	}
	return success(c, user)
}

func (ctrl *UserController) Login(c echo.Context) error {
	var req param.LoginRequest
	if err := c.Bind(&req); err != nil {
		return fail(c, http.StatusBadRequest, "参数错误")
	}
	if err := c.Validate(&req); err != nil {
		return fail(c, http.StatusBadRequest, err.Error())
	}
	token, err := ctrl.svc.Login(service.LoginInput{
		Username: req.Username,
		Password: req.Password,
	})
	if err != nil {
		return fail(c, http.StatusUnauthorized, err.Error())
	}
	return success(c, param.LoginResponse{Token: token})
}

func (ctrl *UserController) GetProfile(c echo.Context) error {
	userID := c.Get("user_id").(uint)
	user, err := ctrl.svc.GetProfile(userID)
	if err != nil {
		return fail(c, http.StatusNotFound, err.Error())
	}
	return success(c, user)
}

func (ctrl *UserController) UpdateProfile(c echo.Context) error {
	userID := c.Get("user_id").(uint)
	var req param.UpdateRequest
	if err := c.Bind(&req); err != nil {
		return fail(c, http.StatusBadRequest, "参数错误")
	}
	if err := c.Validate(&req); err != nil {
		return fail(c, http.StatusBadRequest, err.Error())
	}
	// 普通用户不允许自己改角色
	req.Role = ""
	user, err := ctrl.svc.UpdateProfile(userID, service.UpdateInput{
		Password: req.Password,
		Role:     req.Role,
	})
	if err != nil {
		return fail(c, http.StatusBadRequest, err.Error())
	}
	return success(c, user)
}

func (ctrl *UserController) ListUsers(c echo.Context) error {
	users, err := ctrl.svc.ListUsers()
	if err != nil {
		return fail(c, http.StatusInternalServerError, "查询失败")
	}
	return success(c, users)
}

func (ctrl *UserController) GetUser(c echo.Context) error {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		return fail(c, http.StatusBadRequest, "无效的ID")
	}
	user, err := ctrl.svc.GetUserByID(uint(id))
	if err != nil {
		return fail(c, http.StatusNotFound, err.Error())
	}
	return success(c, user)
}

func (ctrl *UserController) UpdateUser(c echo.Context) error {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		return fail(c, http.StatusBadRequest, "无效的ID")
	}
	var req param.UpdateRequest
	if err := c.Bind(&req); err != nil {
		return fail(c, http.StatusBadRequest, "参数错误")
	}
	if err := c.Validate(&req); err != nil {
		return fail(c, http.StatusBadRequest, err.Error())
	}
	user, err := ctrl.svc.UpdateUser(uint(id), service.UpdateInput{
		Password: req.Password,
		Role:     req.Role,
	})
	if err != nil {
		return fail(c, http.StatusBadRequest, err.Error())
	}
	return success(c, user)
}

func (ctrl *UserController) DeleteUser(c echo.Context) error {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		return fail(c, http.StatusBadRequest, "无效的ID")
	}
	if err := ctrl.svc.DeleteUser(uint(id)); err != nil {
		return fail(c, http.StatusInternalServerError, "删除失败")
	}
	return success(c, nil)
}

func (ctrl *UserController) SendCode(c echo.Context) error {
	var req param.SendCodeRequest
	if err := c.Bind(&req); err != nil {
		return fail(c, http.StatusBadRequest, "参数错误")
	}
	if err := c.Validate(&req); err != nil {
		return fail(c, http.StatusBadRequest, err.Error())
	}
	if err := ctrl.captchaSvc.SendCode(req.Email); err != nil {
		return fail(c, http.StatusTooManyRequests, err.Error())
	}
	return success(c, nil)
}

func (ctrl *UserController) CodeLogin(c echo.Context) error {
	var req param.CodeLoginRequest
	if err := c.Bind(&req); err != nil {
		return fail(c, http.StatusBadRequest, "参数错误")
	}
	if err := c.Validate(&req); err != nil {
		return fail(c, http.StatusBadRequest, err.Error())
	}
	if err := ctrl.captchaSvc.VerifyCode(req.Email, req.Code); err != nil {
		return fail(c, http.StatusUnauthorized, err.Error())
	}
	token, err := ctrl.svc.LoginByEmail(req.Email)
	if err != nil {
		return fail(c, http.StatusUnauthorized, err.Error())
	}
	return success(c, param.LoginResponse{Token: token})
}

func (ctrl *UserController) BindEmail(c echo.Context) error {
	userID := c.Get("user_id").(uint)
	var req param.BindEmailRequest
	if err := c.Bind(&req); err != nil {
		return fail(c, http.StatusBadRequest, "参数错误")
	}
	if err := c.Validate(&req); err != nil {
		return fail(c, http.StatusBadRequest, err.Error())
	}
	if err := ctrl.captchaSvc.VerifyCode(req.Email, req.Code); err != nil {
		return fail(c, http.StatusBadRequest, err.Error())
	}
	if err := ctrl.svc.BindEmail(userID, req.Email); err != nil {
		return fail(c, http.StatusBadRequest, err.Error())
	}
	return success(c, nil)
}
