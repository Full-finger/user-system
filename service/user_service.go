package service

import (
	"errors"
	"time"

	"github.com/full-finger/user-system/config"
	"github.com/full-finger/user-system/model"
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

func (s *UserService) Register(req *model.RegisterRequest) (*model.User, error) {
	var exist model.User
	if err := s.db.Where("username = ?", req.Username).First(&exist).Error; err == nil {
		return nil, errors.New("用户名已存在")
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := &model.User{
		Username: req.Username,
		Password: string(hashed),
		Role:     "user",
	}
	if err := s.db.Create(user).Error; err != nil {
		return nil, err
	}
	return user, nil
}

func (s *UserService) Login(req *model.LoginRequest) (string, error) {
	var user model.User
	if err := s.db.Where("username = ?", req.Username).First(&user).Error; err != nil {
		return "", errors.New("用户名或密码错误")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return "", errors.New("用户名或密码错误")
	}

	claims := jwt.MapClaims{
		"user_id":  user.ID,
		"username": user.Username,
		"role":     user.Role,
		"exp":      time.Now().Add(s.cfg.Expire).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.cfg.Secret))
}

func (s *UserService) GetProfile(userID uint) (*model.User, error) {
	var user model.User
	if err := s.db.First(&user, userID).Error; err != nil {
		return nil, errors.New("用户不存在")
	}
	return &user, nil
}

func (s *UserService) UpdateProfile(userID uint, req *model.UpdateRequest) (*model.User, error) {
	user, err := s.GetProfile(userID)
	if err != nil {
		return nil, err
	}
	return s.updateUser(user, req)
}

func (s *UserService) ListUsers() ([]model.User, error) {
	var users []model.User
	if err := s.db.Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}

func (s *UserService) GetUserByID(id uint) (*model.User, error) {
	return s.GetProfile(id)
}

func (s *UserService) UpdateUser(id uint, req *model.UpdateRequest) (*model.User, error) {
	user, err := s.GetProfile(id)
	if err != nil {
		return nil, err
	}
	return s.updateUser(user, req)
}

func (s *UserService) DeleteUser(id uint) error {
	return s.db.Delete(&model.User{}, id).Error
}

func (s *UserService) updateUser(user *model.User, req *model.UpdateRequest) (*model.User, error) {
	updates := map[string]interface{}{}
	if req.Password != "" {
		hashed, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			return nil, err
		}
		updates["password"] = string(hashed)
	}
	if req.Role != "" {
		updates["role"] = req.Role
	}
	if len(updates) == 0 {
		return user, nil
	}
	if err := s.db.Model(user).Updates(updates).Error; err != nil {
		return nil, err
	}
	return s.GetProfile(user.ID)
}
