package ws

import (
	"log"
	"time"

	"github.com/gorilla/websocket"
)

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
			if websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway) {
				return
			}
			log.Printf("读取错误: %v", err)
			return
		}

		msg, err := ParseMessage(data)
		if err != nil {
			continue
		}

		// 所有业务逻辑交给 business 层处理
		HandleMessage(c, msg)
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

func (c *Client) GetUsername() string {
	return c.Username
}
