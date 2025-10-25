package models

import (
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// User 用户表模型
// 说明：PasswordSalt 使用 32 字节随机盐（以 16 进制存储为 64 个字符），因此列长度设置为 64
type User struct {
	ID           uint      `gorm:"primaryKey;comment:用户ID，自增主键"`
	UUID         string    `gorm:"uniqueIndex;size:36;not null;comment:用户的唯一标识符" json:"uuid"`
	Username     string    `gorm:"uniqueIndex;size:64;not null;comment:用户名，唯一索引"`
	Password     string    `gorm:"size:255;not null;comment:密码哈希值"`
	PasswordSalt string    `gorm:"size:64;not null;comment:密码加密盐值"`
	Role         int       `gorm:"not null;comment:用户角色，0=管理员，1=普通用户"`
	CreatedAt    time.Time `gorm:"comment:创建时间"`
	UpdatedAt    time.Time `gorm:"comment:更新时间"`
}

// BeforeCreate 在创建记录前自动生成UUID
func (user *User) BeforeCreate(tx *gorm.DB) error {
	// 生成UUID
	if user.UUID == "" {
		user.UUID = strings.ToUpper(uuid.New().String())
	}
	return nil
}
