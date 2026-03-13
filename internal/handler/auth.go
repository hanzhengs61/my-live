package handler

import (
	"errors"
	"my-live/internal/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

// RegisterHandler 注册
func RegisterHandler(authSvc *service.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		type req struct {
			Username string `json:"username" binding:"required,min=4,max=50"`
			Password string `json:"password" binding:"required,min=6"`
			Nickname string `json:"nickname" binding:"required,min=2,max=50"`
			Email    string `json:"email" binding:"omitempty,email"`
		}

		var r req
		if err := c.ShouldBindJSON(&r); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		// 注册
		user, err := authSvc.Register(r.Username, r.Password, r.Nickname, r.Email)
		if err != nil {
			if errors.Is(err, service.ErrUsernameExists) {
				c.JSON(http.StatusConflict, gin.H{"error": "用户名已存在"})
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "注册失败"})
			}
			return
		}

		c.JSON(http.StatusCreated, gin.H{
			"message":  "注册成功",
			"user_id":  user.ID,
			"username": user.Username,
			"nickname": user.Nickname,
			"email":    user.Email,
		})
	}
}

// LoginHandler 登录
func LoginHandler(authSvc *service.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		type req struct {
			Username string `json:"username" binding:"required"`
			Password string `json:"password" binding:"required"`
		}

		var r req
		if err := c.ShouldBindJSON(&r); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		token, user, err := authSvc.Login(r.Username, r.Password)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "用户名或密码错误"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"token":    token,
			"user_id":  user.ID,
			"username": user.Username,
			"nickname": user.Nickname,
		})
	}
}
