package service

import (
	"fmt"
	"my-live/config"
	"my-live/internal/model"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

var (
	ErrUsernameExists     = fmt.Errorf("用户名已存在")
	ErrInvalidCredentials = fmt.Errorf("用户名或密码错误")
)

// AuthService 认证服务
type AuthService struct {
	db     *gorm.DB
	config *config.Config
}

// NewAuthService 创建服务实例
func NewAuthService(db *gorm.DB, cfg *config.Config) *AuthService {
	return &AuthService{
		db:     db,
		config: cfg,
	}
}

// Register 注册
func (s *AuthService) Register(username, password, nickname, email string) (*model.User, error) {
	// 检查用户名是否存在
	var count int64
	s.db.Model(&model.User{}).Where("username = ?", username).Count(&count)
	if count > 0 {
		return nil, ErrUsernameExists
	}

	// 加密密码
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := &model.User{
		Username:     username,
		PasswordHash: string(hashed),
		Nickname:     nickname,
		Email:        email,
		CreatedBy:    username,
		UpdatedBy:    username,
	}

	if err := s.db.Debug().Create(user).Error; err != nil {
		return nil, err
	}

	return user, nil
}

// Login 登录，返回 token 和用户信息
func (s *AuthService) Login(username, password string) (string, *model.User, error) {
	var user model.User
	if err := s.db.Where("username = ?", username).First(&user).Error; err != nil {
		return "", nil, ErrInvalidCredentials
	}

	// 验证密码
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return "", nil, ErrInvalidCredentials
	}

	// 生成 JWT
	claims := jwt.MapClaims{
		"user_id":  user.ID,
		"username": user.Username,
		"exp":      time.Now().Add(s.config.GetJWTExpiryDuration()).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(s.config.Server.JwtSecret))
	if err != nil {
		return "", nil, err
	}

	return tokenString, &user, nil
}
