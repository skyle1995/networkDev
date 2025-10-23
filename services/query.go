package services

import (
	"context"
	"fmt"
	"networkDev/models"
	"networkDev/utils"
	"time"

	"gorm.io/gorm"
)



// FindCardByCardNumber 根据卡号查找卡密
// cardNumber: 卡号
// db: 数据库连接
// 返回: 卡密信息和错误
func FindCardByCardNumber(cardNumber string, db *gorm.DB) (*models.Card, error) {
	key := fmt.Sprintf("card:number:%s", cardNumber)
	return utils.RedisGetOrSet(context.Background(), key, 60*time.Second, func() (*models.Card, error) {
		var card models.Card
		err := db.Where("card_number = ?", cardNumber).First(&card).Error
		if err != nil {
			return nil, err
		}
		return &card, nil
	})
}

// FindCardTypeByID 根据ID查找卡密类型
// id: 卡密类型ID
// db: 数据库连接
// 返回: 卡密类型信息和错误
func FindCardTypeByID(id uint, db *gorm.DB) (*models.CardType, error) {
	key := fmt.Sprintf("card_type:id:%d", id)
	return utils.RedisGetOrSet(context.Background(), key, 30*time.Minute, func() (*models.CardType, error) {
		var cardType models.CardType
		err := db.Where("id = ?", id).First(&cardType).Error
		if err != nil {
			return nil, err
		}
		return &cardType, nil
	})
}

// FindSettingByName 根据名称查找设置
// name: 设置名称
// db: 数据库连接
// 返回: 设置信息和错误
func FindSettingByName(name string, db *gorm.DB) (*models.Settings, error) {
	key := fmt.Sprintf("setting:%s", name)
	return utils.RedisGetOrSet(context.Background(), key, 5*time.Minute, func() (*models.Settings, error) {
		var setting models.Settings
		err := db.Where("name = ?", name).First(&setting).Error
		if err != nil {
			return nil, err
		}
		return &setting, nil
	})
}



// UpdateEntityByID 根据ID更新实体
// model: 模型类型
// id: 实体ID
// updates: 更新字段
// db: 数据库连接
// 返回: 错误
func UpdateEntityByID(model interface{}, id uint, updates map[string]interface{}, db *gorm.DB) error {
	return db.Model(model).Where("id = ?", id).Updates(updates).Error
}

// BatchUpdateEntityStatus 批量更新实体状态
// model: 模型类型
// ids: 实体ID列表
// status: 新状态
// db: 数据库连接
// 返回: 错误
func BatchUpdateEntityStatus(model interface{}, ids []uint, status int, db *gorm.DB) error {
	if len(ids) == 0 {
		return nil
	}
	return db.Model(model).Where("id IN ?", ids).Update("status", status).Error
}

// CountEntitiesByCondition 根据条件统计实体数量
// model: 模型类型
// condition: 查询条件
// db: 数据库连接
// args: 查询参数
// 返回: 数量和错误
func CountEntitiesByCondition(model interface{}, condition string, db *gorm.DB, args ...interface{}) (int64, error) {
	var count int64
	err := db.Model(model).Where(condition, args...).Count(&count).Error
	return count, err
}

// FindEntitiesByCondition 根据条件查找实体
// model: 模型类型
// result: 结果容器
// condition: 查询条件
// db: 数据库连接
// args: 查询参数
// 返回: 错误
func FindEntitiesByCondition(model interface{}, result interface{}, condition string, db *gorm.DB, args ...interface{}) error {
	return db.Model(model).Where(condition, args...).Find(result).Error
}

// CheckEntityExists 检查实体是否存在
// model: 模型类型
// condition: 查询条件
// db: 数据库连接
// args: 查询参数
// 返回: 是否存在和错误
func CheckEntityExists(model interface{}, condition string, db *gorm.DB, args ...interface{}) (bool, error) {
	var count int64
	err := db.Model(model).Where(condition, args...).Count(&count).Error
	return count > 0, err
}
