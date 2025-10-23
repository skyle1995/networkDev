package models

import "time"

// User 用户表模型
// 说明：PasswordSalt 使用 32 字节随机盐（以 16 进制存储为 64 个字符），因此列长度设置为 64

type User struct {
	ID           uint      `gorm:"primaryKey;comment:用户ID，自增主键"`
	Username     string    `gorm:"uniqueIndex;size:64;not null;comment:用户名，唯一索引"`
	Password     string    `gorm:"size:255;not null;comment:密码哈希值"`
	PasswordSalt string    `gorm:"size:64;not null;comment:密码加密盐值"`
	Role         int       `gorm:"not null;comment:用户角色，0=管理员，1=普通用户"`
	CreatedAt    time.Time `gorm:"comment:创建时间"`
	UpdatedAt    time.Time `gorm:"comment:更新时间"`
}
