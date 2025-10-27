package admin

import (
	"net/http"
	"networkDev/constants"
	"networkDev/controllers"
	"networkDev/middleware"
	"networkDev/models"
	"networkDev/services"
	"networkDev/utils"
	"networkDev/utils/timeutil"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

// 创建基础控制器实例
var handlersBaseController = controllers.NewBaseController()

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
func AdminIndexHandler(c *gin.Context) {
	if IsAdminAuthenticatedWithCleanup(c) {
		// 直接渲染布局页，保持URL为 /admin
		AdminLayoutHandler(c)
		return
	}
	c.Redirect(http.StatusFound, "/admin/login")
}

// AdminLayoutHandler 后台布局页渲染
// - 渲染 layout.html，包含顶部导航、侧边栏与动态内容容器
func AdminLayoutHandler(c *gin.Context) {
	// 获取或生成CSRF令牌
	var token string
	if existingToken := utils.GetCSRFTokenFromCookie(c); existingToken != "" {
		// 重用现有的Cookie令牌
		token = existingToken
	} else {
		// 生成新的CSRF令牌并设置到Cookie
		newToken, err := utils.GenerateCSRFToken()
		if err != nil {
			handlersBaseController.HandleInternalError(c, "生成CSRF令牌失败", err)
			return
		}
		token = newToken
		utils.SetCSRFToken(c, token)
	}

	// 准备模板数据
	data := handlersBaseController.GetDefaultTemplateData()
	data["CSRFToken"] = token

	// 从数据库读取站点标题，如果失败则使用默认值
	if db, ok := handlersBaseController.GetDB(c); ok {
		if siteTitle, err := services.FindSettingByName("site_title", db); err == nil && siteTitle != nil {
			data["Title"] = siteTitle.Value
		}
	}

	// 合并其他数据（如果有的话）
	extraData := gin.H{}
	for key, value := range extraData {
		data[key] = value
	}

	c.HTML(http.StatusOK, "layout.html", data)
}

// DashboardFragmentHandler 仪表盘片段渲染
// - 展示系统信息：版本、开发模式、数据库类型、启动时长
func DashboardFragmentHandler(c *gin.Context) {
	version := constants.AppVersion
	mode := middleware.IsDevModeFromContext(c)
	dbType := viper.GetString("database.type")
	if dbType == "" {
		dbType = "sqlite"
	}
	uptime := timeutil.GetServerUptimeString()

	data := gin.H{
		"Version": version,
		"Mode":    mode,
		"DBType":  formatDBType(dbType),
		"Uptime":  uptime,
	}

	c.HTML(http.StatusOK, "dashboard.html", data)
}

// SystemInfoHandler 系统信息API接口
// - 返回系统运行状态的JSON数据，用于前端定时刷新
func SystemInfoHandler(c *gin.Context) {
	version := constants.AppVersion
	mode := middleware.IsDevModeFromContext(c)
	dbType := viper.GetString("database.type")
	if dbType == "" {
		dbType = "sqlite"
	}
	uptime := timeutil.GetServerUptimeString()

	data := gin.H{
		"version": version,
		"mode":    mode,
		"db_type": formatDBType(dbType),
		"uptime":  uptime,
	}

	handlersBaseController.HandleSuccess(c, "ok", data)
}

// DashboardStatsHandler 仪表盘统计数据API接口
// - 返回应用统计数据的JSON数据，包括全部/启用/禁用/变量数量
func DashboardStatsHandler(c *gin.Context) {
	// 获取数据库连接
	db, ok := handlersBaseController.GetDB(c)
	if !ok {
		return
	}

	// 统计应用数据
	var totalApps int64
	var enabledApps int64
	var disabledApps int64
	var totalVariables int64

	// 统计全部应用数量
	if err := db.Model(&models.App{}).Count(&totalApps).Error; err != nil {
		handlersBaseController.HandleInternalError(c, "统计应用数量失败", err)
		return
	}

	// 统计启用应用数量
	if err := db.Model(&models.App{}).Where("status = ?", 1).Count(&enabledApps).Error; err != nil {
		handlersBaseController.HandleInternalError(c, "统计启用应用数量失败", err)
		return
	}

	// 统计禁用应用数量
	if err := db.Model(&models.App{}).Where("status = ?", 0).Count(&disabledApps).Error; err != nil {
		handlersBaseController.HandleInternalError(c, "统计禁用应用数量失败", err)
		return
	}

	// 统计变量数量
	if err := db.Model(&models.Variable{}).Count(&totalVariables).Error; err != nil {
		handlersBaseController.HandleInternalError(c, "统计变量数量失败", err)
		return
	}

	data := gin.H{
		"total_apps":      totalApps,
		"enabled_apps":    enabledApps,
		"disabled_apps":   disabledApps,
		"total_variables": totalVariables,
	}

	handlersBaseController.HandleSuccess(c, "ok", data)
}
