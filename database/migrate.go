package database

import (
	"fmt"
	"networkDev/models"
	"strings"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// AutoMigrate 自动迁移数据库模型
// - 会确保必要的数据表结构存在
// - 不会破坏已有数据
func AutoMigrate() error {
	db, err := GetDB()
	if err != nil {
		return err
	}
	if err := db.AutoMigrate(&models.User{}, &models.Settings{}, &models.App{}, &models.API{}); err != nil {
		logrus.WithError(err).Error("AutoMigrate 执行失败")
		return err
	}

	// 兼容迁移：如果 users.password_salt 列长度 < 64，则扩大到 64
	if err := ensureUserPasswordSaltLength(db); err != nil {
		logrus.WithError(err).Error("调整 users.password_salt 列长度失败")
		return err
	}

	// 兼容迁移：确保 tasks.verification_code 字段类型为 LONGTEXT 以支持大图片数据
	if err := ensureVerificationCodeType(db); err != nil {
		logrus.WithError(err).Error("调整 tasks.verification_code 字段类型失败")
		return err
	}

	logrus.Info("AutoMigrate 执行完成")
	return nil
}

// ensureVerificationCodeType 确保tasks.verification_code字段类型为LONGTEXT以支持大图片数据
// 中文注释：检查并修改verification_code字段类型，支持Base64编码的大图片数据存储
func ensureVerificationCodeType(db *gorm.DB) error {
	// 获取数据库方言类型
	dialector := db.Dialector.Name()

	// 根据不同数据库类型执行不同的检查逻辑
	switch dialector {
	case "mysql":
		// MySQL/MariaDB使用INFORMATION_SCHEMA
		var result struct {
			ColumnName string `gorm:"column:COLUMN_NAME"`
			ColumnType string `gorm:"column:COLUMN_TYPE"`
		}

		err := db.Raw("SELECT COLUMN_NAME, COLUMN_TYPE FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_NAME = ? AND COLUMN_NAME = ? LIMIT 1",
			"tasks", "verification_code").Scan(&result).Error

		if err != nil {
			return nil // 查询失败则跳过
		}

		// 检查列类型，如果不是LONGTEXT则修改
		if !strings.Contains(strings.ToLower(result.ColumnType), "longtext") {
			alterSQL := "ALTER TABLE tasks MODIFY COLUMN verification_code LONGTEXT"
			if err := db.Exec(alterSQL).Error; err != nil {
				return fmt.Errorf("修改verification_code字段类型失败: %v", err)
			}
			logrus.Info("verification_code字段类型已更新为LONGTEXT")
		}
	case "sqlite":
		// SQLite使用pragma_table_info检查列信息
		var columns []struct {
			CID       int     `gorm:"column:cid"`
			Name      string  `gorm:"column:name"`
			Type      string  `gorm:"column:type"`
			NotNull   int     `gorm:"column:notnull"`
			DfltValue *string `gorm:"column:dflt_value"`
			PK        int     `gorm:"column:pk"`
		}

		err := db.Raw("PRAGMA table_info(tasks)").Scan(&columns).Error
		if err != nil {
			return nil // 查询失败则跳过
		}

		// 查找verification_code列
		for _, col := range columns {
			if col.Name == "verification_code" {
				// SQLite中，如果列类型不是TEXT，需要重建表
				if !strings.Contains(strings.ToLower(col.Type), "text") {
					// SQLite不支持直接修改列类型，但GORM的AutoMigrate会处理这种情况
					logrus.Info("SQLite检测到verification_code字段类型需要更新，依赖GORM AutoMigrate处理")
				}
				break
			}
		}
	default:
		// 其他数据库类型暂不处理
		logrus.Infof("数据库类型 %s 暂不支持verification_code字段类型检查", dialector)
	}

	return nil
}

// ensureUserPasswordSaltLength 确保users.password_salt列长度至少为64
// 中文注释：检查并修改password_salt列长度，兼容32字节（64十六进制字符）的盐值
func ensureUserPasswordSaltLength(db *gorm.DB) error {
	// 获取数据库方言类型
	dialector := db.Dialector.Name()

	// 根据不同数据库类型执行不同的检查逻辑
	switch dialector {
	case "mysql":
		// MySQL/MariaDB使用INFORMATION_SCHEMA
		var result struct {
			ColumnName string `gorm:"column:COLUMN_NAME"`
			ColumnType string `gorm:"column:COLUMN_TYPE"`
		}

		err := db.Raw("SELECT COLUMN_NAME, COLUMN_TYPE FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_NAME = ? AND COLUMN_NAME = ? LIMIT 1",
			"users", "password_salt").Scan(&result).Error

		if err != nil {
			return nil // 查询失败则跳过
		}

		// 检查列类型，如果长度小于64则修改
		if strings.Contains(strings.ToLower(result.ColumnType), "varchar") {
			if strings.Contains(result.ColumnType, "(32)") || strings.Contains(result.ColumnType, "(16)") {
				alterSQL := "ALTER TABLE users MODIFY COLUMN password_salt VARCHAR(64)"
				if err := db.Exec(alterSQL).Error; err != nil {
					return fmt.Errorf("修改password_salt列长度失败: %v", err)
				}
				logrus.Info("password_salt列长度已更新为64")
			}
		}
	case "sqlite":
		// SQLite使用pragma_table_info检查列信息
		var columns []struct {
			CID       int     `gorm:"column:cid"`
			Name      string  `gorm:"column:name"`
			Type      string  `gorm:"column:type"`
			NotNull   int     `gorm:"column:notnull"`
			DfltValue *string `gorm:"column:dflt_value"`
			PK        int     `gorm:"column:pk"`
		}

		err := db.Raw("PRAGMA table_info(users)").Scan(&columns).Error
		if err != nil {
			return nil // 查询失败则跳过
		}

		// 查找password_salt列
		for _, col := range columns {
			if col.Name == "password_salt" {
				// SQLite中，如果列类型包含长度限制且小于64，需要重建表
				if strings.Contains(strings.ToLower(col.Type), "varchar(32)") ||
					strings.Contains(strings.ToLower(col.Type), "varchar(16)") {
					// SQLite不支持直接修改列类型，但GORM的AutoMigrate会处理这种情况
					logrus.Info("SQLite检测到password_salt列长度需要更新，依赖GORM AutoMigrate处理")
				}
				break
			}
		}
	default:
		// 其他数据库类型暂不处理
		logrus.Infof("数据库类型 %s 暂不支持password_salt列长度检查", dialector)
	}

	return nil
}
