package handler

import (
	"my-live/internal/request"
	"my-live/internal/response"
	"my-live/internal/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

type RechargeHandler struct {
}

func NewRechargeHandler() *RechargeHandler {
	return &RechargeHandler{}
}

func (h *RechargeHandler) Recharge(c *gin.Context) {
	// 从 JWT 中间件获取当前用户 ID
	userId, exists := c.Get("user_id")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "请先登录")
		return
	}
	userID := userId.(uint)

	var r request.UpdateRechargeReq
	if err := c.ShouldBindJSON(&r); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	rechargeSvc := service.NewRechargeService()
	if err := rechargeSvc.Recharge(c.Request.Context(), userID, r.Amount); err != nil {
		response.Error(c, 400, err.Error())
		return
	}

	response.Success(c, "充值成功")
}
