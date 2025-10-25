package admin

import (
	"net/http"
	"networkDev/database"
	"networkDev/services"
	"networkDev/utils"
	"networkDev/utils/timeutil"

	"github.com/spf13/viper"
)

// AdminIndexHandler /admin 与 /admin/ 根路径入口
// - 未登录：重定向到 /admin/login
// - 已登录：渲染后台布局页（或重定向到 /admin/layout）
// - 自动清理失效的JWT Cookie
func AdminIndexHandler(w http.ResponseWriter, r *http.Request) {
	if IsAdminAuthenticatedWithCleanup(w, r) {
		// 直接渲染布局页，保持URL为 /admin
		AdminLayoutHandler(w, r)
		return
	}
	http.Redirect(w, r, "/admin/login", http.StatusFound)
}

// AdminLayoutHandler 后台布局页渲染
// - 渲染 layout.html，包含顶部导航、侧边栏与动态内容容器
func AdminLayoutHandler(w http.ResponseWriter, r *http.Request) {
	// 获取或生成CSRF令牌
	var token string
	if existingToken := utils.GetCSRFTokenFromCookie(r); existingToken != "" {
		// 重用现有的Cookie令牌
		token = existingToken
	} else {
		// 生成新的CSRF令牌并设置到Cookie
		newToken, err := utils.GenerateCSRFToken()
		if err != nil {
			http.Error(w, "生成CSRF令牌失败", http.StatusInternalServerError)
			return
		}
		token = newToken
		utils.SetCSRFToken(w, token)
	}

	// 准备额外的模板数据
	extraData := make(map[string]interface{})

	// 从数据库读取站点标题
	db, dbErr := database.GetDB()
	if dbErr != nil {
		extraData["Title"] = "凌动技术"
	} else {
		siteTitle, settingErr := services.FindSettingByName("site_title", db)
		if settingErr != nil || siteTitle == nil {
			extraData["Title"] = "凌动技术"
		} else {
			extraData["Title"] = siteTitle.Value
		}
	}

	// 准备模板数据
	data := utils.GetDefaultTemplateData()
	data["CSRFToken"] = token
	
	// 合并额外数据
	for key, value := range extraData {
		data[key] = value
	}

	utils.RenderTemplate(w, "layout.html", data)
}

// DashboardFragmentHandler 仪表盘片段渲染
// - 展示系统信息：版本、运行模式、数据库类型、启动时长
func DashboardFragmentHandler(w http.ResponseWriter, r *http.Request) {
	version := "1.0.0"
	mode := viper.GetString("server.mode")
	dbType := viper.GetString("database.type")
	if dbType == "" {
		dbType = "sqlite"
	}
	uptime := timeutil.GetServerUptimeString()

	data := map[string]interface{}{
		"Version": version,
		"Mode":    mode,
		"DBType":  dbType,
		"Uptime":  uptime,
	}

	utils.RenderTemplate(w, "dashboard.html", data)
}

// SystemInfoHandler 系统信息API接口
// - 返回系统运行状态的JSON数据，用于前端定时刷新
func SystemInfoHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	version := "1.0.0"
	mode := viper.GetString("server.mode")
	dbType := viper.GetString("database.type")
	if dbType == "" {
		dbType = "sqlite"
	}
	uptime := timeutil.GetServerUptimeString()

	data := map[string]interface{}{
		"version": version,
		"mode":    mode,
		"db_type": dbType,
		"uptime":  uptime,
	}

	utils.JsonResponse(w, http.StatusOK, true, "ok", data)
}
