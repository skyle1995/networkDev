package database

import (
	"fmt"
	"networkDev/utils"
	"path/filepath"
	"sync"

	"github.com/glebarez/sqlite"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// ============================================================================
// 全局变量
// ============================================================================

var (
	// dbInstance 全局 *gorm.DB 实例，使用单例确保全局复用
	dbInstance *gorm.DB
	// once 确保初始化只执行一次
	once sync.Once
)

// ============================================================================
// 公共函数
// ============================================================================

// Init 初始化数据库连接（根据配置自动选择驱动）
// - 默认使用 SQLite（github.com/glebarez/sqlite）
// - 生产环境支持 MySQL（gorm.io/driver/mysql）
func Init() (*gorm.DB, error) {
	var initErr error
	once.Do(func() {
		dbType := viper.GetString("database.type")
		switch dbType {
		case "mysql":
			initErr = initMySQL()
		default:
			initErr = initSQLite()
		}

		// 如果数据库初始化成功，配置连接池和启动健康检查
		if initErr == nil && dbInstance != nil {
			// 加载数据库配置
			var configPrefix string
			if dbType == "mysql" {
				configPrefix = "database.mysql"
			} else {
				configPrefix = "database.sqlite"
			}

			dbConfig := utils.LoadDatabaseConfig(configPrefix)

			// 验证配置
			if err := utils.ValidateDatabaseConfig(dbConfig); err != nil {
				logrus.WithError(err).Warn("数据库配置验证失败，使用默认配置")
				dbConfig = utils.GetDefaultDatabaseConfig()
			}

			// 配置连接池
			if err := utils.ConfigureConnectionPool(dbInstance, dbConfig); err != nil {
				logrus.WithError(err).Error("配置数据库连接池失败")
			}

			// 启动健康检查
			utils.StartHealthCheck(dbInstance, dbConfig)
		}
	})
	return dbInstance, initErr
}

// GetDB 获取全局 *gorm.DB 实例
// 如果未初始化，会尝试初始化一次
func GetDB() (*gorm.DB, error) {
	if dbInstance != nil {
		return dbInstance, nil
	}
	return Init()
}

// ============================================================================
// 私有函数
// ============================================================================

// initSQLite 初始化 SQLite 数据库
// 使用 viper 中的 database.sqlite.path 作为数据库文件路径
func initSQLite() error {
	path := viper.GetString("database.sqlite.path")
	if path == "" {
		path = "./database.db"
	}
	
	// 确保数据库路径为绝对路径
	absolutePath, err := utils.EnsureAbsolutePath(path)
	if err != nil {
		logrus.WithError(err).Error("转换SQLite数据库路径为绝对路径失败")
		return err
	}
	
	dsn := fmt.Sprintf("file:%s?cache=shared&_busy_timeout=5000&_fk=1", absolutePath)
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		logrus.WithError(err).Error("SQLite 初始化失败")
		return err
	}

	// SQLite 连接池配置（SQLite 对连接池支持有限，但仍可设置基本参数）
	if sqlDB, err := db.DB(); err == nil {
		// SQLite 通常使用单连接，但可以设置一些基本参数
		sqlDB.SetMaxOpenConns(1) // SQLite 建议使用单连接
		sqlDB.SetMaxIdleConns(1)
	}

	dbInstance = db
	// 记录连接成功信息（只显示文件名，不泄露完整路径）
	fileName := filepath.Base(absolutePath)
	logrus.WithField("file", fileName).Info("SQLite 连接已建立")
	return nil
}

// initMySQL 初始化 MySQL 数据库
// 从 viper 读取 database.mysql.* 配置构建 DSN
func initMySQL() error {
	host := viper.GetString("database.mysql.host")
	port := viper.GetInt("database.mysql.port")
	user := viper.GetString("database.mysql.username")
	pass := viper.GetString("database.mysql.password")
	dbname := viper.GetString("database.mysql.database")
	charset := viper.GetString("database.mysql.charset")
	if charset == "" {
		charset = "utf8mb4"
	}

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=True&loc=Local", user, pass, host, port, dbname, charset)
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		logrus.WithError(err).Error("MySQL 初始化失败")
		return err
	}

	dbInstance = db
	logrus.WithField("host", host).WithField("database", dbname).Info("MySQL 连接已建立")
	return nil
}
