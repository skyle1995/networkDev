package config

import (
	"errors"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

// ServerConfig 服务器配置结构体
// 包含HTTP服务器的基本配置信息
type ServerConfig struct {
	Host string `json:"host" mapstructure:"host"` // 服务器监听地址
	Port int    `json:"port" mapstructure:"port"` // 服务器监听端口
	Mode string `json:"mode" mapstructure:"mode"` // 运行模式（debug/release）
	Dist string `json:"dist" mapstructure:"dist"` // 静态文件目录
}

// DatabaseConfig 数据库配置结构体
// 支持MySQL和SQLite两种数据库类型
type DatabaseConfig struct {
	Type   string       `json:"type" mapstructure:"type"`     // 数据库类型（mysql/sqlite）
	MySQL  MySQLConfig  `json:"mysql" mapstructure:"mysql"`   // MySQL配置
	SQLite SQLiteConfig `json:"sqlite" mapstructure:"sqlite"` // SQLite配置
}

// MySQLConfig MySQL数据库配置结构体
// 包含MySQL数据库连接和连接池的配置信息
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
// 包含Redis缓存服务器的连接配置
type RedisConfig struct {
	Host     string `json:"host" mapstructure:"host"`         // Redis服务器地址
	Port     int    `json:"port" mapstructure:"port"`         // Redis服务器端口
	Password string `json:"password" mapstructure:"password"` // Redis密码
	DB       int    `json:"db" mapstructure:"db"`             // Redis数据库编号
}

// LogConfig 日志配置结构体
// 包含日志记录的相关配置信息
type LogConfig struct {
	Level      string `json:"level" mapstructure:"level"`             // 日志级别
	File       string `json:"file" mapstructure:"file"`               // 日志文件路径
	MaxSize    int    `json:"max_size" mapstructure:"max_size"`       // 单个日志文件最大大小(MB)
	MaxBackups int    `json:"max_backups" mapstructure:"max_backups"` // 保留的旧日志文件数量
	MaxAge     int    `json:"max_age" mapstructure:"max_age"`         // 日志文件保留天数
}

// SecurityConfig 安全配置结构体
// 包含应用程序安全相关的配置信息
type SecurityConfig struct {
	JWTSecret                string `json:"jwt_secret" mapstructure:"jwt_secret"`                                   // JWT签名密钥
	EncryptionKey            string `json:"encryption_key" mapstructure:"encryption_key"`                           // 数据加密密钥
	JWTRefreshThresholdHours int    `json:"jwt_refresh_threshold_hours" mapstructure:"jwt_refresh_threshold_hours"` // JWT令牌刷新阈值（小时）
}

// AppConfig 应用配置结构体
type AppConfig struct {
	Server   ServerConfig   `json:"server" mapstructure:"server"`
	Database DatabaseConfig `json:"database" mapstructure:"database"`
	Redis    RedisConfig    `json:"redis" mapstructure:"redis"`
	Log      LogConfig      `json:"log" mapstructure:"log"`
	Security SecurityConfig `json:"security" mapstructure:"security"`
}

// ValidateAndSetDefaults 验证配置并设置默认值
func ValidateAndSetDefaults() (*AppConfig, error) {
	var config AppConfig

	// 解析配置到结构体
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("解析配置失败: %w", err)
	}

	// 设置默认值
	setDefaults(&config)

	// 验证配置
	if err := validateConfig(&config); err != nil {
		return nil, fmt.Errorf("配置验证失败: %w", err)
	}

	log.Info("配置验证通过")
	return &config, nil
}

// setDefaults 设置默认值
func setDefaults(config *AppConfig) {
	// 服务器默认值
	if config.Server.Host == "" {
		config.Server.Host = "0.0.0.0"
	}
	if config.Server.Port == 0 {
		config.Server.Port = 8080
	}
	if config.Server.Mode == "" {
		config.Server.Mode = "debug"
	}

	// 数据库默认值
	if config.Database.Type == "" {
		config.Database.Type = "sqlite"
	}
	if config.Database.MySQL.Port == 0 {
		config.Database.MySQL.Port = 3306
	}
	if config.Database.MySQL.Charset == "" {
		config.Database.MySQL.Charset = "utf8mb4"
	}
	if config.Database.MySQL.MaxIdleConns == 0 {
		config.Database.MySQL.MaxIdleConns = 10
	}
	if config.Database.MySQL.MaxOpenConns == 0 {
		config.Database.MySQL.MaxOpenConns = 100
	}
	if config.Database.SQLite.Path == "" {
		config.Database.SQLite.Path = "./recharge.db"
	}

	// Redis默认值
	if config.Redis.Host == "" {
		config.Redis.Host = "localhost"
	}
	if config.Redis.Port == 0 {
		config.Redis.Port = 6379
	}

	// 日志默认值
	if config.Log.Level == "" {
		config.Log.Level = "info"
	}
	// 不为空的日志文件路径设置默认值，保持为空表示只输出到控制台
	if config.Log.MaxSize == 0 {
		config.Log.MaxSize = 100
	}
	if config.Log.MaxBackups == 0 {
		config.Log.MaxBackups = 5
	}
	if config.Log.MaxAge == 0 {
		config.Log.MaxAge = 30
	}

	// 安全配置默认值
	if config.Security.JWTSecret == "" || config.Security.JWTSecret == "your-jwt-secret-key" {
		config.Security.JWTSecret = "default-jwt-secret-change-in-production"
	}
	if config.Security.EncryptionKey == "" || config.Security.EncryptionKey == "your-encryption-key" {
		config.Security.EncryptionKey = "default-encryption-key-change-in-production"
	}
	if config.Security.JWTRefreshThresholdHours == 0 {
		config.Security.JWTRefreshThresholdHours = 6
	}
}

// validateConfig 验证配置
func validateConfig(config *AppConfig) error {
	// 验证服务器配置
	if err := validateServerConfig(&config.Server); err != nil {
		return fmt.Errorf("服务器配置错误: %w", err)
	}

	// 验证数据库配置
	if err := validateDatabaseConfig(&config.Database); err != nil {
		return fmt.Errorf("数据库配置错误: %w", err)
	}

	// 验证Redis配置
	if err := validateRedisConfig(&config.Redis); err != nil {
		return fmt.Errorf("redis配置错误: %w", err)
	}

	// 验证日志配置
	if err := validateLogConfig(&config.Log); err != nil {
		return fmt.Errorf("日志配置错误: %w", err)
	}

	// 验证安全配置
	if err := validateSecurityConfig(&config.Security); err != nil {
		return fmt.Errorf("安全配置错误: %w", err)
	}

	return nil
}

// validateServerConfig 验证服务器配置
func validateServerConfig(config *ServerConfig) error {
	// 验证主机地址
	if config.Host != "" {
		if ip := net.ParseIP(config.Host); ip == nil && config.Host != "localhost" {
			return fmt.Errorf("无效的主机地址: %s", config.Host)
		}
	}

	// 验证端口
	if config.Port < 1 || config.Port > 65535 {
		return fmt.Errorf("无效的端口号: %d，端口号必须在1-65535之间", config.Port)
	}

	// 验证运行模式
	validModes := []string{"debug", "release", "test"}
	if !contains(validModes, config.Mode) {
		return fmt.Errorf("无效的运行模式: %s，支持的模式: %s", config.Mode, strings.Join(validModes, ", "))
	}

	return nil
}

// validateDatabaseConfig 验证数据库配置
func validateDatabaseConfig(config *DatabaseConfig) error {
	// 验证数据库类型
	validTypes := []string{"mysql", "sqlite"}
	if !contains(validTypes, config.Type) {
		return fmt.Errorf("不支持的数据库类型: %s，支持的类型: %s", config.Type, strings.Join(validTypes, ", "))
	}

	// 根据类型验证具体配置
	switch config.Type {
	case "mysql":
		return validateMySQLConfig(&config.MySQL)
	case "sqlite":
		return validateSQLiteConfig(&config.SQLite)
	}

	return nil
}

// validateMySQLConfig 验证MySQL配置
func validateMySQLConfig(config *MySQLConfig) error {
	if config.Host == "" {
		return errors.New("MySQL主机地址不能为空")
	}
	if config.Port < 1 || config.Port > 65535 {
		return fmt.Errorf("无效的MySQL端口号: %d", config.Port)
	}
	if config.Username == "" {
		return errors.New("MySQL用户名不能为空")
	}
	if config.Database == "" {
		return errors.New("MySQL数据库名不能为空")
	}
	if config.MaxIdleConns < 0 {
		return errors.New("MySQL最大空闲连接数不能为负数")
	}
	if config.MaxOpenConns < 0 {
		return errors.New("MySQL最大打开连接数不能为负数")
	}
	return nil
}

// validateSQLiteConfig 验证SQLite配置
func validateSQLiteConfig(config *SQLiteConfig) error {
	if config.Path == "" {
		return errors.New("SQLite数据库路径不能为空")
	}

	// 检查目录是否存在，不存在则创建
	dir := filepath.Dir(config.Path)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("创建SQLite数据库目录失败: %w", err)
		}
	}

	return nil
}

// validateRedisConfig 验证Redis配置
func validateRedisConfig(config *RedisConfig) error {
	if config.Host == "" {
		return errors.New("Redis主机地址不能为空")
	}
	if config.Port < 1 || config.Port > 65535 {
		return fmt.Errorf("无效的Redis端口号: %d", config.Port)
	}
	if config.DB < 0 || config.DB > 15 {
		return fmt.Errorf("无效的Redis数据库索引: %d，必须在0-15之间", config.DB)
	}
	return nil
}

// validateLogConfig 验证日志配置
func validateLogConfig(config *LogConfig) error {
	// 验证日志级别
	validLevels := []string{"trace", "debug", "info", "warn", "error", "fatal", "panic"}
	if !contains(validLevels, config.Level) {
		return fmt.Errorf("无效的日志级别: %s，支持的级别: %s", config.Level, strings.Join(validLevels, ", "))
	}

	// 检查日志文件目录（仅当日志文件路径不为空时）
	if config.File != "" {
		dir := filepath.Dir(config.File)
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			if err := os.MkdirAll(dir, 0755); err != nil {
				return fmt.Errorf("创建日志目录失败: %w", err)
			}
		}
	}
	// 当日志文件路径为空时，不进行目录检查和创建

	// 验证日志轮转配置
	if config.MaxSize <= 0 {
		return errors.New("日志文件最大大小必须大于0")
	}
	if config.MaxBackups < 0 {
		return errors.New("日志备份文件数量不能为负数")
	}
	if config.MaxAge < 0 {
		return errors.New("日志文件保留天数不能为负数")
	}

	return nil
}

// validateSecurityConfig 验证安全配置
func validateSecurityConfig(config *SecurityConfig) error {
	if len(config.JWTSecret) < 16 {
		return errors.New("JWT密钥长度不能少于16个字符")
	}

	if len(config.EncryptionKey) < 16 {
		return errors.New("加密密钥长度不能少于16个字符")
	}

	if config.JWTRefreshThresholdHours < 1 || config.JWTRefreshThresholdHours > 23 {
		return errors.New("JWT令牌刷新阈值必须在1-23小时之间")
	}

	// 检查是否使用默认值（生产环境警告）
	if strings.Contains(config.JWTSecret, "default") {
		log.Warn("检测到使用默认JWT密钥，生产环境请更换为安全的密钥")
	}

	if strings.Contains(config.EncryptionKey, "default") {
		log.Warn("检测到使用默认加密密钥，生产环境请更换为安全的密钥")
	}

	return nil
}

// contains 检查切片是否包含指定元素
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// GetConfigValue 获取配置值，支持类型转换和默认值
func GetConfigValue[T any](key string, defaultValue T) T {
	if !viper.IsSet(key) {
		return defaultValue
	}

	value := viper.Get(key)
	if result, ok := value.(T); ok {
		return result
	}

	// 尝试类型转换
	if converted, err := convertValue[T](value); err == nil {
		return converted
	}

	return defaultValue
}

// convertValue 尝试类型转换
func convertValue[T any](value interface{}) (T, error) {
	var zero T
	str := fmt.Sprintf("%v", value)

	switch any(zero).(type) {
	case int:
		if i, err := strconv.Atoi(str); err == nil {
			return any(i).(T), nil
		}
	case string:
		return any(str).(T), nil
	case bool:
		if b, err := strconv.ParseBool(str); err == nil {
			return any(b).(T), nil
		}
	}

	return zero, fmt.Errorf("无法转换类型")
}
