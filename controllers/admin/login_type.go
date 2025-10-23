package admin

import (
	"encoding/json"
	"net/http"
	"networkDev/database"
	"networkDev/models"
	"networkDev/utils"
	"strconv"
	"strings"
)

// LoginTypesFragmentHandler 登录方式管理片段渲染
// - 渲染 login_types.html 列表与表单界面
func LoginTypesFragmentHandler(w http.ResponseWriter, r *http.Request) {
	utils.RenderTemplate(w, "login_types.html", map[string]interface{}{})
}

// LoginTypesListHandler 获取登录方式列表
// - 支持GET
// - 支持分页和筛选
func LoginTypesListHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	// 获取查询参数
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))
	keyword := r.URL.Query().Get("keyword")
	statusStr := r.URL.Query().Get("status")

	// 设置默认分页参数
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}

	db, err := database.GetDB()
	if err != nil {
		utils.JsonResponse(w, http.StatusInternalServerError, false, "数据库连接失败", nil)
		return
	}

	// 构建查询条件
	query := db.Model(&models.LoginType{})

	// 筛选条件
	if keyword != "" {
		query = query.Where("name LIKE ?", "%"+keyword+"%")
	}
	if statusStr != "" {
		if status, err := strconv.Atoi(statusStr); err == nil {
			query = query.Where("status = ?", status)
		}
	}

	// 计算总数
	var total int64
	if err := query.Count(&total).Error; err != nil {
		utils.JsonResponse(w, http.StatusInternalServerError, false, "统计总数失败", nil)
		return
	}

	// 分页查询
	var items []models.LoginType
	offset := (page - 1) * pageSize
	if err := query.Order("id asc").Offset(offset).Limit(pageSize).Find(&items).Error; err != nil {
		utils.JsonResponse(w, http.StatusInternalServerError, false, "查询失败", nil)
		return
	}

	// 返回分页数据
	result := map[string]interface{}{
		"items":     items,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
		"pages":     (total + int64(pageSize) - 1) / int64(pageSize),
	}
	utils.JsonResponse(w, http.StatusOK, true, "ok", result)
}

// LoginTypeCreateHandler 新增登录方式
// - 接收JSON: {name, description, status}
// - Name 必填且唯一
func LoginTypeCreateHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
	type reqBody struct {
		Name        string `json:"name"`
		VerifyTypes string `json:"verify_types"`
		Status      int    `json:"status"`
	}
	var body reqBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		utils.JsonResponse(w, http.StatusBadRequest, false, "请求体错误", nil)
		return
	}
	if body.Name == "" {
		utils.JsonResponse(w, http.StatusBadRequest, false, "名称不能为空", nil)
		return
	}
	db, err := database.GetDB()
	if err != nil {
		utils.JsonResponse(w, http.StatusInternalServerError, false, "数据库连接失败", nil)
		return
	}
	item := models.LoginType{
		Name:        body.Name,
		Status:      body.Status,
		VerifyTypes: body.VerifyTypes,
	}
	if item.Status != 0 {
		item.Status = 1
	}
	if err := db.Create(&item).Error; err != nil {
		utils.JsonResponse(w, http.StatusBadRequest, false, "创建失败，可能是名称重复", nil)
		return
	}
	utils.JsonResponse(w, http.StatusOK, true, "创建成功", item)
}

// checkLoginTypeInUse 检查登录类型是否被卡密类型使用
// - 检查 card_types 表中的 login_types 字段是否包含该登录类型名称
// - 返回是否被使用和使用该登录类型的卡密类型名称列表
func checkLoginTypeInUse(loginTypeName string) (bool, []string, error) {
	db, err := database.GetDB()
	if err != nil {
		return false, nil, err
	}

	var cardTypes []models.CardType
	// 查询包含该登录类型名称的卡密类型
	if err := db.Where("login_types LIKE ?", "%"+loginTypeName+"%").Find(&cardTypes).Error; err != nil {
		return false, nil, err
	}

	var usingCardTypes []string
	for _, cardType := range cardTypes {
		// 精确匹配登录类型名称（避免部分匹配）
		loginTypes := strings.Split(cardType.LoginTypes, ",")
		for _, lt := range loginTypes {
			if strings.TrimSpace(lt) == loginTypeName {
				usingCardTypes = append(usingCardTypes, cardType.Name)
				break
			}
		}
	}

	return len(usingCardTypes) > 0, usingCardTypes, nil
}

// checkLoginTypesByIDsInUse 批量检查登录类型ID是否被使用
// - 先查询登录类型ID对应的名称，再检查是否被使用
func checkLoginTypesByIDsInUse(loginTypeIDs []uint) (bool, map[uint][]string, error) {
	db, err := database.GetDB()
	if err != nil {
		return false, nil, err
	}

	// 查询登录类型名称
	var loginTypes []models.LoginType
	if err := db.Where("id IN ?", loginTypeIDs).Find(&loginTypes).Error; err != nil {
		return false, nil, err
	}

	hasUsage := false
	usageMap := make(map[uint][]string)

	for _, loginType := range loginTypes {
		inUse, usingCardTypes, err := checkLoginTypeInUse(loginType.Name)
		if err != nil {
			return false, nil, err
		}
		if inUse {
			hasUsage = true
			usageMap[loginType.ID] = usingCardTypes
		}
	}

	return hasUsage, usageMap, nil
}

// LoginTypeUpdateHandler 更新登录方式
// - 接收JSON: {id, name, description, status}
func LoginTypeUpdateHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
	type reqBody struct {
		ID          uint   `json:"id"`
		Name        string `json:"name"`
		VerifyTypes string `json:"verify_types"`
		Status      int    `json:"status"`
	}
	var body reqBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		utils.JsonResponse(w, http.StatusBadRequest, false, "请求体错误", nil)
		return
	}
	if body.ID == 0 {
		utils.JsonResponse(w, http.StatusBadRequest, false, "缺少ID", nil)
		return
	}

	db, err := database.GetDB()
	if err != nil {
		utils.JsonResponse(w, http.StatusInternalServerError, false, "数据库连接失败", nil)
		return
	}

	// 始终查询原始记录，便于后续校验（重命名/禁用）
	var originalLoginType models.LoginType
	if err := db.First(&originalLoginType, body.ID).Error; err != nil {
		utils.JsonResponse(w, http.StatusBadRequest, false, "登录类型不存在", nil)
		return
	}

	// 如果名称发生变化，检查原名称是否被使用（与删除逻辑一致）
	if body.Name != "" && originalLoginType.Name != body.Name {
		inUse, usingCardTypes, err := checkLoginTypeInUse(originalLoginType.Name)
		if err != nil {
			utils.JsonResponse(w, http.StatusInternalServerError, false, "检查使用状态失败", nil)
			return
		}
		if inUse {
			utils.JsonResponse(w, http.StatusBadRequest, false, "该登录类型正在被以下卡密类型使用，无法修改名称："+strings.Join(usingCardTypes, "、"), nil)
			return
		}
	}

	// 当尝试禁用（status=0）时，如被卡密类型使用则禁止禁用
	if body.Status == 0 && originalLoginType.Status != 0 {
		inUse, usingCardTypes, err := checkLoginTypeInUse(originalLoginType.Name)
		if err != nil {
			utils.JsonResponse(w, http.StatusInternalServerError, false, "检查使用状态失败", nil)
			return
		}
		if inUse {
			utils.JsonResponse(w, http.StatusBadRequest, false, "该登录类型正在被以下卡密类型使用，无法禁用："+strings.Join(usingCardTypes, "、"), nil)
			return
		}
	}

	updates := map[string]interface{}{}
	if body.Name != "" {
		updates["name"] = body.Name
	}
	updates["status"] = body.Status
	updates["verify_types"] = body.VerifyTypes
	if err := db.Model(&models.LoginType{}).Where("id = ?", body.ID).Updates(updates).Error; err != nil {
		utils.JsonResponse(w, http.StatusBadRequest, false, "更新失败，可能是名称重复", nil)
		return
	}
	utils.JsonResponse(w, http.StatusOK, true, "更新成功", nil)
}

// LoginTypeDeleteHandler 删除单个登录方式
// - 接收JSON: {id}
func LoginTypeDeleteHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
	var body struct {
		ID uint `json:"id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.ID == 0 {
		utils.JsonResponse(w, http.StatusBadRequest, false, "参数错误", nil)
		return
	}

	db, err := database.GetDB()
	if err != nil {
		utils.JsonResponse(w, http.StatusInternalServerError, false, "数据库连接失败", nil)
		return
	}

	// 查询登录类型名称
	var loginType models.LoginType
	if dbErr := db.First(&loginType, body.ID).Error; dbErr != nil {
		utils.JsonResponse(w, http.StatusBadRequest, false, "登录类型不存在", nil)
		return
	}

	// 检查是否被卡密类型使用
	inUse, usingCardTypes, err := checkLoginTypeInUse(loginType.Name)
	if err != nil {
		utils.JsonResponse(w, http.StatusInternalServerError, false, "检查使用状态失败", nil)
		return
	}
	if inUse {
		utils.JsonResponse(w, http.StatusBadRequest, false, "该登录类型正在被以下卡密类型使用，无法删除："+strings.Join(usingCardTypes, "、"), nil)
		return
	}

	if err := db.Delete(&models.LoginType{}, body.ID).Error; err != nil {
		utils.JsonResponse(w, http.StatusInternalServerError, false, "删除失败", nil)
		return
	}
	utils.JsonResponse(w, http.StatusOK, true, "删除成功", nil)
}

// LoginTypesBatchDeleteHandler 批量删除登录方式
// - 接收JSON: {ids: []}
func LoginTypesBatchDeleteHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
	var body struct {
		IDs []uint `json:"ids"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || len(body.IDs) == 0 {
		utils.JsonResponse(w, http.StatusBadRequest, false, "参数错误", nil)
		return
	}

	// 检查批量删除的登录类型是否被使用
	hasUsage, usageMap, err := checkLoginTypesByIDsInUse(body.IDs)
	if err != nil {
		utils.JsonResponse(w, http.StatusInternalServerError, false, "检查使用状态失败", nil)
		return
	}
	if hasUsage {
		// 构建详细的错误信息
		var errorMessages []string
		db, _ := database.GetDB()
		for loginTypeID, usingCardTypes := range usageMap {
			var loginType models.LoginType
			if db.First(&loginType, loginTypeID).Error == nil {
				errorMessages = append(errorMessages, loginType.Name+"（被"+strings.Join(usingCardTypes, "、")+"使用）")
			}
		}
		utils.JsonResponse(w, http.StatusBadRequest, false, "以下登录类型正在被使用，无法删除："+strings.Join(errorMessages, "；"), nil)
		return
	}

	db, err := database.GetDB()
	if err != nil {
		utils.JsonResponse(w, http.StatusInternalServerError, false, "数据库连接失败", nil)
		return
	}
	if err := db.Delete(&models.LoginType{}, body.IDs).Error; err != nil {
		utils.JsonResponse(w, http.StatusInternalServerError, false, "批量删除失败", nil)
		return
	}
	utils.JsonResponse(w, http.StatusOK, true, "批量删除成功", nil)
}

// LoginTypesBatchEnableHandler 批量启用
// - 接收JSON: {ids: []}
func LoginTypesBatchEnableHandler(w http.ResponseWriter, r *http.Request) {
	batchUpdateLoginTypeStatus(w, r, 1)
}

// LoginTypesBatchDisableHandler 批量禁用
// - 接收JSON: {ids: []}
func LoginTypesBatchDisableHandler(w http.ResponseWriter, r *http.Request) {
	batchUpdateLoginTypeStatus(w, r, 0)
}

// batchUpdateLoginTypeStatus 批量更新登录方式状态的通用函数
// - status: 1 启用，0 禁用
func batchUpdateLoginTypeStatus(w http.ResponseWriter, r *http.Request, status int) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
	var body struct {
		IDs []uint `json:"ids"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || len(body.IDs) == 0 {
		utils.JsonResponse(w, http.StatusBadRequest, false, "参数错误", nil)
		return
	}
	db, err := database.GetDB()
	if err != nil {
		utils.JsonResponse(w, http.StatusInternalServerError, false, "数据库连接失败", nil)
		return
	}
	if err := db.Model(&models.LoginType{}).Where("id IN ?", body.IDs).Update("status", status).Error; err != nil {
		utils.JsonResponse(w, http.StatusInternalServerError, false, "批量更新失败", nil)
		return
	}
	utils.JsonResponse(w, http.StatusOK, true, "操作成功", nil)
}
