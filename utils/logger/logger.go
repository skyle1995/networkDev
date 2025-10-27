package logger

import (
	log "github.com/sirupsen/logrus"
)

// ============================================================================
// 结构体定义
// ============================================================================

// Logger 日志工具结构体
// 封装logrus.Logger，提供统一的日志接口
type Logger struct {
	*log.Logger // 嵌入logrus.Logger，继承其所有方法
}

// ============================================================================
// 构造函数
// ============================================================================

// NewLogger 创建新的日志实例，使用全局logrus配置
// 返回: 新的Logger实例
func NewLogger() *Logger {
	// 使用全局logrus实例而不是创建新实例，确保配置一致性
	return &Logger{Logger: log.StandardLogger()}
}

// InitLogger 初始化HTTP日志处理器
// 创建专门用于HTTP请求日志的Logger实例，使用全局logrus配置
// 返回: 初始化后的Logger实例
func InitLogger() *Logger {
	logger := NewLogger()

	// HTTP日志使用全局logrus的配置
	// 通过使用log.StandardLogger()确保与全局配置保持一致

	// 更新全局日志实例
	SetGlobalLogger(logger)

	return logger
}

// ============================================================================
// 方法函数
// ============================================================================

// WithFields 添加字段到日志条目
// fields: 要添加的字段映射
// 返回: 包含字段的日志条目
func (l *Logger) WithFields(fields log.Fields) *log.Entry {
	return l.Logger.WithFields(fields)
}

// WithField 添加单个字段到日志条目
// key: 字段名
// value: 字段值
// 返回: 包含字段的日志条目
func (l *Logger) WithField(key string, value interface{}) *log.Entry {
	return l.Logger.WithField(key, value)
}

// WithError 添加错误字段到日志条目
// err: 要记录的错误
// 返回: 包含错误信息的日志条目
func (l *Logger) WithError(err error) *log.Entry {
	return l.Logger.WithError(err)
}

// InfoWithFields 记录带字段的信息级别日志
// msg: 日志消息
// fields: 附加字段
func (l *Logger) InfoWithFields(msg string, fields log.Fields) {
	l.WithFields(fields).Info(msg)
}

// ErrorWithFields 记录带字段的错误级别日志
// msg: 日志消息
// fields: 附加字段
func (l *Logger) ErrorWithFields(msg string, fields log.Fields) {
	l.WithFields(fields).Error(msg)
}

// WarnWithFields 记录带字段的警告级别日志
// msg: 日志消息
// fields: 附加字段
func (l *Logger) WarnWithFields(msg string, fields log.Fields) {
	l.WithFields(fields).Warn(msg)
}

// DebugWithFields 记录带字段的调试级别日志
// msg: 日志消息
// fields: 附加字段
func (l *Logger) DebugWithFields(msg string, fields log.Fields) {
	l.WithFields(fields).Debug(msg)
}

// LogError 记录错误日志
// err: 错误对象
// msg: 日志消息
func (l *Logger) LogError(err error, msg string) {
	l.WithError(err).Error(msg)
}

// ============================================================================
// 全局变量
// ============================================================================

// GlobalLogger 全局日志实例
// 提供全局访问的日志记录器
var GlobalLogger *Logger

// init 包初始化函数
// 创建全局日志实例，使用全局logrus配置
func init() {
	GlobalLogger = NewLogger()
}

// ============================================================================
// 全局函数
// ============================================================================

// GetLogger 获取全局日志实例
// 返回: 全局Logger实例
func GetLogger() *Logger {
	return GlobalLogger
}

// SetGlobalLogger 设置全局日志实例
// logger: 要设置的Logger实例
func SetGlobalLogger(logger *Logger) {
	GlobalLogger = logger
}
