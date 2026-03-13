package db

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

var DB *gorm.DB // 全局数据库实例

// Init 初始化全局 DB，并自动注入 gin 上下文
func Init(database *gorm.DB) {
	DB = database

	// 全局回调：在每次操作前自动注入上下文
	DB.Callback().Create().Before("gorm:create").Register("auto_context", injectContext)
	DB.Callback().Update().Before("gorm:update").Register("auto_context", injectContext)
	DB.Callback().Query().Before("gorm:query").Register("auto_context", injectContext)
}

func injectContext(db *gorm.DB) {
	// 从当前请求上下文取 gin.Context
	if ginCtxVal := db.Statement.Context.Value("gin_context"); ginCtxVal != nil {
		if ginCtx, ok := ginCtxVal.(*gin.Context); ok && ginCtx != nil {
			db.Statement.Context = ginCtx.Request.Context()
		}
	}
}
