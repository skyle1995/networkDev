package middleware

import (
	"net/http"
	"strings"
	"time"

	"networkDev/utils/logger"
)

// LoggingMiddleware HTTP请求日志中间件
// 记录每个HTTP请求的详细信息，包括方法、路径、状态码和响应时间
type LoggingMiddleware struct {
	logger *logger.Logger
}

// NewLoggingMiddleware 创建新的日志中间件实例
func NewLoggingMiddleware(logger *logger.Logger) *LoggingMiddleware {
	return &LoggingMiddleware{
		logger: logger,
	}
}

// responseWriter 包装http.ResponseWriter以捕获状态码
type responseWriter struct {
	http.ResponseWriter
	statusCode int
	written    bool
}

// newResponseWriter 创建新的响应写入器包装器
func newResponseWriter(w http.ResponseWriter) *responseWriter {
	return &responseWriter{
		ResponseWriter: w,
		statusCode:     http.StatusOK, // 默认状态码
	}
}

// WriteHeader 重写WriteHeader方法以捕获状态码
func (rw *responseWriter) WriteHeader(code int) {
	if !rw.written {
		rw.statusCode = code
		rw.written = true
		rw.ResponseWriter.WriteHeader(code)
	}
}

// Write 重写Write方法以确保状态码被设置
func (rw *responseWriter) Write(b []byte) (int, error) {
	if !rw.written {
		rw.WriteHeader(http.StatusOK)
	}
	return rw.ResponseWriter.Write(b)
}

// Handler 中间件处理器函数
// 包装HTTP处理器以添加请求日志记录功能
func (lm *LoggingMiddleware) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// 包装响应写入器以捕获状态码
		wrapped := newResponseWriter(w)

		// 调用下一个处理器
		next.ServeHTTP(wrapped, r)

		// 计算响应时间
		duration := time.Since(start)

		// 记录请求日志
		lm.logger.LogRequestWithHeaders(
			r.Method,
			r.URL.Path,
			getClientIP(r),
			wrapped.statusCode,
			duration,
			"-",
			r.Header.Get("User-Agent"),
		)
	})
}

// getClientIP 获取客户端真实IP地址
// 优先从X-Forwarded-For、X-Real-IP等头部获取，最后使用RemoteAddr
func getClientIP(r *http.Request) string {
	// 检查X-Forwarded-For头部
	xForwardedFor := r.Header.Get("X-Forwarded-For")
	if xForwardedFor != "" {
		// X-Forwarded-For可能包含多个IP，取第一个
		ips := strings.Split(xForwardedFor, ",")
		if len(ips) > 0 {
			return strings.TrimSpace(ips[0])
		}
	}

	// 检查X-Real-IP头部
	xRealIP := r.Header.Get("X-Real-IP")
	if xRealIP != "" {
		return xRealIP
	}

	// 检查X-Forwarded头部
	xForwarded := r.Header.Get("X-Forwarded")
	if xForwarded != "" {
		return xForwarded
	}

	// 使用RemoteAddr
	remoteAddr := r.RemoteAddr
	if strings.Contains(remoteAddr, ":") {
		// 移除端口号
		if idx := strings.LastIndex(remoteAddr, ":"); idx != -1 {
			return remoteAddr[:idx]
		}
	}

	return remoteAddr
}

// WrapHandler 包装HTTP处理器以添加日志记录功能
// 使用全局日志记录器创建日志中间件
func WrapHandler(handler http.Handler) http.Handler {
	logger := logger.GetLogger()
	middleware := NewLoggingMiddleware(logger)
	return middleware.Handler(handler)
}

// WrapHandlerFunc 包装HTTP处理器函数以添加日志记录功能
// 将HandlerFunc转换为Handler并添加日志中间件
func WrapHandlerFunc(handlerFunc http.HandlerFunc) http.Handler {
	return WrapHandler(handlerFunc)
}
