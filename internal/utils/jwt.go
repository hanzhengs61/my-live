package utils

import (
	"errors"
	"my-live/config"

	"github.com/golang-jwt/jwt/v5"
)

// ValidateJWT 验证 jwt
func ValidateJWT(tokenStr string) (*jwt.MapClaims, error) {
	// 获取配置
	cfg := config.LoadConfig()

	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		// 验证签名方法
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("无效的签名方法")
		}
		return []byte(cfg.Server.JwtSecret), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return &claims, nil
	}

	return nil, errors.New("token 无效")
}
