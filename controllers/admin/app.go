package admin

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"networkDev/database"
	"networkDev/models"
	"networkDev/utils"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// AppsFragmentHandler 应用列表页面片段处理器
func AppsFragmentHandler(w http.ResponseWriter, r *http.Request) {
	utils.RenderTemplate(w, "apps.html", map[string]interface{}{
		"Title": "应用管理",
	})
}

// AppsListHandler 应用列表API处理器
func AppsListHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// 获取分页参数
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page <= 0 {
		page = 1
	}
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit <= 0 {
		limit = 10
	}

	// 获取搜索参数
	search := strings.TrimSpace(r.URL.Query().Get("search"))

	// 构建查询
	db, err := database.GetDB()
	if err != nil {
		logrus.WithError(err).Error("Failed to get database connection")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	var apps []models.App
	var total int64

	query := db.Model(&models.App{})

	// 如果有搜索条件
	if search != "" {
		query = query.Where("name LIKE ? OR uuid LIKE ?", "%"+search+"%", "%"+search+"%")
	}

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		logrus.WithError(err).Error("Failed to count apps")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// 分页查询
	offset := (page - 1) * limit
	if err := query.Order("created_at DESC").Offset(offset).Limit(limit).Find(&apps).Error; err != nil {
		logrus.WithError(err).Error("Failed to query apps")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// 返回结果
	response := map[string]interface{}{
		"code":  0,
		"msg":   "success",
		"count": total,
		"data":  apps,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// AppGetAnnouncementHandler 获取应用程序公告处理器
func AppGetAnnouncementHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// 获取UUID参数
	uuid := r.URL.Query().Get("uuid")
	if uuid == "" {
		response := map[string]interface{}{
			"code": 1,
			"msg":  "应用UUID不能为空",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
		return
	}

	// 获取数据库连接
	db, err := database.GetDB()
	if err != nil {
		logrus.WithError(err).Error("Failed to get database connection")
		response := map[string]interface{}{
			"code": 1,
			"msg":  "数据库连接失败",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
		return
	}

	// 查找应用
	var app models.App
	if err := db.Where("uuid = ?", uuid).First(&app).Error; err != nil {
		logrus.WithError(err).Error("Failed to find app")
		response := map[string]interface{}{
			"code": 1,
			"msg":  "应用不存在",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
		return
	}

	// 解码base64公告内容
	var announcement string
	if app.Announcement != "" {
		decodedBytes, err := base64.StdEncoding.DecodeString(app.Announcement)
		if err != nil {
			logrus.WithError(err).Error("Failed to decode announcement")
			// 如果解码失败，返回空字符串
			announcement = ""
		} else {
			announcement = string(decodedBytes)
		}
	}

	response := map[string]interface{}{
		"code": 0,
		"msg":  "获取成功",
		"data": map[string]interface{}{
			"announcement": announcement,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// AppResetSecretHandler 重置应用密钥API处理器
func AppResetSecretHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		UUID string `json:"uuid"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if req.UUID == "" {
		http.Error(w, "应用UUID不能为空", http.StatusBadRequest)
		return
	}

	db, err := database.GetDB()
	if err != nil {
		logrus.WithError(err).Error("Failed to get database connection")
		http.Error(w, "数据库连接失败", http.StatusInternalServerError)
		return
	}

	// 查找应用
	var app models.App
	if err := db.Where("uuid = ?", req.UUID).First(&app).Error; err != nil {
		logrus.WithError(err).Error("Failed to find app by UUID")
		http.Error(w, "应用不存在", http.StatusNotFound)
		return
	}

	// 生成新的密钥
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		logrus.WithError(err).Error("Failed to generate random secret")
		http.Error(w, "生成密钥失败", http.StatusInternalServerError)
		return
	}
	newSecret := strings.ToUpper(hex.EncodeToString(bytes))

	// 更新密钥
	if err := db.Model(&app).Update("secret", newSecret).Error; err != nil {
		logrus.WithError(err).Error("Failed to update app secret")
		http.Error(w, "更新密钥失败", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"code": 0,
		"msg":  "密钥重置成功",
		"data": map[string]interface{}{
			"uuid":   app.UUID,
			"secret": newSecret,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// AppCreateHandler 创建应用API处理器
func AppCreateHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Name         string `json:"name"`
		Version      string `json:"version"`
		Status       int    `json:"status"`
		DownloadType int    `json:"download_type"`
		ForceUpdate  int    `json:"force_update"`
		DownloadURL  string `json:"download_url"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logrus.WithError(err).Error("Failed to decode JSON request")
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// 验证必填字段
	if strings.TrimSpace(req.Name) == "" {
		logrus.Error("App name is empty")
		http.Error(w, "应用名称不能为空", http.StatusBadRequest)
		return
	}

	// 设置默认值
	if req.Version == "" {
		req.Version = "1.0.0"
	}

	logrus.WithFields(logrus.Fields{
		"name":          req.Name,
		"version":       req.Version,
		"status":        req.Status,
		"download_type": req.DownloadType,
		"download_url":  req.DownloadURL,
		"force_update":  req.ForceUpdate,
	}).Info("Received app create request")

	// 创建应用
	app := models.App{
		Name:         strings.TrimSpace(req.Name),
		Version:      req.Version,
		Status:       req.Status,
		DownloadType: req.DownloadType,
		DownloadURL:  strings.TrimSpace(req.DownloadURL),
		ForceUpdate:  req.ForceUpdate,
	}

	// 确保UUID和Secret被设置（虽然BeforeCreate钩子应该处理这些，但为了保险起见）
	if app.UUID == "" {
		app.UUID = strings.ToUpper(uuid.New().String())
	}
	if app.Secret == "" {
		// 生成32位大写16进制随机字符
		bytes := make([]byte, 16) // 16字节 = 32位16进制字符
		rand.Read(bytes)
		app.Secret = strings.ToUpper(hex.EncodeToString(bytes))
	}

	db, err := database.GetDB()
	if err != nil {
		logrus.WithError(err).Error("Failed to get database connection")
		http.Error(w, "数据库连接失败", http.StatusInternalServerError)
		return
	}

	if err := db.Create(&app).Error; err != nil {
		logrus.WithError(err).Error("Failed to create app")
		http.Error(w, "创建应用失败", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"code": 0,
		"msg":  "创建成功",
		"data": app,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// AppUpdateHandler 更新应用API处理器
func AppUpdateHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		ID           uint   `json:"id"`
		Name         string `json:"name"`
		Version      string `json:"version"`
		Status       int    `json:"status"`
		DownloadType int    `json:"download_type"`
		DownloadURL  string `json:"download_url"`
		ForceUpdate  int    `json:"force_update"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// 验证必填字段
	if req.ID == 0 {
		http.Error(w, "应用ID不能为空", http.StatusBadRequest)
		return
	}
	if strings.TrimSpace(req.Name) == "" {
		http.Error(w, "应用名称不能为空", http.StatusBadRequest)
		return
	}

	db, err := database.GetDB()
	if err != nil {
		logrus.WithError(err).Error("Failed to get database connection")
		http.Error(w, "数据库连接失败", http.StatusInternalServerError)
		return
	}

	// 查找应用
	var app models.App
	if err := db.First(&app, req.ID).Error; err != nil {
		http.Error(w, "应用不存在", http.StatusNotFound)
		return
	}

	// 更新字段
	app.Name = strings.TrimSpace(req.Name)
	app.Version = req.Version
	app.Status = req.Status
	app.DownloadType = req.DownloadType
	app.DownloadURL = strings.TrimSpace(req.DownloadURL)
	app.ForceUpdate = req.ForceUpdate

	if err := db.Save(&app).Error; err != nil {
		logrus.WithError(err).Error("Failed to update app")
		http.Error(w, "更新应用失败", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"code": 0,
		"msg":  "更新成功",
		"data": app,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// AppDeleteHandler 删除应用API处理器
func AppDeleteHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		ID uint `json:"id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if req.ID == 0 {
		http.Error(w, "应用ID不能为空", http.StatusBadRequest)
		return
	}

	db, err := database.GetDB()
	if err != nil {
		logrus.WithError(err).Error("Failed to get database connection")
		http.Error(w, "数据库连接失败", http.StatusInternalServerError)
		return
	}

	// 删除应用
	if err := db.Delete(&models.App{}, req.ID).Error; err != nil {
		logrus.WithError(err).Error("Failed to delete app")
		http.Error(w, "删除应用失败", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"code": 0,
		"msg":  "删除成功",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// AppsBatchDeleteHandler 批量删除应用API处理器
func AppsBatchDeleteHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		IDs []uint `json:"ids"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if len(req.IDs) == 0 {
		http.Error(w, "请选择要删除的应用", http.StatusBadRequest)
		return
	}

	db, err := database.GetDB()
	if err != nil {
		logrus.WithError(err).Error("Failed to get database connection")
		http.Error(w, "数据库连接失败", http.StatusInternalServerError)
		return
	}

	// 批量删除
	if err := db.Delete(&models.App{}, req.IDs).Error; err != nil {
		logrus.WithError(err).Error("Failed to batch delete apps")
		http.Error(w, "批量删除失败", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"code": 0,
		"msg":  "批量删除成功",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// AppsBatchUpdateStatusHandler 批量更新应用状态API处理器
func AppsBatchUpdateStatusHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		IDs    []uint `json:"ids"`
		Status int    `json:"status"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if len(req.IDs) == 0 {
		http.Error(w, "请选择要更新的应用", http.StatusBadRequest)
		return
	}

	if req.Status != 0 && req.Status != 1 {
		http.Error(w, "状态值无效", http.StatusBadRequest)
		return
	}

	db, err := database.GetDB()
	if err != nil {
		logrus.WithError(err).Error("Failed to get database connection")
		http.Error(w, "数据库连接失败", http.StatusInternalServerError)
		return
	}

	// 批量更新状态
	if err := db.Model(&models.App{}).Where("id IN ?", req.IDs).Update("status", req.Status).Error; err != nil {
		logrus.WithError(err).Error("Failed to batch update app status")
		http.Error(w, "批量更新状态失败", http.StatusInternalServerError)
		return
	}

	statusText := "禁用"
	if req.Status == 1 {
		statusText = "启用"
	}

	response := map[string]interface{}{
		"code": 0,
		"msg":  "批量" + statusText + "成功",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// AppUpdateAnnouncementHandler 更新应用程序公告处理器
func AppUpdateAnnouncementHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// 解析请求体
	var req struct {
		UUID         string `json:"uuid"`
		Announcement string `json:"announcement"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logrus.WithError(err).Error("Failed to decode request body")
		response := map[string]interface{}{
			"code": 1,
			"msg":  "请求参数格式错误",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
		return
	}

	// 验证UUID
	if req.UUID == "" {
		response := map[string]interface{}{
			"code": 1,
			"msg":  "应用UUID不能为空",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
		return
	}

	// 验证UUID格式
	if _, err := uuid.Parse(req.UUID); err != nil {
		logrus.WithError(err).Error("Invalid UUID format")
		response := map[string]interface{}{
			"code": 1,
			"msg":  "无效的UUID格式",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
		return
	}

	// 获取数据库连接
	db, err := database.GetDB()
	if err != nil {
		logrus.WithError(err).Error("Failed to get database connection")
		response := map[string]interface{}{
			"code": 1,
			"msg":  "数据库连接失败",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
		return
	}

	// 查找应用
	var app models.App
	if err := db.Where("uuid = ?", req.UUID).First(&app).Error; err != nil {
		logrus.WithError(err).Error("Failed to find app")
		response := map[string]interface{}{
			"code": 1,
			"msg":  "应用不存在",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
		return
	}

	// 对公告内容进行base64编码
	encodedAnnouncement := base64.StdEncoding.EncodeToString([]byte(req.Announcement))

	// 更新应用的公告内容
	if err := db.Model(&app).Update("announcement", encodedAnnouncement).Error; err != nil {
		logrus.WithError(err).Error("Failed to update app announcement")
		response := map[string]interface{}{
			"code": 1,
			"msg":  "更新程序公告失败",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
		return
	}

	logrus.WithFields(logrus.Fields{
		"app_uuid": req.UUID,
		"app_name": app.Name,
	}).Info("App announcement updated successfully")

	response := map[string]interface{}{
		"code": 0,
		"msg":  "程序公告更新成功",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// ... existing code ...

// AppGetMultiConfigHandler 获取应用多开配置
func AppGetMultiConfigHandler(w http.ResponseWriter, r *http.Request) {
	appUUID := r.URL.Query().Get("uuid")
	if appUUID == "" {
		http.Error(w, "缺少应用UUID", http.StatusBadRequest)
		return
	}

	// 验证UUID格式
	if _, err := uuid.Parse(appUUID); err != nil {
		http.Error(w, "无效的UUID格式", http.StatusBadRequest)
		return
	}

	db, err := database.GetDB()
	if err != nil {
		logrus.WithError(err).Error("Failed to get database connection")
		http.Error(w, "数据库连接失败", http.StatusInternalServerError)
		return
	}

	var app models.App
	if err := db.Where("uuid = ?", appUUID).First(&app).Error; err != nil {
		logrus.WithError(err).Error("Failed to find app")
		http.Error(w, "应用不存在", http.StatusNotFound)
		return
	}

	// 返回多开配置信息
	response := map[string]interface{}{
		"login_type":       app.LoginType,
		"multi_open_scope": app.MultiOpenScope,
		"clean_interval":   app.CleanInterval,
		"check_interval":   app.CheckInterval,
		"multi_open_count": app.MultiOpenCount,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// AppUpdateMultiConfigHandler 更新应用多开配置
func AppUpdateMultiConfigHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		UUID            string `json:"uuid"`
		LoginType       int    `json:"login_type"`
		MultiOpenScope  int    `json:"multi_open_scope"`
		CleanInterval   int    `json:"clean_interval"`
		CheckInterval   int    `json:"check_interval"`
		MultiOpenCount  int    `json:"multi_open_count"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// 验证UUID格式
	if _, err := uuid.Parse(req.UUID); err != nil {
		http.Error(w, "无效的UUID格式", http.StatusBadRequest)
		return
	}

	// 验证参数范围
	if req.LoginType < 0 || req.LoginType > 1 {
		http.Error(w, "登录方式参数无效", http.StatusBadRequest)
		return
	}
	if req.MultiOpenScope < 0 || req.MultiOpenScope > 2 {
		http.Error(w, "多开范围参数无效", http.StatusBadRequest)
		return
	}
	if req.CleanInterval < 1 {
		http.Error(w, "清理间隔必须大于0", http.StatusBadRequest)
		return
	}
	if req.CheckInterval < 1 {
		http.Error(w, "校验间隔必须大于0", http.StatusBadRequest)
		return
	}
	if req.MultiOpenCount < 1 {
		http.Error(w, "多开数量必须大于0", http.StatusBadRequest)
		return
	}

	db, err := database.GetDB()
	if err != nil {
		logrus.WithError(err).Error("Failed to get database connection")
		http.Error(w, "数据库连接失败", http.StatusInternalServerError)
		return
	}

	// 查找应用
	var app models.App
	if err := db.Where("uuid = ?", req.UUID).First(&app).Error; err != nil {
		logrus.WithError(err).Error("Failed to find app")
		http.Error(w, "应用不存在", http.StatusNotFound)
		return
	}

	// 更新多开配置
	updates := map[string]interface{}{
		"login_type":       req.LoginType,
		"multi_open_scope": req.MultiOpenScope,
		"clean_interval":   req.CleanInterval,
		"check_interval":   req.CheckInterval,
		"multi_open_count": req.MultiOpenCount,
	}

	if err := db.Model(&app).Updates(updates).Error; err != nil {
		logrus.WithError(err).Error("Failed to update app multi config")
		http.Error(w, "更新多开配置失败", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "多开配置更新成功"})
}
