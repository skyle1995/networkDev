package models

import (
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Variable 变量表模型
// 用于管理应用程序的变量数据
// UUID 为变量的唯一标识符，自动生成并转换为大写
// Alias 为变量别名，便于识别和管理
// Data 为变量数据内容
// Remark 为备注信息，用于描述变量用途
// CreatedAt/UpdatedAt 由 GORM 自动维护

type Variable struct {
	// ID：主键，自增
	ID uint `gorm:"primaryKey;comment:变量ID，自增主键" json:"id"`

	// UUID：变量的唯一标识符，36位字符串
	UUID string `gorm:"uniqueIndex;size:36;not null;comment:变量的唯一标识符" json:"uuid"`

	// Number：变量编号，时间戳+6位随机数字格式
	Number string `gorm:"uniqueIndex;size:20;not null;comment:变量编号，时间戳+6位随机数字格式" json:"number"`

	// Alias：变量别名，便于识别和管理
	Alias string `gorm:"size:100;not null;comment:变量别名" json:"alias"`

	// Data：变量数据内容
	Data string `gorm:"type:text;comment:变量数据" json:"data"`

	// Remark：备注信息，用于描述变量用途
	Remark string `gorm:"type:text;comment:备注信息" json:"remark"`

	// 时间字段
	CreatedAt time.Time `gorm:"comment:创建时间" json:"created_at"`
	UpdatedAt time.Time `gorm:"comment:更新时间" json:"updated_at"`
}

// BeforeCreate 在创建记录前自动生成UUID和Number
func (variable *Variable) BeforeCreate(tx *gorm.DB) error {
	// 生成UUID
	if variable.UUID == "" {
		variable.UUID = strings.ToUpper(uuid.New().String())
	}
	
	// 生成Number：使用时间戳格式
	variable.Number = time.Now().Format("20060102150405")
	return nil
}

// TableName 指定表名
func (Variable) TableName() string {
	return "variables"
}