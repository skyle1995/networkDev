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

// CardTypesFragmentHandler 卡密类型管理片段渲染
// - 渲染 card_types.html 列表与表单界面
func CardTypesFragmentHandler(w http.ResponseWriter, r *http.Request) {
	utils.RenderTemplate(w, "card_types.html", map[string]interface{}{})
}

// CardTypesListHandler 获取卡密类型列表
// - 支持GET
// - 支持分页和筛选
func CardTypesListHandler(w http.ResponseWriter, r *http.Request) {
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
	query := db.Model(&models.CardType{})

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
	var items []models.CardType
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

// CardTypeCreateHandler 新增卡密类型
// - 接收JSON: {name, status, login_types}
// - Name 必填且唯一
func CardTypeCreateHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
	type reqBody struct {
		Name       string `json:"name"`
		Status     int    `json:"status"`
		LoginTypes string `json:"login_types"`
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

	// 校验登录方式ID是否存在
	if errMsg := validateLoginTypes(body.LoginTypes); errMsg != "" {
		utils.JsonResponse(w, http.StatusBadRequest, false, errMsg, nil)
		return
	}
	db, err := database.GetDB()
	if err != nil {
		utils.JsonResponse(w, http.StatusInternalServerError, false, "数据库连接失败", nil)
		return
	}
	item := models.CardType{
		Name:       body.Name,
		Status:     body.Status,
		LoginTypes: body.LoginTypes,
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

// checkCardTypeInUse 检查卡密类型是否被卡密使用
// - 通过 cards 表中 card_type_id 外键计数
// - 返回是否被使用以及被使用的数量
func checkCardTypeInUse(cardTypeID uint) (bool, int64, error) {
	db, err := database.GetDB()
	if err != nil {
		return false, 0, err
	}
	var count int64
	if err := db.Model(&models.Card{}).Where("card_type_id = ?", cardTypeID).Count(&count).Error; err != nil {
		return false, 0, err
	}
	return count > 0, count, nil
}

// CardTypeUpdateHandler 更新卡密类型
// - 接收JSON: {id, name, status, login_types}
func CardTypeUpdateHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
	type reqBody struct {
		ID         uint   `json:"id"`
		Name       string `json:"name"`
		Status     int    `json:"status"`
		LoginTypes string `json:"login_types"`
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

	// 校验登录方式名称是否存在且未被禁用
	if errMsg := validateLoginTypes(body.LoginTypes); errMsg != "" {
		utils.JsonResponse(w, http.StatusBadRequest, false, errMsg, nil)
		return
	}
	db, err := database.GetDB()
	if err != nil {
		utils.JsonResponse(w, http.StatusInternalServerError, false, "数据库连接失败", nil)
		return
	}

	// 查询原始记录，便于后续在用校验（重命名/禁用）
	var original models.CardType
	if err := db.First(&original, body.ID).Error; err != nil {
		utils.JsonResponse(w, http.StatusBadRequest, false, "卡密类型不存在", nil)
		return
	}

	// 如果名称发生变化且该卡密类型已被卡密使用，则不允许修改名称
	if body.Name != "" && body.Name != original.Name {
		inUse, count, err := checkCardTypeInUse(body.ID)
		if err != nil {
			utils.JsonResponse(w, http.StatusInternalServerError, false, "检查使用状态失败", nil)
			return
		}
		if inUse {
			utils.JsonResponse(w, http.StatusBadRequest, false, "该卡密类型已被卡密使用（数量："+strconv.FormatInt(count, 10)+"），无法修改名称", nil)
			return
		}
	}

	// 当尝试禁用（status=0）且原状态不是禁用时，如该类型已被卡密使用则禁止禁用
	if body.Status == 0 && original.Status != 0 {
		inUse, count, err := checkCardTypeInUse(body.ID)
		if err != nil {
			utils.JsonResponse(w, http.StatusInternalServerError, false, "检查使用状态失败", nil)
			return
		}
		if inUse {
			utils.JsonResponse(w, http.StatusBadRequest, false, "该卡密类型已被卡密使用（数量："+strconv.FormatInt(count, 10)+"），无法禁用", nil)
			return
		}
	}

	// 构建更新字段
	updates := map[string]interface{}{}
	if body.Name != "" {
		updates["name"] = body.Name
	}
	updates["status"] = body.Status
	updates["login_types"] = body.LoginTypes
	if err := db.Model(&models.CardType{}).Where("id = ?", body.ID).Updates(updates).Error; err != nil {
		utils.JsonResponse(w, http.StatusBadRequest, false, "更新失败，可能是名称重复", nil)
		return
	}
	utils.JsonResponse(w, http.StatusOK, true, "更新成功", nil)
}

// CardTypeDeleteHandler 删除单个卡密类型
// - 接收JSON: {id}
func CardTypeDeleteHandler(w http.ResponseWriter, r *http.Request) {
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

	// 在用校验
	inUse, count, err := checkCardTypeInUse(body.ID)
	if err != nil {
		utils.JsonResponse(w, http.StatusInternalServerError, false, "检查使用状态失败", nil)
		return
	}
	if inUse {
		utils.JsonResponse(w, http.StatusBadRequest, false, "该卡密类型已被卡密使用（数量："+strconv.FormatInt(count, 10)+"），无法删除", nil)
		return
	}

	if err := db.Delete(&models.CardType{}, body.ID).Error; err != nil {
		utils.JsonResponse(w, http.StatusInternalServerError, false, "删除失败", nil)
		return
	}
	utils.JsonResponse(w, http.StatusOK, true, "删除成功", nil)
}

// CardTypesBatchDeleteHandler 批量删除卡密类型
// - 接收JSON: {ids: []}
func CardTypesBatchDeleteHandler(w http.ResponseWriter, r *http.Request) {
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

	// 批量在用校验
	var blocking []string
	for _, id := range body.IDs {
		inUse, count, err := checkCardTypeInUse(id)
		if err != nil {
			utils.JsonResponse(w, http.StatusInternalServerError, false, "检查使用状态失败", nil)
			return
		}
		if inUse {
			var ct models.CardType
			if db.First(&ct, id).Error == nil {
				blocking = append(blocking, ct.Name+"（数量："+strconv.FormatInt(count, 10)+"）")
			} else {
				blocking = append(blocking, strconv.FormatUint(uint64(id), 10)+"（数量："+strconv.FormatInt(count, 10)+"）")
			}
		}
	}
	if len(blocking) > 0 {
		utils.JsonResponse(w, http.StatusBadRequest, false, "以下卡密类型已被卡密使用，无法删除："+strings.Join(blocking, "；"), nil)
		return
	}

	if err := db.Delete(&models.CardType{}, body.IDs).Error; err != nil {
		utils.JsonResponse(w, http.StatusInternalServerError, false, "批量删除失败", nil)
		return
	}
	utils.JsonResponse(w, http.StatusOK, true, "批量删除成功", nil)
}

// validateLoginTypes 校验登录方式名称是否存在且未被禁用
// - 接收逗号分隔的登录方式名称字符串
// - 检查登录方式是否存在且状态为启用(status=1)
// - 返回错误信息，如果所有名称都存在且启用则返回空字符串
func validateLoginTypes(loginTypesStr string) string {
	if loginTypesStr == "" {
		return ""
	}

	// 分割登录方式名称字符串
	nameStrs := strings.Split(loginTypesStr, ",")
	var names []string

	// 去重并清理空格
	nameSet := make(map[string]bool)
	for _, nameStr := range nameStrs {
		nameStr = strings.TrimSpace(nameStr)
		if nameStr == "" {
			continue
		}
		nameSet[nameStr] = true
	}

	// 转换为切片
	for name := range nameSet {
		names = append(names, name)
	}

	if len(names) == 0 {
		return ""
	}

	// 查询数据库检查名称是否存在
	db, err := database.GetDB()
	if err != nil {
		return "数据库连接失败"
	}

	// 查询所有匹配的登录方式，包括状态信息
	var loginTypes []models.LoginType
	if err := db.Where("name IN ?", names).Find(&loginTypes).Error; err != nil {
		return "查询登录方式失败"
	}

	// 检查是否有不存在的名称和被禁用的登录方式
	existingSet := make(map[string]bool)
	disabledNames := []string{}
	for _, loginType := range loginTypes {
		existingSet[loginType.Name] = true
		// 检查登录方式是否被禁用 (status != 1 表示禁用)
		if loginType.Status != 1 {
			disabledNames = append(disabledNames, loginType.Name)
		}
	}

	// 检查不存在的名称
	var notFoundNames []string
	for _, name := range names {
		if !existingSet[name] {
			notFoundNames = append(notFoundNames, name)
		}
	}

	// 返回错误信息
	if len(notFoundNames) > 0 {
		return "以下登录方式名称不存在: " + strings.Join(notFoundNames, ", ")
	}
	if len(disabledNames) > 0 {
		return "以下登录方式已被禁用，无法使用: " + strings.Join(disabledNames, ", ")
	}

	return ""
}

// CardTypesBatchEnableHandler 批量启用
// - 接收JSON: {ids: []}
func CardTypesBatchEnableHandler(w http.ResponseWriter, r *http.Request) {
	batchUpdateStatus(w, r, 1)
}

// CardTypesBatchDisableHandler 批量禁用
// - 接收JSON: {ids: []}
func CardTypesBatchDisableHandler(w http.ResponseWriter, r *http.Request) {
	batchUpdateStatus(w, r, 0)
}

// batchUpdateStatus 批量更新状态的通用函数
// - status: 1 启用，0 禁用
func batchUpdateStatus(w http.ResponseWriter, r *http.Request, status int) {
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
	if err := db.Model(&models.CardType{}).Where("id IN ?", body.IDs).Update("status", status).Error; err != nil {
		utils.JsonResponse(w, http.StatusInternalServerError, false, "批量更新失败", nil)
		return
	}
	utils.JsonResponse(w, http.StatusOK, true, "操作成功", nil)
}
