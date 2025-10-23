package models

import (
	"time"
)

// Card 卡密模型
// 用于存储和管理系统中的卡密信息，包括卡密号码、状态、使用情况等
type Card struct {
	// ID：主键，自增
	ID uint `gorm:"primaryKey;comment:卡密ID，自增主键" json:"id"`
	// CardNumber：卡密号码，唯一且非空
	CardNumber string `gorm:"size:200;not null;comment:卡密号码（十六进制字符串）" json:"card_number"`
	// CardTypeID：所属卡密类型ID（外键）
	CardTypeID uint `gorm:"not null;index;comment:所属卡密类型ID（外键）" json:"card_type_id"`
	// Status：状态（0=未使用，1=已使用，2=禁用）
	Status int `gorm:"default:0;not null;comment:状态，0=未使用，1=已使用，2=禁用" json:"status"`
	// Batch：批次标识，用于区分导入或生成批次
	Batch string `gorm:"size:100;comment:批次标识" json:"batch"`
	// Remark：备注信息
	Remark string `gorm:"size:255;comment:备注信息" json:"remark"`
	// UsedAt：使用时间，未使用为NULL（调整到创建时间前面，以便前端展示顺序一致）
	UsedAt *time.Time `gorm:"comment:使用时间" json:"used_at"`
	// CreatedAt/UpdatedAt：时间字段
	CreatedAt time.Time `gorm:"comment:创建时间" json:"created_at"`
	UpdatedAt time.Time `gorm:"comment:更新时间" json:"updated_at"`
}
