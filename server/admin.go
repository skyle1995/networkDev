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

	// 后台布局页（需要管理员认证）
	mux.HandleFunc("/admin/layout", adminctl.AdminAuthRequired(adminctl.AdminLayoutHandler))

	// 片段路由（需要管理员认证）
	mux.HandleFunc("/admin/dashboard", adminctl.AdminAuthRequired(adminctl.DashboardFragmentHandler))
	mux.HandleFunc("/admin/user", adminctl.AdminAuthRequired(adminctl.UserFragmentHandler))
	mux.HandleFunc("/admin/settings", adminctl.AdminAuthRequired(adminctl.SettingsFragmentHandler))
	mux.HandleFunc("/admin/apps", adminctl.AdminAuthRequired(adminctl.AppsFragmentHandler))
	mux.HandleFunc("/admin/logintypes", adminctl.AdminAuthRequired(adminctl.LoginTypesFragmentHandler))
	mux.HandleFunc("/admin/cardtypes", adminctl.AdminAuthRequired(adminctl.CardTypesFragmentHandler))
	mux.HandleFunc("/admin/cards", adminctl.AdminAuthRequired(adminctl.CardsFragmentHandler))

	// 个人资料API
	mux.HandleFunc("/admin/api/user/profile", adminctl.AdminAuthRequired(adminctl.UserProfileQueryHandler))
	mux.HandleFunc("/admin/api/user/profile/update", adminctl.AdminAuthRequired(adminctl.UserProfileUpdateHandler))
	mux.HandleFunc("/admin/api/user/password", adminctl.AdminAuthRequired(adminctl.UserPasswordUpdateHandler))
	// 设置API（需要管理员认证）
	mux.HandleFunc("/admin/api/settings", adminctl.AdminAuthRequired(adminctl.SettingsQueryHandler))
	mux.HandleFunc("/admin/api/settings/update", adminctl.AdminAuthRequired(adminctl.SettingsUpdateHandler))

	// 供前端下拉选择卡密类型
	mux.HandleFunc("/admin/api/cards/types", adminctl.AdminAuthRequired(adminctl.GetCardTypesHandler))
	// 应用管理API
	mux.HandleFunc("/admin/api/apps/list", adminctl.AdminAuthRequired(adminctl.AppsListHandler))
	mux.HandleFunc("/admin/api/apps/create", adminctl.AdminAuthRequired(adminctl.AppCreateHandler))
	mux.HandleFunc("/admin/api/apps/update", adminctl.AdminAuthRequired(adminctl.AppUpdateHandler))
	mux.HandleFunc("/admin/api/apps/delete", adminctl.AdminAuthRequired(adminctl.AppDeleteHandler))
	mux.HandleFunc("/admin/api/apps/batch_delete", adminctl.AdminAuthRequired(adminctl.AppsBatchDeleteHandler))
	mux.HandleFunc("/admin/api/apps/batch_update_status", adminctl.AdminAuthRequired(adminctl.AppsBatchUpdateStatusHandler))
	// 登录方式管理API
	mux.HandleFunc("/admin/api/login_types/list", adminctl.AdminAuthRequired(adminctl.LoginTypesListHandler))
	mux.HandleFunc("/admin/api/login_types/create", adminctl.AdminAuthRequired(adminctl.LoginTypeCreateHandler))
	mux.HandleFunc("/admin/api/login_types/update", adminctl.AdminAuthRequired(adminctl.LoginTypeUpdateHandler))
	mux.HandleFunc("/admin/api/login_types/delete", adminctl.AdminAuthRequired(adminctl.LoginTypeDeleteHandler))
	mux.HandleFunc("/admin/api/login_types/batch_delete", adminctl.AdminAuthRequired(adminctl.LoginTypesBatchDeleteHandler))
	mux.HandleFunc("/admin/api/login_types/batch_enable", adminctl.AdminAuthRequired(adminctl.LoginTypesBatchEnableHandler))
	mux.HandleFunc("/admin/api/login_types/batch_disable", adminctl.AdminAuthRequired(adminctl.LoginTypesBatchDisableHandler))
	// 卡密类型管理API
	mux.HandleFunc("/admin/api/card_types/list", adminctl.AdminAuthRequired(adminctl.CardTypesListHandler))
	mux.HandleFunc("/admin/api/card_types/create", adminctl.AdminAuthRequired(adminctl.CardTypeCreateHandler))
	mux.HandleFunc("/admin/api/card_types/update", adminctl.AdminAuthRequired(adminctl.CardTypeUpdateHandler))
	mux.HandleFunc("/admin/api/card_types/delete", adminctl.AdminAuthRequired(adminctl.CardTypeDeleteHandler))
	mux.HandleFunc("/admin/api/card_types/batch_delete", adminctl.AdminAuthRequired(adminctl.CardTypesBatchDeleteHandler))
	mux.HandleFunc("/admin/api/card_types/batch_enable", adminctl.AdminAuthRequired(adminctl.CardTypesBatchEnableHandler))
	mux.HandleFunc("/admin/api/card_types/batch_disable", adminctl.AdminAuthRequired(adminctl.CardTypesBatchDisableHandler))
	// 卡密管理API
	mux.HandleFunc("/admin/api/cards/list", adminctl.AdminAuthRequired(adminctl.CardsListHandler))
	mux.HandleFunc("/admin/api/cards/create", adminctl.AdminAuthRequired(adminctl.CardCreateHandler))
	mux.HandleFunc("/admin/api/cards/update", adminctl.AdminAuthRequired(adminctl.CardUpdateHandler))
	mux.HandleFunc("/admin/api/cards/delete", adminctl.AdminAuthRequired(adminctl.CardDeleteHandler))
	mux.HandleFunc("/admin/api/cards/batch_delete", adminctl.AdminAuthRequired(adminctl.CardsBatchDeleteHandler))
	mux.HandleFunc("/admin/api/cards/batch_update_status", adminctl.AdminAuthRequired(adminctl.CardsBatchUpdateStatusHandler))
	// 新增：卡密导出API（CSV下载）
	mux.HandleFunc("/admin/api/cards/export", adminctl.AdminAuthRequired(adminctl.CardsExportHandler))
	// 新增：导出选中卡密API
	mux.HandleFunc("/admin/api/cards/export_selected", adminctl.AdminAuthRequired(adminctl.CardsExportSelectedHandler))

	// 系统信息API（用于仪表盘定时刷新）
	mux.HandleFunc("/admin/api/system/info", adminctl.AdminAuthRequired(adminctl.SystemInfoHandler))

	// 卡密统计API（用于仪表盘统计显示）
	mux.HandleFunc("/admin/api/cards/stats_overview", adminctl.AdminAuthRequired(adminctl.CardStatsOverviewHandler))
	mux.HandleFunc("/admin/api/cards/trend_30days", adminctl.AdminAuthRequired(adminctl.CardStatsTrend30DaysHandler))
	mux.HandleFunc("/admin/api/cards/stats_simple", adminctl.AdminAuthRequired(adminctl.CardStatsSimpleHandler))
}
