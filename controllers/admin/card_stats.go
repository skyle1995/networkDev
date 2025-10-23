package admin

import (
	"net/http"
	"networkDev/constants"
	"networkDev/database"
	"networkDev/models"
	"networkDev/utils"
	"time"
)

// CardStatsOverviewHandler 卡密统计概览API
// - 返回当日和所有卡密的统计信息
// - 包括：总数、使用/未使用/禁用状态分布
func CardStatsOverviewHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	db, err := database.GetDB()
	if err != nil {
		utils.JsonResponse(w, http.StatusInternalServerError, false, "数据库连接失败", nil)
		return
	}

	// 获取当日统计
	today := time.Now().Format("2006-01-02")
	todayStart := today + " 00:00:00"
	todayEnd := today + " 23:59:59"

	// 当日卡密统计
	var todayTotal int64
	var todayByStatus = make(map[int]int64)

	// 当日总数
	db.Model(&models.Card{}).Where("created_at >= ? AND created_at <= ?", todayStart, todayEnd).Count(&todayTotal)

	// 当日按状态分布
	var todayStatusCounts []struct {
		Status int   `json:"status"`
		Count  int64 `json:"count"`
	}
	db.Model(&models.Card{}).
		Select("status, count(*) as count").
		Where("created_at >= ? AND created_at <= ?", todayStart, todayEnd).
		Group("status").
		Find(&todayStatusCounts)

	for _, sc := range todayStatusCounts {
		todayByStatus[sc.Status] = sc.Count
	}

	// 所有卡密统计
	var allTotal int64
	var allByStatus = make(map[int]int64)

	// 总数
	db.Model(&models.Card{}).Count(&allTotal)

	// 按状态分布
	var allStatusCounts []struct {
		Status int   `json:"status"`
		Count  int64 `json:"count"`
	}
	db.Model(&models.Card{}).
		Select("status, count(*) as count").
		Group("status").
		Find(&allStatusCounts)

	for _, sc := range allStatusCounts {
		allByStatus[sc.Status] = sc.Count
	}

	// 构建响应数据
	data := map[string]interface{}{
		"today": map[string]interface{}{
			"total":     todayTotal,
			"by_status": todayByStatus,
		},
		"all": map[string]interface{}{
			"total":     allTotal,
			"by_status": allByStatus,
		},
	}

	utils.JsonResponse(w, http.StatusOK, true, "获取成功", data)
}

// CardStatsTrend30DaysHandler 卡密30天趋势API
// - 返回近30天的卡密创建和使用趋势
func CardStatsTrend30DaysHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	db, err := database.GetDB()
	if err != nil {
		utils.JsonResponse(w, http.StatusInternalServerError, false, "数据库连接失败", nil)
		return
	}

	// 生成近30天的日期列表
	var dates []string
	var totalCounts []int64
	var usedCounts []int64
	var unusedCounts []int64

	for i := 29; i >= 0; i-- {
		date := time.Now().AddDate(0, 0, -i).Format("2006-01-02")
		dates = append(dates, date)

		dayStart := date + " 00:00:00"
		dayEnd := date + " 23:59:59"

		// 当天创建的卡密总数
		var totalCount int64
		db.Model(&models.Card{}).Where("created_at >= ? AND created_at <= ?", dayStart, dayEnd).Count(&totalCount)
		totalCounts = append(totalCounts, totalCount)

		// 当天创建且已使用的卡密数
		var usedCount int64
		db.Model(&models.Card{}).
			Where("created_at >= ? AND created_at <= ? AND status = ?", dayStart, dayEnd, constants.CardStatusUsed).
			Count(&usedCount)
		usedCounts = append(usedCounts, usedCount)

		// 当天创建且未使用的卡密数
		var unusedCount int64
		db.Model(&models.Card{}).
			Where("created_at >= ? AND created_at <= ? AND status = ?", dayStart, dayEnd, constants.CardStatusUnused).
			Count(&unusedCount)
		unusedCounts = append(unusedCounts, unusedCount)
	}

	// 构建响应数据
	data := map[string]interface{}{
		"dates":  dates,
		"total":  totalCounts,
		"used":   usedCounts,
		"unused": unusedCounts,
	}

	utils.JsonResponse(w, http.StatusOK, true, "获取成功", data)
}

// CardStatsSimpleHandler 简单卡密统计API
// - 返回卡密的基本统计信息：总数、已使用、未使用、禁用
func CardStatsSimpleHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	db, err := database.GetDB()
	if err != nil {
		utils.JsonResponse(w, http.StatusInternalServerError, false, "数据库连接失败", nil)
		return
	}

	// 统计各状态的卡密数量
	var total int64
	var used int64
	var unused int64
	var disabled int64

	db.Model(&models.Card{}).Count(&total)
	db.Model(&models.Card{}).Where("status = ?", constants.CardStatusUsed).Count(&used)
	db.Model(&models.Card{}).Where("status = ?", constants.CardStatusUnused).Count(&unused)
	db.Model(&models.Card{}).Where("status = ?", constants.CardStatusDisabled).Count(&disabled)

	data := map[string]interface{}{
		"total":    total,
		"used":     used,
		"unused":   unused,
		"disabled": disabled,
	}

	utils.JsonResponse(w, http.StatusOK, true, "获取成功", data)
}