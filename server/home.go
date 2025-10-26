package server

import (
	"networkDev/controllers/home"

	"github.com/gin-gonic/gin"
)

// RegisterHomeRoutes 注册主页路由
// 只包含根路径，用于主页功能
func RegisterHomeRoutes(router *gin.Engine) {
	// 根路径 - 主页
	router.GET("/", home.RootHandler)
}
