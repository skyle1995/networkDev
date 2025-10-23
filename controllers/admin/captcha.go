package admin

import (
	"encoding/base64"
	"encoding/json"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"github.com/mojocn/base64Captcha"
)

// 全局验证码存储器
var store = base64Captcha.DefaultMemStore

// CaptchaHandler 生成验证码图片
// GET /admin/captcha - 返回验证码图片
func CaptchaHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	// 随机生成4-6位长度
	// Go 1.20+ 无需手动设置随机种子，使用默认全局随机源即可
	captchaLength := 4 + rand.Intn(3) // 4-6位随机长度

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
		http.Error(w, "生成验证码失败", http.StatusInternalServerError)
		return
	}

	// 将验证码ID存储到session中（这里简化处理，实际项目中应该使用更安全的方式）
	// 设置cookie来存储验证码ID
	cookie := &http.Cookie{
		Name:     "captcha_id",
		Value:    id,
		Path:     "/",
		HttpOnly: true,
		Secure:   false, // 生产环境应设置为true
		MaxAge:   300,   // 5分钟过期
		Expires:  time.Now().Add(5 * time.Minute),
	}
	http.SetCookie(w, cookie)

	// 解码base64图片数据并返回
	w.Header().Set("Content-Type", "image/png")
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")

	// 直接返回base64编码的图片数据，让浏览器解析
	// 但是我们需要返回实际的图片数据，所以需要解码base64

	// 去掉data:image/png;base64,前缀
	b64s = strings.TrimPrefix(b64s, "data:image/png;base64,")

	imgData, err := base64.StdEncoding.DecodeString(b64s)
	if err != nil {
		http.Error(w, "解码验证码图片失败", http.StatusInternalServerError)
		return
	}

	w.Write(imgData)
}

// VerifyCaptcha 验证验证码
// 这个函数将在登录处理中被调用
// 支持大小写不敏感匹配
func VerifyCaptcha(r *http.Request, captchaValue string) bool {
	// 从cookie中获取验证码ID
	cookie, err := r.Cookie("captcha_id")
	if err != nil {
		return false
	}

	captchaId := cookie.Value
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
func CaptchaAPIHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	var body struct {
		Captcha string `json:"captcha"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"code": 1,
			"msg":  "请求参数错误",
		})
		return
	}

	isValid := VerifyCaptcha(r, body.Captcha)

	w.Header().Set("Content-Type", "application/json")
	if isValid {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"code": 0,
			"msg":  "验证码正确",
		})
	} else {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"code": 1,
			"msg":  "验证码错误",
		})
	}
}
