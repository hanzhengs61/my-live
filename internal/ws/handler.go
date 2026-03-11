package ws

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

// 升级 HTTP 请求为 websocket 请求
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// 开发阶段允许所有来源
		return true
	},
}

// WsHandler WebSocket入口 handler
func WsHandler(w http.ResponseWriter, r *http.Request) {
	// 升级 HTTP 请求为 websocket 请求
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("升级WebSocket失败: %v", err)
		return
	}

	// 创建一个 Client 对象
	client := &Client{
		conn: conn,
		send: make(chan []byte, 256),
		hub:  GetHub(),
	}

	// 告诉消息消息中心Hub：新客户端来了
	client.hub.register <- client

	// 启动两个 goroutine
	// 1. 持续读取客户端发送的消息（也就是浏览器发消息给我们）
	// 2. 持续把消息写给客户端（消息中心Hub要广播给我们）
	go client.readPump()
	go client.writePump()
}
