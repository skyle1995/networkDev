package admin

import (
	"crypto/rand"
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
		app.UUID = uuid.New().String()
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
