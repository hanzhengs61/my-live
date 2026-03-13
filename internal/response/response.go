package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Response 统一响应结构（真实项目标准格式）
type Response struct {
	Code    int         `json:"code"`    // 0=成功，非0=失败
	Message string      `json:"message"` // 提示信息
	Data    interface{} `json:"data"`    // 业务数据（可空）
}

// Success 成功响应
func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code:    0,
		Message: "success",
		Data:    data,
	})
}

// Error 失败响应（带自定义错误码）
func Error(c *gin.Context, code int, message string) {
	c.JSON(http.StatusOK, Response{ // 注意：HTTP状态码仍用200，前端通过code判断
		Code:    code,
		Message: message,
		Data:    nil,
	})
}

// BadRequest 参数错误
func BadRequest(c *gin.Context, message string) {
	Error(c, 400, message)
}
