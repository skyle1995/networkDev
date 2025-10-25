package utils

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"net/http"
)

const (
	CSRFTokenLength = 32
	CSRFCookieName  = "csrf_token"
	CSRFHeaderName  = "X-CSRF-Token"
	CSRFFormField   = "csrf_token"
)

// generateRandomBytes 生成指定长度的随机字节
func generateRandomBytes(length int) ([]byte, error) {
	bytes := make([]byte, length)
	_, err := rand.Read(bytes)
	if err != nil {
		return nil, err
	}
	return bytes, nil
}

// GenerateCSRFToken 生成CSRF令牌
func GenerateCSRFToken() (string, error) {
	bytes, err := generateRandomBytes(CSRFTokenLength)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}

// SetCSRFToken 设置CSRF令牌到Cookie和响应头
func SetCSRFToken(w http.ResponseWriter, token string) {
	// 设置CSRF令牌到Cookie
	cookie := CreateSecureCookie(CSRFCookieName, token, 3600) // 1小时过期
	http.SetCookie(w, cookie)

	// 设置CSRF令牌到响应头，方便JavaScript获取
	w.Header().Set("X-CSRF-Token", token)
}

// GetCSRFTokenFromRequest 从请求中获取CSRF令牌
// 优先级：Header > Form > Cookie
func GetCSRFTokenFromRequest(r *http.Request) string {
	// 1. 从Header获取
	if token := r.Header.Get(CSRFHeaderName); token != "" {
		return token
	}

	// 2. 从Form获取
	if token := r.FormValue(CSRFFormField); token != "" {
		return token
	}

	// 3. 从Cookie获取（作为备选）
	if cookie, err := r.Cookie(CSRFCookieName); err == nil {
		return cookie.Value
	}

	return ""
}

// GetCSRFTokenFromCookie 从Cookie中获取CSRF令牌
func GetCSRFTokenFromCookie(r *http.Request) string {
	cookie, err := r.Cookie(CSRFCookieName)
	if err != nil {
		return ""
	}
	return cookie.Value
}

// ValidateCSRFToken 验证CSRF令牌
func ValidateCSRFToken(r *http.Request) bool {
	// 获取Cookie中的令牌（服务器端存储的）
	cookieToken := GetCSRFTokenFromCookie(r)
	if cookieToken == "" {
		return false
	}

	// 获取请求中的令牌（客户端提交的）
	requestToken := GetCSRFTokenFromRequest(r)
	if requestToken == "" {
		return false
	}

	// 使用常量时间比较防止时序攻击
	return subtle.ConstantTimeCompare([]byte(cookieToken), []byte(requestToken)) == 1
}

// CSRFProtection CSRF保护中间件
func CSRFProtection(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 对于GET、HEAD、OPTIONS请求，只生成令牌，不验证
		if r.Method == http.MethodGet || r.Method == http.MethodHead || r.Method == http.MethodOptions {
			// 生成新的CSRF令牌
			token, err := GenerateCSRFToken()
			if err != nil {
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}
			SetCSRFToken(w, token)
			next(w, r)
			return
		}

		// 对于POST、PUT、DELETE等修改性请求，验证CSRF令牌
		if !ValidateCSRFToken(r) {
			JsonResponse(w, http.StatusForbidden, false, "CSRF令牌验证失败", nil)
			return
		}

		// 验证通过，继续处理请求
		next(w, r)
	}
}

// RequireCSRFToken 要求CSRF令牌的中间件（用于特定路由）
func RequireCSRFToken(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !ValidateCSRFToken(r) {
			JsonResponse(w, http.StatusForbidden, false, "CSRF令牌验证失败", nil)
			return
		}
		next(w, r)
	}
}

// GetCSRFTokenForTemplate 获取用于模板的CSRF令牌
func GetCSRFTokenForTemplate(r *http.Request) string {
	// 尝试从Cookie获取现有令牌
	if token := GetCSRFTokenFromCookie(r); token != "" {
		return token
	}

	// 如果没有现有令牌，生成新的（但不设置到响应中）
	token, err := GenerateCSRFToken()
	if err != nil {
		return ""
	}
	return token
}

// CSRFTokenHandler 专门用于获取CSRF令牌的API端点
func CSRFTokenHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		JsonResponse(w, http.StatusMethodNotAllowed, false, "只支持GET请求", nil)
		return
	}

	// 生成新的CSRF令牌
	token, err := GenerateCSRFToken()
	if err != nil {
		JsonResponse(w, http.StatusInternalServerError, false, "生成CSRF令牌失败", nil)
		return
	}

	// 设置令牌到Cookie和响应头
	SetCSRFToken(w, token)

	// 返回令牌给前端
	JsonResponse(w, http.StatusOK, true, "CSRF令牌获取成功", map[string]interface{}{
		"csrf_token": token,
	})
}