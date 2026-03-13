package handler

import (
	"errors"
	"my-live/internal/request"
	"my-live/internal/response"
	"my-live/internal/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	authSvc *service.AuthService
}

func NewAuthHandler(authSvc *service.AuthService) *AuthHandler {
	return &AuthHandler{
		authSvc: authSvc,
	}
}

// RegisterHandler 注册
func (h *AuthHandler) RegisterHandler(c *gin.Context) {
	var r request.CreateUserReq
	if err := c.ShouldBindJSON(&r); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	// 注册
	err := h.authSvc.Register(c, r)
	if err != nil {
		if errors.Is(err, service.ErrUsernameExists) {
			c.JSON(http.StatusConflict, gin.H{"error": "用户名已存在"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "注册失败"})
		}
		return
	}
	response.Success(c, nil)
}

// LoginHandler 登录
func (h *AuthHandler) LoginHandler(c *gin.Context) {
	var r request.CreateUserReq
	if err := c.ShouldBindJSON(&r); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	token, err := h.authSvc.Login(c, r)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户名或密码错误"})
		return
	}

	response.Success(c, token)
}
