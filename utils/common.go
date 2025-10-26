package utils

import (
	"encoding/json"
	"net/http"
	"networkDev/web"
	"strings"
)

// JsonResponse 通用JSON响应函数
// 将 success 转换为 code：true -> 0, false -> 1，并输出 data
func JsonResponse(w http.ResponseWriter, status int, success bool, message string, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	// 将success转换为code格式：true -> 0, false -> 1
	code := 1
	if success {
		code = 0
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"code": code,
		"msg":  message,
		"data": data,
	})
}

// RenderTemplate 通用模板渲染函数
// templateName: 模板文件名
// data: 模板数据
// w: HTTP响应写入器
func RenderTemplate(w http.ResponseWriter, templateName string, data map[string]interface{}) error {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	tmpl, err := web.ParseTemplates()
	if err != nil {
		http.Error(w, "模板解析失败", http.StatusInternalServerError)
		return err
	}

	if err := tmpl.ExecuteTemplate(w, templateName, data); err != nil {
		http.Error(w, "模板渲染失败", http.StatusInternalServerError)
		return err
	}
	return nil
}

// GetDefaultTemplateData 获取默认模板数据
// 返回包含系统基础信息的数据映射
func GetDefaultTemplateData() map[string]interface{} {
	return map[string]interface{}{
		"SystemName": "网络验证系统",
		"FooterText": "© 2025 凌动技术 保留所有权利",
	}
}

// GetTemplateDataWithCSRF 获取包含CSRF令牌的模板数据
// 合并默认数据和CSRF令牌，用于需要CSRF保护的页面
func GetTemplateDataWithCSRF(r *http.Request, additionalData map[string]interface{}) map[string]interface{} {
	// 获取默认模板数据
	data := GetDefaultTemplateData()

	// 添加CSRF令牌
	data["CSRFToken"] = GetCSRFTokenForTemplate(r)

	// 合并额外数据
	for key, value := range additionalData {
		data[key] = value
	}

	return data
}

// GetClientIP 获取客户端IP地址
// 优先从 X-Forwarded-For 和 X-Real-IP 头部获取，否则使用 RemoteAddr
func GetClientIP(r *http.Request) string {
	// 检查 X-Forwarded-For 头部
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		// X-Forwarded-For 可能包含多个IP，取第一个
		if idx := strings.Index(xff, ","); idx != -1 {
			return strings.TrimSpace(xff[:idx])
		}
		return strings.TrimSpace(xff)
	}

	// 检查 X-Real-IP 头部
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return strings.TrimSpace(xri)
	}

	// 使用 RemoteAddr
	if idx := strings.LastIndex(r.RemoteAddr, ":"); idx != -1 {
		return r.RemoteAddr[:idx]
	}
	return r.RemoteAddr
}
