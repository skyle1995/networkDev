package models

import (
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ============================================================================
// 结构体定义
// ============================================================================

// User 用户表模型
// 此表只存储普通用户，管理员账号存储在settings表中
// CreatedAt/UpdatedAt 由 GORM 自动维护
type User struct {
	ID           uint      `gorm:"primaryKey;comment:用户ID，自增主键"`
	UUID         string    `gorm:"uniqueIndex;size:36;not null;comment:用户的唯一标识符" json:"uuid"`
	Username     string    `gorm:"uniqueIndex;size:64;not null;comment:用户名，唯一索引"`
	Password     string    `gorm:"size:255;not null;comment:密码哈希值"`
	PasswordSalt string    `gorm:"size:64;not null;comment:密码加密盐值"`
	CreatedAt    time.Time `gorm:"comment:创建时间"`
	UpdatedAt    time.Time `gorm:"comment:更新时间"`
}

// ============================================================================
// 结构体方法
// ============================================================================

// BeforeCreate 在创建记录前自动生成UUID
func (user *User) BeforeCreate(tx *gorm.DB) error {
	// 生成UUID
	if user.UUID == "" {
		user.UUID = strings.ToUpper(uuid.New().String())
	}
	return nil
}
