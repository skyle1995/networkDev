package admin

import (
	"net/http"
	"networkDev/controllers"
	"networkDev/models"
	"networkDev/utils"
	"strings"

	"github.com/gin-gonic/gin"
)

// 创建基础控制器实例
var baseController = controllers.NewBaseController()

// UserFragmentHandler 个人资料片段渲染
// - 渲染个人资料与修改密码表单
func UserFragmentHandler(c *gin.Context) {
	c.HTML(http.StatusOK, "user.html", gin.H{})
}

// UserProfileQueryHandler 获取当前登录管理员的用户名
// - 返回 JSON: {username}
// - 直接从JWT获取用户名信息
func UserProfileQueryHandler(c *gin.Context) {
	claims, _, err := GetCurrentAdminUserWithRefresh(c)
	if err != nil {
		baseController.HandleValidationError(c, "未登录或会话已过期")
		return
	}

	baseController.HandleSuccess(c, "ok", gin.H{
		"username": claims.Username,
	})
}

// UserPasswordUpdateHandler 修改当前登录管理员的密码
// - 接收 JSON: {old_password, new_password, confirm_password}
// - 校验旧密码正确性、新密码与确认一致性
// - 成功后更新密码哈希
// - 自动刷新接近过期的JWT令牌
func UserPasswordUpdateHandler(c *gin.Context) {
	claims, _, err := GetCurrentAdminUserWithRefresh(c)
	if err != nil {
		baseController.HandleValidationError(c, "未登录或会话已过期")
		return
	}

	var body struct {
		OldPassword     string `json:"old_password"`
		NewPassword     string `json:"new_password"`
		ConfirmPassword string `json:"confirm_password"`
	}
	
	if !baseController.BindJSON(c, &body) {
		return
	}

	// 基础校验
	if !baseController.ValidateRequired(c, map[string]interface{}{
		"旧密码": body.OldPassword,
		"新密码": body.NewPassword,
		"确认密码": body.ConfirmPassword,
	}) {
		return
	}
	
	if len(body.NewPassword) < 6 {
		baseController.HandleValidationError(c, "新密码长度不能少于6位")
		return
	}
	if body.NewPassword != body.ConfirmPassword {
		baseController.HandleValidationError(c, "两次输入的新密码不一致")
		return
	}
	if body.NewPassword == body.OldPassword {
		baseController.HandleValidationError(c, "新密码不能与旧密码相同")
		return
	}

	// 注释：由于使用了AdminAuthRequired中间件，已确保是管理员用户

	// 获取数据库连接
	db, ok := baseController.GetDB(c)
	if !ok {
		return
	}

	// 通过前缀匹配一次性获取所有管理员相关设置
	var adminSettings []models.Settings
	if err = db.Where("name LIKE ?", "admin_%").Find(&adminSettings).Error; err != nil {
		baseController.HandleInternalError(c, "获取管理员设置失败", err)
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
		baseController.HandleInternalError(c, "管理员密码设置不完整", nil)
		return
	}

	// 校验旧密码
	if !utils.VerifyPasswordWithSalt(body.OldPassword, adminPasswordSalt, adminPassword) {
		baseController.HandleValidationError(c, "旧密码不正确")
		return
	}

	// 生成新的密码盐值
	newSalt, err := utils.GenerateRandomSalt()
	if err != nil {
		baseController.HandleInternalError(c, "生成密码盐失败", err)
		return
	}

	// 生成新密码哈希
	newPasswordHash, err := utils.HashPasswordWithSalt(body.NewPassword, newSalt)
	if err != nil {
		baseController.HandleInternalError(c, "生成密码哈希失败", err)
		return
	}

	// 更新settings中的管理员密码和盐值
	if err = db.Model(&models.Settings{}).Where("name = ?", "admin_password").Update("value", newPasswordHash).Error; err != nil {
		baseController.HandleInternalError(c, "更新密码失败", err)
		return
	}
	if err = db.Model(&models.Settings{}).Where("name = ?", "admin_password_salt").Update("value", newSalt).Error; err != nil {
		baseController.HandleInternalError(c, "更新密码盐值失败", err)
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
		baseController.HandleInternalError(c, "生成新令牌失败", err)
		return
	}

	// 更新Cookie（使用安全配置）
	c.SetCookie("admin_session", newToken, utils.GetDefaultCookieMaxAge(), "/", "", false, true)

	// 密码修改成功，已重新生成JWT令牌
	baseController.HandleSuccess(c, "密码修改成功", nil)
}

// UserProfileUpdateHandler 修改当前登录管理员的用户名
// - 接收 JSON: {username}
// - 校验用户名非空、长度与唯一性
// - 更新数据库后重新签发JWT并写入 Cookie，保持前端展示的一致性
// - 自动刷新接近过期的JWT令牌
func UserProfileUpdateHandler(c *gin.Context) {
	_, _, err := GetCurrentAdminUserWithRefresh(c)
	if err != nil {
		baseController.HandleValidationError(c, "未登录或会话已过期")
		return
	}

	var body struct {
		Username    string `json:"username"`
		OldPassword string `json:"old_password"`
	}
	if !baseController.BindJSON(c, &body) {
		return
	}

	username := strings.TrimSpace(body.Username)
	if username == "" {
		baseController.HandleValidationError(c, "用户名不能为空")
		return
	}
	if len(username) > 64 {
		baseController.HandleValidationError(c, "用户名长度不能超过64字符")
		return
	}

	db, ok := baseController.GetDB(c)
	if !ok {
		return
	}

	// 注释：由于使用了AdminAuthRequired中间件，已确保是管理员用户

	// 获取所有管理员相关设置
	var adminSettings []models.Settings
	if dbErr := db.Where("name LIKE ?", "admin_%").Find(&adminSettings).Error; dbErr != nil {
		baseController.HandleInternalError(c, "获取管理员设置失败", dbErr)
		return
	}

	// 转换为map便于查找
	settingsMap := make(map[string]string)
	for _, setting := range adminSettings {
		settingsMap[setting.Name] = setting.Value
	}

	adminUsername, exists := settingsMap["admin_username"]
	if !exists {
		baseController.HandleInternalError(c, "管理员用户名设置不存在", nil)
		return
	}

	adminPassword, exists := settingsMap["admin_password"]
	if !exists {
		baseController.HandleInternalError(c, "管理员密码设置不存在", nil)
		return
	}

	adminPasswordSalt, exists := settingsMap["admin_password_salt"]
	if !exists {
		baseController.HandleInternalError(c, "管理员密码盐值设置不存在", nil)
		return
	}

	// 如果用户名未变化则直接返回成功（无需校验旧密码）
	if strings.EqualFold(username, adminUsername) {
		baseController.HandleSuccess(c, "保存成功", gin.H{
			"username": username,
		})
		return
	}

	// 修改用户名需要进行当前密码校验
	if strings.TrimSpace(body.OldPassword) == "" {
		baseController.HandleValidationError(c, "修改用户名需要提供当前密码")
		return
	}

	// 使用盐值验证当前密码
	if !utils.VerifyPasswordWithSalt(body.OldPassword, adminPasswordSalt, adminPassword) {
		baseController.HandleValidationError(c, "当前密码不正确")
		return
	}

	// 更新管理员用户名设置
	if dbErr := db.Model(&models.Settings{}).Where("name = ?", "admin_username").Update("value", username).Error; dbErr != nil {
		baseController.HandleInternalError(c, "更新管理员用户名失败", dbErr)
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
		baseController.HandleInternalError(c, "生成新令牌失败", err)
		return
	}
	c.SetCookie("admin_session", token, utils.GetDefaultCookieMaxAge(), "/", "", false, true)

	baseController.HandleSuccess(c, "保存成功", gin.H{
		"username": username,
	})
}
