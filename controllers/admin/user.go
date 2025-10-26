package admin

import (
	"encoding/json"
	"net/http"
	"networkDev/database"
	"networkDev/models"
	"networkDev/utils"
	"strings"
)

// UserFragmentHandler 个人资料片段渲染
// - 渲染个人资料与修改密码表单
func UserFragmentHandler(w http.ResponseWriter, r *http.Request) {
	utils.RenderTemplate(w, "user.html", map[string]interface{}{})
}

// UserProfileQueryHandler 获取当前登录管理员的用户名
// - 返回 JSON: {username}
// - 直接从JWT获取用户名信息
func UserProfileQueryHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	claims, _, err := GetCurrentAdminUserWithRefresh(w, r)
	if err != nil {
		utils.JsonResponse(w, http.StatusUnauthorized, false, "未登录或会话已过期", nil)
		return
	}

	utils.JsonResponse(w, http.StatusOK, true, "ok", map[string]interface{}{
		"username": claims.Username,
	})
}

// UserPasswordUpdateHandler 修改当前登录管理员的密码
// - 接收 JSON: {old_password, new_password, confirm_password}
// - 校验旧密码正确性、新密码与确认一致性
// - 成功后更新密码哈希
// - 自动刷新接近过期的JWT令牌
func UserPasswordUpdateHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	claims, _, err := GetCurrentAdminUserWithRefresh(w, r)
	if err != nil {
		utils.JsonResponse(w, http.StatusUnauthorized, false, "未登录或会话已过期", nil)
		return
	}

	var body struct {
		OldPassword     string `json:"old_password"`
		NewPassword     string `json:"new_password"`
		ConfirmPassword string `json:"confirm_password"`
	}
	var decodeErr error
	if decodeErr = json.NewDecoder(r.Body).Decode(&body); decodeErr != nil {
		utils.JsonResponse(w, http.StatusBadRequest, false, "请求参数错误", nil)
		return
	}

	// 基础校验
	if body.OldPassword == "" || body.NewPassword == "" || body.ConfirmPassword == "" {
		utils.JsonResponse(w, http.StatusBadRequest, false, "旧密码/新密码/确认密码均不能为空", nil)
		return
	}
	if len(body.NewPassword) < 6 {
		utils.JsonResponse(w, http.StatusBadRequest, false, "新密码长度不能少于6位", nil)
		return
	}
	if body.NewPassword != body.ConfirmPassword {
		utils.JsonResponse(w, http.StatusBadRequest, false, "两次输入的新密码不一致", nil)
		return
	}
	if body.NewPassword == body.OldPassword {
		utils.JsonResponse(w, http.StatusBadRequest, false, "新密码不能与旧密码相同", nil)
		return
	}

	// 确认是管理员
	if !claims.IsAdmin {
		utils.JsonResponse(w, http.StatusForbidden, false, "权限不足", nil)
		return
	}

	// 获取数据库连接
	db, err := database.GetDB()
	if err != nil {
		utils.JsonResponse(w, http.StatusInternalServerError, false, "数据库连接失败", nil)
		return
	}

	// 通过前缀匹配一次性获取所有管理员相关设置
	var adminSettings []models.Settings
	if err = db.Where("name LIKE ?", "admin_%").Find(&adminSettings).Error; err != nil {
		utils.JsonResponse(w, http.StatusInternalServerError, false, "获取管理员设置失败", nil)
		return
	}

	// 将设置转换为map便于查找
	settingsMap := make(map[string]string)
	for _, setting := range adminSettings {
		settingsMap[setting.Name] = setting.Value
	}

	// 检查必要的设置是否存在
	adminPassword, hasPassword := settingsMap["admin_password"]
	adminPasswordSalt, hasSalt := settingsMap["admin_password_salt"]
	if !hasPassword || !hasSalt {
		utils.JsonResponse(w, http.StatusInternalServerError, false, "管理员密码设置不完整", nil)
		return
	}

	// 校验旧密码
	if !utils.VerifyPasswordWithSalt(body.OldPassword, adminPasswordSalt, adminPassword) {
		utils.JsonResponse(w, http.StatusUnauthorized, false, "旧密码不正确", nil)
		return
	}

	// 生成新的密码盐值
	newSalt, err := utils.GenerateRandomSalt()
	if err != nil {
		utils.JsonResponse(w, http.StatusInternalServerError, false, "生成密码盐失败", nil)
		return
	}

	// 生成新密码哈希
	newPasswordHash, err := utils.HashPasswordWithSalt(body.NewPassword, newSalt)
	if err != nil {
		utils.JsonResponse(w, http.StatusInternalServerError, false, "生成密码哈希失败", nil)
		return
	}

	// 更新settings中的管理员密码和盐值
	if err = db.Model(&models.Settings{}).Where("name = ?", "admin_password").Update("value", newPasswordHash).Error; err != nil {
		utils.JsonResponse(w, http.StatusInternalServerError, false, "更新密码失败", nil)
		return
	}
	if err = db.Model(&models.Settings{}).Where("name = ?", "admin_password_salt").Update("value", newSalt).Error; err != nil {
		utils.JsonResponse(w, http.StatusInternalServerError, false, "更新密码盐值失败", nil)
		return
	}

	// 重新生成JWT令牌（包含新的密码哈希摘要）
	adminUser := models.User{
		Username:     claims.Username,
		Password:     newPasswordHash,
		PasswordSalt: newSalt,
	}
	newToken, err := generateJWTTokenForAdmin(adminUser)
	if err != nil {
		utils.JsonResponse(w, http.StatusInternalServerError, false, "生成新令牌失败", nil)
		return
	}

	// 更新Cookie（使用安全配置）
	cookie := utils.CreateSecureCookie("admin_session", newToken, utils.GetDefaultCookieMaxAge())
	http.SetCookie(w, cookie)

	// 密码修改成功，已重新生成JWT令牌
	utils.JsonResponse(w, http.StatusOK, true, "密码修改成功", nil)
}

// UserProfileUpdateHandler 修改当前登录管理员的用户名
// - 接收 JSON: {username}
// - 校验用户名非空、长度与唯一性
// - 更新数据库后重新签发JWT并写入 Cookie，保持前端展示的一致性
// - 自动刷新接近过期的JWT令牌
func UserProfileUpdateHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	claims, _, err := GetCurrentAdminUserWithRefresh(w, r)
	if err != nil {
		utils.JsonResponse(w, http.StatusUnauthorized, false, "未登录或会话已过期", nil)
		return
	}

	var body struct {
		Username    string `json:"username"`
		OldPassword string `json:"old_password"`
	}
	if decodeErr := json.NewDecoder(r.Body).Decode(&body); decodeErr != nil {
		utils.JsonResponse(w, http.StatusBadRequest, false, "请求参数错误", nil)
		return
	}

	username := strings.TrimSpace(body.Username)
	if username == "" {
		utils.JsonResponse(w, http.StatusBadRequest, false, "用户名不能为空", nil)
		return
	}
	if len(username) > 64 {
		utils.JsonResponse(w, http.StatusBadRequest, false, "用户名长度不能超过64字符", nil)
		return
	}

	db, err := database.GetDB()
	if err != nil {
		utils.JsonResponse(w, http.StatusInternalServerError, false, "数据库连接失败", nil)
		return
	}

	// 确认当前用户是管理员
	if !claims.IsAdmin {
		utils.JsonResponse(w, http.StatusForbidden, false, "权限不足", nil)
		return
	}

	// 获取所有管理员相关设置
	var adminSettings []models.Settings
	if dbErr := db.Where("name LIKE ?", "admin_%").Find(&adminSettings).Error; dbErr != nil {
		utils.JsonResponse(w, http.StatusInternalServerError, false, "获取管理员设置失败", nil)
		return
	}

	// 转换为map便于查找
	settingsMap := make(map[string]string)
	for _, setting := range adminSettings {
		settingsMap[setting.Name] = setting.Value
	}

	adminUsername, exists := settingsMap["admin_username"]
	if !exists {
		utils.JsonResponse(w, http.StatusInternalServerError, false, "管理员用户名设置不存在", nil)
		return
	}

	adminPassword, exists := settingsMap["admin_password"]
	if !exists {
		utils.JsonResponse(w, http.StatusInternalServerError, false, "管理员密码设置不存在", nil)
		return
	}

	adminPasswordSalt, exists := settingsMap["admin_password_salt"]
	if !exists {
		utils.JsonResponse(w, http.StatusInternalServerError, false, "管理员密码盐值设置不存在", nil)
		return
	}

	// 如果用户名未变化则直接返回成功（无需校验旧密码）
	if strings.EqualFold(username, adminUsername) {
		utils.JsonResponse(w, http.StatusOK, true, "保存成功", map[string]interface{}{
			"username": username,
		})
		return
	}

	// 修改用户名需要进行当前密码校验
	if strings.TrimSpace(body.OldPassword) == "" {
		utils.JsonResponse(w, http.StatusBadRequest, false, "修改用户名需要提供当前密码", nil)
		return
	}

	// 使用盐值验证当前密码
	if !utils.VerifyPasswordWithSalt(body.OldPassword, adminPasswordSalt, adminPassword) {
		utils.JsonResponse(w, http.StatusUnauthorized, false, "当前密码不正确", nil)
		return
	}

	// 更新管理员用户名设置
	if dbErr := db.Model(&models.Settings{}).Where("name = ?", "admin_username").Update("value", username).Error; dbErr != nil {
		utils.JsonResponse(w, http.StatusInternalServerError, false, "更新管理员用户名失败", nil)
		return
	}

	// 重新签发JWT并写入Cookie
	// 创建虚拟用户对象用于生成JWT令牌
	adminUser := models.User{
		Username:     username,        // 使用新的用户名
		Password:     adminPassword,
		PasswordSalt: adminPasswordSalt,
	}
	token, err := generateJWTTokenForAdmin(adminUser)
	if err != nil {
		utils.JsonResponse(w, http.StatusInternalServerError, false, "生成新令牌失败", nil)
		return
	}
	cookie := utils.CreateSecureCookie("admin_session", token, utils.GetDefaultCookieMaxAge())
	http.SetCookie(w, cookie)

	utils.JsonResponse(w, http.StatusOK, true, "保存成功", map[string]interface{}{
		"username": username,
	})
}
