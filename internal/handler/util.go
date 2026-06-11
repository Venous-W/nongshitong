package handler

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

// apiOK 统一成功响应格式：{"code":0,"data":...,"msg":"ok"}
func apiOK(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": data, "msg": "ok"})
}

// apiFail 统一失败响应格式：{"code":1,"data":null,"msg":"错误原因"}
func apiFail(c *gin.Context, msg string) {
	c.JSON(http.StatusOK, gin.H{"code": 1, "data": nil, "msg": msg})
}

// parseID 从 URL 路径参数中解析整数 id，解析失败时直接响应错误并返回 false。
func parseID(c *gin.Context, param string) (int64, bool) {
	var id int64
	if _, err := fmt.Sscan(c.Param(param), &id); err != nil || id <= 0 {
		apiFail(c, "无效的 id 参数")
		return 0, false
	}
	return id, true
}
