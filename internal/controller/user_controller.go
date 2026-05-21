package controller

import (
	"net/http"
	"strconv"

	"github.com/full-finger/user-system/internal/apperror"
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
		"code": 200, "message": "success", "data": data,
	})
}

func (ctrl *UserController) Register(c echo.Context) error {
	var req param.RegisterRequest
	if err := c.Bind(&req); err != nil {
		return apperror.BadRequest("参数错误")
	}
	if err := c.Validate(&req); err != nil {
		return apperror.BadRequest(err.Error())
	}
	user, err := ctrl.svc.Register(service.RegisterInput{
		Username: req.Username,
		Password: req.Password,
	})
	if err != nil {
		return err
	}
	return success(c, user)
}

func (ctrl *UserController) Login(c echo.Context) error {
	var req param.LoginRequest
	if err := c.Bind(&req); err != nil {
		return apperror.BadRequest("参数错误")
	}
	if err := c.Validate(&req); err != nil {
		return apperror.BadRequest(err.Error())
	}
	token, err := ctrl.svc.Login(service.LoginInput{
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
	user, err := ctrl.svc.GetProfile(userID)
	if err != nil {
		return err
	}
	return success(c, user)
}

func (ctrl *UserController) UpdateProfile(c echo.Context) error {
	userID, ok := c.Get("user_id").(uint)
	if !ok {
		return apperror.Unauthorized("未认证")
	}
	var req param.UpdateRequest
	if err := c.Bind(&req); err != nil {
		return apperror.BadRequest("参数错误")
	}
	if err := c.Validate(&req); err != nil {
		return apperror.BadRequest(err.Error())
	}
	req.Role = ""
	user, err := ctrl.svc.UpdateProfile(userID, service.UpdateInput{
		Password: req.Password,
		Role:     req.Role,
	})
	if err != nil {
		return err
	}
	return success(c, user)
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
	users, total, err := ctrl.svc.ListUsers(page, pageSize)
	if err != nil {
		return err
	}
	return success(c, map[string]any{
		"list":      users,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

func (ctrl *UserController) GetUser(c echo.Context) error {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		return apperror.BadRequest("无效的ID")
	}
	user, err := ctrl.svc.GetUserByID(uint(id))
	if err != nil {
		return err
	}
	return success(c, user)
}

func (ctrl *UserController) UpdateUser(c echo.Context) error {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		return apperror.BadRequest("无效的ID")
	}
	var req param.UpdateRequest
	if err := c.Bind(&req); err != nil {
		return apperror.BadRequest("参数错误")
	}
	if err := c.Validate(&req); err != nil {
		return apperror.BadRequest(err.Error())
	}
	user, err := ctrl.svc.UpdateUser(uint(id), service.UpdateInput{
		Password: req.Password,
		Role:     req.Role,
	})
	if err != nil {
		return err
	}
	return success(c, user)
}

func (ctrl *UserController) DeleteUser(c echo.Context) error {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		return apperror.BadRequest("无效的ID")
	}
	if err := ctrl.svc.DeleteUser(uint(id)); err != nil {
		return err
	}
	return success(c, nil)
}

func (ctrl *UserController) SendCode(c echo.Context) error {
	var req param.SendCodeRequest
	if err := c.Bind(&req); err != nil {
		return apperror.BadRequest("参数错误")
	}
	if err := c.Validate(&req); err != nil {
		return apperror.BadRequest(err.Error())
	}
	if err := ctrl.captchaSvc.SendCode(req.Email); err != nil {
		return err
	}
	return success(c, nil)
}

func (ctrl *UserController) CodeLogin(c echo.Context) error {
	var req param.CodeLoginRequest
	if err := c.Bind(&req); err != nil {
		return apperror.BadRequest("参数错误")
	}
	if err := c.Validate(&req); err != nil {
		return apperror.BadRequest(err.Error())
	}
	if err := ctrl.captchaSvc.VerifyCode(req.Email, req.Code); err != nil {
		return err
	}
	token, err := ctrl.svc.LoginByEmail(req.Email)
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
	if err := c.Bind(&req); err != nil {
		return apperror.BadRequest("参数错误")
	}
	if err := c.Validate(&req); err != nil {
		return apperror.BadRequest(err.Error())
	}
	if err := ctrl.captchaSvc.VerifyCode(req.Email, req.Code); err != nil {
		return err
	}
	if err := ctrl.svc.BindEmail(userID, req.Email); err != nil {
		return err
	}
	return success(c, nil)
}
