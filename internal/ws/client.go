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

// readPump 持续读取客户端发送的消息（也就说浏览器发消息给我们）
func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c // 告诉消息中心，这个客户端已经断开
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
			// 切换房间
			if c.roomID != "" {
				c.hub.unregister <- c
				msg.Nickname = c.Nickname
				c.hub.broadcast <- msg
			}
			c.roomID = msg.RoomID
			// 加入新房间
			c.hub.register <- c
			// 广播加入消息
			c.hub.broadcast <- Message{
				Type:   TypeJoin,
				RoomID: c.roomID,
			}
		case TypeChat, TypeGift:
			// 普通消息直接广播到本房间
			if c.roomID != "" {
				c.hub.broadcast <- msg
			}

		case TypePong:
			// 客户端回复pong（我们只处理，不广播）
		}
	}
}

// writePump 持续把消息写给客户端（消息中心Hub要广播给我们）
func (c *Client) writePump() {
	// 每30秒发送一次ping
	ticker := time.NewTicker(30 * time.Second)
	defer func() {
		ticker.Stop()
		_ = c.conn.Close()
		close(c.send)
	}()

	for {
		select {
		case message, ok := <-c.send:
			// 设置写入超时
			_ = c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				// 监听通道关闭，就关闭 WebSocket 连接
				_ = c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			// 真正的把消息发送给浏览器
			if err := c.conn.WriteMessage(websocket.TextMessage, message); err != nil {
				log.Printf("发送消息错误: %v", err)
				return
			}

		case <-ticker.C:
			// 发送心跳 ping
			_ = c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.conn.WriteMessage(websocket.PingMessage, []byte{}); err != nil {
				log.Printf("发送心跳错误: %v", err)
				return
			}
		}
	}
}
