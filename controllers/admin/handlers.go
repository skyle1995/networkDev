package admin

import (
	"net/http"
	"networkDev/database"
	"networkDev/models"
	"networkDev/services"
	"networkDev/utils"
	"networkDev/utils/timeutil"

	"github.com/spf13/viper"
)

// formatDBType 格式化数据库类型显示
// 将配置文件中的小写类型转换为友好的显示格式
func formatDBType(dbType string) string {
	switch dbType {
	case "mysql":
		return "MySQL"
	case "sqlite":
		return "SQLite"
	case "postgresql", "postgres":
		return "PostgreSQL"
	case "sqlserver":
		return "SQL Server"
	default:
		return "SQLite" // 默认显示
	}
}

// AdminIndexHandler 后台首页处理器/admin 与 /admin/ 根路径入口
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
// - 展示系统信息：版本、开发模式、数据库类型、启动时长
func DashboardFragmentHandler(w http.ResponseWriter, r *http.Request) {
	version := "1.0.0"
	mode := viper.GetBool("server.dev_mode")
	dbType := viper.GetString("database.type")
	if dbType == "" {
		dbType = "sqlite"
	}
	uptime := timeutil.GetServerUptimeString()

	data := map[string]interface{}{
		"Version": version,
		"Mode":    mode,
		"DBType":  formatDBType(dbType),
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
	mode := viper.GetBool("server.dev_mode")
	dbType := viper.GetString("database.type")
	if dbType == "" {
		dbType = "sqlite"
	}
	uptime := timeutil.GetServerUptimeString()

	data := map[string]interface{}{
		"version": version,
		"mode":    mode,
		"db_type": formatDBType(dbType),
		"uptime":  uptime,
	}

	utils.JsonResponse(w, http.StatusOK, true, "ok", data)
}

// DashboardStatsHandler 仪表盘统计数据API接口
// - 返回应用统计数据的JSON数据，包括全部/启用/禁用/变量数量
func DashboardStatsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	// 获取数据库连接
	db, err := database.GetDB()
	if err != nil {
		utils.JsonResponse(w, http.StatusInternalServerError, false, "数据库连接失败", nil)
		return
	}

	// 统计应用数据
	var totalApps int64
	var enabledApps int64
	var disabledApps int64
	var totalVariables int64

	// 统计全部应用数量
	if err := db.Model(&models.App{}).Count(&totalApps).Error; err != nil {
		utils.JsonResponse(w, http.StatusInternalServerError, false, "统计应用数量失败", nil)
		return
	}

	// 统计启用应用数量
	if err := db.Model(&models.App{}).Where("status = ?", 1).Count(&enabledApps).Error; err != nil {
		utils.JsonResponse(w, http.StatusInternalServerError, false, "统计启用应用数量失败", nil)
		return
	}

	// 统计禁用应用数量
	if err := db.Model(&models.App{}).Where("status = ?", 0).Count(&disabledApps).Error; err != nil {
		utils.JsonResponse(w, http.StatusInternalServerError, false, "统计禁用应用数量失败", nil)
		return
	}

	// 统计变量数量
	if err := db.Model(&models.Variable{}).Count(&totalVariables).Error; err != nil {
		utils.JsonResponse(w, http.StatusInternalServerError, false, "统计变量数量失败", nil)
		return
	}

	data := map[string]interface{}{
		"total_apps":      totalApps,
		"enabled_apps":    enabledApps,
		"disabled_apps":   disabledApps,
		"total_variables": totalVariables,
	}

	utils.JsonResponse(w, http.StatusOK, true, "ok", data)
}
