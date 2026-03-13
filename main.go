package main

import (
	"fmt"
	"log"
	"my-live/config"
	mydb "my-live/internal/db"
	"my-live/internal/handler"
	"my-live/internal/middleware"
	"my-live/internal/model"
	"my-live/internal/service"
	"my-live/internal/ws"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	// 加载配置
	cfg := config.LoadConfig()

	// 初始化数据库连接
	db, err := gorm.Open(postgres.Open(cfg.Database.DSN), &gorm.Config{})
	if err != nil {
		log.Fatal("数据库连接失败:", err)
	}

	// 自动迁移
	err = db.AutoMigrate(&model.User{}, &model.Room{})
	if err != nil {
		log.Fatal("数据库迁移失败:", err)
	}
	mydb.InitDB(db)
	// 创建 auth 服务
	authSvc := service.NewAuthService(db, cfg)
	authHandler := handler.NewAuthHandler(authSvc)
	roomSvc := service.NewRoomService()
	roomHandler := handler.NewRoomHandler(roomSvc)

	// 启动WebSocket的Hub（消息中心）
	// Hub会在后台一直运行，负责把消息广播给所有客户端
	ws.GetHub().Start()

	// 创建 gin 引擎（默认待 Logger 和 Recovery 中间件）
	r := gin.Default()

	// 注册 WebSocket 路由
	r.GET("/ws", func(c *gin.Context) {
		ws.WsHandler(c.Writer, c.Request)
	})

	r.POST("/api/register", authHandler.RegisterHandler)
	r.POST("/api/login", authHandler.LoginHandler)

	api := r.Group("/api")
	api.Use(middleware.AuthMiddleware())

	api.POST("/createRoom", roomHandler.CreateRoom)
	api.POST("/listRooms", roomHandler.ListRooms)

	fmt.Printf("服务器启动成功！监听端口 %s\n", cfg.Server.Port)
	if runErr := r.Run(cfg.Server.Port); runErr != nil {
		log.Fatalf("服务器启动失败: %v", runErr)
	}
}
