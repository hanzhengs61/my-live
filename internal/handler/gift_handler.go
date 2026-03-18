package handler

import (
	"my-live/internal/request"
	"my-live/internal/response"
	"my-live/internal/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

type GiftHandler struct {
	giftSvc *service.GiftService
}

func NewGiftHandler(giftSvc *service.GiftService) *GiftHandler {
	return &GiftHandler{giftSvc: giftSvc}
}

func (h *GiftHandler) SendGift(c *gin.Context) {
	// 从 JWT 中间件获取当前用户 ID
	userId, exists := c.Get("user_id")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "请先登录")
		return
	}
	userID := userId.(uint)

	var req request.SendGiftReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	gift, err := h.giftSvc.SendGift(c.Request.Context(), userID, req)
	if err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	response.Success(c, gift)
}
