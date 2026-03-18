package ws

import (
	"log"
	"my-live/internal/utils"
	"net/http"

	"github.com/gin-gonic/gin"
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
func WsHandler(c *gin.Context, w http.ResponseWriter, r *http.Request) {
	// 从 query 获取 token
	tokenStr := r.URL.Query().Get("token")
	if tokenStr == "" {
		http.Error(w, "缺少 token 参数", http.StatusUnauthorized)
		return
	}

	// 解析和验证 jwt
	claims, err := utils.ValidateJWT(tokenStr)
	if err != nil {
		http.Error(w, "token 无效或已过期: "+err.Error(), http.StatusUnauthorized)
		return
	}

	// 升级 HTTP 请求为 websocket 请求
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("升级WebSocket失败: %v", err)
		return
	}

	// 创建一个 Client 对象
	client := &Client{
		conn:     conn,
		send:     make(chan []byte, 256),
		hub:      GetHub(),
		UserID:   uint((*claims)["user_id"].(float64)), // jwt claims 是 float64，转 uint
		Username: (*claims)["username"].(string),
		Nickname: (*claims)["nickname"].(string), // 如果你 jwt 里带了 nickname
		IsAuthed: true,
	}

	// 启动两个 goroutine
	// 1. 持续读取客户端发送的消息（也就是浏览器发消息给我们）
	// 2. 持续把消息写给客户端（消息中心Hub要广播给我们）
	go client.readPump()
	go client.writePump()
}
