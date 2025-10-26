package services

import (
	"networkDev/database"
	"networkDev/models"
	"strconv"
	"sync"

	"github.com/sirupsen/logrus"
)

// SettingsService 设置服务
type SettingsService struct {
	mu    sync.RWMutex
	cache map[string]string
}

var settingsService *SettingsService
var settingsOnce sync.Once

// GetSettingsService 获取设置服务单例
func GetSettingsService() *SettingsService {
	settingsOnce.Do(func() {
		settingsService = &SettingsService{
			cache: make(map[string]string),
		}
		// 初始化时加载所有设置
		settingsService.loadAllSettings()
	})
	return settingsService
}

// loadAllSettings 从数据库加载所有设置到缓存
func (s *SettingsService) loadAllSettings() {
	db, err := database.GetDB()
	if err != nil {
		logrus.WithError(err).Error("获取数据库连接失败")
		return
	}

	var settings []models.Settings
	if err := db.Find(&settings).Error; err != nil {
		logrus.WithError(err).Error("加载设置失败")
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	for _, setting := range settings {
		s.cache[setting.Name] = setting.Value
	}

	logrus.WithField("count", len(settings)).Info("设置缓存加载完成")
}

// GetString 获取字符串类型的设置值
func (s *SettingsService) GetString(name, defaultValue string) string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if value, exists := s.cache[name]; exists {
		return value
	}
	return defaultValue
}

// GetInt 获取整数类型的设置值
func (s *SettingsService) GetInt(name string, defaultValue int) int {
	strValue := s.GetString(name, "")
	if strValue == "" {
		return defaultValue
	}

	if intValue, err := strconv.Atoi(strValue); err == nil {
		return intValue
	}
	return defaultValue
}

// GetBool 获取布尔类型的设置值
func (s *SettingsService) GetBool(name string, defaultValue bool) bool {
	strValue := s.GetString(name, "")
	if strValue == "" {
		return defaultValue
	}

	return strValue == "1" || strValue == "true"
}

// RefreshCache 刷新设置缓存
func (s *SettingsService) RefreshCache() {
	s.loadAllSettings()
}

// GetSessionTimeout 获取会话超时时间（秒）
func (s *SettingsService) GetSessionTimeout() int {
	return s.GetInt("session_timeout", 3600) // 默认1小时
}

// IsMaintenanceMode 检查是否开启维护模式
func (s *SettingsService) IsMaintenanceMode() bool {
	return s.GetBool("maintenance_mode", false)
}
