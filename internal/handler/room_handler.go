package handler

import (
	"errors"
	"my-live/internal/request"
	"my-live/internal/response"
	"my-live/internal/service"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type RoomHandler struct {
	roomSvc *service.RoomService
}

func NewRoomHandler(roomSvc *service.RoomService) *RoomHandler {
	return &RoomHandler{roomSvc: roomSvc}
}

func (h *RoomHandler) CreateRoom(ctx *gin.Context) {
	userIDVal, exists := ctx.Get("user_id")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "请先登录"})
		return
	}
	hostID := userIDVal.(uint)

	var req request.CreateRoomReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.IsPrivate && req.Password == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "私密房间必须设置密码"})
		return
	}

	c := ctx.Request.Context()
	room, err := h.roomSvc.CreateRoom(c, hostID, req)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "创建房间失败: " + err.Error()})
		return
	}
	response.Success(ctx, room)
}

// ListRooms 获取房间列表
func (h *RoomHandler) ListRooms(ctx *gin.Context) {
	c := ctx.Request.Context()
	rooms, err := h.roomSvc.GetAllRooms(c)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "获取房间列表失败"})
		return
	}
	response.Success(ctx, rooms)
}

// GetRoom 获取房间详情
func (h *RoomHandler) GetRoom(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(ctx, "无效的房间 ID")
		return
	}

	c := ctx.Request.Context()
	room, err := h.roomSvc.GetRoom(c, uint(id))

	if errors.Is(err, gorm.ErrRecordNotFound) {
		response.Error(ctx, 404, "房间不存在")
		return
	}

	if room.IsPrivate {
		room.Password = "" // 前端不需要看到密码
	}
	response.Success(ctx, room)
}
