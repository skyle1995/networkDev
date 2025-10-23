package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// API 接口表模型
// 用于管理API接口的配置信息
// 包含加密算法配置、密钥管理等功能
// 支持多种加密算法：不加密、RC4、RSA、RSA（动态）

type API struct {
	// ID：主键，自增
	ID uint `gorm:"primaryKey;comment:API接口ID，自增主键" json:"id"`
	
	// API类型（int型）
	APIType int `gorm:"not null;comment:API类型" json:"api_type"`
	
	// API密钥
	APIKey string `gorm:"size:255;not null;uniqueIndex;comment:API密钥，唯一标识" json:"api_key"`
	
	// 应用UUID，关联到App表
	AppUUID string `gorm:"size:36;not null;index;comment:关联的应用UUID" json:"app_uuid"`
	
	// 接口状态（1=启用，0=禁用）
	Status int `gorm:"default:1;not null;comment:接口状态，1=启用，0=禁用" json:"status"`
	
	// 接口提交算法
	// 支持的算法：0=不加密，1=RC4，2=RSA，3=RSA（动态）
	SubmitAlgorithm int `gorm:"default:0;not null;comment:提交算法，0=不加密，1=RC4，2=RSA，3=RSA动态" json:"submit_algorithm"`
	
	// 接口返回算法
	// 支持的算法：0=不加密，1=RC4，2=RSA，3=RSA（动态）
	ReturnAlgorithm int `gorm:"default:0;not null;comment:返回算法，0=不加密，1=RC4，2=RSA，3=RSA动态" json:"return_algorithm"`
	
	// 提交算法公钥（base64编码存储）
	SubmitPublicKey string `gorm:"type:text;comment:提交算法公钥，base64编码" json:"submit_public_key"`
	
	// 提交算法私钥（base64编码存储）
	SubmitPrivateKey string `gorm:"type:text;comment:提交算法私钥，base64编码" json:"submit_private_key"`
	
	// 返回算法公钥（base64编码存储）
	ReturnPublicKey string `gorm:"type:text;comment:返回算法公钥，base64编码" json:"return_public_key"`
	
	// 返回算法私钥（base64编码存储）
	ReturnPrivateKey string `gorm:"type:text;comment:返回算法私钥，base64编码" json:"return_private_key"`
	
	// 时间字段
	CreatedAt time.Time `gorm:"comment:创建时间" json:"created_at"`
	UpdatedAt time.Time `gorm:"comment:更新时间" json:"updated_at"`
}

// BeforeCreate 在创建记录前自动生成API密钥
func (api *API) BeforeCreate(tx *gorm.DB) error {
	if api.APIKey == "" {
		// 生成唯一的API密钥
		api.APIKey = "api_" + uuid.New().String()
	}
	return nil
}

// TableName 指定表名
func (API) TableName() string {
	return "apis"
}

// 算法类型常量
const (
	AlgorithmNone      = 0 // 不加密
	AlgorithmRC4       = 1 // RC4
	AlgorithmRSA       = 2 // RSA
	AlgorithmRSADynamic = 3 // RSA（动态）
)

// GetAlgorithmName 获取算法名称
func GetAlgorithmName(algorithm int) string {
	switch algorithm {
	case AlgorithmNone:
		return "不加密"
	case AlgorithmRC4:
		return "RC4"
	case AlgorithmRSA:
		return "RSA"
	case AlgorithmRSADynamic:
		return "RSA（动态）"
	default:
		return "未知算法"
	}
}

// IsValidAlgorithm 验证算法类型是否有效
func IsValidAlgorithm(algorithm int) bool {
	return algorithm >= AlgorithmNone && algorithm <= AlgorithmRSADynamic
}