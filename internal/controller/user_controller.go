package controller

import (
	"net/http"
	"strconv"

	"github.com/full-finger/user-system/internal/model"
	"github.com/full-finger/user-system/internal/service"
	"github.com/labstack/echo/v4"
)

type UserController struct {
	svc *service.UserService
}

func NewUserController(svc *service.UserService) *UserController {
	return &UserController{svc: svc}
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
	var req model.RegisterRequest
	if err := c.Bind(&req); err != nil {
		return fail(c, http.StatusBadRequest, "参数错误")
	}
	user, err := ctrl.svc.Register(&req)
	if err != nil {
		return fail(c, http.StatusBadRequest, err.Error())
	}
	return success(c, user)
}

func (ctrl *UserController) Login(c echo.Context) error {
	var req model.LoginRequest
	if err := c.Bind(&req); err != nil {
		return fail(c, http.StatusBadRequest, "参数错误")
	}
	token, err := ctrl.svc.Login(&req)
	if err != nil {
		return fail(c, http.StatusUnauthorized, err.Error())
	}
	return success(c, model.LoginResponse{Token: token})
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
	var req model.UpdateRequest
	if err := c.Bind(&req); err != nil {
		return fail(c, http.StatusBadRequest, "参数错误")
	}
	// 普通用户不允许自己改角色
	req.Role = ""
	user, err := ctrl.svc.UpdateProfile(userID, &req)
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
	var req model.UpdateRequest
	if err := c.Bind(&req); err != nil {
		return fail(c, http.StatusBadRequest, "参数错误")
	}
	user, err := ctrl.svc.UpdateUser(uint(id), &req)
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
