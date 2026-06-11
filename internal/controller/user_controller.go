// Package controller 处理 HTTP 请求，调用 service 层完成业务。
package controller

import (
	"net/http"
	"regexp"
	"strconv"

	"github.com/full-finger/user-system/internal/apperror"
	"github.com/full-finger/user-system/internal/auth"
	"github.com/full-finger/user-system/internal/config"
	"github.com/full-finger/user-system/internal/controller/param"
	"github.com/full-finger/user-system/internal/service"
	"github.com/labstack/echo/v4"
)

var usernameRe = regexp.MustCompile(`^[a-zA-Z0-9_]{3,30}$`)

// UserController 用户相关接口的处理器。
type UserController struct {
	svc        *service.UserService
	captchaSvc *service.CaptchaService
	guestCfg   *config.GuestJWTConfig
}

func NewUserController(svc *service.UserService, captchaSvc *service.CaptchaService, guestCfg *config.GuestJWTConfig) *UserController {
	return &UserController{svc: svc, captchaSvc: captchaSvc, guestCfg: guestCfg}
}

func success(c echo.Context, data any) error {
	return c.JSON(http.StatusOK, map[string]any{
		"code": 200, "message": "success", "data": data,
	})
}

func (ctrl *UserController) GuestToken(c echo.Context) error {
	uc := auth.GetUserContext(c)
	token, err := auth.GenerateGuestToken(uc.DeviceID, ctrl.guestCfg)
	if err != nil {
		return apperror.Internal("生成游客令牌失败")
	}
	return success(c, param.LoginResponse{Token: token})
}

func (ctrl *UserController) CheckUsername(c echo.Context) error {
	username := c.QueryParam("username")
	if username == "" {
		return apperror.BadRequest("请输入用户名")
	}
	if !usernameRe.MatchString(username) {
		return apperror.BadRequest("用户名仅限字母、数字和下划线，3-30 位")
	}
	if err := ctrl.svc.CheckUsername(c.Request().Context(), username); err != nil {
		return err
	}
	return success(c, map[string]bool{"available": true})
}

func (ctrl *UserController) Register(c echo.Context) error {
	var req param.RegisterRequest
	if err := param.BindAndValidate(c, &req); err != nil {
		return err
	}
	if err := ctrl.captchaSvc.VerifyCode(c.Request().Context(), req.Email, req.Code); err != nil {
		return err
	}
	user, err := ctrl.svc.Register(c.Request().Context(), service.RegisterInput{
		Username: req.Username,
		Password: req.Password,
		Nickname: req.Nickname,
		Email:    req.Email,
	})
	if err != nil {
		return err
	}
	return success(c, param.ToUserResponse(user))
}

func (ctrl *UserController) Login(c echo.Context) error {
	var req param.LoginRequest
	if err := param.BindAndValidate(c, &req); err != nil {
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
	uc := auth.GetUserContext(c)
	user, err := ctrl.svc.GetProfile(c.Request().Context(), uc)
	if err != nil {
		return err
	}
	return success(c, param.ToUserResponse(user))
}

func (ctrl *UserController) UpdateProfile(c echo.Context) error {
	uc := auth.GetUserContext(c)
	var req param.UpdateRequest
	if err := param.BindAndValidate(c, &req); err != nil {
		return err
	}
	user, err := ctrl.svc.UpdateProfile(c.Request().Context(), uc, service.ProfileUpdateInput{
		Password: req.Password,
		Nickname: req.Nickname,
	})
	if err != nil {
		return err
	}
	return success(c, param.ToUserResponse(user))
}

func (ctrl *UserController) ListUsers(c echo.Context) error {
	uc := auth.GetUserContext(c)
	page, pageSize := parsePage(c)
	users, total, err := ctrl.svc.ListUsers(c.Request().Context(), uc, page, pageSize)
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
	uc := auth.GetUserContext(c)
	user, err := ctrl.svc.GetProfile(c.Request().Context(), uc)
	if err != nil {
		return err
	}
	// 管理员查看他人资料，用 id 查询
	if uint(id) != uc.UserID {
		user, err = ctrl.svc.FindByID(c.Request().Context(), uint(id))
		if err != nil {
			return err
		}
	}
	return success(c, param.ToUserResponse(user))
}

func (ctrl *UserController) UpdateUser(c echo.Context) error {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		return apperror.BadRequest("无效的ID")
	}
	var req param.UpdateRequest
	if err := param.BindAndValidate(c, &req); err != nil {
		return err
	}
	uc := auth.GetUserContext(c)
	user, err := ctrl.svc.UpdateUser(c.Request().Context(), uc, uint(id), service.UpdateInput{
		Password: req.Password,
		Nickname: req.Nickname,
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
	uc := auth.GetUserContext(c)
	if err := ctrl.svc.DeleteUser(c.Request().Context(), uc, uint(id)); err != nil {
		return err
	}
	return success(c, nil)
}

func (ctrl *UserController) SendCode(c echo.Context) error {
	var req param.SendCodeRequest
	if err := param.BindAndValidate(c, &req); err != nil {
		return err
	}
	if err := ctrl.captchaSvc.SendCode(c.Request().Context(), req.Email); err != nil {
		return err
	}
	return success(c, nil)
}

func (ctrl *UserController) CodeLogin(c echo.Context) error {
	var req param.CodeLoginRequest
	if err := param.BindAndValidate(c, &req); err != nil {
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

func (ctrl *UserController) AppointModerator(c echo.Context) error {
	var req param.AppointModeratorRequest
	if err := param.BindAndValidate(c, &req); err != nil {
		return err
	}
	uc := auth.GetUserContext(c)
	user, err := ctrl.svc.AppointModerator(c.Request().Context(), uc, req.UserID, req.NodeIDs)
	if err != nil {
		return err
	}
	return success(c, param.ToUserResponse(user))
}

func (ctrl *UserController) BindEmail(c echo.Context) error {
	uc := auth.GetUserContext(c)
	var req param.BindEmailRequest
	if err := param.BindAndValidate(c, &req); err != nil {
		return err
	}
	if err := ctrl.captchaSvc.VerifyCode(c.Request().Context(), req.Email, req.Code); err != nil {
		return err
	}
	if err := ctrl.svc.BindEmail(c.Request().Context(), uc, req.Email); err != nil {
		return err
	}
	return success(c, nil)
}
