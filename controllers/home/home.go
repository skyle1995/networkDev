package home

import (
	"net/http"
	"networkDev/database"
	"networkDev/models"
	"networkDev/services"
	"networkDev/utils"
)

// RootHandler 主页处理器
func RootHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	// 获取数据库连接
	db, err := database.GetDB()
	if err != nil {
		http.Error(w, "数据库连接失败", http.StatusInternalServerError)
		return
	}

	// 从数据库获取站点标题和页脚文本
	siteTitle, err := services.FindSettingByName("site_title", db)
	if err != nil {
		siteTitle = &models.Settings{Value: "凌动技术"}
	}

	footerText, err := services.FindSettingByName("footer_text", db)
	if err != nil {
		footerText = &models.Settings{Value: "© 2025 凌动技术 保留所有权利"}
	}

	// 从数据库获取备案信息
	icpRecord, err := services.FindSettingByName("icp_record", db)
	if err != nil {
		icpRecord = &models.Settings{Value: ""}
	}

	icpRecordLink, err := services.FindSettingByName("icp_record_link", db)
	if err != nil {
		icpRecordLink = &models.Settings{Value: "https://beian.miit.gov.cn"}
	}

	// 从数据库获取公安备案信息
	psbRecord, err := services.FindSettingByName("psb_record", db)
	if err != nil {
		psbRecord = &models.Settings{Value: ""}
	}

	psbRecordLink, err := services.FindSettingByName("psb_record_link", db)
	if err != nil {
		psbRecordLink = &models.Settings{Value: "https://www.beian.gov.cn"}
	}

	// 准备模板数据
	data := map[string]interface{}{
		"SystemName":    siteTitle.Value,
		"FooterText":    footerText.Value,
		"ICPRecord":     icpRecord.Value,
		"ICPRecordLink": icpRecordLink.Value,
		"PSBRecord":     psbRecord.Value,
		"PSBRecordLink": psbRecordLink.Value,
		"title":         "主页",
	}

	if err := utils.RenderTemplate(w, "index.html", data); err != nil {
		http.Error(w, "页面加载失败", http.StatusInternalServerError)
		return
	}
}
