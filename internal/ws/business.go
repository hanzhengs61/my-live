package ws

import (
	"context"
	"log"
	"my-live/internal/redis"
	"my-live/internal/request"
	"my-live/internal/service"
	"strconv"
	"time"
)

var roomService *service.RoomService

func InitServices(rs *service.RoomService) {
	roomService = rs
}

// HandleMessage 统一处理所有消息
func HandleMessage(c *Client, msg Message) {
	switch msg.Type {
	case TypeJoin:
		handleJoin(c, msg)
	case TypeChat:
		handleChat(c, msg)
	case TypeGift:
		handleGift(c, msg)
	case TypeLeave:
		handleLeave(c)
	case TypePong: // 心跳
	case TypePrivateChat:
		handlePrivateChat(c, msg)
	default:
		log.Printf("未知消息类型: %s", msg.Type)
	}
}

func handlePrivateChat(c *Client, msg Message) {
	if c.roomID == "" {
		c.send <- Message{Type: "error", Content: "请先加入房间"}.ToJSON()
		return
	}

	targetUserID := msg.TargetID
	if targetUserID <= 0 {
		c.send <- Message{Type: "error", Content: "无效的目标用户 ID"}.ToJSON()
		return
	}

	// 构造私聊消息（只发给目标用户）
	privateMsg := Message{
		Type:      TypePrivateChat,
		RoomID:    c.roomID,
		UserID:    int64(c.UserID),
		Nickname:  c.Nickname,
		Content:   msg.Content,
		TargetID:  targetUserID,
		Timestamp: time.Now().UnixMilli(),
	}

	found := false
	c.hub.mu.RLock()
	for client := range c.hub.rooms[c.roomID] {
		if client.UserID == uint(targetUserID) {
			client.send <- privateMsg.ToJSON()
			found = true
			break
		}
	}
	c.hub.mu.RUnlock()
	if !found {
		c.send <- Message{Type: "error", Content: "目标用户不在当前房间"}.ToJSON()
		return
	}

	c.send <- Message{
		Type:      "system",
		Content:   "私聊消息:" + privateMsg.Content,
		Timestamp: time.Now().UnixMilli(),
	}.ToJSON()
}

// ==================== 具体业务处理 ====================

func handleJoin(c *Client, msg Message) {
	roomIDUint, err := strconv.ParseUint(msg.RoomID, 10, 32)
	if err != nil {
		c.send <- Message{Type: "error", Content: "无效的房间 ID"}.ToJSON()
		return
	}

	room, err := roomService.VerifyJoinRoom(uint(roomIDUint), msg.Content)
	if err != nil {
		log.Printf("加入失败: roomID=%d, user=%s, err=%v", roomIDUint, c.Nickname, err)
		c.send <- Message{Type: "error", Content: err.Error()}.ToJSON()
		return
	}

	// 离开旧房间
	if c.roomID != "" {
		c.hub.broadcast <- Message{
			Type:     TypeLeave,
			RoomID:   c.roomID,
			Nickname: c.Nickname,
			Content:  c.Nickname + " 离开了房间",
		}
		c.hub.unregister <- c
	}

	c.roomID = msg.RoomID
	c.hub.register <- c

	_ = redis.JoinRoom(c.roomID, c.UserID)

	c.hub.broadcast <- Message{
		Type:      TypeJoin,
		RoomID:    c.roomID,
		UserID:    int64(c.UserID),
		Nickname:  c.Nickname,
		Content:   c.Nickname + " 进入了房间",
		Timestamp: time.Now().UnixMilli(),
	}

	c.send <- Message{
		Type:      "system",
		Content:   "欢迎来到《" + room.Title + "》",
		Timestamp: time.Now().UnixMilli(),
	}.ToJSON()
}

func handleChat(c *Client, msg Message) {
	if c.roomID == "" {
		c.send <- Message{Type: "error", Content: "请先加入房间"}.ToJSON()
		return
	}

	msg.RoomID = c.roomID
	msg.UserID = int64(c.UserID)
	msg.Nickname = c.Nickname
	msg.Timestamp = time.Now().UnixMilli()

	c.hub.broadcast <- msg
}

func handleGift(c *Client, msg Message) {
	if c.roomID == "" {
		c.send <- Message{Type: "error", Content: "请先加入房间才能送礼物"}.ToJSON()
		return
	}

	count := int64(1)
	if msg.Count > 0 {
		count = int64(msg.Count)
	}
	roomIDInt, _ := strconv.ParseInt(c.roomID, 10, 64)
	req := request.SendGiftReq{
		RoomID: roomIDInt,
		GiftID: msg.GiftID,
		Count:  count,
	}
	giftSvc := service.NewGiftService()
	ctx := context.WithValue(context.Background(), "auditor", c)
	giftName, err := giftSvc.SendGift(ctx, int64(c.UserID), req)
	if err != nil {
		c.send <- Message{Type: "error", Content: err.Error()}.ToJSON()
		return
	}

	// 广播礼物特效
	go func() {
		c.hub.broadcast <- Message{
			Type:      TypeGift,
			RoomID:    c.roomID,
			UserID:    int64(c.UserID),
			Nickname:  c.Nickname,
			Content:   giftName,
			Timestamp: time.Now().UnixMilli(),
		}
	}()
}

func handleLeave(c *Client) {
	if c.roomID == "" {
		return
	}

	_ = redis.LeaveRoom(c.roomID, c.UserID)

	c.hub.broadcast <- Message{
		Type:      TypeLeave,
		RoomID:    c.roomID,
		Nickname:  c.Nickname,
		Content:   c.Nickname + " 离开了房间",
		Timestamp: time.Now().UnixMilli(),
	}

	count, _ := redis.GetOnlineCount(c.roomID)
	c.hub.broadcast <- Message{
		Type:        TypeOnline,
		RoomID:      c.roomID,
		OnlineCount: int(count),
		Timestamp:   time.Now().UnixMilli(),
	}

	c.hub.unregister <- c
	c.roomID = ""
}
