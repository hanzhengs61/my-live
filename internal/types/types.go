package types

import (
	"context"

	"github.com/gin-gonic/gin"
)

// Auditor 审计信息提供者
type Auditor interface {
	GetUsername() string
}

// GetUsernameFromCtx 从 context 中获取用户名（支持多种来源）
func GetUsernameFromCtx(ctx context.Context) string {
	if ctx == nil {
		return "system"
	}

	// HTTP
	if gc, ok := ctx.Value("gin_context").(*gin.Context); ok && gc != nil {
		if u, exists := gc.Get("username"); exists {
			if username, ok := u.(string); ok && username != "" {
				return username
			}
		}
	}

	// WebSocket
	if auditor, ok := ctx.Value("auditor").(Auditor); ok && auditor != nil {
		if username := auditor.GetUsername(); username != "" {
			return username
		}
	}

	return "system"
}
