package service

import (
	"strings"
	"time"

	"github.com/full-finger/user-system/internal/apperror"
	"github.com/full-finger/user-system/internal/config"
	"github.com/full-finger/user-system/internal/model"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type UserService struct {
	db  *gorm.DB
	cfg *config.JWTConfig
}

func NewUserService(db *gorm.DB, cfg *config.JWTConfig) *UserService {
	return &UserService{db: db, cfg: cfg}
}

func (s *UserService) Register(in RegisterInput) (*model.User, error) {
	hashed, err := bcrypt.GenerateFromPassword([]byte(in.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, apperror.Internal("密码加密失败")
	}

	user := &model.User{
		Username: in.Username,
		Password: string(hashed),
		Role:     "user",
	}
	if err := s.db.Create(user).Error; err != nil {
		if strings.Contains(err.Error(), "duplicate key") || strings.Contains(err.Error(), "23505") {
			return nil, apperror.BadRequest("用户名已存在")
		}
		return nil, apperror.Internal("注册失败")
	}
	return user, nil
}

func (s *UserService) Login(in LoginInput) (string, error) {
	var user model.User
	if err := s.db.Where("username = ? OR email = ?", in.Username, in.Username).First(&user).Error; err != nil {
		return "", apperror.Unauthorized("用户名或密码错误")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(in.Password)); err != nil {
		return "", apperror.Unauthorized("用户名或密码错误")
	}
	return s.generateToken(&user)
}

func (s *UserService) LoginByEmail(email string) (string, error) {
	var user model.User
	if err := s.db.Where("email = ?", email).First(&user).Error; err != nil {
		return "", apperror.Unauthorized("用户不存在")
	}
	return s.generateToken(&user)
}

func (s *UserService) BindEmail(userID uint, email string) error {
	var exist model.User
	if err := s.db.Where("email = ?", email).First(&exist).Error; err == nil {
		return apperror.BadRequest("该邮箱已被绑定")
	}
	if err := s.db.Model(&model.User{}).Where("id = ?", userID).Update("email", email).Error; err != nil {
		return apperror.Internal("绑定邮箱失败")
	}
	return nil
}

func (s *UserService) GetProfile(userID uint) (*model.User, error) {
	var user model.User
	if err := s.db.First(&user, userID).Error; err != nil {
		return nil, apperror.NotFound("用户不存在")
	}
	return &user, nil
}

func (s *UserService) UpdateProfile(userID uint, in UpdateInput) (*model.User, error) {
	user, err := s.GetProfile(userID)
	if err != nil {
		return nil, err
	}
	return s.updateUser(user, in)
}

func (s *UserService) ListUsers(page, pageSize int) ([]model.User, int64, error) {
	var users []model.User
	var total int64
	s.db.Model(&model.User{}).Count(&total)
	offset := (page - 1) * pageSize
	if err := s.db.Offset(offset).Limit(pageSize).Find(&users).Error; err != nil {
		return nil, 0, apperror.Internal("查询失败")
	}
	return users, total, nil
}

func (s *UserService) GetUserByID(id uint) (*model.User, error) {
	return s.GetProfile(id)
}

func (s *UserService) UpdateUser(id uint, in UpdateInput) (*model.User, error) {
	user, err := s.GetProfile(id)
	if err != nil {
		return nil, err
	}
	return s.updateUser(user, in)
}

func (s *UserService) DeleteUser(id uint) error {
	if err := s.db.Delete(&model.User{}, id).Error; err != nil {
		return apperror.Internal("删除失败")
	}
	return nil
}

func (s *UserService) updateUser(user *model.User, in UpdateInput) (*model.User, error) {
	updates := map[string]interface{}{}
	if in.Password != "" {
		hashed, err := bcrypt.GenerateFromPassword([]byte(in.Password), bcrypt.DefaultCost)
		if err != nil {
			return nil, apperror.Internal("密码加密失败")
		}
		updates["password"] = string(hashed)
	}
	if in.Role != "" {
		updates["role"] = in.Role
	}
	if len(updates) == 0 {
		return user, nil
	}
	if err := s.db.Model(user).Updates(updates).Error; err != nil {
		return nil, apperror.Internal("更新失败")
	}
	return s.GetProfile(user.ID)
}

func (s *UserService) generateToken(user *model.User) (string, error) {
	claims := jwt.MapClaims{
		"user_id":  user.ID,
		"username": user.Username,
		"role":     user.Role,
		"exp":      jwt.NewNumericDate(time.Now().Add(s.cfg.Expire)).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.cfg.Secret))
}
