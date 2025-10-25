package admin

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"networkDev/database"
	"networkDev/models"
	"networkDev/utils"

	"github.com/golang-jwt/jwt/v5"
	"github.com/spf13/viper"
)

// LoginPageHandler 管理员登录页渲染处理器
// - 如果已登录则重定向到 /admin
// - 否则渲染 web/template/admin/login.html 模板
// - 自动清理失效的JWT Cookie，避免刷新时的问题
func LoginPageHandler(w http.ResponseWriter, r *http.Request) {
	// 使用带清理功能的JWT校验，避免失效Cookie在登录页面造成问题
	if IsAdminAuthenticatedWithCleanup(w, r) {
		http.Redirect(w, r, "/admin", http.StatusFound)
		return
	}

	// 获取或生成CSRF令牌
	var token string
	if existingToken := utils.GetCSRFTokenFromCookie(r); existingToken != "" {
		// 重用现有的Cookie令牌
		token = existingToken
	} else {
		// 生成新的CSRF令牌并设置到Cookie
		newToken, err := utils.GenerateCSRFToken()
		if err != nil {
			http.Error(w, "生成CSRF令牌失败", http.StatusInternalServerError)
			return
		}
		token = newToken
		utils.SetCSRFToken(w, token)
	}

	// 准备模板数据
	extraData := map[string]interface{}{
		"Title": "管理员登录",
	}
	data := utils.GetDefaultTemplateData()
	data["CSRFToken"] = token
	
	// 合并额外数据
	for key, value := range extraData {
		data[key] = value
	}

	utils.RenderTemplate(w, "login.html", data)
}

// LoginHandler 管理员登录接口
// - 接收JSON: {username, password}
// - 验证用户存在与密码正确性
// - 仅允许 Role=0 的管理员登录
// - 成功后设置简单的会话Cookie（后续可切换为JWT或更完善的Session）
func LoginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	var body struct {
		Username string `json:"username"`
		Password string `json:"password"`
		Captcha  string `json:"captcha"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		utils.JsonResponse(w, http.StatusBadRequest, false, "请求参数错误", nil)
		return
	}
	if body.Username == "" || body.Password == "" {
		utils.JsonResponse(w, http.StatusBadRequest, false, "用户名和密码不能为空", nil)
		return
	}
	if body.Captcha == "" {
		utils.JsonResponse(w, http.StatusBadRequest, false, "验证码不能为空", nil)
		return
	}

	// 验证验证码
	if !VerifyCaptcha(r, body.Captcha) {
		utils.JsonResponse(w, http.StatusBadRequest, false, "验证码错误", nil)
		return
	}

	db, err := database.GetDB()
	if err != nil {
		utils.JsonResponse(w, http.StatusInternalServerError, false, "数据库连接失败", nil)
		return
	}

	var user models.User
	dbErr := db.Where("username = ?", body.Username).First(&user).Error
	if dbErr != nil {
		utils.JsonResponse(w, http.StatusUnauthorized, false, "用户不存在或密码错误", nil)
		return
	}
	if user.Role != 0 {
		utils.JsonResponse(w, http.StatusForbidden, false, "非管理员账号不可登录后台", nil)
		return
	}

	// 使用盐值验证密码
	if !utils.VerifyPasswordWithSalt(body.Password, user.PasswordSalt, user.Password) {
		utils.JsonResponse(w, http.StatusUnauthorized, false, "用户不存在或密码错误", nil)
		return
	}

	// 生成JWT令牌
	token, err := generateJWTToken(user)
	if err != nil {
		utils.JsonResponse(w, http.StatusInternalServerError, false, "生成令牌失败", nil)
		return
	}

	// 设置JWT Cookie（使用安全配置）
	cookie := utils.CreateSecureCookie("admin_session", token, utils.GetDefaultCookieMaxAge())
	http.SetCookie(w, cookie)

	utils.JsonResponse(w, http.StatusOK, true, "登录成功", map[string]interface{}{
		"redirect": "/admin",
	})
}

// LogoutHandler 管理员登出
// - 清理JWT Cookie会话
// - 确保令牌完全失效
func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	// 清理JWT Cookie
	clearInvalidJWTCookie(w)

	// 可选：将JWT令牌加入黑名单（需要Redis或数据库支持）
	// 这里可以实现JWT黑名单机制

	utils.JsonResponse(w, http.StatusOK, true, "已退出登录", map[string]interface{}{
		"redirect": "/admin/login",
	})
}

// clearInvalidJWTCookie 清理失效的JWT Cookie
// - 统一的Cookie清理函数，确保一致性
// - 在JWT校验失败时自动调用，提升安全性和用户体验
func clearInvalidJWTCookie(w http.ResponseWriter) {
	cookie := utils.CreateExpiredCookie("admin_session")
	http.SetCookie(w, cookie)
}

// JWT密钥（生产环境应从配置文件或环境变量读取）
var jwtSecret = []byte(viper.GetString("security.jwt_secret"))

// JWTClaims JWT载荷结构
type JWTClaims struct {
	UserUUID     string `json:"user_uuid"`
	Username     string `json:"username"`
	Role         int    `json:"role"`
	PasswordHash string `json:"password_hash"` // 密码哈希摘要，用于验证密码是否被修改
	jwt.RegisteredClaims
}

// generateJWTToken 生成JWT令牌
// - 包含用户ID、用户名、角色信息
// - 设置24小时过期时间
// - 使用HMAC-SHA256签名
func generateJWTToken(user models.User) (string, error) {
	// 生成密码哈希摘要（使用SHA256）
	passwordHashDigest := utils.GenerateSHA256Hash(user.Password)
	
	claims := JWTClaims{
		UserUUID:     user.UUID,
		Username:     user.Username,
		Role:         user.Role,
		PasswordHash: passwordHashDigest, // 包含密码哈希摘要
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "凌动技术",
			Subject:   user.UUID,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
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
		return jwtSecret, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*JWTClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, fmt.Errorf("invalid token")
}

// IsAdminAuthenticated 判断管理员是否已认证（导出）
// - 检查admin_session Cookie中的JWT令牌
// - 验证令牌签名、过期时间和用户角色
func IsAdminAuthenticated(r *http.Request) bool {
	cookie, err := r.Cookie("admin_session")
	if err != nil || cookie.Value == "" {
		return false
	}

	// 解析并验证JWT令牌
	claims, err := parseJWTToken(cookie.Value)
	if err != nil {
		return false
	}

	// 验证用户角色（只允许管理员角色=0）
	if claims.Role != 0 {
		return false
	}

	// 验证用户是否仍然存在于数据库中
	db, err := database.GetDB()
	if err != nil {
		return false
	}

	var user models.User
	if dbErr := db.Where("uuid = ? AND role = 0", claims.UserUUID).First(&user).Error; dbErr != nil {
		// 记录安全事件：用户不存在但持有有效JWT令牌
		fmt.Printf("[SECURITY WARNING] Invalid JWT token detected - User not found: UUID=%s, Username=%s, IP=%s\n", 
			claims.UserUUID, claims.Username, r.RemoteAddr)
		return false
	}

	// 验证用户名是否匹配（防止用户名被修改后仍使用旧令牌）
	if user.Username != claims.Username {
		// 记录安全事件：用户名不匹配
		fmt.Printf("[SECURITY WARNING] Username mismatch detected - Token username=%s, DB username=%s, UUID=%s, IP=%s\n", 
			claims.Username, user.Username, claims.UserUUID, r.RemoteAddr)
		return false
	}

	// 验证密码哈希是否匹配（防止密码被修改后仍使用旧令牌）
	currentPasswordHash := utils.GenerateSHA256Hash(user.Password)
	if claims.PasswordHash != currentPasswordHash {
		// 记录安全事件：密码哈希不匹配，可能密码已被修改
		fmt.Printf("[SECURITY WARNING] Password hash mismatch detected - Token may be invalid due to password change: UUID=%s, Username=%s, IP=%s\n", 
			claims.UserUUID, claims.Username, r.RemoteAddr)
		return false
	}

	return true
}

// IsAdminAuthenticatedWithCleanup 带自动清理功能的JWT校验函数
// - 当JWT校验失败时，自动清理失效的Cookie
// - 适用于API接口等需要清理失效令牌的场景
func IsAdminAuthenticatedWithCleanup(w http.ResponseWriter, r *http.Request) bool {
	cookie, err := r.Cookie("admin_session")
	if err != nil || cookie.Value == "" {
		return false
	}

	// 解析并验证JWT令牌
	claims, err := parseJWTToken(cookie.Value)
	if err != nil {
		// JWT解析失败，清理失效Cookie
		clearInvalidJWTCookie(w)
		return false
	}

	// 验证用户角色（只允许管理员角色=0）
	if claims.Role != 0 {
		clearInvalidJWTCookie(w)
		return false
	}

	// 验证用户是否仍然存在于数据库中
	db, err := database.GetDB()
	if err != nil {
		return false
	}

	var user models.User
	if dbErr := db.Where("uuid = ? AND role = 0", claims.UserUUID).First(&user).Error; dbErr != nil {
		// 记录安全事件并清理失效Cookie
		fmt.Printf("[SECURITY WARNING] Invalid JWT token detected - User not found: UUID=%s, Username=%s, IP=%s\n", 
			claims.UserUUID, claims.Username, r.RemoteAddr)
		clearInvalidJWTCookie(w)
		return false
	}

	// 验证用户名是否匹配（防止用户名被修改后仍使用旧令牌）
	if user.Username != claims.Username {
		// 记录安全事件并清理失效Cookie
		fmt.Printf("[SECURITY WARNING] Username mismatch detected - Token username=%s, DB username=%s, UUID=%s, IP=%s\n", 
			claims.Username, user.Username, claims.UserUUID, r.RemoteAddr)
		clearInvalidJWTCookie(w)
		return false
	}

	// 验证密码哈希是否匹配（防止密码被修改后仍使用旧令牌）
	currentPasswordHash := utils.GenerateSHA256Hash(user.Password)
	if claims.PasswordHash != currentPasswordHash {
		// 记录安全事件并清理失效Cookie
		fmt.Printf("[SECURITY WARNING] Password hash mismatch detected - Token may be invalid due to password change: UUID=%s, Username=%s, IP=%s\n", 
			claims.UserUUID, claims.Username, r.RemoteAddr)
		clearInvalidJWTCookie(w)
		return false
	}

	return true
}

// GetCurrentAdminUser 获取当前登录的管理员用户信息
// - 从JWT令牌中提取用户信息
// - 自动刷新接近过期的令牌（剩余时间少于6小时时刷新）
// - 返回用户ID、用户名和角色
func GetCurrentAdminUser(r *http.Request) (*JWTClaims, error) {
	cookie, err := r.Cookie("admin_session")
	if err != nil {
		return nil, fmt.Errorf("未找到会话信息")
	}

	claims, err := parseJWTToken(cookie.Value)
	if err != nil {
		return nil, fmt.Errorf("无效的会话信息")
	}

	if claims.Role != 0 {
		return nil, fmt.Errorf("权限不足")
	}

	// 验证用户是否仍然存在于数据库中
	db, err := database.GetDB()
	if err != nil {
		return nil, fmt.Errorf("数据库连接失败")
	}

	var user models.User
	if dbErr := db.Where("uuid = ? AND role = 0", claims.UserUUID).First(&user).Error; dbErr != nil {
		// 记录安全事件：用户不存在但持有有效JWT令牌
		fmt.Printf("[SECURITY WARNING] Invalid JWT token detected in GetCurrentAdminUser - User not found: UUID=%s, Username=%s, IP=%s\n", 
			claims.UserUUID, claims.Username, r.RemoteAddr)
		return nil, fmt.Errorf("用户不存在或权限已变更")
	}

	// 验证用户名是否匹配（防止用户名被修改后仍使用旧令牌）
	if user.Username != claims.Username {
		// 记录安全事件：用户名不匹配
		fmt.Printf("[SECURITY WARNING] Username mismatch detected in GetCurrentAdminUser - Token username=%s, DB username=%s, UUID=%s, IP=%s\n", 
			claims.Username, user.Username, claims.UserUUID, r.RemoteAddr)
		return nil, fmt.Errorf("用户信息已变更，请重新登录")
	}

	// 验证密码哈希是否匹配（防止密码被修改后仍使用旧令牌）
	currentPasswordHash := utils.GenerateSHA256Hash(user.Password)
	if claims.PasswordHash != currentPasswordHash {
		// 记录安全事件：密码哈希不匹配，可能密码已被修改
		fmt.Printf("[SECURITY WARNING] Password hash mismatch detected in GetCurrentAdminUser - Token may be invalid due to password change: UUID=%s, Username=%s, IP=%s\n", 
			claims.UserUUID, claims.Username, r.RemoteAddr)
		return nil, fmt.Errorf("密码已变更，请重新登录")
	}

	return claims, nil
}

// GetCurrentAdminUserWithRefresh 获取当前登录的管理员用户信息并自动刷新令牌
// - 从JWT令牌中提取用户信息
// - 自动刷新接近过期的令牌（剩余时间少于6小时时刷新）
// - 返回用户ID、用户名、角色和是否刷新了令牌
func GetCurrentAdminUserWithRefresh(w http.ResponseWriter, r *http.Request) (*JWTClaims, bool, error) {
	cookie, err := r.Cookie("admin_session")
	if err != nil {
		return nil, false, fmt.Errorf("未找到会话信息")
	}

	claims, err := parseJWTToken(cookie.Value)
	if err != nil {
		return nil, false, fmt.Errorf("无效的会话信息")
	}

	if claims.Role != 0 {
		return nil, false, fmt.Errorf("权限不足")
	}

	// 验证用户是否仍然存在于数据库中
	db, err := database.GetDB()
	if err != nil {
		return nil, false, fmt.Errorf("数据库连接失败")
	}

	var user models.User
	if dbErr := db.Where("uuid = ? AND role = 0", claims.UserUUID).First(&user).Error; dbErr != nil {
		// 记录安全事件：用户不存在但持有有效JWT令牌
		fmt.Printf("[SECURITY WARNING] Invalid JWT token detected in GetCurrentAdminUserWithRefresh - User not found: UUID=%s, Username=%s, IP=%s\n", 
			claims.UserUUID, claims.Username, r.RemoteAddr)
		return nil, false, fmt.Errorf("用户不存在或权限已变更")
	}

	// 验证用户名是否匹配（防止用户名被修改后仍使用旧令牌）
	if user.Username != claims.Username {
		// 记录安全事件：用户名不匹配
		fmt.Printf("[SECURITY WARNING] Username mismatch detected in GetCurrentAdminUserWithRefresh - Token username=%s, DB username=%s, UUID=%s, IP=%s\n", 
			claims.Username, user.Username, claims.UserUUID, r.RemoteAddr)
		return nil, false, fmt.Errorf("用户信息已变更，请重新登录")
	}

	// 验证密码哈希是否匹配（防止密码被修改后仍使用旧令牌）
	currentPasswordHash := utils.GenerateSHA256Hash(user.Password)
	if claims.PasswordHash != currentPasswordHash {
		// 记录安全事件：密码哈希不匹配，可能密码已被修改
		fmt.Printf("[SECURITY WARNING] Password hash mismatch detected in GetCurrentAdminUserWithRefresh - Token may be invalid due to password change: UUID=%s, Username=%s, IP=%s\n", 
			claims.UserUUID, claims.Username, r.RemoteAddr)
		return nil, false, fmt.Errorf("密码已变更，请重新登录")
	}

	// 检查是否需要刷新令牌（根据配置的阈值）
	refreshed := false
	refreshThreshold := time.Duration(viper.GetInt("security.jwt_refresh_threshold_hours")) * time.Hour
	if time.Until(claims.ExpiresAt.Time) < refreshThreshold {
		// 生成新的JWT令牌
		user := models.User{
			UUID:     claims.UserUUID,
			Username: claims.Username,
			Role:     claims.Role,
		}
		newToken, err := generateJWTToken(user)
		if err == nil {
			// 更新Cookie（使用安全配置）
			newCookie := utils.CreateSecureCookie("admin_session", newToken, utils.GetDefaultCookieMaxAge())
			http.SetCookie(w, newCookie)
			refreshed = true

			// 更新claims的过期时间
			claims.ExpiresAt = jwt.NewNumericDate(time.Now().Add(24 * time.Hour))
			claims.IssuedAt = jwt.NewNumericDate(time.Now())
		}
	}

	return claims, refreshed, nil
}

// AdminAuthRequired 管理员认证拦截中间件
// - 未登录：重定向到 /admin/login
// - 已登录：自动刷新接近过期的令牌，然后放行到后续处理器
func AdminAuthRequired(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 尝试获取用户信息并自动刷新令牌
		claims, refreshed, err := GetCurrentAdminUserWithRefresh(w, r)
		if err != nil {
			// 自动清理失效的JWT Cookie，提升安全性和用户体验
			clearInvalidJWTCookie(w)
			
			// 中文注释：区分普通页面请求与AJAX/JSON请求
			// - 对 AJAX/JSON：直接返回 401 JSON，便于前端处理（如提示重新登录）
			// - 对普通页面：保持原有重定向到登录页
			accept := r.Header.Get("Accept")
			xrw := strings.ToLower(strings.TrimSpace(r.Header.Get("X-Requested-With")))
			if strings.Contains(accept, "application/json") || xrw == "xmlhttprequest" {
				utils.JsonResponse(w, http.StatusUnauthorized, false, "未登录或会话已过期", nil)
				return
			}
			http.Redirect(w, r, "/admin/login", http.StatusFound)
			return
		}

		// 如果令牌被刷新，可以在这里记录日志（可选）
		if refreshed {
			// 可以添加日志记录令牌刷新事件
			_ = claims // 避免未使用变量警告
		}

		next(w, r)
	}
}
