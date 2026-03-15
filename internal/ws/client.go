package ws

import (
	"log"
	"my-live/internal/redis"
	"my-live/internal/service"
	"strconv"
	"time"

	"github.com/gorilla/websocket"
)

var roomService *service.RoomService

func InitServices(rs *service.RoomService) {
	roomService = rs
}

// Client 代表一个浏览器/WebSocket连接
type Client struct {
	conn   *websocket.Conn // 真正的 WebSocket 连接对象
	send   chan []byte     // 这个客户端要发送的消息队列（缓冲通道）
	hub    *Hub            // 指向消息中心
	roomID string          // 当前所在房间 ID

	// ---------------认证字段---------------
	UserID   uint   `json:"user_id"`  // 来自 JWT 的 user_id
	Username string `json:"username"` // 用户名
	Nickname string `json:"nickname"` // 昵称
	IsAuthed bool   // 是否已通过认证
}

// readPump 持续读取客户端发送的消息（浏览器发消息给我们）
func (c *Client) readPump() {
	defer func() {
		if c.roomID != "" {
			_ = redis.LeaveRoom(c.roomID, c.UserID)

			c.hub.broadcast <- Message{
				Type:     TypeLeave,
				RoomID:   c.roomID,
				Nickname: c.Nickname,
				Content:  c.Nickname + " 离开了房间",
			}

			count, _ := redis.GetOnlineCount(c.roomID)
			c.hub.broadcast <- Message{
				Type:        TypeOnline,
				RoomID:      c.roomID,
				OnlineCount: int(count),
			}

			c.hub.unregister <- c
		}
		_ = c.conn.Close()
	}()

	// 设置读取超时（防止客户端假死 60秒）
	_ = c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.conn.SetPongHandler(func(string) error {
		_ = c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		_, data, err := c.conn.ReadMessage()
		if err != nil {
			// 判断是否是正常关闭（浏览器关闭页面就是1001）
			if websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway) ||
				websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) == false {
				// 正常关闭，不打印错误日志
				return
			}
			log.Printf("读取错误: %v", err)
			return
		}

		msg, err := ParseMessage(data)
		if err != nil {
			continue
		}

		switch msg.Type {
		case TypeJoin:
			// 1. 获取房间ID
			roomID, err := strconv.ParseUint(msg.RoomID, 10, 32)
			if err != nil {
				c.send <- Message{Type: "error", Content: "无效的房间 ID"}.ToJSON()
				continue
			}
			// 2. 校验是否能加入
			password := msg.Content
			room, errV := roomService.VerifyJoinRoom(uint(roomID), password)
			if errV != nil {
				log.Printf("加入房间失败 - roomID=%d, user=%s, err=%v", roomID, c.Nickname, err)
				c.send <- Message{Type: "error", Content: err.Error()}.ToJSON()
				continue
			}

			// 如果之前在房间，先退出
			if c.roomID != "" {
				c.hub.unregister <- c
			}

			// 设置新房间
			c.roomID = msg.RoomID

			// 注册到 Hub
			c.hub.register <- c

			// Redis 记录在线
			_ = redis.JoinRoom(c.roomID, c.UserID)

			// 广播加入
			c.hub.broadcast <- Message{
				Type:      TypeJoin,
				RoomID:    c.roomID,
				UserID:    int64(c.UserID),
				Nickname:  c.Nickname,
				Content:   c.Nickname + " 进入了房间",
				Timestamp: time.Now().UnixMilli(),
			}

			// 给自己发欢迎消息
			c.send <- Message{
				Type:      "system",
				RoomID:    c.roomID,
				Nickname:  c.Nickname,
				Content:   "欢迎来到《" + room.Title + "》",
				Timestamp: time.Now().UnixMilli(),
			}.ToJSON()

		case TypeChat, TypeGift:
			if c.roomID == "" {
				c.send <- Message{
					Type:      "error",
					Content:   "请先加入房间才能发送消息",
					Timestamp: time.Now().UnixMilli(),
				}.ToJSON()
				continue
			}

			msg.RoomID = c.roomID
			msg.UserID = int64(c.UserID)
			msg.Nickname = c.Nickname
			msg.Timestamp = time.Now().UnixMilli()
			// 广播消息
			c.hub.broadcast <- msg
		case TypePong:
			// 客户端回复 pong，不广播
		case TypeLeave:
			if c.roomID == "" {
				c.send <- Message{Type: "error", Content: "当前不在任何房间"}.ToJSON()
				continue
			}

			_ = redis.LeaveRoom(c.roomID, c.UserID)

			c.hub.broadcast <- Message{
				Type:      TypeLeave,
				RoomID:    c.roomID,
				UserID:    int64(c.UserID),
				Nickname:  c.Nickname,
				Content:   c.Nickname + " 离开了房间",
				Timestamp: time.Now().UnixMilli(),
			}

			c.hub.unregister <- c
			c.roomID = ""
		}
	}
}

// writePump 持续把消息写给客户端（消息中心Hub要广播给我们）
func (c *Client) writePump() {
	ticker := time.NewTicker(30 * time.Second) // 每30秒发一次 ping
	defer func() {
		ticker.Stop()
		_ = c.conn.Close()
		close(c.send)
	}()

	for {
		select {
		case message, ok := <-c.send:
			_ = c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				_ = c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			if err := c.conn.WriteMessage(websocket.TextMessage, message); err != nil {
				log.Printf("发送消息错误: %v", err)
				return
			}

		case <-ticker.C:
			_ = c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.conn.WriteMessage(websocket.PingMessage, []byte{}); err != nil {
				log.Printf("发送心跳错误: %v", err)
				return
			}
		}
	}
}
