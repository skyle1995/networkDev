package database

import (
	"networkDev/models"
	"networkDev/utils"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
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
			Value:       "验证,网络,管理系统,网络验证,卡密管理,账户管理",
			Description: "网站关键词，用于SEO优化，多个关键词用逗号分隔",
		},
		{
			Name:        "site_description",
			Value:       "专业的网络验证管理系统，提供便捷的在线网络验证服务和设备管理功能",
			Description: "网站描述，用于SEO优化和社交媒体分享",
		},
		{
			Name:        "site_logo",
			Value:       "/favicon.ico",
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
			Description: "维护模式，0=关闭维护模式，1=开启维护模式",
		},
		// ===== 管理员账号相关默认项 =====
		{
			Name:        "admin_username",
			Value:       "admin",
			Description: "管理员用户名",
		},
		{
			Name:        "admin_password",
			Value:       "",
			Description: "管理员密码哈希值",
		},
		{
			Name:        "admin_password_salt",
			Value:       "",
			Description: "管理员密码加密盐值",
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

	// 初始化默认管理员账号（如果密码为空）
	if err := initDefaultAdmin(db); err != nil {
		return err
	}

	logrus.Info("默认系统设置初始化完成")
	return nil
}

// initDefaultAdmin 初始化默认管理员账号
// 如果admin_password为空，则生成默认密码admin123的哈希值
func initDefaultAdmin(db *gorm.DB) error {
	var passwordSetting models.Settings
	if err := db.Where("name = ?", "admin_password").First(&passwordSetting).Error; err != nil {
		logrus.WithError(err).Error("获取管理员密码设置失败")
		return err
	}

	// 如果密码已设置，跳过初始化
	if passwordSetting.Value != "" {
		logrus.Debug("管理员密码已设置，跳过默认密码初始化")
		return nil
	}

	// 生成密码盐值
	salt, err := utils.GenerateRandomSalt()
	if err != nil {
		logrus.WithError(err).Error("生成密码盐值失败")
		return err
	}

	// 使用盐值生成密码哈希（默认密码：admin123）
	hash, err := utils.HashPasswordWithSalt("admin123", salt)
	if err != nil {
		logrus.WithError(err).Error("生成密码哈希失败")
		return err
	}

	// 更新密码和盐值
	if err := db.Model(&models.Settings{}).Where("name = ?", "admin_password").Update("value", hash).Error; err != nil {
		logrus.WithError(err).Error("更新管理员密码失败")
		return err
	}

	if err := db.Model(&models.Settings{}).Where("name = ?", "admin_password_salt").Update("value", salt).Error; err != nil {
		logrus.WithError(err).Error("更新管理员密码盐值失败")
		return err
	}

	logrus.Info("默认管理员账号初始化完成，用户名: admin, 密码: admin123")
	return nil
}
