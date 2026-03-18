package ws

import (
	"context"
	"encoding/json"
	"log"
	"my-live/internal/model"
	"my-live/internal/redis"
	"my-live/internal/request"
	"my-live/internal/service"
	"strconv"
	"time"
)

var (
	roomService *service.RoomService
	chatService *service.ChatService
)

func InitServices(rs *service.RoomService, cs *service.ChatService) {
	roomService = rs
	chatService = cs
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
	case TypeRoomUpdate:
		handleRoomUpdate(c, msg)
	case TypeBanUser:
		handleBanUser(c, msg)
	case TypeKickUser:
		handleKickUser(c, msg)
	case TypeCloseRoom:
		handleCloseRoom(c, msg)
	default:
		log.Printf("未知消息类型: %s", msg.Type)
	}
}

// ==================== 具体业务处理 ====================

// 加入房间
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
		IsHost:    room.HostID == int64(c.UserID),
	}.ToJSON()
	// 广播用户列表更新
	c.hub.broadcastUserList(c.roomID)
	// 广播最新礼物排行榜
	rankSvc := service.NewRankService()
	roomIDInt, _ := strconv.ParseInt(c.roomID, 10, 64)
	ctx := context.WithValue(context.Background(), "auditor", c)
	rankList, _ := rankSvc.GetTopGiftRank(ctx, roomIDInt, 5)
	c.hub.broadcast <- Message{
		Type:      "gift_rank",
		RoomID:    c.roomID,
		Extra:     rankList,
		Timestamp: time.Now().UnixMilli(),
	}
	// 加载并发送历史消息
	if chatService != nil {
		messages, chatErr := chatService.GetRecentMessages(ctx, roomIDInt, 50)
		if chatErr == nil {
			for i := len(messages) - 1; i >= 0; i-- { // 倒序发送，保证时间顺序
				m := messages[i]
				c.send <- Message{
					Type:      MessageType(m.Type),
					RoomID:    strconv.FormatInt(m.RoomID, 10),
					UserID:    m.UserID,
					Nickname:  m.Nickname,
					Content:   m.Content,
					Timestamp: m.CreatedAt,
				}.ToJSON()
			}
		}
	}
}

// 发送弹幕
func handleChat(c *Client, msg Message) {
	if c.roomID == "" {
		c.send <- Message{Type: "error", Content: "请先加入房间"}.ToJSON()
		return
	}

	msg.RoomID = c.roomID
	msg.UserID = int64(c.UserID)
	msg.Nickname = c.Nickname
	// 广播给房间所有人
	c.hub.broadcast <- msg
	if chatService != nil {
		roomIDInt, _ := strconv.ParseInt(c.roomID, 10, 64)
		record := &model.ChatMessage{
			Type:     "chat",
			RoomID:   roomIDInt,
			UserID:   msg.UserID,
			Nickname: msg.Nickname,
			Content:  msg.Content,
		}
		go func() {
			ctx := context.WithValue(context.Background(), "auditor", c)
			if err := chatService.SaveMessage(ctx, record); err != nil {
				log.Printf("保存聊天消息失败: %v", err)
			}
		}()
	}
}

func handleGift(c *Client, msg Message) {
	if c.roomID == "" {
		c.send <- Message{Type: "error", Content: "请先加入房间才能送礼物"}.ToJSON()
		return
	}

	giftID := msg.GiftID
	count := msg.Count
	if count <= 0 {
		count = 1
	}
	roomIDInt, _ := strconv.ParseInt(c.roomID, 10, 64)
	req := request.SendGiftReq{
		RoomID: roomIDInt,
		GiftID: giftID,
		Count:  int64(count),
	}
	giftSvc := service.NewGiftService()
	ctx := context.WithValue(context.Background(), "auditor", c)
	giftName, err := giftSvc.SendGift(ctx, c.UserID, req)
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
			GiftID:    giftID,
			Content:   "送出礼物" + giftName + " x" + strconv.FormatInt(int64(count), 10),
			Timestamp: time.Now().UnixMilli(),
		}
	}()

	// 广播最新礼物排行榜
	rankSvc := service.NewRankService()
	rankList, _ := rankSvc.GetTopGiftRank(ctx, roomIDInt, 5)
	c.hub.broadcast <- Message{
		Type:      "gift_rank",
		RoomID:    c.roomID,
		Extra:     rankList,
		Timestamp: time.Now().UnixMilli(),
	}
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

	// 广播用户列表更新
	c.hub.broadcastUserList(c.roomID)
}

func handleRoomUpdate(c *Client, msg Message) {
	roomIDInt, _ := strconv.ParseInt(c.roomID, 10, 64)

	update := request.UpdateRoomReq{}
	if err := json.Unmarshal([]byte(msg.Content), &update); err != nil {
		c.send <- Message{Type: "error", Content: "格式错误"}.ToJSON()
		return
	}

	roomSvc := service.NewRoomService()
	ctx := context.WithValue(context.Background(), "auditor", c)
	if err := roomSvc.UpdateRoom(ctx, roomIDInt, int64(c.UserID), update.Title, update.Announcement, false, ""); err != nil {
		c.send <- Message{Type: "error", Content: err.Error()}.ToJSON()
		return
	}

	// 广播更新（全房间）
	c.hub.broadcast <- Message{
		Type:      "system",
		RoomID:    c.roomID,
		Content:   "房间信息已更新",
		Timestamp: time.Now().UnixMilli(),
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

func handleBanUser(c *Client, msg Message) {
	targetID := msg.TargetID
	if targetID <= 0 {
		return
	}
	c.hub.broadcast <- Message{
		Type:      "system",
		RoomID:    c.roomID,
		Content:   "用户已被禁言 5 分钟",
		Timestamp: time.Now().UnixMilli(),
	}
}

func handleKickUser(c *Client, msg Message) {
	targetID := msg.TargetID
	if targetID <= 0 {
		return
	}

	c.hub.mu.RLock()
	for client := range c.hub.rooms[c.roomID] {
		if client.UserID == uint(targetID) {
			client.send <- Message{Type: "error", Content: "您已被踢出房间"}.ToJSON()
			c.hub.unregister <- client
			break
		}
	}
	c.hub.mu.RUnlock()

	c.hub.broadcast <- Message{
		Type:      "system",
		RoomID:    c.roomID,
		Content:   "用户已被踢出房间",
		Timestamp: time.Now().UnixMilli(),
	}
}

func handleCloseRoom(c *Client, msg Message) {
	c.hub.broadcast <- Message{
		Type:      "system",
		RoomID:    c.roomID,
		Content:   "房间已被主播关闭",
		Timestamp: time.Now().UnixMilli(),
	}
	// 可选：清空房间所有客户端
}
