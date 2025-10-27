package middleware

import (
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"networkDev/utils/logger"
)

// ============================================================================
// 结构体定义
// ============================================================================

// LoggingMiddleware 日志记录中间件结构体
// 用于记录HTTP请求的详细信息，包括方法、路径、状态码和响应时间
type LoggingMiddleware struct {
	logger *logger.Logger
}

// ============================================================================
// 构造函数
// ============================================================================

// NewLoggingMiddleware 创建新的日志记录中间件实例
func NewLoggingMiddleware(logger *logger.Logger) *LoggingMiddleware {
	return &LoggingMiddleware{
		logger: logger,
	}
}

// ============================================================================
// 中间件函数
// ============================================================================

// Handler 返回Gin中间件函数，用于记录HTTP请求日志
// 记录格式遵循Apache Common Log Format
func (lm *LoggingMiddleware) Handler() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 记录开始时间
		start := time.Now()

		// 处理请求
		c.Next()

		// 计算处理时间
		duration := time.Since(start)

		// 获取客户端IP
		clientIP := getClientIP(c)

		// 记录日志 - Apache Common Log Format
		// 使用专门的HTTP日志方法避免User-Agent中的反斜杠被转义
		lm.logger.LogRequestWithHeaders(
			c.Request.Method,
			c.Request.RequestURI,
			clientIP,
			c.Writer.Status(),
			duration,
			"-", // referer (已废弃)
			c.Request.UserAgent(),
		)
	}
}

// ============================================================================
// 私有函数
// ============================================================================

// getClientIP 获取客户端真实IP地址
// 优先从X-Forwarded-For、X-Real-IP等头部获取，最后使用RemoteAddr
func getClientIP(c *gin.Context) string {
	// 检查X-Forwarded-For头部
	xForwardedFor := c.GetHeader("X-Forwarded-For")
	if xForwardedFor != "" {
		// X-Forwarded-For可能包含多个IP，取第一个
		ips := strings.Split(xForwardedFor, ",")
		if len(ips) > 0 {
			return strings.TrimSpace(ips[0])
		}
	}

	// 检查X-Real-IP头部
	xRealIP := c.GetHeader("X-Real-IP")
	if xRealIP != "" {
		return xRealIP
	}

	// 检查X-Forwarded头部
	xForwarded := c.GetHeader("X-Forwarded")
	if xForwarded != "" {
		return xForwarded
	}

	// 使用RemoteAddr
	remoteAddr := c.Request.RemoteAddr
	if strings.Contains(remoteAddr, ":") {
		// 移除端口号
		if idx := strings.LastIndex(remoteAddr, ":"); idx != -1 {
			return remoteAddr[:idx]
		}
	}

	return remoteAddr
}

// ============================================================================
// 公共函数
// ============================================================================

// WrapHandler 创建Gin日志中间件
// 使用全局日志记录器创建日志中间件
func WrapHandler() gin.HandlerFunc {
	logger := logger.GetLogger()
	middleware := NewLoggingMiddleware(logger)
	return middleware.Handler()
}
