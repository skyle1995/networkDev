package constants



// 卡密状态常量
// CardStatus 定义卡密的状态
const (
	// CardStatusUnused 未使用
	CardStatusUnused = 0
	// CardStatusUsed 已使用
	CardStatusUsed = 1
	// CardStatusDisabled 禁用
	CardStatusDisabled = 2
)

// 登录类型状态常量
// LoginTypeStatus 定义登录类型的状态
const (
	// LoginTypeStatusDisabled 禁用
	LoginTypeStatusDisabled = 0
	// LoginTypeStatusEnabled 启用
	LoginTypeStatusEnabled = 1
)

// 卡密类型状态常量
// CardTypeStatus 定义卡密类型的状态
const (
	// CardTypeStatusDisabled 禁用
	CardTypeStatusDisabled = 0
	// CardTypeStatusEnabled 启用
	CardTypeStatusEnabled = 1
)

// 验证码类型常量
// VerificationCodeType 定义验证码的类型
const (
	// VerificationCodeTypeText 文本验证码
	VerificationCodeTypeText = 1
	// VerificationCodeTypeImage 图片验证码
	VerificationCodeTypeImage = 2
)
