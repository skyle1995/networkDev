package server

import (
	"net/http"
	"networkDev/controllers/home"
)

// RegisterHomeRoutes 注册主页路由
// 只包含根路径，用于主页功能
func RegisterHomeRoutes(mux *http.ServeMux) {
	// 根路径 - 主页
	mux.HandleFunc("/", home.RootHandler)
}
