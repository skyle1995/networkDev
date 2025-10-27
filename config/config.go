package config

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/fs"
	"os"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

// ============================================================================
// 结构体定义
// ============================================================================

// ServerConfig 服务器配置结构体
// 包含服务器运行相关的配置信息
type ServerConfig struct {
	Host    string `json:"host" mapstructure:"host"`         // 服务器监听地址
	Port    int    `json:"port" mapstructure:"port"`         // 服务器监听端口
	Dist    string `json:"dist" mapstructure:"dist"`         // 静态文件目录
	DevMode bool   `json:"dev_mode" mapstructure:"dev_mode"` // 开发模式（跳过验证码等）
}

// DatabaseConfig 数据库配置结构体
// 包含数据库连接相关的配置信息
type DatabaseConfig struct {
	Type   string       `json:"type" mapstructure:"type"`     // 数据库类型（mysql/sqlite）
	MySQL  MySQLConfig  `json:"mysql" mapstructure:"mysql"`   // MySQL配置
	SQLite SQLiteConfig `json:"sqlite" mapstructure:"sqlite"` // SQLite配置
}

// MySQLConfig MySQL数据库配置结构体
// 包含MySQL数据库连接的详细配置信息
type MySQLConfig struct {
	Host         string `json:"host" mapstructure:"host"`                     // 数据库主机地址
	Port         int    `json:"port" mapstructure:"port"`                     // 数据库端口
	Username     string `json:"username" mapstructure:"username"`             // 数据库用户名
	Password     string `json:"password" mapstructure:"password"`             // 数据库密码
	Database     string `json:"database" mapstructure:"database"`             // 数据库名称
	Charset      string `json:"charset" mapstructure:"charset"`               // 字符集
	MaxIdleConns int    `json:"max_idle_conns" mapstructure:"max_idle_conns"` // 最大空闲连接数
	MaxOpenConns int    `json:"max_open_conns" mapstructure:"max_open_conns"` // 最大打开连接数
}

// SQLiteConfig SQLite数据库配置结构体
// 包含SQLite数据库文件路径配置
type SQLiteConfig struct {
	Path string `json:"path" mapstructure:"path"` // 数据库文件路径
}

// RedisConfig Redis配置结构体
// 包含Redis连接相关的配置信息
type RedisConfig struct {
	Host     string `json:"host" mapstructure:"host"`         // Redis服务器地址
	Port     int    `json:"port" mapstructure:"port"`         // Redis服务器端口
	Password string `json:"password" mapstructure:"password"` // Redis密码
	DB       int    `json:"db" mapstructure:"db"`             // Redis数据库编号
}

// LogConfig 日志配置结构体
// 包含日志记录相关的配置信息
type LogConfig struct {
	Level      string `json:"level" mapstructure:"level"`             // 日志级别
	File       string `json:"file" mapstructure:"file"`               // 日志文件路径
	MaxSize    int    `json:"max_size" mapstructure:"max_size"`       // 单个日志文件最大大小(MB)
	MaxBackups int    `json:"max_backups" mapstructure:"max_backups"` // 保留的旧日志文件数量
	MaxAge     int    `json:"max_age" mapstructure:"max_age"`         // 日志文件保留天数
}

// CookieConfig Cookie配置结构体
// 包含Cookie相关的安全配置信息
type CookieConfig struct {
	Secure   bool   `json:"secure" mapstructure:"secure"`       // 是否只在HTTPS下发送Cookie
	SameSite string `json:"same_site" mapstructure:"same_site"` // SameSite属性（Strict/Lax/None）
	Domain   string `json:"domain" mapstructure:"domain"`       // Cookie域名
	MaxAge   int    `json:"max_age" mapstructure:"max_age"`     // Cookie最大存活时间（秒）
}

// SecurityConfig 安全配置结构体
// 包含应用程序安全相关的配置信息
type SecurityConfig struct {
	JWTSecret     string       `json:"jwt_secret" mapstructure:"jwt_secret"`         // JWT签名密钥
	EncryptionKey string       `json:"encryption_key" mapstructure:"encryption_key"` // 数据加密密钥
	JWTRefresh    int          `json:"jwt_refresh" mapstructure:"jwt_refresh"`       // JWT令牌刷新阈值（小时）
	Cookie        CookieConfig `json:"cookie" mapstructure:"cookie"`                 // Cookie配置
}

// AppConfig 应用配置结构体
type AppConfig struct {
	Server   ServerConfig   `json:"server" mapstructure:"server"`
	Database DatabaseConfig `json:"database" mapstructure:"database"`
	Redis    RedisConfig    `json:"redis" mapstructure:"redis"`
	Log      LogConfig      `json:"log" mapstructure:"log"`
	Security SecurityConfig `json:"security" mapstructure:"security"`
}

// ============================================================================
// 公共函数
// ============================================================================

// GetDefaultAppConfig 获取默认应用配置
func GetDefaultAppConfig() *AppConfig {
	return &AppConfig{
		Server: ServerConfig{
			Host:    "0.0.0.0",
			Port:    8080,
			Dist:    "",
			DevMode: false,
		},
		Database: DatabaseConfig{
			Type: "sqlite",
			MySQL: MySQLConfig{
				Host:         "localhost",
				Port:         3306,
				Username:     "root",
				Password:     "password",
				Database:     "networkdev",
				Charset:      "utf8mb4",
				MaxIdleConns: 10,
				MaxOpenConns: 100,
			},
			SQLite: SQLiteConfig{
				Path: "./database.db",
			},
		},
		Redis: RedisConfig{
			Host:     "localhost",
			Port:     6379,
			Password: "",
			DB:       0,
		},
		Log: LogConfig{
			Level:      "info",
			File:       "./logs/app.log",
			MaxSize:    100,
			MaxBackups: 5,
			MaxAge:     30,
		},
		Security: SecurityConfig{
			JWTSecret:     "",
			EncryptionKey: "",
			JWTRefresh:    6,
			Cookie: CookieConfig{
				Secure:   true,
				SameSite: "Lax",
				Domain:   "",
				MaxAge:   86400,
			},
		},
	}
}

// GetSecureDefaultAppConfig 获取带有安全密钥的默认应用配置
func GetSecureDefaultAppConfig() (*AppConfig, error) {
	config := GetDefaultAppConfig()

	// 生成安全密钥
	jwtSecret, encryptionKey, err := GenerateSecureKeys()
	if err != nil {
		return nil, err
	}

	// 设置安全密钥
	config.Security.JWTSecret = jwtSecret
	config.Security.EncryptionKey = encryptionKey

	return config, nil
}

// Init 初始化配置文件
func Init(cfgFilePath string) {
	viper.SetConfigFile(cfgFilePath)
	viper.SetConfigType("json")
	viper.AddConfigPath(".")

	if err := viper.ReadInConfig(); err != nil {
		var pathError *fs.PathError
		if errors.As(err, &pathError) {
			log.Warn("未找到配置文件，使用默认配置")

			// 生成带有安全密钥的默认配置
			defaultConfig, configErr := GetSecureDefaultAppConfig()
			if configErr != nil {
				log.WithFields(
					log.Fields{
						"err": configErr,
					},
				).Error("生成安全配置失败，使用基础默认配置")
				defaultConfig = GetDefaultAppConfig()
			}

			// 将配置结构体转换为JSON
			configBytes, marshalErr := json.MarshalIndent(defaultConfig, "", "  ")
			if marshalErr != nil {
				log.WithFields(
					log.Fields{
						"err": marshalErr,
					},
				).Fatal("序列化默认配置失败")
				return
			}

			// 写入配置文件
			err = os.WriteFile(cfgFilePath, configBytes, 0o644)
			if err != nil {
				log.WithFields(
					log.Fields{
						"err": err,
					},
				).Error("写入默认配置文件失败")
			} else {
				log.WithFields(
					log.Fields{
						"file": cfgFilePath,
					},
				).Info("写入默认配置文件成功（已生成安全密钥）")
			}

			// 将配置加载到viper中
			err = viper.ReadConfig(bytes.NewBuffer(configBytes))
			if err != nil {
				log.WithFields(
					log.Fields{
						"err": err,
					},
				).Error("读取默认配置失败")
			} else {
				log.Info("已成功读取默认配置")
			}
		} else {
			log.WithFields(
				log.Fields{
					"err": err,
				},
			).Fatal("配置文件解析错误")
		}
	}
	log.WithFields(
		log.Fields{
			"file": viper.ConfigFileUsed(),
		},
	).Info("使用配置文件")

	// 验证配置
	if _, err := ValidateConfig(); err != nil {
		log.WithFields(
			log.Fields{
				"err": err,
			},
		).Fatal("配置验证失败")
	}
}

// CreateDefaultConfig 创建默认配置文件
func CreateDefaultConfig(filePath string) error {
	// 生成带有安全密钥的默认配置
	defaultConfig, err := GetSecureDefaultAppConfig()
	if err != nil {
		log.WithFields(
			log.Fields{
				"err": err,
			},
		).Error("生成安全配置失败，使用基础默认配置")
		defaultConfig = GetDefaultAppConfig()
	}

	// 将配置结构体转换为JSON
	configBytes, err := json.MarshalIndent(defaultConfig, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filePath, configBytes, 0o644)
}
