package server

import (
	"net/http"
	adminctl "networkDev/controllers/admin"
	"networkDev/utils"
)

// RegisterAdminRoutes 注册管理员后台相关路由
// - /admin/login: 支持GET渲染登录页、POST提交登录
// - /admin/logout: 管理员退出登录
// - /admin/dashboard: 管理员仪表盘（示例）
// - /admin/fragment/*: 布局内动态片段加载
// - /admin/api/settings*: 设置接口（查询/更新）
func RegisterAdminRoutes(mux *http.ServeMux) {
	// /admin 根与前缀统一入口：根据是否登录跳转
	mux.HandleFunc("/admin", adminctl.AdminIndexHandler)
	mux.HandleFunc("/admin/", adminctl.AdminIndexHandler)

	// Admin 认证相关路由
	mux.HandleFunc("/admin/login", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			adminctl.LoginPageHandler(w, r)
			return
		}
		if r.Method == http.MethodPost {
			// 应用CSRF保护
			utils.RequireCSRFToken(adminctl.LoginHandler)(w, r)
			return
		}
		w.WriteHeader(http.StatusMethodNotAllowed)
	})

	// 退出登录（无需拦截，幂等清理）
	mux.HandleFunc("/admin/logout", adminctl.LogoutHandler)

	// 验证码生成路由（无需认证）
	mux.HandleFunc("/admin/captcha", adminctl.CaptchaHandler)

	// CSRF令牌获取API（无需认证，但需要在登录页面等地方获取）
	mux.HandleFunc("/admin/api/csrf-token", utils.CSRFTokenHandler)

	// 后台布局页（需要管理员认证）
	mux.HandleFunc("/admin/layout", adminctl.AdminAuthRequired(adminctl.AdminLayoutHandler))

	// 片段路由（需要管理员认证）
	mux.HandleFunc("/admin/dashboard", adminctl.AdminAuthRequired(adminctl.DashboardFragmentHandler))
	mux.HandleFunc("/admin/user", adminctl.AdminAuthRequired(adminctl.UserFragmentHandler))
	mux.HandleFunc("/admin/settings", adminctl.AdminAuthRequired(adminctl.SettingsFragmentHandler))
	mux.HandleFunc("/admin/apps", adminctl.AdminAuthRequired(adminctl.AppsFragmentHandler))
	mux.HandleFunc("/admin/apis", adminctl.AdminAuthRequired(adminctl.APIFragmentHandler))
	mux.HandleFunc("/admin/variables", adminctl.AdminAuthRequired(adminctl.VariableFragmentHandler))

	// 系统信息API（用于仪表盘定时刷新）
	mux.HandleFunc("/admin/api/system/info", adminctl.AdminAuthRequired(adminctl.SystemInfoHandler))

	// 仪表盘统计数据API
	mux.HandleFunc("/admin/api/dashboard/stats", adminctl.AdminAuthRequired(adminctl.DashboardStatsHandler))

	// 个人资料API
	mux.HandleFunc("/admin/api/user/profile", adminctl.AdminAuthRequired(adminctl.UserProfileQueryHandler))
	mux.HandleFunc("/admin/api/user/profile/update", adminctl.AdminAuthRequired(utils.RequireCSRFToken(adminctl.UserProfileUpdateHandler)))
	mux.HandleFunc("/admin/api/user/password", adminctl.AdminAuthRequired(utils.RequireCSRFToken(adminctl.UserPasswordUpdateHandler)))

	// 系统设置API
	mux.HandleFunc("/admin/api/settings", adminctl.AdminAuthRequired(adminctl.SettingsQueryHandler))
	mux.HandleFunc("/admin/api/settings/update", adminctl.AdminAuthRequired(utils.RequireCSRFToken(adminctl.SettingsUpdateHandler)))

	// 应用管理API
	mux.HandleFunc("/admin/api/apps/list", adminctl.AdminAuthRequired(adminctl.AppsListHandler))
	mux.HandleFunc("/admin/api/apps/create", adminctl.AdminAuthRequired(utils.RequireCSRFToken(adminctl.AppCreateHandler)))
	mux.HandleFunc("/admin/api/apps/update", adminctl.AdminAuthRequired(utils.RequireCSRFToken(adminctl.AppUpdateHandler)))
	mux.HandleFunc("/admin/api/apps/delete", adminctl.AdminAuthRequired(utils.RequireCSRFToken(adminctl.AppDeleteHandler)))
	mux.HandleFunc("/admin/api/apps/batch_delete", adminctl.AdminAuthRequired(utils.RequireCSRFToken(adminctl.AppsBatchDeleteHandler)))
	mux.HandleFunc("/admin/api/apps/batch_update_status", adminctl.AdminAuthRequired(utils.RequireCSRFToken(adminctl.AppsBatchUpdateStatusHandler)))
	mux.HandleFunc("/admin/api/apps/reset_secret", adminctl.AdminAuthRequired(utils.RequireCSRFToken(adminctl.AppResetSecretHandler)))
	mux.HandleFunc("/admin/api/apps/get_app_data", adminctl.AdminAuthRequired(adminctl.AppGetAppDataHandler))
	mux.HandleFunc("/admin/api/apps/update_app_data", adminctl.AdminAuthRequired(utils.RequireCSRFToken(adminctl.AppUpdateAppDataHandler)))
	mux.HandleFunc("/admin/api/apps/get_announcement", adminctl.AdminAuthRequired(adminctl.AppGetAnnouncementHandler))
	mux.HandleFunc("/admin/api/apps/update_announcement", adminctl.AdminAuthRequired(utils.RequireCSRFToken(adminctl.AppUpdateAnnouncementHandler)))
	mux.HandleFunc("/admin/api/apps/get_multi_config", adminctl.AdminAuthRequired(adminctl.AppGetMultiConfigHandler))
	mux.HandleFunc("/admin/api/apps/update_multi_config", adminctl.AdminAuthRequired(utils.RequireCSRFToken(adminctl.AppUpdateMultiConfigHandler)))
	mux.HandleFunc("/admin/api/apps/get_bind_config", adminctl.AdminAuthRequired(adminctl.AppGetBindConfigHandler))
	mux.HandleFunc("/admin/api/apps/update_bind_config", adminctl.AdminAuthRequired(utils.RequireCSRFToken(adminctl.AppUpdateBindConfigHandler)))
	mux.HandleFunc("/admin/api/apps/get_register_config", adminctl.AdminAuthRequired(adminctl.AppGetRegisterConfigHandler))
	mux.HandleFunc("/admin/api/apps/update_register_config", adminctl.AdminAuthRequired(utils.RequireCSRFToken(adminctl.AppUpdateRegisterConfigHandler)))

	// API接口管理API
	mux.HandleFunc("/admin/api/apis/list", adminctl.AdminAuthRequired(adminctl.APIListHandler))
	mux.HandleFunc("/admin/api/apis/update", adminctl.AdminAuthRequired(utils.RequireCSRFToken(adminctl.APIUpdateHandler)))
	mux.HandleFunc("/admin/api/apis/apps", adminctl.AdminAuthRequired(adminctl.APIGetAppsHandler))
	mux.HandleFunc("/admin/api/apis/types", adminctl.AdminAuthRequired(adminctl.APIGetTypesHandler))
	mux.HandleFunc("/admin/api/apis/generate_keys", adminctl.AdminAuthRequired(utils.RequireCSRFToken(adminctl.APIGenerateKeysHandler)))

	// 变量管理API
	mux.HandleFunc("/admin/variable/list", adminctl.AdminAuthRequired(adminctl.VariableListHandler))
	mux.HandleFunc("/admin/variable/apps", adminctl.AdminAuthRequired(adminctl.VariableGetAppsHandler))
	mux.HandleFunc("/admin/variable/create", adminctl.AdminAuthRequired(utils.RequireCSRFToken(adminctl.VariableCreateHandler)))
	mux.HandleFunc("/admin/variable/update", adminctl.AdminAuthRequired(utils.RequireCSRFToken(adminctl.VariableUpdateHandler)))
	mux.HandleFunc("/admin/variable/delete", adminctl.AdminAuthRequired(utils.RequireCSRFToken(adminctl.VariableDeleteHandler)))
	mux.HandleFunc("/admin/variable/batch_delete", adminctl.AdminAuthRequired(utils.RequireCSRFToken(adminctl.VariablesBatchDeleteHandler)))
}
