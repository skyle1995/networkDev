package models

import "time"

// Settings 系统设置表模型
// 用于存储应用的配置参数
// Name 为配置项名称，唯一索引
// Value 为配置项的值
// Description 为配置项描述说明
// CreatedAt/UpdatedAt 由 GORM 自动维护

type Settings struct {
	ID          uint      `gorm:"primaryKey;comment:设置ID，自增主键"`
	Name        string    `gorm:"uniqueIndex;size:64;not null;comment:配置项名称，唯一索引"`
	Value       string    `gorm:"type:text;comment:配置项的值"`
	Description string    `gorm:"size:255;comment:配置项描述说明"`
	CreatedAt   time.Time `gorm:"comment:创建时间"`
	UpdatedAt   time.Time `gorm:"comment:更新时间"`
}
