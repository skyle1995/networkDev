package admin

import (
	"encoding/json"
	"net/http"
	"networkDev/database"
	"networkDev/models"
	"networkDev/utils"
	"regexp"
	"strconv"
	"strings"

	"github.com/sirupsen/logrus"
)

// VariableFragmentHandler 变量列表页面片段处理器
func VariableFragmentHandler(w http.ResponseWriter, r *http.Request) {
	utils.RenderTemplate(w, "variables", map[string]interface{}{
		"Title": "变量管理",
	})
}

// VariableListHandler 变量列表API处理器
func VariableListHandler(w http.ResponseWriter, r *http.Request) {
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

	// 获取应用UUID参数（用于按应用筛选变量）
	appUUID := strings.TrimSpace(r.URL.Query().Get("app_uuid"))

	// 获取别名搜索参数
	alias := strings.TrimSpace(r.URL.Query().Get("alias"))

	// 构建查询
	db, err := database.GetDB()
	if err != nil {
		logrus.WithError(err).Error("Failed to get database connection")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// 构建基础查询
	query := db.Model(&models.Variable{})

	// 如果指定了应用UUID，则按应用筛选
	if appUUID != "" {
		query = query.Where("app_uuid = ?", appUUID)
	}

	// 如果指定了别名搜索，则按别名模糊搜索
	if alias != "" {
		query = query.Where("alias LIKE ?", "%"+alias+"%")
	}

	// 获取总数
	var total int64
	if err := query.Count(&total).Error; err != nil {
		logrus.WithError(err).Error("Failed to count variables")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// 获取分页数据
	var variables []models.Variable
	offset := (page - 1) * limit
	if err := query.Offset(offset).Limit(limit).Order("created_at DESC").Find(&variables).Error; err != nil {
		logrus.WithError(err).Error("Failed to fetch variables")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// 获取关联的应用信息
	var appUUIDs []string
	for _, variable := range variables {
		appUUIDs = append(appUUIDs, variable.AppUUID)
	}

	var apps []models.App
	if len(appUUIDs) > 0 {
		if err := db.Where("uuid IN ?", appUUIDs).Find(&apps).Error; err != nil {
			logrus.WithError(err).Error("Failed to fetch related apps")
		}
	}

	// 创建应用UUID到应用名称的映射
	appMap := make(map[string]string)
	for _, app := range apps {
		appMap[app.UUID] = app.Name
	}

	// 构建响应数据
	type VariableResponse struct {
		ID        uint   `json:"id"`
		UUID      string `json:"uuid"`
		Number    string `json:"number"`
		AppUUID   string `json:"app_uuid"`
		AppName   string `json:"app_name"`
		Alias     string `json:"alias"`
		Data      string `json:"data"`
		Remark    string `json:"remark"`
		CreatedAt string `json:"created_at"`
		UpdatedAt string `json:"updated_at"`
	}

	var responseData []VariableResponse
	for _, variable := range variables {
		appName := appMap[variable.AppUUID]
		if appName == "" {
			appName = "未知应用"
		}

		responseData = append(responseData, VariableResponse{
			ID:        variable.ID,
			UUID:      variable.UUID,
			Number:    variable.Number,
			AppUUID:   variable.AppUUID,
			AppName:   appName,
			Alias:     variable.Alias,
			Data:      variable.Data,
			Remark:    variable.Remark,
			CreatedAt: variable.CreatedAt.Format("2006-01-02 15:04:05"),
			UpdatedAt: variable.UpdatedAt.Format("2006-01-02 15:04:05"),
		})
	}

	response := map[string]interface{}{
		"code":  0,
		"msg":   "success",
		"count": total,
		"data":  responseData,
	}

	w.Header().Set("Content-Type", "application/json")
	utils.WriteJSONResponse(w, http.StatusOK, response)
}

// VariableCreateHandler 新增变量API处理器
func VariableCreateHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		AppUUID string `json:"app_uuid"`
		Alias   string `json:"alias"`
		Data    string `json:"data"`
		Remark  string `json:"remark"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logrus.WithError(err).Error("Failed to decode JSON request")
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// 验证必填字段
	if strings.TrimSpace(req.AppUUID) == "" {
		http.Error(w, "应用UUID不能为空", http.StatusBadRequest)
		return
	}
	if strings.TrimSpace(req.Alias) == "" {
		http.Error(w, "变量别名不能为空", http.StatusBadRequest)
		return
	}

	// 验证别名格式：必须以英文字母开头，只能包含数字和英文字母
	aliasPattern := regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9]*$`)
	if !aliasPattern.MatchString(req.Alias) {
		http.Error(w, "别名必须以英文字母开头，只能包含数字和英文字母", http.StatusBadRequest)
		return
	}

	db, err := database.GetDB()
	if err != nil {
		logrus.WithError(err).Error("Failed to get database connection")
		http.Error(w, "数据库连接失败", http.StatusInternalServerError)
		return
	}

	// 验证应用是否存在
	var app models.App
	if err := db.Where("uuid = ?", req.AppUUID).First(&app).Error; err != nil {
		http.Error(w, "应用不存在", http.StatusBadRequest)
		return
	}

	// 创建变量
	variable := models.Variable{
		AppUUID: strings.TrimSpace(req.AppUUID),
		Alias:   strings.TrimSpace(req.Alias),
		Data:    req.Data,
		Remark:  strings.TrimSpace(req.Remark),
	}

	if err := db.Create(&variable).Error; err != nil {
		logrus.WithError(err).Error("Failed to create variable")
		http.Error(w, "创建变量失败", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"code": 0,
		"msg":  "创建成功",
		"data": variable,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// VariableUpdateHandler 更新变量API处理器
func VariableUpdateHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		UUID    string `json:"uuid"`
		AppUUID string `json:"app_uuid"`
		Alias   string `json:"alias"`
		Data    string `json:"data"`
		Remark  string `json:"remark"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// 验证必填字段
	if strings.TrimSpace(req.UUID) == "" {
		http.Error(w, "变量UUID不能为空", http.StatusBadRequest)
		return
	}
	if strings.TrimSpace(req.AppUUID) == "" {
		http.Error(w, "应用UUID不能为空", http.StatusBadRequest)
		return
	}
	if strings.TrimSpace(req.Alias) == "" {
		http.Error(w, "变量别名不能为空", http.StatusBadRequest)
		return
	}

	// 验证别名格式：必须以英文字母开头，只能包含数字和英文字母
	aliasPattern := regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9]*$`)
	if !aliasPattern.MatchString(req.Alias) {
		http.Error(w, "别名必须以英文字母开头，只能包含数字和英文字母", http.StatusBadRequest)
		return
	}

	db, err := database.GetDB()
	if err != nil {
		logrus.WithError(err).Error("Failed to get database connection")
		http.Error(w, "数据库连接失败", http.StatusInternalServerError)
		return
	}

	// 验证应用是否存在
	var app models.App
	if err := db.Where("uuid = ?", req.AppUUID).First(&app).Error; err != nil {
		http.Error(w, "应用不存在", http.StatusBadRequest)
		return
	}

	// 通过uuid字段查找变量
	var variable models.Variable
	if err := db.Where("uuid = ?", strings.TrimSpace(req.UUID)).First(&variable).Error; err != nil {
		http.Error(w, "变量不存在", http.StatusNotFound)
		return
	}

	// 更新字段
	variable.AppUUID = strings.TrimSpace(req.AppUUID)
	variable.Alias = strings.TrimSpace(req.Alias)
	variable.Data = req.Data
	variable.Remark = strings.TrimSpace(req.Remark)

	if err := db.Save(&variable).Error; err != nil {
		logrus.WithError(err).Error("Failed to update variable")
		http.Error(w, "更新变量失败", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"code": 0,
		"msg":  "更新成功",
		"data": variable,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// VariableDeleteHandler 删除变量API处理器
func VariableDeleteHandler(w http.ResponseWriter, r *http.Request) {
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
		http.Error(w, "变量ID不能为空", http.StatusBadRequest)
		return
	}

	db, err := database.GetDB()
	if err != nil {
		logrus.WithError(err).Error("Failed to get database connection")
		http.Error(w, "数据库连接失败", http.StatusInternalServerError)
		return
	}

	// 删除变量
	if err := db.Delete(&models.Variable{}, req.ID).Error; err != nil {
		logrus.WithError(err).Error("Failed to delete variable")
		http.Error(w, "删除变量失败", http.StatusInternalServerError)
		return
	}

	logrus.WithField("variable_id", req.ID).Info("Successfully deleted variable")

	response := map[string]interface{}{
		"code": 0,
		"msg":  "删除成功",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// VariablesBatchDeleteHandler 批量删除变量API处理器
func VariablesBatchDeleteHandler(w http.ResponseWriter, r *http.Request) {
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
		http.Error(w, "请选择要删除的变量", http.StatusBadRequest)
		return
	}

	db, err := database.GetDB()
	if err != nil {
		logrus.WithError(err).Error("Failed to get database connection")
		http.Error(w, "数据库连接失败", http.StatusInternalServerError)
		return
	}

	// 批量删除变量
	if err := db.Delete(&models.Variable{}, req.IDs).Error; err != nil {
		logrus.WithError(err).Error("Failed to batch delete variables")
		http.Error(w, "批量删除失败", http.StatusInternalServerError)
		return
	}

	logrus.WithField("variable_ids", req.IDs).Info("Successfully batch deleted variables")

	response := map[string]interface{}{
		"code": 0,
		"msg":  "批量删除成功",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// VariableGetAppsHandler 获取应用列表（用于筛选下拉框）
func VariableGetAppsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	db, err := database.GetDB()
	if err != nil {
		logrus.WithError(err).Error("Failed to get database connection")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	var apps []models.App
	if err := db.Select("uuid, name").Find(&apps).Error; err != nil {
		logrus.WithError(err).Error("Failed to fetch apps")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"code": 0,
		"msg":  "success",
		"data": apps,
	}

	w.Header().Set("Content-Type", "application/json")
	utils.WriteJSONResponse(w, http.StatusOK, response)
}