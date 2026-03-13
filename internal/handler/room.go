package handler

import (
	"my-live/internal/request"
	"my-live/internal/response"
	"my-live/internal/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

type RoomHandler struct {
	roomSvc *service.RoomService
}

func NewRoomHandler(roomSvc *service.RoomService) *RoomHandler {
	return &RoomHandler{roomSvc: roomSvc}
}

func (h *RoomHandler) CreateRoom(c *gin.Context) {
	userIDVal, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "请先登录"})
		return
	}
	hostID := userIDVal.(uint)

	var req request.CreateRoomReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.IsPrivate && req.Password == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "私密房间必须设置密码"})
		return
	}

	room, err := h.roomSvc.CreateRoom(c, hostID, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建房间失败: " + err.Error()})
		return
	}
	response.Success(c, room)
}

// ListRooms 获取房间列表
func (h *RoomHandler) ListRooms(c *gin.Context) {
	rooms, err := h.roomSvc.GetAllRooms(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取房间列表失败"})
		return
	}
	response.Success(c, rooms)
}
