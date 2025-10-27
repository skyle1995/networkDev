package admin

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"networkDev/controllers"
	"networkDev/database"
	"networkDev/models"
	"networkDev/utils"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/spf13/viper"
)

// ============================================================================
// 全局变量
// ============================================================================

// 创建BaseController实例
var authBaseController = controllers.NewBaseController()

// ============================================================================
// 页面处理器
// ============================================================================

// LoginPageHandler 管理员登录页渲染处理器
// - 如果已登录则重定向到 /admin
// - 否则渲染 web/template/admin/login.html 模板
// - 自动清理失效的JWT Cookie，避免刷新时的问题
func LoginPageHandler(c *gin.Context) {
	// 使用带清理功能的JWT校验，避免失效Cookie在登录页面造成问题
	if IsAdminAuthenticatedWithCleanup(c) {
		c.Redirect(http.StatusFound, "/admin")
		return
	}

	// 获取或生成CSRF令牌
	var token string
	if existingToken := utils.GetCSRFTokenFromCookie(c); existingToken != "" {
		// 重用现有的Cookie令牌
		token = existingToken
	} else {
		// 生成新的CSRF令牌并设置到Cookie
		newToken, err := utils.GenerateCSRFToken()
		if err != nil {
			c.HTML(http.StatusInternalServerError, "error.html", gin.H{
				"Error": "生成CSRF令牌失败",
			})
			return
		}
		token = newToken
		utils.SetCSRFToken(c, token)
	}

	// 准备模板数据
	extraData := gin.H{
		"Title": "管理员登录",
	}
	data := authBaseController.GetDefaultTemplateData()
	data["CSRFToken"] = token

	// 合并额外数据
	for key, value := range extraData {
		data[key] = value
	}

	c.HTML(http.StatusOK, "login.html", data)
}

// ============================================================================
// API处理器
// ============================================================================

// LoginHandler 管理员登录接口
// - 接收JSON: {username, password}
// - 验证用户存在与密码正确性
// - 仅允许 Role=0 的管理员登录
// - 成功后设置简单的会话Cookie（后续可切换为JWT或更完善的Session）
func LoginHandler(c *gin.Context) {
	var body struct {
		Username string `json:"username"`
		Password string `json:"password"`
		Captcha  string `json:"captcha"`
	}

	if !authBaseController.BindJSON(c, &body) {
		return
	}

	if !authBaseController.ValidateRequired(c, map[string]interface{}{
		"用户名": body.Username,
		"密码":  body.Password,
		"验证码": body.Captcha,
	}) {
		return
	}

	// 验证验证码
	if !VerifyCaptcha(c, body.Captcha) {
		authBaseController.HandleValidationError(c, "验证码错误")
		return
	}

	// 获取数据库连接
	db, ok := authBaseController.GetDB(c)
	if !ok {
		return
	}

	// 通过前缀匹配一次性获取所有管理员相关设置
	var adminSettings []models.Settings
	if err := db.Where("name LIKE ?", "admin_%").Find(&adminSettings).Error; err != nil {
		authBaseController.HandleValidationError(c, "用户不存在或密码错误")
		return
	}

	// 将设置转换为map便于查找
	settingsMap := make(map[string]string)
	for _, setting := range adminSettings {
		settingsMap[setting.Name] = setting.Value
	}

	// 检查必要的设置是否存在
	adminUsername, hasUsername := settingsMap["admin_username"]
	adminPassword, hasPassword := settingsMap["admin_password"]
	adminPasswordSalt, hasSalt := settingsMap["admin_password_salt"]

	if !hasUsername || !hasPassword || !hasSalt {
		authBaseController.HandleValidationError(c, "用户不存在或密码错误")
		return
	}

	// 验证用户名
	if body.Username != adminUsername {
		authBaseController.HandleValidationError(c, "用户不存在或密码错误")
		return
	}

	// 验证密码为空的情况（首次登录需要初始化）
	if adminPassword == "" || adminPasswordSalt == "" {
		authBaseController.HandleInternalError(c, "管理员账号未初始化，请联系系统管理员", nil)
		return
	}

	// 使用盐值验证密码
	if !utils.VerifyPasswordWithSalt(body.Password, adminPasswordSalt, adminPassword) {
		authBaseController.HandleValidationError(c, "用户不存在或密码错误")
		return
	}

	// 创建虚拟用户对象用于生成JWT令牌
	adminUser := models.User{
		Username:     adminUsername,
		Password:     adminPassword,
		PasswordSalt: adminPasswordSalt,
	}

	// 生成JWT令牌
	token, err := generateJWTTokenForAdmin(adminUser)
	if err != nil {
		authBaseController.HandleInternalError(c, "生成令牌失败", err)
		return
	}

	// 设置JWT Cookie（使用安全配置）
	cookie := utils.CreateSecureCookie("admin_session", token, utils.GetDefaultCookieMaxAge())
	c.SetCookie(cookie.Name, cookie.Value, cookie.MaxAge, cookie.Path, cookie.Domain, cookie.Secure, cookie.HttpOnly)

	authBaseController.HandleSuccess(c, "登录成功", gin.H{
		"redirect": "/admin",
	})
}

// LogoutHandler 管理员登出
// - 清理JWT Cookie
// - 确保令牌完全失效
func LogoutHandler(c *gin.Context) {
	// 清理JWT Cookie
	clearInvalidJWTCookie(c)

	// 可选：将JWT令牌加入黑名单（需要Redis或数据库支持）
	// 这里可以实现JWT黑名单机制

	authBaseController.HandleSuccess(c, "已退出登录", gin.H{
		"redirect": "/admin/login",
	})
}

// ============================================================================
// 辅助函数
// ============================================================================

// clearInvalidJWTCookie 清理无效的JWT Cookie
// - 统一的Cookie清理函数，确保一致性
// - 在JWT校验失败时自动调用，提升安全性和用户体验
func clearInvalidJWTCookie(c *gin.Context) {
	cookie := utils.CreateExpiredCookie("admin_session")
	c.SetCookie(cookie.Name, cookie.Value, cookie.MaxAge, cookie.Path, cookie.Domain, cookie.Secure, cookie.HttpOnly)
}

// getJWTSecret 动态获取当前的JWT密钥
// 修复安全漏洞：确保每次都从最新配置中获取密钥，而不是使用启动时的全局变量
func getJWTSecret() []byte {
	return []byte(viper.GetString("security.jwt_secret"))
}

// ============================================================================
// 结构体定义
// ============================================================================

// JWTClaims JWT载荷结构体
type JWTClaims struct {
	Username     string `json:"username"`
	PasswordHash string `json:"password_hash"` // 密码哈希摘要，用于验证密码是否被修改
	jwt.RegisteredClaims
}

// generateJWTTokenForAdmin 生成管理员JWT令牌
// - 包含管理员UUID、用户名信息
// - 设置24小时过期时间
// - 使用HMAC-SHA256签名
func generateJWTTokenForAdmin(adminUser models.User) (string, error) {
	// 生成密码哈希摘要（使用SHA256）
	passwordHashDigest := utils.GenerateSHA256Hash(adminUser.Password)

	claims := JWTClaims{
		Username:     adminUser.Username,
		PasswordHash: passwordHashDigest, // 包含密码哈希摘要
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "凌动技术",
			Subject:   adminUser.Username,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(getJWTSecret())
}

// parseJWTToken 解析并验证JWT令牌
// - 验证签名有效性
// - 检查过期时间
// - 返回用户信息
func parseJWTToken(tokenString string) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return getJWTSecret(), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*JWTClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, fmt.Errorf("invalid token")
}

// getJWTCookie 获取JWT cookie的通用函数
func getJWTCookie(c *gin.Context) (string, error) {
	return c.Cookie("admin_session")
}

// validateAdminPasswordHash 验证管理员密码哈希的通用函数
func validateAdminPasswordHash(claims *JWTClaims, c *gin.Context) bool {
	// 【安全修复】验证数据库中的当前密码哈希
	// 这确保了密码修改后，旧的JWT令牌会失效
	db, err := database.GetDB()
	if err != nil {
		fmt.Printf("[SECURITY WARNING] Database connection failed during auth - Username=%s, IP=%s\n",
			claims.Username, c.ClientIP())
		return false
	}

	// 获取当前数据库中的管理员密码
	var adminPassword models.Settings
	if err := db.Where("name = ?", "admin_password").First(&adminPassword).Error; err != nil {
		fmt.Printf("[SECURITY WARNING] Admin password not found in database - Username=%s, IP=%s\n",
			claims.Username, c.ClientIP())
		return false
	}

	// 生成当前数据库密码的哈希摘要
	currentPasswordHash := utils.GenerateSHA256Hash(adminPassword.Value)

	// 验证JWT中的密码哈希是否与当前数据库中的密码哈希一致
	if claims.PasswordHash != currentPasswordHash {
		fmt.Printf("[SECURITY WARNING] Password hash mismatch - JWT token invalidated - Username=%s, IP=%s\n",
			claims.Username, c.ClientIP())
		return false
	}

	return true
}

// IsAdminAuthenticated 判断管理员是否已认证（导出）
// - 检查admin_session Cookie中的JWT令牌
// - 验证令牌签名、过期时间和用户角色
func IsAdminAuthenticated(c *gin.Context) bool {
	cookie, err := getJWTCookie(c)
	if err != nil || cookie == "" {
		return false
	}

	// 解析并验证JWT令牌
	claims, err := parseJWTToken(cookie)
	if err != nil {
		return false
	}

	// 注释：由于这是管理员专用认证函数，不需要额外的角色验证

	// 验证密码哈希
	return validateAdminPasswordHash(claims, c)
}

// IsAdminAuthenticatedWithCleanup 带自动清理功能的JWT校验函数
// - 当JWT校验失败时，自动清理失效的Cookie
// - 适用于API接口等需要清理失效令牌的场景
func IsAdminAuthenticatedWithCleanup(c *gin.Context) bool {
	cookie, err := getJWTCookie(c)
	if err != nil || cookie == "" {
		return false
	}

	// 解析并验证JWT令牌
	claims, err := parseJWTToken(cookie)
	if err != nil {
		// JWT解析失败，清理失效Cookie
		clearInvalidJWTCookie(c)
		return false
	}

	// 注释：由于这是管理员专用认证函数，不需要额外的角色验证

	// 验证密码哈希
	if !validateAdminPasswordHash(claims, c) {
		clearInvalidJWTCookie(c)
		return false
	}

	return true
}

// GetCurrentAdminUser 获取当前登录的管理员用户信息
// - 从JWT令牌中提取用户信息
// - 自动刷新接近过期的令牌（剩余时间少于6小时时刷新）
// - 返回用户ID、用户名和角色
func GetCurrentAdminUser(c *gin.Context) (*JWTClaims, error) {
	cookie, err := getJWTCookie(c)
	if err != nil {
		return nil, fmt.Errorf("未找到会话信息")
	}

	claims, err := parseJWTToken(cookie)
	if err != nil {
		return nil, fmt.Errorf("无效的会话信息")
	}

	// 注释：由于这是管理员专用函数，不需要额外的角色验证

	return claims, nil
}

// GetCurrentAdminUserWithRefresh 获取当前登录的管理员用户信息并自动刷新令牌
// - 从JWT令牌中提取用户信息
// - 自动刷新接近过期的令牌（剩余时间少于6小时时刷新）
// - 返回用户ID、用户名、角色和是否刷新了令牌
func GetCurrentAdminUserWithRefresh(c *gin.Context) (*JWTClaims, bool, error) {
	cookie, err := getJWTCookie(c)
	if err != nil {
		return nil, false, fmt.Errorf("未找到会话信息")
	}

	claims, err := parseJWTToken(cookie)
	if err != nil {
		return nil, false, fmt.Errorf("无效的会话信息")
	}

	// 验证密码哈希
	if !validateAdminPasswordHash(claims, c) {
		return nil, false, fmt.Errorf("会话已失效，请重新登录")
	}

	// 检查是否需要刷新令牌
	refreshed := false
	refreshThreshold := time.Duration(viper.GetInt("security.jwt_refresh")) * time.Hour
	if time.Until(claims.ExpiresAt.Time) < refreshThreshold {
		adminUser := models.User{
			Username: claims.Username,
		}
		newToken, err := generateJWTTokenForAdmin(adminUser)
		if err == nil {
			c.SetCookie("admin_session", newToken, utils.GetDefaultCookieMaxAge(), "/", "", false, true)
			refreshed = true

			claims.ExpiresAt = jwt.NewNumericDate(time.Now().Add(24 * time.Hour))
			claims.IssuedAt = jwt.NewNumericDate(time.Now())
		}
	}

	return claims, refreshed, nil
}

// AdminAuthRequired 管理员认证拦截中间件
// - 未登录：重定向到 /admin/login
// - 已登录：自动刷新接近过期的令牌，然后放行到后续处理器
func AdminAuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 尝试获取用户信息并自动刷新令牌
		claims, refreshed, err := GetCurrentAdminUserWithRefresh(c)
		if err != nil {
			// 自动清理失效的JWT Cookie，提升安全性和用户体验
			clearInvalidJWTCookie(c)

			// 中文注释：区分普通页面请求与AJAX/JSON请求
			// - 对 AJAX/JSON：直接返回 401 JSON，便于前端处理（如提示重新登录）
			// - 对普通页面：保持原有重定向到登录页
			accept := c.GetHeader("Accept")
			xrw := strings.ToLower(strings.TrimSpace(c.GetHeader("X-Requested-With")))
			if strings.Contains(accept, "application/json") || xrw == "xmlhttprequest" {
				c.JSON(http.StatusUnauthorized, gin.H{
					"success": false,
					"message": "未登录或会话已过期",
					"data":    nil,
				})
				c.Abort()
				return
			}
			c.Redirect(http.StatusFound, "/admin/login")
			c.Abort()
			return
		}

		// 如果令牌被刷新，可以在这里记录日志（可选）
		if refreshed {
			// 可以添加日志记录令牌刷新事件
			_ = claims // 避免未使用变量警告
		}

		c.Next()
	}
}
