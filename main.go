package main

import (
	"fmt"
	"log"
	"my-live/internal/ws"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

func main() {
	// 启动WebSocket的Hub（消息中心）
	// Hub会在后台一直运行，负责把消息广播给所有客户端
	ws.GetHub().Start()

	// 创建 gin 引擎（默认待 Logger 和 Recovery 中间件）
	r := gin.Default()

	// 注册 WebSocket 路由
	r.GET("/ws", func(c *gin.Context) {
		ws.WsHandler(c.Writer, c.Request)
	})

	// 一个测试用的 HTTP 接口（后续扩展用户、房间等）
	r.GET("/api/health", func(c *gin.Context) {
		// 可以在这里返回一些基本信息，方便调试
		onlineRooms := ws.GetHub().Rooms()
		c.JSON(http.StatusOK, gin.H{
			"status":       "healthy",
			"message":      "成人直播系统原型运行中",
			"online_rooms": onlineRooms,
			"timestamp":    time.Now().Format(time.DateTime),
		})
	})

	port := ":8080"
	fmt.Printf("服务器启动成功！监听端口 %s\n", port)
	fmt.Println("WebSocket 地址: ws://localhost:8080/ws")
	fmt.Println("健康检查: http://localhost:8080/api/health")
	if err := r.Run(port); err != nil {
		log.Fatalf("服务器启动失败: %v", err)
	}
}
