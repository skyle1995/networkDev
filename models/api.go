package models

import (
	"strings"
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

	// UUID：API接口唯一标识符，自动生成
	UUID string `gorm:"uniqueIndex;size:36;not null;comment:API接口UUID，唯一标识符" json:"uuid"`

	// API类型（int型）
	APIType int `gorm:"not null;comment:API类型" json:"api_type"`

	// 应用UUID，关联到App表
	AppUUID string `gorm:"size:36;not null;index;comment:关联的应用UUID" json:"app_uuid"`

	// 接口状态（1=启用，0=禁用）
	Status int `gorm:"default:0;not null;comment:接口状态，1=启用，0=禁用" json:"status"`

	// 接口提交算法
	// 支持的算法：0=不加密，1=RC4，2=RSA，3=RSA（动态），4=易加密
	SubmitAlgorithm int `gorm:"default:0;not null;comment:提交算法，0=不加密，1=RC4，2=RSA，3=RSA动态，4=易加密" json:"submit_algorithm"`

	// 接口返回算法
	// 支持的算法：0=不加密，1=RC4，2=RSA，3=RSA（动态），4=易加密
	ReturnAlgorithm int `gorm:"default:0;not null;comment:返回算法，0=不加密，1=RC4，2=RSA，3=RSA动态，4=易加密" json:"return_algorithm"`

	// 提交算法公钥（明文PEM存储）
	SubmitPublicKey string `gorm:"type:text;comment:提交算法公钥，明文PEM" json:"submit_public_key"`

	// 提交算法私钥（明文PEM存储）
	SubmitPrivateKey string `gorm:"type:text;comment:提交算法私钥，明文PEM" json:"submit_private_key"`

	// 返回算法公钥（明文PEM存储）
	ReturnPublicKey string `gorm:"type:text;comment:返回算法公钥，明文PEM" json:"return_public_key"`

	// 返回算法私钥（明文PEM存储）
	ReturnPrivateKey string `gorm:"type:text;comment:返回算法私钥，明文PEM" json:"return_private_key"`

	// 时间字段
	CreatedAt time.Time `gorm:"comment:创建时间" json:"created_at"`
	UpdatedAt time.Time `gorm:"comment:更新时间" json:"updated_at"`
}

// BeforeCreate 在创建记录前自动生成UUID
func (api *API) BeforeCreate(tx *gorm.DB) error {
	if api.UUID == "" {
		api.UUID = strings.ToUpper(uuid.New().String())
	}
	return nil
}

// TableName 指定表名
func (API) TableName() string {
	return "apis"
}

// API类型常量定义
const (
	// 基础信息获取类API
	APITypeGetBulletin     = 1 // 获取程序公告
	APITypeGetUpdateUrl    = 2 // 获取更新地址
	APITypeCheckAppVersion = 3 // 检测最新版本
	APITypeGetCardInfo     = 4 // 获取卡密信息

	// 登录相关API
	APITypeSingleLogin = 10 // 卡密登录

	// 用户账号管理API
	APITypeUserLogin    = 20 // 用户登录
	APITypeUserRegin    = 21 // 用户注册
	APITypeUserRecharge = 22 // 用户充值
	APITypeCardRegin    = 23 // 卡密注册

	// 登出API
	APITypeLogOut = 30 // 退出登录

	// 用户状态查询API
	APITypeGetExpired      = 40 // 获取到期时间
	APITypeCheckUserStatus = 41 // 检测账号状态
	APITypeGetAppData      = 42 // 获取程序数据
	APITypeGetVariable     = 43 // 获取变量数据

	// 用户操作API
	APITypeUpdatePwd     = 50 // 修改账号密码
	APITypeMacChangeBind = 51 // 机器码转绑
	APITypeIPChangeBind  = 52 // IP转绑

	// 管理员操作API
	APITypeDisableUser      = 60 // 封停用户
	APITypeBlackUser        = 61 // 添加黑名单
	APITypeUserDeductedTime = 62 // 扣除时间
)

// 算法类型常量
const (
	AlgorithmNone       = 0 // 不加密
	AlgorithmRC4        = 1 // RC4
	AlgorithmRSA        = 2 // RSA
	AlgorithmRSADynamic = 3 // RSA（动态）
	AlgorithmEasy       = 4 // 易加密
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
	case AlgorithmEasy:
		return "易加密"
	default:
		return "未知算法"
	}
}

// IsValidAlgorithm 验证算法类型是否有效
func IsValidAlgorithm(algorithm int) bool {
	return algorithm >= AlgorithmNone && algorithm <= AlgorithmEasy
}

// GetAPITypeName 获取API类型名称
func GetAPITypeName(apiType int) string {
	switch apiType {
	// 基础信息获取类API
	case APITypeGetBulletin:
		return "获取程序公告"
	case APITypeGetUpdateUrl:
		return "获取更新地址"
	case APITypeCheckAppVersion:
		return "检测最新版本"
	case APITypeGetCardInfo:
		return "获取卡密信息"

	// 登录相关API
	case APITypeSingleLogin:
		return "卡密登录"

	// 用户账号管理API
	case APITypeUserLogin:
		return "用户登录"
	case APITypeUserRegin:
		return "用户注册"
	case APITypeUserRecharge:
		return "用户充值"
	case APITypeCardRegin:
		return "卡密注册"

	// 登出API
	case APITypeLogOut:
		return "退出登录"

	// 用户状态查询API
	case APITypeGetExpired:
		return "获取到期时间"
	case APITypeCheckUserStatus:
		return "检测账号状态"
	case APITypeGetAppData:
		return "获取程序数据"
	case APITypeGetVariable:
		return "获取变量数据"

	// 用户操作API
	case APITypeUpdatePwd:
		return "修改账号密码"
	case APITypeMacChangeBind:
		return "机器码转绑"
	case APITypeIPChangeBind:
		return "IP转绑"

	// 管理员操作API
	case APITypeDisableUser:
		return "封停用户"
	case APITypeBlackUser:
		return "添加黑名单"
	case APITypeUserDeductedTime:
		return "扣除时间"

	default:
		return "未知API类型"
	}
}

// IsValidAPIType 验证API类型是否有效
func IsValidAPIType(apiType int) bool {
	validTypes := []int{
		APITypeGetBulletin, APITypeGetUpdateUrl, APITypeCheckAppVersion, APITypeGetCardInfo,
		APITypeSingleLogin,
		APITypeUserLogin, APITypeUserRegin, APITypeUserRecharge, APITypeCardRegin,
		APITypeLogOut,
		APITypeGetExpired, APITypeCheckUserStatus, APITypeGetAppData, APITypeGetVariable,
		APITypeUpdatePwd, APITypeMacChangeBind, APITypeIPChangeBind,
		APITypeDisableUser, APITypeBlackUser, APITypeUserDeductedTime,
	}

	for _, validType := range validTypes {
		if apiType == validType {
			return true
		}
	}
	return false
}

// GetAPITypesByCategory 根据分类获取API类型列表
func GetAPITypesByCategory() map[string][]int {
	return map[string][]int{
		"基础信息获取": {APITypeGetBulletin, APITypeGetUpdateUrl, APITypeCheckAppVersion, APITypeGetCardInfo},
		"登录相关":   {APITypeSingleLogin},
		"用户账号管理": {APITypeUserLogin, APITypeUserRegin, APITypeUserRecharge, APITypeCardRegin},
		"登出":     {APITypeLogOut},
		"用户状态查询": {APITypeGetExpired, APITypeCheckUserStatus, APITypeGetAppData, APITypeGetVariable},
		"用户操作":   {APITypeUpdatePwd, APITypeMacChangeBind, APITypeIPChangeBind},
		"管理员操作":  {APITypeDisableUser, APITypeBlackUser, APITypeUserDeductedTime},
	}
}
