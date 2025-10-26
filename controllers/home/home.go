package home

import (
	"net/http"
	"networkDev/controllers"
	"networkDev/database"
	"networkDev/services"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

var homeBaseController = controllers.NewBaseController()

// getSettingValue 获取配置值，优先从数据库获取，不存在时使用默认值
func getSettingValue(settingName string, defaultValue string, db *gorm.DB) string {
	if setting, err := services.FindSettingByName(settingName, db); err == nil {
		return setting.Value
	}
	return defaultValue
}

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
	defaultData := homeBaseController.GetDefaultTemplateData()

	// 准备模板数据，优先使用数据库配置，不存在时使用默认值
	data := gin.H{
		"SystemName":    getSettingValue("site_title", defaultData["SystemName"].(string), db),
		"FooterText":    getSettingValue("footer_text", defaultData["FooterText"].(string), db),
		"ICPRecord":     getSettingValue("icp_record", defaultData["ICPRecord"].(string), db),
		"ICPRecordLink": getSettingValue("icp_record_link", defaultData["ICPRecordLink"].(string), db),
		"PSBRecord":     getSettingValue("psb_record", defaultData["PSBRecord"].(string), db),
		"PSBRecordLink": getSettingValue("psb_record_link", defaultData["PSBRecordLink"].(string), db),
		"title":         "主页",
	}

	c.HTML(http.StatusOK, "index.html", data)
}
