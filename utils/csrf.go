package utils

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"net/http"

	"github.com/gin-gonic/gin"
)

// ============================================================================
// 常量定义
// ============================================================================

const (
	CSRFTokenLength = 32
	CSRFCookieName  = "csrf_token"
	CSRFHeaderName  = "X-CSRF-Token"
	CSRFFormField   = "csrf_token"
)

// ============================================================================
// 私有函数
// ============================================================================

// generateRandomBytes 生成指定长度的随机字节
func generateRandomBytes(length int) ([]byte, error) {
	bytes := make([]byte, length)
	_, err := rand.Read(bytes)
	if err != nil {
		return nil, err
	}
	return bytes, nil
}

// ============================================================================
// 公共函数
// ============================================================================

// GenerateCSRFToken 生成CSRF令牌
func GenerateCSRFToken() (string, error) {
	bytes, err := generateRandomBytes(CSRFTokenLength)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}

// SetCSRFToken 设置CSRF令牌到Cookie和响应头
func SetCSRFToken(c *gin.Context, token string) {
	c.SetCookie(CSRFCookieName, token, 3600*24, "/", "", false, true)
	c.Header(CSRFHeaderName, token)
}

// GetCSRFTokenFromRequest 从Gin请求中获取CSRF令牌
// 优先级：Header > Form > Cookie
func GetCSRFTokenFromRequest(c *gin.Context) string {
	// 1. 从Header获取
	if token := c.GetHeader(CSRFHeaderName); token != "" {
		return token
	}

	// 2. 从Form获取
	if token := c.PostForm(CSRFFormField); token != "" {
		return token
	}

	// 3. 从Cookie获取（作为备选）
	if cookie, err := c.Cookie(CSRFCookieName); err == nil {
		return cookie
	}

	return ""
}

// GetCSRFTokenFromCookie 从Cookie中获取CSRF令牌
func GetCSRFTokenFromCookie(c *gin.Context) string {
	cookie, err := c.Cookie(CSRFCookieName)
	if err != nil {
		return ""
	}
	return cookie
}

// ValidateCSRFToken 验证CSRF令牌
func ValidateCSRFToken(c *gin.Context) bool {
	// 获取Cookie中的令牌（服务器端存储的）
	cookieToken := GetCSRFTokenFromCookie(c)
	if cookieToken == "" {
		return false
	}

	// 获取请求中的令牌（客户端提交的）
	requestToken := GetCSRFTokenFromRequest(c)
	if requestToken == "" {
		return false
	}

	// 使用常量时间比较防止时序攻击
	return subtle.ConstantTimeCompare([]byte(cookieToken), []byte(requestToken)) == 1
}

// CSRFProtection CSRF保护中间件
func CSRFProtection() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 对于GET、HEAD、OPTIONS请求，只生成令牌，不验证
		if c.Request.Method == http.MethodGet || c.Request.Method == http.MethodHead || c.Request.Method == http.MethodOptions {
			// 生成新的CSRF令牌
			token, err := GenerateCSRFToken()
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"code": 1,
					"msg":  "Internal Server Error",
					"data": nil,
				})
				c.Abort()
				return
			}
			SetCSRFToken(c, token)
			c.Next()
			return
		}

		// 对于POST、PUT、DELETE等修改性请求，验证CSRF令牌
		if !ValidateCSRFToken(c) {
			c.JSON(http.StatusForbidden, gin.H{
				"code": 1,
				"msg":  "CSRF令牌验证失败",
				"data": nil,
			})
			c.Abort()
			return
		}

		// 验证通过，继续处理请求
		c.Next()
	}
}

// RequireCSRFToken 要求CSRF令牌的中间件（用于特定路由）
func RequireCSRFToken() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !ValidateCSRFToken(c) {
			c.JSON(http.StatusForbidden, gin.H{
				"code": 1,
				"msg":  "CSRF令牌验证失败",
				"data": nil,
			})
			c.Abort()
			return
		}
		c.Next()
	}
}

// GetCSRFTokenForTemplate 获取用于模板的CSRF令牌
func GetCSRFTokenForTemplate(c *gin.Context) string {
	// 尝试从Cookie获取现有令牌
	if token := GetCSRFTokenFromCookie(c); token != "" {
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
func CSRFTokenHandler(c *gin.Context) {
	if c.Request.Method != http.MethodGet {
		c.JSON(http.StatusMethodNotAllowed, gin.H{
			"code": 1,
			"msg":  "只支持GET请求",
			"data": nil,
		})
		return
	}

	// 生成新的CSRF令牌
	token, err := GenerateCSRFToken()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 1,
			"msg":  "生成CSRF令牌失败",
			"data": nil,
		})
		return
	}

	// 设置令牌到Cookie和响应头
	SetCSRFToken(c, token)

	// 返回令牌给前端
	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"msg":  "CSRF令牌生成成功",
		"data": gin.H{
			"csrf_token": token,
		},
	})
}
