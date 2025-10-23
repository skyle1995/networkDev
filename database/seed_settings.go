package database

import (
	"networkDev/models"

	"github.com/sirupsen/logrus"
)

// SeedDefaultSettings 初始化默认系统设置
// - 检查各项设置是否已存在，如不存在则创建默认值
// - 包含站点基本信息、SEO设置等常用配置项
func SeedDefaultSettings() error {
	db, err := GetDB()
	if err != nil {
		return err
	}

	// 定义默认设置项
	defaultSettings := []models.Settings{
		{
			Name:        "site_title",
			Value:       "凌动技术",
			Description: "网站标题，显示在浏览器标题栏和页面顶部",
		},
		{
			Name:        "site_keywords",
			Value:       "验证,网络,管理系统,网络验证,账户管理",
			Description: "网站关键词，用于SEO优化，多个关键词用逗号分隔",
		},
		{
			Name:        "site_description",
			Value:       "专业的网络验证管理系统，提供便捷的在线网络验证服务和账户管理功能",
			Description: "网站描述，用于SEO优化和社交媒体分享",
		},
		{
			Name:        "site_logo",
			Value:       "/assets/logo.png",
			Description: "网站Logo图片路径",
		},
		{
			Name:        "contact_email",
			Value:       "admin@example.com",
			Description: "联系邮箱，用于客服和业务咨询",
		},
		{
			Name:        "max_upload_size",
			Value:       "10485760",
			Description: "文件上传最大尺寸（字节），默认10MB",
		},
		{
			Name:        "default_user_role",
			Value:       "1",
			Description: "新用户默认角色，0=管理员，1=普通用户",
		},
		{
			Name:        "session_timeout",
			Value:       "3600",
			Description: "会话超时时间（秒），默认1小时",
		},
		{
			Name:        "maintenance_mode",
			Value:       "0",
			Description: "系统开关，0=开启系统，1=关闭系统",
		},
		// ===== 页脚与备案相关默认项 =====
		{
			Name:        "footer_text",
			Value:       "Copyright © 2025 凌动技术. All Rights Reserved.",
			Description: "页脚展示的版权或说明信息",
		},
		{
			Name:        "icp_record",
			Value:       "京ICP备12345678号",
			Description: "ICP备案号，留空则不显示",
		},
		{
			Name:        "icp_record_link",
			Value:       "https://beian.miit.gov.cn",
			Description: "工信部ICP备案查询链接，留空则不显示",
		},
		{
			Name:        "psb_record",
			Value:       "京公网安备 11000002000001号",
			Description: "公安备案号，留空则不显示",
		},
		{
			Name:        "psb_record_link",
			Value:       "https://www.beian.gov.cn/portal/registerSystemInfo?recordcode=11000002000001",
			Description: "公安备案查询链接，留空则不显示",
		},
		{
			Name:        "card_batch_counter",
			Value:       "0",
			Description: "卡密批次号计数器（用于记录上次生成批次号的序号，自增使用）",
		},
	}

	// 逐个检查并创建不存在的设置项
	for _, setting := range defaultSettings {
		var count int64
		if err := db.Model(&models.Settings{}).Where("name = ?", setting.Name).Count(&count).Error; err != nil {
			return err
		}

		if count == 0 {
			if err := db.Create(&setting).Error; err != nil {
				logrus.WithError(err).WithField("name", setting.Name).Error("创建默认设置失败")
				return err
			}
			logrus.WithField("name", setting.Name).WithField("value", setting.Value).Info("创建默认设置项")
		}
	}

	logrus.Info("默认系统设置初始化完成")
	return nil
}
