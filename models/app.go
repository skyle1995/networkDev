package models

import (
	"crypto/rand"
	"encoding/hex"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// App 应用表模型
// 用于管理应用程序的基本信息
// UUID 为应用的唯一标识符，自动生成
// Status 为应用状态（1:启用 0:禁用），默认为1
// Name 为应用名称
// Secret 为应用密钥，用于API认证
// Version 为应用版本号
// CreatedAt/UpdatedAt 由 GORM 自动维护

type App struct {
	// ID：主键，自增，同时通过 json 标签保证前端接收为 id
	ID uint `gorm:"primaryKey;comment:应用ID，自增主键" json:"id"`
	// UUID：应用唯一标识符，自动生成
	UUID string `gorm:"uniqueIndex;size:36;not null;comment:应用UUID，唯一标识符" json:"uuid"`
	// Status：状态（1=启用，0=禁用）；json 名称与前端一致
	Status int `gorm:"default:1;not null;comment:应用状态，1=启用，0=禁用" json:"status"`
	// Name：应用名称；json 名称与前端一致
	Name string `gorm:"size:100;not null;comment:应用名称" json:"name"`
	// Secret：应用密钥，用于API认证
	Secret string `gorm:"size:255;not null;comment:应用密钥，用于API认证" json:"secret"`
	// Version：应用版本号
	Version string `gorm:"size:50;default:'1.0.0';comment:应用版本号" json:"version"`
	// ForceUpdate：强制更新（0=不开启，1=开启）
	ForceUpdate int `gorm:"default:0;not null;comment:强制更新，0=不开启，1=开启" json:"force_update"`
	// DownloadType：下载方式（0=不启用更新，1=自动更新，2=手动下载）
	DownloadType int `gorm:"default:0;not null;comment:更新方式，0=不启用更新，1=自动更新，2=手动下载" json:"download_type"`
	// DownloadURL：下载地址
	DownloadURL string `gorm:"size:500;comment:下载地址" json:"download_url"`
	// Announcement：程序公告内容（base64编码存储）
	Announcement string `gorm:"type:text;comment:程序公告内容，base64编码存储" json:"announcement"`
	// LoginType：登陆方式（0=顶号登录（默认），1=非顶号登录）
	LoginType int `gorm:"default:0;not null;comment:登陆方式，0=顶号登录，1=非顶号登录" json:"login_type"`
	// MultiOpenScope：多开范围（0=单电脑，1=单IP，2=全部电脑（默认））
	MultiOpenScope int `gorm:"default:2;not null;comment:多开范围，0=单电脑，1=单IP，2=全部电脑" json:"multi_open_scope"`
	// CleanInterval：清理间隔（单位：小时，默认1小时）
	CleanInterval int `gorm:"default:1;not null;comment:清理间隔，单位小时" json:"clean_interval"`
	// CheckInterval：校验间隔（单位：分钟，默认10分钟）
	CheckInterval int `gorm:"default:10;not null;comment:校验间隔，单位分钟" json:"check_interval"`
	// MultiOpenCount：多开数量（默认1）
	MultiOpenCount int `gorm:"default:1;not null;comment:多开数量" json:"multi_open_count"`
	// CreatedAt/UpdatedAt：时间字段，返回为 created_at/updated_at，便于前端展示
	CreatedAt time.Time `gorm:"comment:创建时间" json:"created_at"`
	UpdatedAt time.Time `gorm:"comment:更新时间" json:"updated_at"`
}

// BeforeCreate 在创建记录前自动生成UUID和密钥
func (app *App) BeforeCreate(tx *gorm.DB) error {
	if app.UUID == "" {
		app.UUID = uuid.New().String()
	}
	if app.Secret == "" {
		// 生成32位大写16进制随机字符
		bytes := make([]byte, 16) // 16字节 = 32位16进制字符
		rand.Read(bytes)
		app.Secret = strings.ToUpper(hex.EncodeToString(bytes))
	}
	return nil
}

// TableName 指定表名
func (App) TableName() string {
	return "apps"
}
