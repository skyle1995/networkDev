package admin

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"net/http"
	"networkDev/controllers"
	"networkDev/models"
	"networkDev/services"
	"networkDev/utils"
)

// 创建基础控制器实例
var settingsBaseController = controllers.NewBaseController()

// SettingsFragmentHandler 设置片段渲染
// - 渲染设置表单（通过前端JS调用API加载/保存）
func SettingsFragmentHandler(c *gin.Context) {
	c.HTML(http.StatusOK, "settings.html", gin.H{})
}

// SettingsQueryHandler 设置查询API
// - 返回所有设置项的 name:value 映射
func SettingsQueryHandler(c *gin.Context) {
	db, ok := settingsBaseController.GetDB(c)
	if !ok {
		return
	}
	var list []models.Settings
	if err := db.Find(&list).Error; err != nil {
		settingsBaseController.HandleInternalError(c, "查询失败", err)
		return
	}
	res := map[string]string{}
	for _, s := range list {
		res[s.Name] = s.Value
	}
	settingsBaseController.HandleSuccess(c, "ok", res)
}

// SettingsUpdateHandler 更新系统设置处理器
// - 接收JSON格式的设置数据，支持两种格式：
//  1. 直接字段格式: {"site_title": "值", "site_keywords": "值"}
//  2. 嵌套格式: {"settings": {"site_title": "值", "site_keywords": "值"}}
//
// - 自动创建不存在的设置项
// - 更新已存在的设置项
// - 更新完成后：
//  1. 删除对应的Redis缓存键，确保后续读取走数据库并重建缓存
//  2. 刷新SettingsService内存缓存
func SettingsUpdateHandler(c *gin.Context) {
	// 先尝试解析为直接字段格式
	var directBody map[string]interface{}
	if !settingsBaseController.BindJSON(c, &directBody) {
		return
	}

	// 提取设置数据
	var settingsData map[string]string

	// 检查是否为嵌套格式（包含settings字段）
	if settings, exists := directBody["settings"]; exists {
		if settingsMap, ok := settings.(map[string]interface{}); ok {
			settingsData = make(map[string]string)
			for k, v := range settingsMap {
				if str, ok := v.(string); ok {
					settingsData[k] = str
				}
			}
		} else {
			settingsBaseController.HandleValidationError(c, "settings字段格式错误")
			return
		}
	} else {
		// 直接字段格式
		settingsData = make(map[string]string)
		for k, v := range directBody {
			if str, ok := v.(string); ok {
				settingsData[k] = str
			} else if v != nil {
				// 转换其他类型为字符串
				settingsData[k] = fmt.Sprintf("%v", v)
			}
		}
	}

	if len(settingsData) == 0 {
		settingsBaseController.HandleValidationError(c, "无设置项")
		return
	}

	db, ok := settingsBaseController.GetDB(c)
	if !ok {
		return
	}

	// 记录需要失效的缓存键，统一删除，减少与Redis交互次数
	keysToDel := make([]string, 0, len(settingsData))

	// 批量处理设置项
	for k, v := range settingsData {
		var s models.Settings
		if err := db.Where("name = ?", k).First(&s).Error; err != nil {
			// 不存在则创建
			s = models.Settings{Name: k, Value: v}
			if err := db.Create(&s).Error; err != nil {
				logrus.WithError(err).WithField("setting_name", k).Error("创建设置失败")
				settingsBaseController.HandleInternalError(c, fmt.Sprintf("保存设置 %s 失败", k), err)
				return
			}

		} else {
			// 存在则更新
			if err := db.Model(&models.Settings{}).Where("id = ?", s.ID).Update("value", v).Error; err != nil {
				logrus.WithError(err).WithField("setting_name", k).Error("更新设置失败")
				settingsBaseController.HandleInternalError(c, fmt.Sprintf("更新设置 %s 失败", k), err)
				return
			}

		}
		// 收集对应的Redis缓存键（与services/query.go中的键命名保持一致）
		keysToDel = append(keysToDel, fmt.Sprintf("setting:%s", k))
	}

	// 删除Redis缓存键（如果Redis不可用则静默跳过）
	_ = utils.RedisDel(context.Background(), keysToDel...)

	// 刷新内存中的设置缓存，保证后续读取一致
	services.GetSettingsService().RefreshCache()

	settingsBaseController.HandleSuccess(c, "保存成功", nil)
}
