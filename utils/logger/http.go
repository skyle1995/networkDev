package logger

import (
	"fmt"
	"os"
	"time"
)

// LogRequest 记录HTTP请求日志 - 使用标准Apache Common Log Format
// 格式: IP - - [timestamp] "METHOD path HTTP/1.1" status_code response_size
// method: HTTP请求方法
// path: 请求路径
// clientIP: 客户端IP地址
// statusCode: HTTP状态码
// duration: 请求处理时长
func (l *Logger) LogRequest(method, path, clientIP string, statusCode int, duration time.Duration) {
	l.LogRequestWithHeaders(method, path, clientIP, statusCode, duration, "-", "-")
}

// LogRequestWithHeaders 记录HTTP请求日志 - 使用修改的Apache Log Format（移除Referer字段）
// 直接输出标准格式，不通过logrus格式化器
// method: HTTP请求方法
// path: 请求路径
// clientIP: 客户端IP地址
// statusCode: HTTP状态码
// duration: 请求处理时长
// referer: 引用页面（已废弃，保留参数兼容性）
// userAgent: 用户代理字符串
func (l *Logger) LogRequestWithHeaders(method, path, clientIP string, statusCode int, duration time.Duration, referer, userAgent string) {
	// 格式化时间戳为Apache标准格式
	timestamp := time.Now().Format("02/Jan/2006:15:04:05 -0700")

	// 处理空值
	if userAgent == "" {
		userAgent = "-"
	}

	// 构建修改的HTTP Log格式（完全移除Referer字段）
	logLine := fmt.Sprintf(`%s - - [%s] "%s %s HTTP/1.1" %d - "%s" %dms`,
		clientIP,
		timestamp,
		method,
		path,
		statusCode,
		userAgent,
		duration.Milliseconds(),
	)

	// 直接输出到标准输出和日志文件，不使用logrus格式化
	l.writeHTTPLog(logLine)
}

// writeHTTPLog 直接输出HTTP日志到标准输出
// 避免Logrus的任何格式化和转义，保持Apache日志格式的原始性
// logLine: 格式化后的日志行
func (l *Logger) writeHTTPLog(logLine string) {
	// 直接输出到标准输出，避免Logrus的转义处理
	fmt.Fprintln(os.Stdout, logLine)
}
