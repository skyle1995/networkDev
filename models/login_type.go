package models

import "time"

// LoginType 登录类型表模型
// 用于管理不同的登录方式，如直登、Google、Microsoft、Apple等
// ID 为自增主键
// Name 为登录类型名称，唯一索引
// Status 为状态（1:启用 0:禁用），默认为1
// CreatedAt/UpdatedAt 由 GORM 自动维护

type LoginType struct {
	// ID：主键，自增，同时通过 json 标签保证前端接收为 id
	ID uint `gorm:"primaryKey;comment:登录类型ID，自增主键" json:"id"`
	// Name：名称，唯一；json 名称与前端一致
	Name string `gorm:"uniqueIndex;size:100;not null;comment:登录类型名称，唯一索引" json:"name"`
	// Status：状态（1=启用，0=禁用）；json 名称与前端一致
	Status int `gorm:"default:1;not null;comment:状态，1=启用，0=禁用" json:"status"`
	// VerifyTypes：验证方式（逗号分隔）；json 使用 verify_types；用于记录多种验证方式，输入内容用多个用逗号分隔
	VerifyTypes string `gorm:"type:varchar(500);default:'';comment:验证方式，输入内容用多个用逗号分隔" json:"verify_types"`
	// CreatedAt/UpdatedAt：时间字段，返回为 created_at/updated_at，便于前端展示
	CreatedAt time.Time `gorm:"comment:创建时间" json:"created_at"`
	UpdatedAt time.Time `gorm:"comment:更新时间" json:"updated_at"`
}
