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
func (h *AuthHandler) RegisterHandler(ctx *gin.Context) {
	var r request.CreateUserReq
	if err := ctx.ShouldBindJSON(&r); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	// 注册
	c := ctx.Request.Context()
	err := h.authSvc.Register(c, r)
	if err != nil {
		if errors.Is(err, service.ErrUsernameExists) {
			ctx.JSON(http.StatusConflict, gin.H{"error": "用户名已存在"})
		} else {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "注册失败"})
		}
		return
	}
	response.Success(ctx, nil)
}

// LoginHandler 登录
func (h *AuthHandler) LoginHandler(ctx *gin.Context) {
	var r request.LoginReq
	if err := ctx.ShouldBindJSON(&r); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	token, err := h.authSvc.Login(ctx, r)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "用户名或密码错误"})
		return
	}

	response.Success(ctx, token)
}
