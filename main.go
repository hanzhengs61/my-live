package main

import (
	"fmt"
	"log"
	"my-live/internal/ws"
	"net/http"
)

func main() {
	// 启动WebSocket的Hub（消息中心）
	// Hub会在后台一直运行，负责把消息广播给所有客户端
	ws.GetHub().Start()

	// 注册WebSocket路由（任何人访问 /ws 都会进入WebSocket连接）
	http.HandleFunc("/ws", ws.WsHandler)

	fmt.Println("🚀 WebSocket服务器启动成功！")
	fmt.Println("访问地址: http://localhost:8080/ws")
	fmt.Println("测试页面: 把下面的 test-ws.html 用浏览器打开")

	// 启动HTTP服务器，监听8080端口（你可以改成你的项目端口）
	log.Fatal(http.ListenAndServe(":8080", nil))
}
