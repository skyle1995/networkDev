package server

import (
	adminctl "networkDev/controllers/admin"
	"networkDev/utils"

	"github.com/gin-gonic/gin"
)

// RegisterAdminRoutes 注册管理员后台相关路由
// - /admin/login: 支持GET渲染登录页、POST提交登录
// - /admin/logout: 管理员退出登录
// - /admin/dashboard: 管理员仪表盘（示例）
// - /admin/fragment/*: 布局内动态片段加载
// - /admin/api/settings*: 设置接口（查询/更新）
func RegisterAdminRoutes(router *gin.Engine) {
	// /admin 根与前缀统一入口：根据是否登录跳转
	router.GET("/admin", adminctl.AdminIndexHandler)
	router.GET("/admin/", adminctl.AdminIndexHandler)

	// Admin 认证相关路由
	router.GET("/admin/login", adminctl.LoginPageHandler)
	router.POST("/admin/login", adminctl.LoginHandler) // CSRF验证在控制器内部处理

	// 退出登录（无需拦截，幂等清理）
	router.POST("/admin/logout", adminctl.LogoutHandler)

	// 验证码生成路由（无需认证）
	router.GET("/admin/captcha", adminctl.CaptchaHandler)

	// CSRF令牌获取API（无需认证，但需要在登录页面等地方获取）
	router.GET("/admin/api/csrf-token", func(c *gin.Context) {
		// 生成新的CSRF令牌
		token, err := utils.GenerateCSRFToken()
		if err != nil {
			c.JSON(500, gin.H{"success": false, "message": "生成CSRF令牌失败"})
			return
		}

		// 设置令牌到Cookie和响应头
		utils.SetCSRFToken(c, token)

		// 返回令牌给前端
		c.JSON(200, gin.H{
			"success":    true,
			"message":    "CSRF令牌生成成功",
			"csrf_token": token,
		})
	})

	// 后台布局页（需要管理员认证）
	router.GET("/admin/layout", adminctl.AdminAuthRequired(), adminctl.AdminLayoutHandler)

	// 片段路由（需要管理员认证）
	router.GET("/admin/dashboard", adminctl.AdminAuthRequired(), adminctl.DashboardFragmentHandler)
	router.GET("/admin/user", adminctl.AdminAuthRequired(), adminctl.UserFragmentHandler)
	router.GET("/admin/settings", adminctl.AdminAuthRequired(), adminctl.SettingsFragmentHandler)
	router.GET("/admin/apps", adminctl.AdminAuthRequired(), adminctl.AppsFragmentHandler)
	router.GET("/admin/apis", adminctl.AdminAuthRequired(), adminctl.APIFragmentHandler)
	router.GET("/admin/variables", adminctl.AdminAuthRequired(), adminctl.VariableFragmentHandler)

	// 系统信息API（用于仪表盘定时刷新）
	router.GET("/admin/api/system/info", adminctl.AdminAuthRequired(), adminctl.SystemInfoHandler)

	// 仪表盘统计数据API
	router.GET("/admin/api/dashboard/stats", adminctl.AdminAuthRequired(), adminctl.DashboardStatsHandler)

	// 个人资料API
	router.GET("/admin/api/user/profile", adminctl.AdminAuthRequired(), adminctl.UserProfileQueryHandler)
	router.POST("/admin/api/user/profile/update", adminctl.AdminAuthRequired(), adminctl.UserProfileUpdateHandler)
	router.POST("/admin/api/user/password", adminctl.AdminAuthRequired(), adminctl.UserPasswordUpdateHandler)

	// 系统设置API
	router.GET("/admin/api/settings", adminctl.AdminAuthRequired(), adminctl.SettingsQueryHandler)
	router.POST("/admin/api/settings/update", adminctl.AdminAuthRequired(), adminctl.SettingsUpdateHandler)

	// 应用管理API
	router.GET("/admin/api/apps/list", adminctl.AdminAuthRequired(), adminctl.AppsListHandler)
	router.POST("/admin/api/apps/create", adminctl.AdminAuthRequired(), adminctl.AppCreateHandler)
	router.POST("/admin/api/apps/update", adminctl.AdminAuthRequired(), adminctl.AppUpdateHandler)
	router.POST("/admin/api/apps/delete", adminctl.AdminAuthRequired(), adminctl.AppDeleteHandler)
	router.POST("/admin/api/apps/batch_delete", adminctl.AdminAuthRequired(), adminctl.AppsBatchDeleteHandler)
	router.POST("/admin/api/apps/batch_update_status", adminctl.AdminAuthRequired(), adminctl.AppsBatchUpdateStatusHandler)
	router.POST("/admin/api/apps/update_status", adminctl.AdminAuthRequired(), adminctl.AppUpdateStatusHandler)
	router.POST("/admin/api/apps/reset_secret", adminctl.AdminAuthRequired(), adminctl.AppResetSecretHandler)
	router.GET("/admin/api/apps/get_app_data", adminctl.AdminAuthRequired(), adminctl.AppGetAppDataHandler)
	router.POST("/admin/api/apps/update_app_data", adminctl.AdminAuthRequired(), adminctl.AppUpdateAppDataHandler)
	router.GET("/admin/api/apps/get_announcement", adminctl.AdminAuthRequired(), adminctl.AppGetAnnouncementHandler)
	router.POST("/admin/api/apps/update_announcement", adminctl.AdminAuthRequired(), adminctl.AppUpdateAnnouncementHandler)
	router.GET("/admin/api/apps/get_multi_config", adminctl.AdminAuthRequired(), adminctl.AppGetMultiConfigHandler)
	router.POST("/admin/api/apps/update_multi_config", adminctl.AdminAuthRequired(), adminctl.AppUpdateMultiConfigHandler)
	router.GET("/admin/api/apps/get_bind_config", adminctl.AdminAuthRequired(), adminctl.AppGetBindConfigHandler)
	router.POST("/admin/api/apps/update_bind_config", adminctl.AdminAuthRequired(), adminctl.AppUpdateBindConfigHandler)
	router.GET("/admin/api/apps/get_register_config", adminctl.AdminAuthRequired(), adminctl.AppGetRegisterConfigHandler)
	router.POST("/admin/api/apps/update_register_config", adminctl.AdminAuthRequired(), adminctl.AppUpdateRegisterConfigHandler)

	// API接口管理API
	router.GET("/admin/api/apis/list", adminctl.AdminAuthRequired(), adminctl.APIListHandler)
	router.POST("/admin/api/apis/update", adminctl.AdminAuthRequired(), adminctl.APIUpdateHandler)
	router.POST("/admin/api/apis/update_status", adminctl.AdminAuthRequired(), adminctl.APIUpdateStatusHandler)
	router.GET("/admin/api/apis/apps", adminctl.AdminAuthRequired(), adminctl.APIGetAppsHandler)
	router.GET("/admin/api/apis/types", adminctl.AdminAuthRequired(), adminctl.APIGetTypesHandler)
	router.POST("/admin/api/apis/generate_keys", adminctl.AdminAuthRequired(), adminctl.APIGenerateKeysHandler)

	// 变量管理API
	router.GET("/admin/variable/list", adminctl.AdminAuthRequired(), adminctl.VariableListHandler)
	router.POST("/admin/variable/create", adminctl.AdminAuthRequired(), adminctl.VariableCreateHandler)
	router.POST("/admin/variable/update", adminctl.AdminAuthRequired(), adminctl.VariableUpdateHandler)
	router.POST("/admin/variable/delete", adminctl.AdminAuthRequired(), adminctl.VariableDeleteHandler)
	router.POST("/admin/variable/batch_delete", adminctl.AdminAuthRequired(), adminctl.VariablesBatchDeleteHandler)
}
