package service

import (
	"context"
	"fmt"
	"my-live/config"
	"my-live/internal/db"
	"my-live/internal/model"
	"my-live/internal/request"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrUsernameExists     = fmt.Errorf("用户名已存在")
	ErrInvalidCredentials = fmt.Errorf("用户名或密码错误")
)

// AuthService 认证服务
type AuthService struct {
	config *config.Config
}

// NewAuthService 创建服务实例
func NewAuthService(cfg *config.Config) *AuthService {
	return &AuthService{
		config: cfg,
	}
}

// Register 注册
func (s *AuthService) Register(ctx context.Context, r request.CreateUserReq) error {
	// 检查用户名是否存在
	var count int64
	db.DB.WithContext(ctx).Model(&model.User{}).Where("username = ?", r.Username).Count(&count)
	if count > 0 {
		return ErrUsernameExists
	}

	// 加密密码
	hashed, err := bcrypt.GenerateFromPassword([]byte(r.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	user := &model.User{
		Username:     r.Username,
		PasswordHash: string(hashed),
		Nickname:     r.Nickname,
		Email:        r.Email,
	}

	if err := db.DB.WithContext(ctx).Debug().Create(user).Error; err != nil {
		return err
	}
	return nil
}

// Login 登录，返回 token 和用户信息
func (s *AuthService) Login(ctx *gin.Context, r request.LoginReq) (string, error) {
	var user model.User
	if err := db.DB.WithContext(ctx).Where("username = ?", r.Username).First(&user).Error; err != nil {
		return "", ErrInvalidCredentials
	}

	// 验证密码
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(r.Password)); err != nil {
		return "", ErrInvalidCredentials
	}

	// 生成 JWT
	claims := jwt.MapClaims{
		"user_id":  user.ID,
		"username": user.Username,
		"nickname": user.Nickname,
		"exp":      time.Now().Add(s.config.GetJWTExpiryDuration()).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(s.config.Server.JwtSecret))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}
