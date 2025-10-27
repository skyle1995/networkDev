package utils

import (
	"encoding/json"
	"fmt"
	"log"
	"runtime"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// ============================================================================
// 结构体定义
// ============================================================================

// ErrorResponse 统一的错误响应结构
// 用于标准化API错误响应格式
type ErrorResponse struct {
	Success   bool        `json:"success"`              // 请求是否成功，错误响应时固定为false
	Message   string      `json:"message"`              // 错误消息描述
	ErrorCode string      `json:"error_code,omitempty"` // 错误代码，用于客户端识别错误类型
	Data      interface{} `json:"data"`                 // 附加数据，可为空
	Timestamp int64       `json:"timestamp"`            // 响应时间戳
}

// SuccessResponse 统一的成功响应结构
// 用于标准化API成功响应格式
type SuccessResponse struct {
	Success   bool        `json:"success"`   // 请求是否成功，成功响应时固定为true
	Message   string      `json:"message"`   // 成功消息描述
	Data      interface{} `json:"data"`      // 响应数据
	Timestamp int64       `json:"timestamp"` // 响应时间戳
}

// ============================================================================
// 常量定义
// ============================================================================

// ErrorCode 错误代码常量
// 定义标准化的错误代码，用于客户端识别和处理不同类型的错误
const (
	ErrCodeInvalidRequest   = "INVALID_REQUEST"   // 无效请求，通常是请求参数格式错误
	ErrCodeUnauthorized     = "UNAUTHORIZED"      // 未授权，需要登录或token无效
	ErrCodeForbidden        = "FORBIDDEN"         // 禁止访问，权限不足
	ErrCodeNotFound         = "NOT_FOUND"         // 资源不存在
	ErrCodeConflict         = "CONFLICT"          // 资源冲突，如重复创建
	ErrCodeInternalError    = "INTERNAL_ERROR"    // 服务器内部错误
	ErrCodeDatabaseError    = "DATABASE_ERROR"    // 数据库操作错误
	ErrCodeValidationError  = "VALIDATION_ERROR"  // 数据验证错误
	ErrCodeTokenExpired     = "TOKEN_EXPIRED"     // 令牌已过期
	ErrCodeInsufficientData = "INSUFFICIENT_DATA" // 数据不足，缺少必要信息
)

// LogLevel 日志级别
// 定义不同的日志记录级别
type LogLevel int

const (
	LogLevelInfo  LogLevel = iota // 信息级别，记录一般信息
	LogLevelWarn                  // 警告级别，记录潜在问题
	LogLevelError                 // 错误级别，记录错误信息
	LogLevelDebug                 // 调试级别，记录调试信息
)

// LogEntry 日志条目结构
// 包含完整的日志信息，用于结构化日志记录
type LogEntry struct {
	Level     LogLevel    `json:"level"`             // 日志级别
	Message   string      `json:"message"`           // 日志消息
	Error     string      `json:"error,omitempty"`   // 错误信息，仅在错误日志中存在
	Context   interface{} `json:"context,omitempty"` // 上下文信息，额外的结构化数据
	Timestamp time.Time   `json:"timestamp"`         // 日志时间戳
	File      string      `json:"file"`              // 源文件路径
	Line      int         `json:"line"`              // 源文件行号
}

// ============================================================================
// 响应函数
// ============================================================================

// WriteErrorResponse 写入错误响应
// c: Gin上下文
// statusCode: HTTP状态码
// message: 错误消息
// errorCode: 错误代码
// data: 附加数据
func WriteErrorResponse(c *gin.Context, statusCode int, message, errorCode string, data interface{}) {
	response := ErrorResponse{
		Success:   false,
		Message:   message,
		ErrorCode: errorCode,
		Data:      data,
		Timestamp: time.Now().Unix(),
	}

	c.JSON(statusCode, response)
}

// WriteSuccessResponse 写入成功响应
// c: Gin上下文
// statusCode: HTTP状态码
// message: 成功消息
// data: 响应数据
func WriteSuccessResponse(c *gin.Context, statusCode int, message string, data interface{}) {
	response := SuccessResponse{
		Success:   true,
		Message:   message,
		Data:      data,
		Timestamp: time.Now().Unix(),
	}

	c.JSON(statusCode, response)
}

// ============================================================================
// 错误处理函数
// ============================================================================

// HandleDatabaseError 处理数据库错误
// c: Gin上下文
// err: 数据库错误
// operation: 操作描述
func HandleDatabaseError(c *gin.Context, err error, operation string) {
	if err == gorm.ErrRecordNotFound {
		LogWarn(fmt.Sprintf("Record not found during %s", operation), map[string]interface{}{
			"operation": operation,
			"error":     err.Error(),
		})
		WriteErrorResponse(c, 404, "记录不存在", ErrCodeNotFound, nil)
		return
	}

	LogError(fmt.Sprintf("Database error during %s", operation), err, map[string]interface{}{
		"operation": operation,
	})
	WriteErrorResponse(c, 500, "数据库操作失败", ErrCodeDatabaseError, nil)
}

// HandleValidationError 处理验证错误
// c: Gin上下文
// message: 验证错误消息
// details: 验证错误详情
func HandleValidationError(c *gin.Context, message string, details interface{}) {
	LogWarn("Validation error: "+message, map[string]interface{}{
		"details": details,
	})
	WriteErrorResponse(c, 400, message, ErrCodeValidationError, details)
}

// HandleUnauthorizedError 处理未授权错误
// c: Gin上下文
// message: 错误消息
func HandleUnauthorizedError(c *gin.Context, message string) {
	LogWarn("Unauthorized access: "+message, nil)
	WriteErrorResponse(c, 401, message, ErrCodeUnauthorized, nil)
}

// HandleInternalError 处理内部错误
// c: Gin上下文
// err: 错误
// operation: 操作描述
func HandleInternalError(c *gin.Context, err error, operation string) {
	LogError(fmt.Sprintf("Internal error during %s", operation), err, map[string]interface{}{
		"operation": operation,
	})
	WriteErrorResponse(c, 500, "服务器内部错误", ErrCodeInternalError, nil)
}

// ============================================================================
// 日志函数
// ============================================================================

// LogInfo 记录信息日志
// message: 日志消息
// context: 上下文信息
func LogInfo(message string, context interface{}) {
	logEntry := createLogEntry(LogLevelInfo, message, nil, context)
	printLog(logEntry)
}

// LogWarn 记录警告日志
// message: 日志消息
// context: 上下文信息
func LogWarn(message string, context interface{}) {
	logEntry := createLogEntry(LogLevelWarn, message, nil, context)
	printLog(logEntry)
}

// LogError 记录错误日志
// message: 日志消息
// err: 错误对象
// context: 上下文信息
func LogError(message string, err error, context interface{}) {
	errorStr := ""
	if err != nil {
		errorStr = err.Error()
	}
	logEntry := createLogEntry(LogLevelError, message, &errorStr, context)
	printLog(logEntry)
}

// LogDebug 记录调试日志
// message: 日志消息
// context: 上下文信息
func LogDebug(message string, context interface{}) {
	logEntry := createLogEntry(LogLevelDebug, message, nil, context)
	printLog(logEntry)
}

// ============================================================================
// 私有函数
// ============================================================================

// createLogEntry 创建日志条目
// level: 日志级别
// message: 日志消息
// errorStr: 错误字符串
// context: 上下文信息
// 返回: 日志条目
func createLogEntry(level LogLevel, message string, errorStr *string, context interface{}) LogEntry {
	_, file, line, _ := runtime.Caller(2)

	entry := LogEntry{
		Level:     level,
		Message:   message,
		Context:   context,
		Timestamp: time.Now(),
		File:      file,
		Line:      line,
	}

	if errorStr != nil {
		entry.Error = *errorStr
	}

	return entry
}

// printLog 打印日志
// entry: 日志条目
func printLog(entry LogEntry) {
	levelStr := getLevelString(entry.Level)
	timestamp := entry.Timestamp.Format("2006-01-02 15:04:05")

	logMessage := fmt.Sprintf("[%s] %s %s", levelStr, timestamp, entry.Message)

	if entry.Error != "" {
		logMessage += fmt.Sprintf(" | Error: %s", entry.Error)
	}

	if entry.Context != nil {
		contextJSON, _ := json.Marshal(entry.Context)
		logMessage += fmt.Sprintf(" | Context: %s", string(contextJSON))
	}

	logMessage += fmt.Sprintf(" | %s:%d", entry.File, entry.Line)

	log.Println(logMessage)
}

// getLevelString 获取日志级别字符串
// level: 日志级别
// 返回: 级别字符串
func getLevelString(level LogLevel) string {
	switch level {
	case LogLevelInfo:
		return "INFO"
	case LogLevelWarn:
		return "WARN"
	case LogLevelError:
		return "ERROR"
	case LogLevelDebug:
		return "DEBUG"
	default:
		return "UNKNOWN"
	}
}
