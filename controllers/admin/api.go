package admin

import (
	"encoding/json"
	"net/http"
	"networkDev/database"
	"networkDev/models"
	"networkDev/utils"
	"networkDev/utils/encrypt"
	"strconv"
	"strings"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/hex"
	"encoding/pem"

	"github.com/sirupsen/logrus"
)

// APIFragmentHandler 接口列表页面片段处理器
func APIFragmentHandler(w http.ResponseWriter, r *http.Request) {
	utils.RenderTemplate(w, "apis.html", map[string]interface{}{
		"Title": "接口管理",
	})
}

// APIListHandler 接口列表API处理器
func APIListHandler(w http.ResponseWriter, r *http.Request) {
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

	// 获取应用UUID参数（用于按应用筛选接口）
	appUUID := strings.TrimSpace(r.URL.Query().Get("app_uuid"))

	// 获取搜索参数
	search := strings.TrimSpace(r.URL.Query().Get("search"))

	// 构建查询
	db, err := database.GetDB()
	if err != nil {
		logrus.WithError(err).Error("Failed to get database connection")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// 构建基础查询
	query := db.Model(&models.API{})

	// 如果指定了应用UUID，则按应用筛选
	if appUUID != "" {
		query = query.Where("app_uuid = ?", appUUID)
	}

	// 如果有搜索条件，添加搜索
	if search != "" {
		query = query.Where("api_key LIKE ? OR app_uuid LIKE ?", "%"+search+"%", "%"+search+"%")
	}

	// 获取总数
	var total int64
	if err := query.Count(&total).Error; err != nil {
		logrus.WithError(err).Error("Failed to count APIs")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// 获取分页数据
	var apis []models.API
	offset := (page - 1) * limit
	if err := query.Offset(offset).Limit(limit).Order("created_at DESC").Find(&apis).Error; err != nil {
		logrus.WithError(err).Error("Failed to fetch APIs")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// 获取关联的应用信息
	var appUUIDs []string
	for _, api := range apis {
		appUUIDs = append(appUUIDs, api.AppUUID)
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
	type APIResponse struct {
		models.API
		AppName        string `json:"app_name"`
		APITypeName    string `json:"api_type_name"`
		StatusName     string `json:"status_name"`
		AlgorithmNames struct {
			Submit string `json:"submit"`
			Return string `json:"return"`
		} `json:"algorithm_names"`
	}

	var responseAPIs []APIResponse
	for _, api := range apis {
		responseAPI := APIResponse{
			API:         api,
			AppName:     appMap[api.AppUUID],
			APITypeName: models.GetAPITypeName(api.APIType),
			StatusName:  getAPIStatusName(api.Status),
		}
		responseAPI.AlgorithmNames.Submit = models.GetAlgorithmName(api.SubmitAlgorithm)
		responseAPI.AlgorithmNames.Return = models.GetAlgorithmName(api.ReturnAlgorithm)
		responseAPIs = append(responseAPIs, responseAPI)
	}

	// 计算分页信息
	totalPages := (total + int64(limit) - 1) / int64(limit)

	response := map[string]interface{}{
		"success": true,
		"data": map[string]interface{}{
			"apis":        responseAPIs,
			"total":       total,
			"page":        page,
			"limit":       limit,
			"total_pages": totalPages,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// getAPIStatusName 获取API状态名称
func getAPIStatusName(status int) string {
	switch status {
	case 1:
		return "启用"
	case 0:
		return "禁用"
	default:
		return "未知"
	}
}

// APIUpdateHandler 更新接口处理器
func APIUpdateHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		ID               uint   `json:"id"`
		Status           int    `json:"status"`
		SubmitAlgorithm  int    `json:"submit_algorithm"`
		ReturnAlgorithm  int    `json:"return_algorithm"`
		SubmitPublicKey  string `json:"submit_public_key"`
		SubmitPrivateKey string `json:"submit_private_key"`
		ReturnPublicKey  string `json:"return_public_key"`
		ReturnPrivateKey string `json:"return_private_key"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// 验证必填字段
	if req.ID == 0 {
		http.Error(w, "接口ID不能为空", http.StatusBadRequest)
		return
	}

	if req.Status != 0 && req.Status != 1 {
		http.Error(w, "无效的状态值", http.StatusBadRequest)
		return
	}

	if !models.IsValidAlgorithm(req.SubmitAlgorithm) || !models.IsValidAlgorithm(req.ReturnAlgorithm) {
		http.Error(w, "无效的算法类型", http.StatusBadRequest)
		return
	}

	// 获取数据库连接
	db, err := database.GetDB()
	if err != nil {
		logrus.WithError(err).Error("Failed to get database connection")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// 查找并更新API记录
	var api models.API
	if err := db.First(&api, req.ID).Error; err != nil {
		http.Error(w, "接口不存在", http.StatusNotFound)
		return
	}

	// 更新字段（不允许修改 APIType）
	api.Status = req.Status
	api.SubmitAlgorithm = req.SubmitAlgorithm
	api.ReturnAlgorithm = req.ReturnAlgorithm

	// 可选更新密钥/证书（当提供时）
	if req.SubmitPublicKey != "" || req.SubmitPrivateKey != "" {
		api.SubmitPublicKey = req.SubmitPublicKey
		api.SubmitPrivateKey = req.SubmitPrivateKey
	}
	if req.ReturnPublicKey != "" || req.ReturnPrivateKey != "" {
		api.ReturnPublicKey = req.ReturnPublicKey
		api.ReturnPrivateKey = req.ReturnPrivateKey
	}

	if err := db.Save(&api).Error; err != nil {
		logrus.WithError(err).Error("Failed to update API")
		http.Error(w, "更新接口失败", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"success": true,
		"message": "接口更新成功",
		"data":    api,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// APIGetAppsHandler 获取应用列表（用于接口页面的应用选择器）
func APIGetAppsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// 获取数据库连接
	db, err := database.GetDB()
	if err != nil {
		logrus.WithError(err).Error("Failed to get database connection")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// 获取所有应用
	var apps []models.App
	if err := db.Select("uuid, name").Order("created_at ASC").Find(&apps).Error; err != nil {
		logrus.WithError(err).Error("Failed to fetch apps")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"success": true,
		"data":    apps,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func APIGenerateKeysHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Side      string `json:"side"`           // submit | return
		Algorithm int    `json:"algorithm"`      // 与 models.Algorithm* 对应
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	if req.Side != "submit" && req.Side != "return" {
		http.Error(w, "side参数必须为submit或return", http.StatusBadRequest)
		return
	}
	if !models.IsValidAlgorithm(req.Algorithm) {
		http.Error(w, "无效的算法类型", http.StatusBadRequest)
		return
	}

	// 根据算法生成密钥/证书
	result := map[string]interface{}{}

	switch req.Algorithm {
	case models.AlgorithmNone:
		// 不加密不生成任何密钥
		result["public_key"] = ""
		result["private_key"] = ""
	case models.AlgorithmRC4:
		// 生成16字节随机密钥并返回16位十六进制（大写）
		bytes := make([]byte, 8)
		if _, err := rand.Read(bytes); err != nil {
			logrus.WithError(err).Error("Failed to generate RC4 key")
			http.Error(w, "生成RC4密钥失败", http.StatusInternalServerError)
			return
		}
		result["public_key"] = ""
		result["private_key"] = strings.ToUpper(hex.EncodeToString(bytes))
	case models.AlgorithmRSA, models.AlgorithmRSADynamic:
		// 生成RSA 2048密钥对，返回PEM明文字符串
		key, err := rsa.GenerateKey(rand.Reader, 2048)
		if err != nil {
			logrus.WithError(err).Error("Failed to generate RSA key pair")
			http.Error(w, "生成RSA密钥失败", http.StatusInternalServerError)
			return
		}
		privBytes := x509.MarshalPKCS1PrivateKey(key)
		pubBytes := x509.MarshalPKCS1PublicKey(&key.PublicKey)
		privPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: privBytes})
		pubPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PUBLIC KEY", Bytes: pubBytes})
		result["private_key"] = string(privPEM)
		result["public_key"] = string(pubPEM)
	case models.AlgorithmEasy:
		// 生成易加密密钥对，返回逗号分隔的整数数组字符串
		encryptKey, _, err := encrypt.GenerateEasyKey()
		if err != nil {
			logrus.WithError(err).Error("Failed to generate Easy encryption key")
			http.Error(w, "生成易加密密钥失败", http.StatusInternalServerError)
			return
		}
		result["public_key"] = ""
		result["private_key"] = encrypt.FormatKeyAsString(encryptKey)
	default:
		http.Error(w, "不支持的算法类型", http.StatusBadRequest)
		return
	}

	response := map[string]interface{}{
		"success": true,
		"message": "生成成功",
		"data":    result,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func APIResetKeyHandler(w http.ResponseWriter, r *http.Request) {
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
        http.Error(w, "接口ID不能为空", http.StatusBadRequest)
        return
    }

    db, err := database.GetDB()
    if err != nil {
        logrus.WithError(err).Error("Failed to get database connection")
        http.Error(w, "Internal server error", http.StatusInternalServerError)
        return
    }

    var api models.API
    if err := db.First(&api, req.ID).Error; err != nil {
        http.Error(w, "接口不存在", http.StatusNotFound)
        return
    }

    // 生成新的16位大写十六进制密钥
    bytes := make([]byte, 8)
    if _, err := rand.Read(bytes); err != nil {
        logrus.WithError(err).Error("Failed to generate random API key")
        http.Error(w, "生成密钥失败", http.StatusInternalServerError)
        return
    }
    newKey := strings.ToUpper(hex.EncodeToString(bytes))

    if err := db.Model(&api).Update("api_key", newKey).Error; err != nil {
        logrus.WithError(err).Error("Failed to update API key")
        http.Error(w, "更新密钥失败", http.StatusInternalServerError)
        return
    }

    response := map[string]interface{}{
        "success": true,
        "message": "接口密钥重置成功",
        "data": map[string]interface{}{
            "id":      api.ID,
            "api_key": newKey,
        },
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(response)
}
