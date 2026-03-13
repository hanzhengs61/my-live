package db

import (
	"context"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// DB 全局数据库实例（带上下文）
var DB *gorm.DB

// InitDB 初始化全局 DB，并注册回调以自动携带 gin 上下文
func InitDB(db *gorm.DB) {
	DB = db

	// 注册全局回调：在每次创建新 session 时，自动从 context 里取 gin.Context
	DB.Callback().Create().Before("gorm:create").Register("audit:before_create", func(db *gorm.DB) {
		if ginCtx, ok := db.Statement.Context.Value("gin_context").(*gin.Context); ok && ginCtx != nil {
			if username, exists := ginCtx.Get("username"); exists {
				if un, ok := username.(string); ok && un != "" {
					if bm, ok := db.Statement.Dest.(interface{ SetCreatedBy(string) }); ok {
						bm.SetCreatedBy(un)
					}
				}
			}
		} else {
			// 默认使用 "system"
			if bm, ok := db.Statement.Dest.(interface{ SetCreatedBy(string) }); ok {
				bm.SetCreatedBy("system")
			}
		}
	})

	DB.Callback().Update().Before("gorm:update").Register("audit:before_update", func(db *gorm.DB) {
		if ginCtx, ok := db.Statement.Context.Value("gin_context").(*gin.Context); ok && ginCtx != nil {
			if username, exists := ginCtx.Get("username"); exists {
				if un, ok := username.(string); ok && un != "" {
					if bm, ok := db.Statement.Dest.(interface{ SetUpdatedBy(string) }); ok {
						bm.SetUpdatedBy(un)
					}
				}
			}
		} else {
			// 默认使用 "system"
			if bm, ok := db.Statement.Dest.(interface{ SetUpdatedBy(string) }); ok {
				bm.SetUpdatedBy("system")
			}
		}
	})
}

// WithContext 为当前 DB 操作绑定 gin 上下文
func WithContext(ctx context.Context) *gorm.DB {
	return DB.WithContext(ctx)
}
