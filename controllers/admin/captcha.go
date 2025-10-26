package admin

import (
	"crypto/rand"
	"encoding/base64"
	"math/big"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"networkDev/controllers"
	"networkDev/utils"

	"github.com/mojocn/base64Captcha"
	"github.com/spf13/viper"
)

// 创建基础控制器实例
var captchaBaseController = controllers.NewBaseController()

// 全局验证码存储器
var store = base64Captcha.DefaultMemStore

// secureRandomInt 生成安全的随机整数，范围 [0, max)
func secureRandomInt(max int) (int, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(int64(max)))
	if err != nil {
		return 0, err
	}
	return int(n.Int64()), nil
}

// CaptchaHandler 生成验证码图片
// GET /admin/captcha - 返回验证码图片
func CaptchaHandler(c *gin.Context) {
	// 随机生成4-6位长度
	// 使用crypto/rand生成安全的随机数
	randomNum, err := secureRandomInt(3)
	if err != nil {
		captchaBaseController.HandleInternalError(c, "生成随机数失败", err)
		return
	}
	captchaLength := 4 + randomNum // 4-6位随机长度

	// 配置验证码参数 - 使用字母数字混合
	driver := base64Captcha.DriverString{
		Height:          60,
		Width:           200,
		NoiseCount:      0,
		ShowLineOptions: 2 | 4,
		Length:          captchaLength,
		Source:          "ABCDEFGHJKMNPQRSTUVWXYZabcdefghjkmnpqrstuvwxyz23456789", // 混合大小写字母和数字，去除易混淆字符
		Fonts:           []string{"wqy-microhei.ttc"},
	}

	// 生成验证码
	captcha := base64Captcha.NewCaptcha(&driver, store)
	id, b64s, _, err := captcha.Generate()
	if err != nil {
		captchaBaseController.HandleInternalError(c, "生成验证码失败", err)
		return
	}

	// 将验证码ID存储到session中（这里简化处理，实际项目中应该使用更安全的方式）
	// 设置cookie来存储验证码ID
	cookie := utils.CreateSecureCookie("captcha_id", id, 300) // 5分钟过期
	c.SetCookie(cookie.Name, cookie.Value, cookie.MaxAge, cookie.Path, cookie.Domain, cookie.Secure, cookie.HttpOnly)

	// 解码base64图片数据并返回
	c.Header("Content-Type", "image/png")
	c.Header("Cache-Control", "no-cache, no-store, must-revalidate")
	c.Header("Pragma", "no-cache")
	c.Header("Expires", "0")

	// 直接返回base64编码的图片数据，让浏览器解析
	// 但是我们需要返回实际的图片数据，所以需要解码base64

	// 去掉data:image/png;base64,前缀
	b64s = strings.TrimPrefix(b64s, "data:image/png;base64,")

	imgData, err := base64.StdEncoding.DecodeString(b64s)
	if err != nil {
		captchaBaseController.HandleInternalError(c, "解码验证码图片失败", err)
		return
	}

	c.Data(http.StatusOK, "image/png", imgData)
}



// VerifyCaptcha 验证验证码
// 这个函数将在登录处理中被调用
// 支持大小写不敏感匹配
func VerifyCaptcha(c *gin.Context, captchaValue string) bool {
	// 检查是否为开发模式，如果是则跳过验证码验证
	if viper.GetBool("server.dev_mode") {
		return true
	}
	
	// 从cookie中获取验证码ID
	captchaId, err := c.Cookie("captcha_id")
	if err != nil {
		return false
	}

	if captchaId == "" {
		return false
	}

	// 先尝试原始值验证
	if store.Verify(captchaId, captchaValue, false) {
		// 验证成功后删除验证码
		store.Verify(captchaId, captchaValue, true)
		return true
	}

	// 如果原始值验证失败，尝试小写验证（因为显示的是大小写混合，但允许用户输入小写）
	if store.Verify(captchaId, strings.ToLower(captchaValue), false) {
		// 验证成功后删除验证码
		store.Verify(captchaId, strings.ToLower(captchaValue), true)
		return true
	}

	// 最后尝试大写验证
	if store.Verify(captchaId, strings.ToUpper(captchaValue), true) {
		return true
	}

	return false
}

// CaptchaAPIHandler 验证码API接口（可选，用于AJAX验证）
// POST /admin/api/captcha/verify - 验证验证码
func CaptchaAPIHandler(c *gin.Context) {
	var body struct {
		Captcha string `json:"captcha"`
	}
	if !captchaBaseController.BindJSON(c, &body) {
		return
	}

	isValid := VerifyCaptcha(c, body.Captcha)

	if isValid {
		captchaBaseController.HandleSuccess(c, "验证码正确", nil)
	} else {
		captchaBaseController.HandleValidationError(c, "验证码错误")
	}
}
