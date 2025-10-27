package server

import (
	adminctl "networkDev/controllers/admin"
	"networkDev/utils"

	"github.com/gin-gonic/gin"
)

// ============================================================================
// 路由注册函数
// ============================================================================

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
	router.GET("/admin/functions", adminctl.AdminAuthRequired(), adminctl.FunctionFragmentHandler)

	// 系统信息API（用于仪表盘定时刷新）
	router.GET("/admin/api/system/info", adminctl.AdminAuthRequired(), adminctl.SystemInfoHandler)

	// 仪表盘统计数据API
	router.GET("/admin/api/dashboard/stats", adminctl.AdminAuthRequired(), adminctl.DashboardStatsHandler)

	// 个人资料API
	userGroup := router.Group("/admin/api/user", adminctl.AdminAuthRequired())
	{
		userGroup.GET("/profile", adminctl.UserProfileQueryHandler)
		userGroup.POST("/profile/update", adminctl.UserProfileUpdateHandler)
		userGroup.POST("/password", adminctl.UserPasswordUpdateHandler)
	}

	// 系统设置API
	settingsGroup := router.Group("/admin/api/settings", adminctl.AdminAuthRequired())
	{
		settingsGroup.GET("", adminctl.SettingsQueryHandler)
		settingsGroup.POST("/update", adminctl.SettingsUpdateHandler)
	}

	// 应用管理API
	appsGroup := router.Group("/admin/api/apps", adminctl.AdminAuthRequired())
	{
		appsGroup.GET("/list", adminctl.AppsListHandler)
		appsGroup.GET("/simple", adminctl.AppsSimpleListHandler)
		appsGroup.POST("/create", adminctl.AppCreateHandler)
		appsGroup.POST("/update", adminctl.AppUpdateHandler)
		appsGroup.POST("/delete", adminctl.AppDeleteHandler)
		appsGroup.POST("/batch_delete", adminctl.AppsBatchDeleteHandler)
		appsGroup.POST("/batch_update_status", adminctl.AppsBatchUpdateStatusHandler)
		appsGroup.POST("/update_status", adminctl.AppUpdateStatusHandler)
		appsGroup.POST("/reset_secret", adminctl.AppResetSecretHandler)
		appsGroup.GET("/get_app_data", adminctl.AppGetAppDataHandler)
		appsGroup.POST("/update_app_data", adminctl.AppUpdateAppDataHandler)
		appsGroup.GET("/get_announcement", adminctl.AppGetAnnouncementHandler)
		appsGroup.POST("/update_announcement", adminctl.AppUpdateAnnouncementHandler)
		appsGroup.GET("/get_multi_config", adminctl.AppGetMultiConfigHandler)
		appsGroup.POST("/update_multi_config", adminctl.AppUpdateMultiConfigHandler)
		appsGroup.GET("/get_bind_config", adminctl.AppGetBindConfigHandler)
		appsGroup.POST("/update_bind_config", adminctl.AppUpdateBindConfigHandler)
		appsGroup.GET("/get_register_config", adminctl.AppGetRegisterConfigHandler)
		appsGroup.POST("/update_register_config", adminctl.AppUpdateRegisterConfigHandler)
	}

	// API接口管理API
	apisGroup := router.Group("/admin/api/apis", adminctl.AdminAuthRequired())
	{
		apisGroup.GET("/list", adminctl.APIListHandler)
		apisGroup.POST("/update", adminctl.APIUpdateHandler)
		apisGroup.POST("/update_status", adminctl.APIUpdateStatusHandler)
		apisGroup.GET("/types", adminctl.APIGetTypesHandler)
		apisGroup.POST("/generate_keys", adminctl.APIGenerateKeysHandler)
	}

	// 变量管理API
	variableGroup := router.Group("/admin/variable", adminctl.AdminAuthRequired())
	{
		variableGroup.GET("/list", adminctl.VariableListHandler)
		variableGroup.POST("/create", adminctl.VariableCreateHandler)
		variableGroup.POST("/update", adminctl.VariableUpdateHandler)
		variableGroup.POST("/delete", adminctl.VariableDeleteHandler)
		variableGroup.POST("/batch_delete", adminctl.VariablesBatchDeleteHandler)
	}

	// 函数管理API
	functionGroup := router.Group("/admin/function", adminctl.AdminAuthRequired())
	{
		functionGroup.GET("/list", adminctl.FunctionListHandler)
		functionGroup.POST("/create", adminctl.FunctionCreateHandler)
		functionGroup.POST("/update", adminctl.FunctionUpdateHandler)
		functionGroup.POST("/delete", adminctl.FunctionDeleteHandler)
		functionGroup.POST("/batch_delete", adminctl.FunctionsBatchDeleteHandler)
	}

}
