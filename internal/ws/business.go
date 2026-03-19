package ws

import (
	"context"
	"encoding/json"
	"fmt"
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

	case TypeJoin: // 加入房间
		handleJoin(c, msg)
	case TypeChat: // 发送弹幕
		handleChat(c, msg)
	case TypeGift: // 发送礼物
		handleGift(c, msg)
	case TypeLeave: // 离开房间
		handleLeave(c)
	case TypePong: // 心跳
	case TypePrivateChat: // 私聊
		handlePrivateChat(c, msg)
	case TypeRoomUpdate: // 房间更新
		handleRoomUpdate(c, msg)
	case TypeBanUser: // 封禁用户
		handleBanUser(c, msg)
	case TypeKickUser: // 踢人
		handleKickUser(c, msg)
	case TypeCloseRoom: // 关闭房间
		handleCloseRoom(c, msg)
	default:
		log.Printf("未知消息类型: %s", msg.Type)
	}
}

// ==================== 具体业务处理 ====================

// 加入房间
func handleJoin(c *Client, msg Message) {

	// 踢人
	roomKey := fmt.Sprintf("live:room:%s:kicked_users", msg.RoomID)
	ctx := context.Background()

	isKicked, err := redis.Client.SIsMember(ctx, roomKey, c.UserID).Result()
	if err == nil && isKicked {
		c.send <- Message{Type: "error", Content: "您已被主播踢出房间，30分钟内无法加入"}.ToJSON()
		return
	}

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

	banKey := fmt.Sprintf("live:room:%s:banned_users", c.roomID)
	ctx := context.WithValue(context.Background(), "auditor", c)
	isBanned, err := redis.Client.SIsMember(ctx, banKey, c.UserID).Result()
	if err == nil && isBanned {
		c.send <- Message{Type: "error", Content: "您已被禁言，无法发言"}.ToJSON()
		return
	}

	msg.RoomID = c.roomID
	msg.UserID = int64(c.UserID)
	msg.Nickname = c.Nickname
	msg.Timestamp = time.Now().Unix()
	// 广播给房间所有人
	c.hub.broadcast <- msg
	if chatService != nil {
		roomIDInt, _ := strconv.ParseInt(c.roomID, 10, 64)
		record := &model.ChatMessage{
			Type:      "chat",
			RoomID:    roomIDInt,
			UserID:    msg.UserID,
			Nickname:  msg.Nickname,
			Content:   msg.Content,
			Timestamp: msg.Timestamp,
		}
		go func() {
			if SaveMessageErr := chatService.SaveMessage(ctx, record); SaveMessageErr != nil {
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

func handleRoomUpdate(c *Client, msg Message) {
	roomIDInt, _ := strconv.ParseInt(c.roomID, 10, 64)

	update := request.UpdateRoomReq{}
	if err := json.Unmarshal([]byte(msg.Content), &update); err != nil {
		c.send <- Message{Type: "error", Content: "格式错误"}.ToJSON()
		return
	}

	ctx := context.WithValue(context.Background(), "auditor", c)
	if err := roomService.UpdateRoom(ctx, roomIDInt, int64(c.UserID), update.Title, update.Announcement, false, ""); err != nil {
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

func handleBanUser(c *Client, msg Message) {
	targetID := msg.TargetID
	if targetID <= 0 {
		return
	}

	banKey := fmt.Sprintf("live:room:%s:banned_users", c.roomID)
	ctx := context.WithValue(context.Background(), "auditor", c)
	_, err := redis.Client.SAdd(ctx, banKey, targetID).Result()
	if err != nil {
		c.send <- Message{Type: "error", Content: "禁言失败"}.ToJSON()
		return
	}
	// 设置过期时间（5分钟后自动解除）
	redis.Client.Expire(ctx, banKey, 5*time.Minute)

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
		c.send <- Message{Type: "error", Content: "无效的用户 ID"}.ToJSON()
		return
	}

	// 踢人
	roomKey := fmt.Sprintf("live:room:%s:kicked_users", c.roomID)
	ctx := context.WithValue(context.Background(), "auditor", c)

	// 添加到踢人集合，并设置 30 分钟过期
	_, err := redis.Client.SAdd(ctx, roomKey, targetID).Result()
	if err != nil {
		c.send <- Message{Type: "error", Content: "踢人失败"}.ToJSON()
		return
	}
	redis.Client.Expire(ctx, roomKey, 30*time.Minute) // 30分钟内无法重连

	// 强制关闭当前连接
	c.hub.mu.RLock()
	for client := range c.hub.rooms[c.roomID] {
		if client.UserID == uint(targetID) {
			client.send <- Message{Type: "error", Content: "您已被主播踢出房间（30分钟内无法重连）"}.ToJSON()
			client.conn.Close()
			c.hub.unregister <- client
			break
		}
	}
	c.hub.mu.RUnlock()

	c.hub.broadcast <- Message{
		Type:      "system",
		RoomID:    c.roomID,
		Content:   "用户已被踢出房间（30分钟内无法重连）",
		Timestamp: time.Now().UnixMilli(),
	}
}

func handleCloseRoom(c *Client, msg Message) {
	c.hub.broadcast <- Message{
		Type:      "system",
		RoomID:    c.roomID,
		Content:   "房间已被主播关闭，所有用户已被请出",
		Timestamp: time.Now().UnixMilli(),
	}

	// 清空房间所有客户端
	c.hub.mu.Lock()
	if clients, ok := c.hub.rooms[c.roomID]; ok {
		for client := range clients {
			client.send <- Message{Type: "error", Content: "房间已被关闭"}.ToJSON()
			c.hub.unregister <- client
		}
		delete(c.hub.rooms, c.roomID)
	}
	c.hub.mu.Unlock()
}
