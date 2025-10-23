package admin

import (
	"crypto/rand"
	// 移除 CSV 导出，改为自定义分隔符文本导出
	// "encoding/csv"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"networkDev/database"
	"networkDev/models"
	"networkDev/utils"
	"strconv"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

// 生成指定长度的十六进制随机字符串
// 入参 n 表示需要的随机字符数（非字节数）；返回小写十六进制字符串
func genRandomHex(n int) string {
	if n <= 0 {
		return ""
	}
	// 由于 hex 每个字节会转成 2 个字符，因此需要 (n+1)/2 个字节
	byteLen := (n + 1) / 2
	b := make([]byte, byteLen)
	if _, err := rand.Read(b); err != nil {
		return ""
	}
	s := hex.EncodeToString(b)
	if len(s) > n {
		s = s[:n]
	}
	return s
}

// 根据前缀和总长度构建卡号
// - totalLen <= 0 时按 18 处理
// - 若前缀长度 >= totalLen，则自动扩展为 前缀长度+18
// - uppercase=true 表示最终结果转为大写；false 表示小写
func buildCardNumber(prefix string, totalLen int, uppercase bool) string {
	if totalLen <= 0 {
		totalLen = 18
	}
	if len(prefix) >= totalLen {
		totalLen = len(prefix) + 18
	}
	rest := totalLen - len(prefix)
	s := prefix + genRandomHex(rest)
	if uppercase {
		return strings.ToUpper(s)
	}
	return strings.ToLower(s)
}

// CardsFragmentHandler 卡密管理片段渲染
// - 渲染 cards.html 列表与表单界面
func CardsFragmentHandler(w http.ResponseWriter, r *http.Request) {
	utils.RenderTemplate(w, "cards.html", map[string]interface{}{})
}

// CardsListHandler 获取卡密列表
// - 支持GET
// - 支持分页查询参数：page、page_size
// - 支持筛选参数：card_type_id、status、batch、keyword(卡号/备注/批次模糊匹配)
// - 返回卡密列表和分页信息
func CardsListHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	// 获取查询参数
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))
	cardTypeIDStr := r.URL.Query().Get("card_type_id")
	statusStr := r.URL.Query().Get("status")
	batch := r.URL.Query().Get("batch")
	// 中文注释：keyword 支持在 card_number、remark、batch 三个字段上进行模糊匹配
	keyword := strings.TrimSpace(r.URL.Query().Get("keyword"))

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

	// 构建查询条件（去除无效的 Preload，前端已通过 card_type_id 自行映射类型名称）
	query := db.Model(&models.Card{})

	// 筛选条件
	if cardTypeIDStr != "" {
		if cardTypeID, err := strconv.Atoi(cardTypeIDStr); err == nil && cardTypeID > 0 {
			query = query.Where("card_type_id = ?", cardTypeID)
		}
	}
	if statusStr != "" {
		if status, err := strconv.Atoi(statusStr); err == nil {
			query = query.Where("status = ?", status)
		}
	}
	if batch != "" {
		query = query.Where("batch LIKE ?", "%"+batch+"%")
	}
	// 中文注释：当提供 keyword 时，在卡号、备注、批次三个字段上进行 OR 模糊匹配
	if keyword != "" {
		kw := "%" + keyword + "%"
		query = query.Where("(card_number LIKE ? OR remark LIKE ? OR batch LIKE ?)", kw, kw, kw)
	}

	// 计算总数
	var total int64
	if err := query.Count(&total).Error; err != nil {
		utils.JsonResponse(w, http.StatusInternalServerError, false, "统计总数失败", nil)
		return
	}

	// 分页查询
	var cards []models.Card
	offset := (page - 1) * pageSize
	if err := query.Order("id desc").Offset(offset).Limit(pageSize).Find(&cards).Error; err != nil {
		utils.JsonResponse(w, http.StatusInternalServerError, false, "查询失败", nil)
		return
	}

	// 中文注释：为每条卡记录补充类型名称，避免前端依赖异步类型映射导致显示“未知类型”
	// 1) 先查询类型列表并构建 id->name 的映射表
	var cardTypeList []models.CardType
	_ = db.Model(&models.CardType{}).Find(&cardTypeList).Error
	typeNameMap := make(map[uint]string, len(cardTypeList))
	for _, t := range cardTypeList {
		typeNameMap[t.ID] = t.Name
	}
	// 2) 将卡列表转换为通用 map 列表，并附加 card_type_name 字段
	items := make([]map[string]interface{}, 0, len(cards))
	for _, c := range cards {
		items = append(items, map[string]interface{}{
			"id":             c.ID,
			"card_number":    c.CardNumber,
			"card_type_id":   c.CardTypeID,
			"card_type_name": typeNameMap[c.CardTypeID],
			"status":         c.Status,
			"batch":          c.Batch,
			"remark":         c.Remark,
			"used_at":        c.UsedAt,
			"created_at":     c.CreatedAt,
		})
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

// CardCreateHandler 新增卡密
// - 接收JSON: {card_type_id, status, remark, prefix, length, uppercase, count}
// - card_number 与 batch 不再由前端传入，后端将自动生成：
//  1. 卡号：按 prefix 与 length 生成随机十六进制字符串，支持大小写控制（uppercase，默认小写）
//  2. 批次：基于设置表 card_batch_counter 自增，格式为 YYYYMMDD-000001
//  3. 生成数量：通过 count 控制一次生成的数量，默认1，最大500
func CardCreateHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
	type reqBody struct {
		CardTypeID uint   `json:"card_type_id"`
		Status     int    `json:"status"`
		Remark     string `json:"remark"`
		Prefix     string `json:"prefix"`
		Length     int    `json:"length"`
		Uppercase  bool   `json:"uppercase"`
		Count      int    `json:"count"`
	}
	var body reqBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		utils.JsonResponse(w, http.StatusBadRequest, false, "请求体错误", nil)
		return
	}
	if body.CardTypeID == 0 {
		utils.JsonResponse(w, http.StatusBadRequest, false, "卡密类型ID不能为空", nil)
		return
	}

	db, err := database.GetDB()
	if err != nil {
		utils.JsonResponse(w, http.StatusInternalServerError, false, "数据库连接失败", nil)
		return
	}

	// 检查卡密类型是否存在且启用
	var cardType models.CardType
	if err := db.First(&cardType, body.CardTypeID).Error; err != nil {
		utils.JsonResponse(w, http.StatusBadRequest, false, "卡密类型不存在", nil)
		return
	}
	// 检查卡密类型是否被禁用
	if cardType.Status != 1 {
		utils.JsonResponse(w, http.StatusBadRequest, false, "卡密类型已被禁用，无法创建卡密", nil)
		return
	}

	// 规范化长度与大小写、生成数量参数
	if body.Length <= 0 {
		body.Length = 18
	}
	if body.Count <= 0 {
		body.Count = 1
	}
	if body.Count > 500 {
		body.Count = 500
	}

	// 生成批次（基于设置表 card_batch_counter 自增）
	// 格式：YYYYMMDD-000001（每天不重置，仅简单自增计数）
	var batch string
	var setting models.Settings
	if err := db.Where("name = ?", "card_batch_counter").First(&setting).Error; err != nil {
		// 若不存在该设置项，则创建并从 1 开始
		setting = models.Settings{Name: "card_batch_counter", Value: "1", Description: "卡密批次号计数器（用于记录上次生成批次号的序号，自增使用）"}
		if e := db.Create(&setting).Error; e != nil {
			utils.JsonResponse(w, http.StatusInternalServerError, false, "初始化批次计数器失败", nil)
			return
		}
		batch = time.Now().Format("20060102") + "-" + fmt.Sprintf("%06d", 1)
	} else {
		cnt, _ := strconv.Atoi(setting.Value)
		cnt++
		newVal := strconv.Itoa(cnt)
		if e := db.Model(&models.Settings{}).Where("id = ?", setting.ID).Update("value", newVal).Error; e != nil {
			utils.JsonResponse(w, http.StatusInternalServerError, false, "更新批次计数器失败", nil)
			return
		}
		batch = time.Now().Format("20060102") + "-" + fmt.Sprintf("%06d", cnt)
	}

	// 中文注释：计算合法状态值（1=已使用，2=禁用，其它按未使用0处理）
	safeStatus := body.Status
	if safeStatus != 1 && safeStatus != 2 {
		safeStatus = 0
	}

	// 中文注释：循环生成 count 条卡密，若单条创建失败则重试最多5次
	success := 0
	for i := 0; i < body.Count; i++ {
		card := models.Card{
			CardNumber: buildCardNumber(body.Prefix, body.Length, body.Uppercase),
			CardTypeID: body.CardTypeID,
			Status:     safeStatus,
			Batch:      batch,
			Remark:     body.Remark,
		}
		var createErr error
		for j := 0; j < 5; j++ {
			createErr = db.Create(&card).Error
			if createErr == nil {
				success++
				break
			}
			// 失败则重新生成一次卡号后重试
			card.CardNumber = buildCardNumber(body.Prefix, body.Length, body.Uppercase)
		}
	}

	if success == 0 {
		utils.JsonResponse(w, http.StatusBadRequest, false, "创建失败，可能是卡密号码重复", nil)
		return
	}

	result := map[string]interface{}{
		"created": success,
		"batch":   batch,
	}
	utils.JsonResponse(w, http.StatusOK, true, fmt.Sprintf("创建成功：%d条", success), result)
}

// CardUpdateHandler 更新卡密
// - 接收JSON: {id, card_number(可选), card_type_id(可选), status, batch(可选), remark}
// - 说明：card_number 与 batch 若未提供或为空，则不会更新对应字段
func CardUpdateHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
	type reqBody struct {
		ID         uint   `json:"id"`
		CardNumber string `json:"card_number"`
		CardTypeID uint   `json:"card_type_id"`
		Status     int    `json:"status"`
		Batch      string `json:"batch"`
		Remark     string `json:"remark"`
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

	// 检查卡密类型是否存在且启用（如果提供了新的卡密类型ID）
	if body.CardTypeID > 0 {
		var cardType models.CardType
		if err := db.First(&cardType, body.CardTypeID).Error; err != nil {
			utils.JsonResponse(w, http.StatusBadRequest, false, "卡密类型不存在", nil)
			return
		}
		// 检查卡密类型是否被禁用
		if cardType.Status != 1 {
			utils.JsonResponse(w, http.StatusBadRequest, false, "卡密类型已被禁用，无法更新为此类型", nil)
			return
		}
	}

	// 中文注释：若尝试将状态置为未使用(0)，则直接允许
	if body.Status == 0 {
		var existing models.Card
		if err := db.First(&existing, body.ID).Error; err != nil {
			utils.JsonResponse(w, http.StatusBadRequest, false, "卡密不存在", nil)
			return
		}
	}

	// 构建更新字段
	updates := map[string]interface{}{}
	if body.CardNumber != "" {
		updates["card_number"] = body.CardNumber
	}
	if body.CardTypeID > 0 {
		updates["card_type_id"] = body.CardTypeID
	}
	updates["status"] = body.Status
	// 仅当提供非空 batch 时才更新，防止被清空
	if strings.TrimSpace(body.Batch) != "" {
		updates["batch"] = body.Batch
	}
	updates["remark"] = body.Remark

	if err := db.Model(&models.Card{}).Where("id = ?", body.ID).Updates(updates).Error; err != nil {
		utils.JsonResponse(w, http.StatusBadRequest, false, "更新失败，可能是卡密号码重复", nil)
		return
	}
	utils.JsonResponse(w, http.StatusOK, true, "更新成功", nil)
}

// CardDeleteHandler 删除单个卡密
// - 接收JSON: {id}
func CardDeleteHandler(w http.ResponseWriter, r *http.Request) {
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
	if err := db.Delete(&models.Card{}, body.ID).Error; err != nil {
		utils.JsonResponse(w, http.StatusInternalServerError, false, "删除失败", nil)
		return
	}
	utils.JsonResponse(w, http.StatusOK, true, "删除成功", nil)
}

// CardsBatchDeleteHandler 批量删除卡密
// - 接收JSON: {ids: []}
func CardsBatchDeleteHandler(w http.ResponseWriter, r *http.Request) {
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
	if err := db.Delete(&models.Card{}, body.IDs).Error; err != nil {
		utils.JsonResponse(w, http.StatusInternalServerError, false, "批量删除失败", nil)
		return
	}
	utils.JsonResponse(w, http.StatusOK, true, "批量删除成功", nil)
}

// CardsBatchUpdateStatusHandler 批量更新卡密状态
// - 接收JSON: {ids: [], status: int}
// - status: 0=未使用，1=已使用，2=禁用
func CardsBatchUpdateStatusHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
	var body struct {
		IDs    []uint `json:"ids"`
		Status int    `json:"status"`
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
	// 中文注释：允许批量重置为未使用(0)
	if err := db.Model(&models.Card{}).Where("id IN ?", body.IDs).Update("status", body.Status).Error; err != nil {
		utils.JsonResponse(w, http.StatusInternalServerError, false, "批量更新失败", nil)
		return
	}
	utils.JsonResponse(w, http.StatusOK, true, "操作成功", nil)
}

// GetCardTypesHandler 获取卡密类型列表（供前端下拉选择）
// - 仅支持GET请求
// - 只返回启用状态的卡密类型，用于前端下拉选择
func GetCardTypesHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
	db, err := database.GetDB()
	if err != nil {
		utils.JsonResponse(w, http.StatusInternalServerError, false, "数据库连接失败", nil)
		return
	}
	var cardTypes []models.CardType
	// 中文注释：根据可选参数 all 决定是否仅返回启用类型
	// - 未提供或为其它值：仅返回启用（status=1）
	// - all=1/true/yes：返回所有状态的类型（用于筛选下拉场景）
	q := db.Model(&models.CardType{})
	all := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("all")))
	if !(all == "1" || all == "true" || all == "yes") {
		q = q.Where("status = ?", 1)
	}
	if err := q.Order("id asc").Find(&cardTypes).Error; err != nil {
		utils.JsonResponse(w, http.StatusInternalServerError, false, "查询失败", nil)
		return
	}
	utils.JsonResponse(w, http.StatusOK, true, "ok", cardTypes)
}

// CardsExportHandler 导出卡密为文本文件
// - 支持GET
// - 筛选参数：card_type_id、status、batch、remark
// - 导出字段（按顺序）：卡号、状态、创建时间；使用“----”分隔
func CardsExportHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	// 解析筛选参数
	cardTypeIDStr := strings.TrimSpace(r.URL.Query().Get("card_type_id"))
	statusStr := strings.TrimSpace(r.URL.Query().Get("status"))
	batch := strings.TrimSpace(r.URL.Query().Get("batch"))
	remark := strings.TrimSpace(r.URL.Query().Get("remark"))

	db, err := database.GetDB()
	if err != nil {
		http.Error(w, "数据库连接失败", http.StatusInternalServerError)
		return
	}

	// 构建查询
	query := db.Model(&models.Card{})
	if cardTypeIDStr != "" {
		if id, err := strconv.Atoi(cardTypeIDStr); err == nil && id > 0 {
			query = query.Where("card_type_id = ?", id)
		}
	}
	if statusStr != "" {
		if s, err := strconv.Atoi(statusStr); err == nil {
			query = query.Where("status = ?", s)
		}
	}
	if batch != "" {
		query = query.Where("batch LIKE ?", "%"+batch+"%")
	}
	if remark != "" {
		query = query.Where("remark LIKE ?", "%"+remark+"%")
	}

	// 查询数据（按ID倒序）
	var cards []models.Card
	if err := query.Order("id desc").Find(&cards).Error; err != nil {
		http.Error(w, "查询失败", http.StatusInternalServerError)
		return
	}

	// 设置响应头（文本下载）
	now := time.Now().Format("20060102150405")
	filename := fmt.Sprintf("cards_%s.txt", now)
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))

	// 写入UTF-8 BOM，避免Excel/记事本中文乱码
	_, _ = w.Write([]byte{0xEF, 0xBB, 0xBF})

	// 写入表头
	_, _ = w.Write([]byte("卡号----状态----创建时间\n"))

	// 时间格式
	const tf = "2006-01-02 15:04:05"

	// 状态转文字
	statusText := func(s int) string {
		switch s {
		case 0:
			return "未使用"
		case 1:
			return "已使用"
		default:
			return "禁用"
		}
	}

	// 写入数据行（以“----”分隔）
	for _, c := range cards {
		record := []string{
			c.CardNumber,
			statusText(c.Status),
			c.CreatedAt.Format(tf),
		}
		line := strings.Join(record, "----") + "\n"
		if _, err := w.Write([]byte(line)); err != nil {
			continue
		}
	}
}

// CardsExportSelectedHandler 导出选中的卡密为文本文件
// - 支持GET
// - 参数：ids（逗号分隔的卡密ID列表）
// - 导出字段（按顺序）：卡号、状态、创建时间；使用"----"分隔
func CardsExportSelectedHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	// 解析选中的卡密ID列表
	idsStr := strings.TrimSpace(r.URL.Query().Get("ids"))
	if idsStr == "" {
		http.Error(w, "请提供要导出的卡密ID列表", http.StatusBadRequest)
		return
	}

	// 解析ID列表
	idStrings := strings.Split(idsStr, ",")
	var ids []uint
	for _, idStr := range idStrings {
		if id, err := strconv.Atoi(strings.TrimSpace(idStr)); err == nil && id > 0 {
			ids = append(ids, uint(id))
		}
	}

	if len(ids) == 0 {
		http.Error(w, "无效的卡密ID列表", http.StatusBadRequest)
		return
	}

	db, err := database.GetDB()
	if err != nil {
		http.Error(w, "数据库连接失败", http.StatusInternalServerError)
		return
	}

	// 查询选中的卡密数据（按ID倒序）
	var cards []models.Card
	if err := db.Where("id IN ?", ids).Order("id desc").Find(&cards).Error; err != nil {
		logrus.WithError(err).Error("查询选中卡密失败")
		http.Error(w, "查询卡密数据失败", http.StatusInternalServerError)
		return
	}

	if len(cards) == 0 {
		http.Error(w, "未找到指定的卡密数据", http.StatusNotFound)
		return
	}

	// 设置响应头，触发下载
	filename := fmt.Sprintf("selected_cards_%s.txt", time.Now().Format("20060102_150405"))
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))

	// 写入数据
	tf := "2006-01-02 15:04:05"
	for _, c := range cards {
		// 状态转换
		var statusText string
		switch c.Status {
		case 0:
			statusText = "未使用"
		case 1:
			statusText = "已使用"
		default:
			statusText = "禁用"
		}

		// 格式：卡号----状态----创建时间
		record := []string{
			c.CardNumber,
			statusText,
			c.CreatedAt.Format(tf),
		}
		line := strings.Join(record, "----") + "\n"
		if _, err := w.Write([]byte(line)); err != nil {
			continue
		}
	}
}
