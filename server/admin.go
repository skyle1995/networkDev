package server

import (
	"net/http"
	adminctl "networkDev/controllers/admin"
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
			adminctl.LoginHandler(w, r)
			return
		}
		w.WriteHeader(http.StatusMethodNotAllowed)
	})

	// 退出登录（无需拦截，幂等清理）
	mux.HandleFunc("/admin/logout", adminctl.LogoutHandler)

	// 验证码生成路由（无需认证）
	mux.HandleFunc("/admin/captcha", adminctl.CaptchaHandler)

	// 后台布局页（需要管理员认证）
	mux.HandleFunc("/admin/layout", adminctl.AdminAuthRequired(adminctl.AdminLayoutHandler))

	// 片段路由（需要管理员认证）
	mux.HandleFunc("/admin/dashboard", adminctl.AdminAuthRequired(adminctl.DashboardFragmentHandler))
	mux.HandleFunc("/admin/user", adminctl.AdminAuthRequired(adminctl.UserFragmentHandler))
	mux.HandleFunc("/admin/settings", adminctl.AdminAuthRequired(adminctl.SettingsFragmentHandler))
	mux.HandleFunc("/admin/apps", adminctl.AdminAuthRequired(adminctl.AppsFragmentHandler))
	mux.HandleFunc("/admin/apis", adminctl.AdminAuthRequired(adminctl.APIFragmentHandler))

	// 个人资料API
	mux.HandleFunc("/admin/api/user/profile", adminctl.AdminAuthRequired(adminctl.UserProfileQueryHandler))
	mux.HandleFunc("/admin/api/user/profile/update", adminctl.AdminAuthRequired(adminctl.UserProfileUpdateHandler))
	mux.HandleFunc("/admin/api/user/password", adminctl.AdminAuthRequired(adminctl.UserPasswordUpdateHandler))
	// 设置API（需要管理员认证）
	mux.HandleFunc("/admin/api/settings", adminctl.AdminAuthRequired(adminctl.SettingsQueryHandler))
	mux.HandleFunc("/admin/api/settings/update", adminctl.AdminAuthRequired(adminctl.SettingsUpdateHandler))

	// 应用管理API
	mux.HandleFunc("/admin/api/apps/list", adminctl.AdminAuthRequired(adminctl.AppsListHandler))
	mux.HandleFunc("/admin/api/apps/create", adminctl.AdminAuthRequired(adminctl.AppCreateHandler))
	mux.HandleFunc("/admin/api/apps/update", adminctl.AdminAuthRequired(adminctl.AppUpdateHandler))
	mux.HandleFunc("/admin/api/apps/delete", adminctl.AdminAuthRequired(adminctl.AppDeleteHandler))
	mux.HandleFunc("/admin/api/apps/batch_delete", adminctl.AdminAuthRequired(adminctl.AppsBatchDeleteHandler))
	mux.HandleFunc("/admin/api/apps/batch_update_status", adminctl.AdminAuthRequired(adminctl.AppsBatchUpdateStatusHandler))
	mux.HandleFunc("/admin/api/apps/reset_secret", adminctl.AdminAuthRequired(adminctl.AppResetSecretHandler))
	mux.HandleFunc("/admin/api/apps/get_app_data", adminctl.AdminAuthRequired(adminctl.AppGetAppDataHandler))
	mux.HandleFunc("/admin/api/apps/update_app_data", adminctl.AdminAuthRequired(adminctl.AppUpdateAppDataHandler))
	mux.HandleFunc("/admin/api/apps/get_announcement", adminctl.AdminAuthRequired(adminctl.AppGetAnnouncementHandler))
	mux.HandleFunc("/admin/api/apps/update_announcement", adminctl.AdminAuthRequired(adminctl.AppUpdateAnnouncementHandler))
	mux.HandleFunc("/admin/api/apps/get_multi_config", adminctl.AdminAuthRequired(adminctl.AppGetMultiConfigHandler))
	mux.HandleFunc("/admin/api/apps/update_multi_config", adminctl.AdminAuthRequired(adminctl.AppUpdateMultiConfigHandler))
	mux.HandleFunc("/admin/api/apps/get_bind_config", adminctl.AdminAuthRequired(adminctl.AppGetBindConfigHandler))
	mux.HandleFunc("/admin/api/apps/update_bind_config", adminctl.AdminAuthRequired(adminctl.AppUpdateBindConfigHandler))
	mux.HandleFunc("/admin/api/apps/get_register_config", adminctl.AdminAuthRequired(adminctl.AppGetRegisterConfigHandler))
	mux.HandleFunc("/admin/api/apps/update_register_config", adminctl.AdminAuthRequired(adminctl.AppUpdateRegisterConfigHandler))

	// API接口管理API
	mux.HandleFunc("/admin/api/apis/list", adminctl.AdminAuthRequired(adminctl.APIListHandler))
	mux.HandleFunc("/admin/api/apis/update", adminctl.AdminAuthRequired(adminctl.APIUpdateHandler))
	mux.HandleFunc("/admin/api/apis/apps", adminctl.AdminAuthRequired(adminctl.APIGetAppsHandler))
	mux.HandleFunc("/admin/api/apis/types", adminctl.AdminAuthRequired(adminctl.APIGetTypesHandler))
	mux.HandleFunc("/admin/api/apis/generate_keys", adminctl.AdminAuthRequired(adminctl.APIGenerateKeysHandler))

	// 系统信息API（用于仪表盘定时刷新）
	mux.HandleFunc("/admin/api/system/info", adminctl.AdminAuthRequired(adminctl.SystemInfoHandler))
}
