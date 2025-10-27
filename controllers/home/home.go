package home

import (
	"net/http"
	"networkDev/controllers"
	"networkDev/database"
	"networkDev/services"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// ============================================================================
// 全局变量
// ============================================================================

var homeBaseController = controllers.NewBaseController()

// ============================================================================
// 辅助函数
// ============================================================================

// getSettingValue 获取配置值，优先从数据库获取，不存在时使用默认值
func getSettingValue(settingName string, defaultValue string, db *gorm.DB) string {
	if setting, err := services.FindSettingByName(settingName, db); err == nil {
		return setting.Value
	}
	return defaultValue
}

// ============================================================================
// 页面处理器
// ============================================================================

// RootHandler 主页处理器
func RootHandler(c *gin.Context) {
	// 获取数据库连接
	db, err := database.GetDB()
	if err != nil {
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{
			"error": "数据库连接失败",
		})
		return
	}

	// 获取默认模板数据
	data := homeBaseController.GetDefaultTemplateData()

	// 从数据库读取设置，优先使用数据库配置，不存在时使用默认值
	data["SystemName"] = getSettingValue("site_title", data["SystemName"].(string), db)
	data["FooterText"] = getSettingValue("footer_text", data["FooterText"].(string), db)
	data["ICPRecord"] = getSettingValue("icp_record", data["ICPRecord"].(string), db)
	data["ICPRecordLink"] = getSettingValue("icp_record_link", data["ICPRecordLink"].(string), db)
	data["PSBRecord"] = getSettingValue("psb_record", data["PSBRecord"].(string), db)
	data["PSBRecordLink"] = getSettingValue("psb_record_link", data["PSBRecordLink"].(string), db)
	data["title"] = "主页"

	c.HTML(http.StatusOK, "index.html", data)
}
