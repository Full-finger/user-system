// Package controller 处理 HTTP 请求，调用 service 层完成业务。
package controller

import (
	"net/http"
	"strconv"

	"github.com/full-finger/user-system/internal/apperror"
	"github.com/full-finger/user-system/internal/controller/param"
	"github.com/full-finger/user-system/internal/service"
	"github.com/labstack/echo/v4"
)

// UserController 用户相关接口的处理器。
type UserController struct {
	svc        *service.UserService
	captchaSvc *service.CaptchaService
}

func NewUserController(svc *service.UserService, captchaSvc *service.CaptchaService) *UserController {
	return &UserController{svc: svc, captchaSvc: captchaSvc}
}

func success(c echo.Context, data any) error {
	return c.JSON(http.StatusOK, map[string]any{
		"code": 200, "message": "success", "data": data,
	})
}

func bindAndValidate(c echo.Context, req any) error {
	if err := c.Bind(req); err != nil {
		return apperror.BadRequest("参数错误")
	}
	if err := c.Validate(req); err != nil {
		return apperror.BadRequest(err.Error())
	}
	return nil
}

func (ctrl *UserController) CheckUsername(c echo.Context) error {
	username := c.QueryParam("username")
	if username == "" {
		return apperror.BadRequest("请输入用户名")
	}
	if len(username) < 3 {
		return apperror.BadRequest("用户名至少 3 个字符")
	}
	if err := ctrl.svc.CheckUsername(c.Request().Context(), username); err != nil {
		return err
	}
	return success(c, map[string]bool{"available": true})
}

func (ctrl *UserController) Register(c echo.Context) error {
	var req param.RegisterRequest
	if err := bindAndValidate(c, &req); err != nil {
		return err
	}
	if err := ctrl.captchaSvc.VerifyCode(c.Request().Context(), req.Email, req.Code); err != nil {
		return err
	}
	user, err := ctrl.svc.Register(c.Request().Context(), service.RegisterInput{
		Username: req.Username,
		Password: req.Password,
		Email:    req.Email,
	})
	if err != nil {
		return err
	}
	return success(c, param.ToUserResponse(user))
}

func (ctrl *UserController) Login(c echo.Context) error {
	var req param.LoginRequest
	if err := bindAndValidate(c, &req); err != nil {
		return err
	}
	token, err := ctrl.svc.Login(c.Request().Context(), service.LoginInput{
		Username: req.Username,
		Password: req.Password,
	})
	if err != nil {
		return err
	}
	return success(c, param.LoginResponse{Token: token})
}

func (ctrl *UserController) GetProfile(c echo.Context) error {
	userID, ok := c.Get("user_id").(uint)
	if !ok {
		return apperror.Unauthorized("未认证")
	}
	user, err := ctrl.svc.GetProfile(c.Request().Context(), userID)
	if err != nil {
		return err
	}
	return success(c, param.ToUserResponse(user))
}

func (ctrl *UserController) UpdateProfile(c echo.Context) error {
	userID, ok := c.Get("user_id").(uint)
	if !ok {
		return apperror.Unauthorized("未认证")
	}
	var req param.UpdateRequest
	if err := bindAndValidate(c, &req); err != nil {
		return err
	}
	user, err := ctrl.svc.UpdateProfile(c.Request().Context(), userID, service.ProfileUpdateInput{
		Password: req.Password,
	})
	if err != nil {
		return err
	}
	return success(c, param.ToUserResponse(user))
}

func (ctrl *UserController) ListUsers(c echo.Context) error {
	page, _ := strconv.Atoi(c.QueryParam("page"))
	pageSize, _ := strconv.Atoi(c.QueryParam("page_size"))
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}
	users, total, err := ctrl.svc.ListUsers(c.Request().Context(), page, pageSize)
	if err != nil {
		return err
	}
	return success(c, param.ToUserListResponse(users, total, page, pageSize))
}

func (ctrl *UserController) GetUser(c echo.Context) error {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		return apperror.BadRequest("无效的ID")
	}
	user, err := ctrl.svc.GetProfile(c.Request().Context(), uint(id))
	if err != nil {
		return err
	}
	return success(c, param.ToUserResponse(user))
}

func (ctrl *UserController) UpdateUser(c echo.Context) error {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		return apperror.BadRequest("无效的ID")
	}
	var req param.UpdateRequest
	if err := bindAndValidate(c, &req); err != nil {
		return err
	}
	user, err := ctrl.svc.UpdateUser(c.Request().Context(), uint(id), service.UpdateInput{
		Password: req.Password,
		Role:     req.Role,
	})
	if err != nil {
		return err
	}
	return success(c, param.ToUserResponse(user))
}

func (ctrl *UserController) DeleteUser(c echo.Context) error {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		return apperror.BadRequest("无效的ID")
	}
	if err := ctrl.svc.DeleteUser(c.Request().Context(), uint(id)); err != nil {
		return err
	}
	return success(c, nil)
}

func (ctrl *UserController) SendCode(c echo.Context) error {
	var req param.SendCodeRequest
	if err := bindAndValidate(c, &req); err != nil {
		return err
	}
	if err := ctrl.captchaSvc.SendCode(c.Request().Context(), req.Email); err != nil {
		return err
	}
	return success(c, nil)
}

func (ctrl *UserController) CodeLogin(c echo.Context) error {
	var req param.CodeLoginRequest
	if err := bindAndValidate(c, &req); err != nil {
		return err
	}
	if err := ctrl.captchaSvc.VerifyCode(c.Request().Context(), req.Email, req.Code); err != nil {
		return err
	}
	token, err := ctrl.svc.LoginByEmail(c.Request().Context(), req.Email)
	if err != nil {
		return err
	}
	return success(c, param.LoginResponse{Token: token})
}

func (ctrl *UserController) BindEmail(c echo.Context) error {
	userID, ok := c.Get("user_id").(uint)
	if !ok {
		return apperror.Unauthorized("未认证")
	}
	var req param.BindEmailRequest
	if err := bindAndValidate(c, &req); err != nil {
		return err
	}
	if err := ctrl.captchaSvc.VerifyCode(c.Request().Context(), req.Email, req.Code); err != nil {
		return err
	}
	if err := ctrl.svc.BindEmail(c.Request().Context(), userID, req.Email); err != nil {
		return err
	}
	return success(c, nil)
}
