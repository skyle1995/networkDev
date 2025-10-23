package database

import (
	"networkDev/models"
	"networkDev/utils"

	"github.com/sirupsen/logrus"
)

// SeedDefaultAdmin 初始化默认管理员账号
// - 如果用户名为 admin 的用户已存在，则跳过
// - 如不存在，则创建用户名为 admin、密码为 admin123（以 bcrypt 哈希存储）、角色 Role=0 的管理员
// - 根据需求：默认 admin 用户的 ID 固定为 10000
func SeedDefaultAdmin() error {
	db, err := GetDB()
	if err != nil {
		return err
	}

	// 检查是否存在用户名为 admin 的用户
	var count int64
	if dbErr := db.Model(&models.User{}).Where("username = ?", "admin").Count(&count).Error; dbErr != nil {
		return dbErr
	}
	if count > 0 {
		return nil
	}

	// 生成密码盐值
	salt, err := utils.GenerateRandomSalt()
	if err != nil {
		return err
	}

	// 使用盐值生成密码哈希（不存明文）
	hash, err := utils.HashPasswordWithSalt("admin123", salt)
	if err != nil {
		return err
	}

	// 创建默认管理员（指定固定 ID=10000）
	admin := models.User{
		ID:           10000,
		Username:     "admin",
		Password:     hash,
		PasswordSalt: salt,
		Role:         0, // 0=管理员
	}
	if err := db.Create(&admin).Error; err != nil {
		return err
	}
	logrus.WithField("username", "admin").WithField("id", admin.ID).Info("默认管理员创建成功")
	return nil
}
