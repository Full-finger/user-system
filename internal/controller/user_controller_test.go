package controller

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/full-finger/user-system/internal/controller/param"
	"github.com/full-finger/user-system/pkg/validator"
	"github.com/labstack/echo/v4"
)

// TestUpdateRequest_RoleValidation 验证 Controller 层 BindAndValidate 对角色的校验。
// 弥补 P04：此前测试只调用 Service 层，绕过了 Controller 验证。
func TestUpdateRequest_RoleValidation(t *testing.T) {
	e := echo.New()
	e.Validator = validator.New()

	tests := []struct {
		name    string
		body    string
		wantErr bool
	}{
		{"user role accepted", `{"role":"user"}`, false},
		{"verified_user accepted", `{"role":"verified_user"}`, false},
		{"admin accepted", `{"role":"admin"}`, false},
		{"moderator rejected by validation", `{"role":"moderator"}`, true},
		{"super_admin rejected by validation", `{"role":"super_admin"}`, true},
		{"nonsense rejected", `{"role":"nonsense"}`, true},
		{"empty body accepted", `{}`, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPut, "/users/1", strings.NewReader(tt.body))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			var upd param.UpdateRequest
			err := param.BindAndValidate(c, &upd)
			if (err != nil) != tt.wantErr {
				t.Errorf("BindAndValidate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
