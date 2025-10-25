package utils

import (
	"net/http"
	"time"

	"github.com/spf13/viper"
)

// CreateSecureCookie 创建安全的Cookie
// name: Cookie名称
// value: Cookie值
// maxAge: 过期时间（秒），0表示会话Cookie，-1表示立即过期
func CreateSecureCookie(name, value string, maxAge int) *http.Cookie {
	cookie := &http.Cookie{
		Name:     name,
		Value:    value,
		Path:     "/",
		HttpOnly: true,
		MaxAge:   maxAge,
	}

	// 从配置读取安全设置
	if viper.GetBool("security.cookie.secure") {
		cookie.Secure = true
	}

	// 设置SameSite属性
	sameSite := viper.GetString("security.cookie.same_site")
	switch sameSite {
	case "Strict":
		cookie.SameSite = http.SameSiteStrictMode
	case "Lax":
		cookie.SameSite = http.SameSiteLaxMode
	case "None":
		cookie.SameSite = http.SameSiteNoneMode
		// SameSite=None 必须配合 Secure=true 使用
		cookie.Secure = true
	default:
		cookie.SameSite = http.SameSiteStrictMode
	}

	// 设置Domain（如果配置了）
	domain := viper.GetString("security.cookie.domain")
	if domain != "" {
		cookie.Domain = domain
	}

	// 如果maxAge > 0，设置Expires时间
	if maxAge > 0 {
		cookie.Expires = time.Now().Add(time.Duration(maxAge) * time.Second)
	} else if maxAge == -1 {
		// 立即过期
		cookie.Expires = time.Unix(0, 0)
	}

	return cookie
}

// CreateSessionCookie 创建会话Cookie（浏览器关闭时过期）
func CreateSessionCookie(name, value string) *http.Cookie {
	return CreateSecureCookie(name, value, 0)
}

// CreateExpiredCookie 创建立即过期的Cookie（用于清理）
func CreateExpiredCookie(name string) *http.Cookie {
	return CreateSecureCookie(name, "", -1)
}

// GetDefaultCookieMaxAge 获取默认Cookie过期时间
func GetDefaultCookieMaxAge() int {
	maxAge := viper.GetInt("security.cookie.max_age")
	if maxAge <= 0 {
		return 86400 // 默认24小时
	}
	return maxAge
}