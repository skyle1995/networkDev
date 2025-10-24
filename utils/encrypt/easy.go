package encrypt

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"strconv"
	"strings"
)

// EasyEncrypt 易加密算法结构体
type EasyEncrypt struct {
	encryptKey []int // 加密密钥
	decryptKey []int // 解密密钥
}

// NewEasyEncrypt 创建新的易加密实例
func NewEasyEncrypt(encryptKey, decryptKey []int) *EasyEncrypt {
	return &EasyEncrypt{
		encryptKey: encryptKey,
		decryptKey: decryptKey,
	}
}

// GenerateEasyKey 生成易加密密钥对
func GenerateEasyKey() ([]int, []int, error) {
	// 使用crypto/rand生成随机长度（15-30位）
	var lengthByte [1]byte
	
	// 生成加密密钥长度
	if _, err := rand.Read(lengthByte[:]); err != nil {
		return nil, nil, err
	}
	encryptKeyLen := 15 + int(lengthByte[0])%16 // 15-30位随机长度
	
	encryptKey := make([]int, encryptKeyLen)
	encryptBytes := make([]byte, encryptKeyLen)
	if _, err := rand.Read(encryptBytes); err != nil {
		return nil, nil, err
	}
	for i, b := range encryptBytes {
		encryptKey[i] = int(b) // 0-255范围
	}

	// 生成解密密钥长度
	if _, err := rand.Read(lengthByte[:]); err != nil {
		return nil, nil, err
	}
	decryptKeyLen := 15 + int(lengthByte[0])%16 // 15-30位随机长度
	
	decryptKey := make([]int, decryptKeyLen)
	decryptBytes := make([]byte, decryptKeyLen)
	if _, err := rand.Read(decryptBytes); err != nil {
		return nil, nil, err
	}
	for i, b := range decryptBytes {
		decryptKey[i] = int(b) // 0-255范围
	}

	return encryptKey, decryptKey, nil
}

// Encrypt 加密函数 - 对应 UserLogin_encrypt_Up_42510
func (e *EasyEncrypt) Encrypt(input string) string {
	if input == "" {
		return ""
	}

	mKeyLen := len(e.encryptKey)
	inputLen := len(input)
	var result strings.Builder

	for i := 0; i < inputLen; i++ {
		mCode := int(input[i])
		mCode = (mCode - 207) ^ e.encryptKey[i%mKeyLen]

		if mCode < 0 {
			mCode = -mCode
			result.WriteString("-")
		}

		// 转换为16进制字符串
		hexStr := strconv.FormatInt(int64(mCode), 16)
		result.WriteString(hexStr)
		result.WriteString(",")
	}

	// Base64编码
	resultStr := result.String()
	return base64.StdEncoding.EncodeToString([]byte(resultStr))
}

// Decrypt 解密函数 - 对应 UserLogin_decrypt_Down_42510
func (e *EasyEncrypt) Decrypt(input string) string {
	if input == "" {
		return ""
	}

	// Base64解码
	decoded, err := base64.StdEncoding.DecodeString(input)
	if err != nil {
		return ""
	}

	decodedStr := string(decoded)
	mKeyLen := len(e.encryptKey)

	// 按逗号分割
	parts := strings.Split(decodedStr, ",")
	var result strings.Builder

	for i, part := range parts {
		if part == "" {
			continue
		}

		var d int
		if strings.HasPrefix(part, "-") {
			// 处理负数
			hexStr := part[1:]
			val, err := strconv.ParseInt(hexStr, 16, 32)
			if err != nil {
				continue
			}
			d = -int(val)
		} else {
			// 处理正数
			val, err := strconv.ParseInt(part, 16, 32)
			if err != nil {
				continue
			}
			d = int(val)
		}

		// 解密计算
		decryptedChar := (d ^ e.encryptKey[i%mKeyLen]) + 207
		result.WriteByte(byte(decryptedChar))
	}

	return result.String()
}

// EncryptWithKey 使用指定密钥加密
func EncryptWithKey(input string, key []int) string {
	if input == "" || len(key) == 0 {
		return ""
	}

	keyLen := len(key)
	inputLen := len(input)
	var result strings.Builder

	for i := 0; i < inputLen; i++ {
		mCode := int(input[i])
		mCode = (mCode - 207) ^ key[i%keyLen]

		if mCode < 0 {
			mCode = -mCode
			result.WriteString("-")
		}

		hexStr := strconv.FormatInt(int64(mCode), 16)
		result.WriteString(hexStr)
		result.WriteString(",")
	}

	resultStr := result.String()
	return base64.StdEncoding.EncodeToString([]byte(resultStr))
}

// DecryptWithKey 使用指定密钥解密
func DecryptWithKey(input string, key []int) string {
	if input == "" || len(key) == 0 {
		return ""
	}

	decoded, err := base64.StdEncoding.DecodeString(input)
	if err != nil {
		return ""
	}

	decodedStr := string(decoded)
	keyLen := len(key)

	parts := strings.Split(decodedStr, ",")
	var result strings.Builder

	for i, part := range parts {
		if part == "" {
			continue
		}

		var d int
		if strings.HasPrefix(part, "-") {
			hexStr := part[1:]
			val, err := strconv.ParseInt(hexStr, 16, 32)
			if err != nil {
				continue
			}
			d = -int(val)
		} else {
			val, err := strconv.ParseInt(part, 16, 32)
			if err != nil {
				continue
			}
			d = int(val)
		}

		decryptedChar := (d ^ key[i%keyLen]) + 40
		result.WriteByte(byte(decryptedChar))
	}

	return result.String()
}

// FormatKeyAsString 将密钥数组格式化为字符串（用于存储）
func FormatKeyAsString(key []int) string {
	var parts []string
	for _, k := range key {
		parts = append(parts, fmt.Sprintf("%d", k))
	}
	return strings.Join(parts, ",")
}

// ParseKeyFromString 从字符串解析密钥数组
func ParseKeyFromString(keyStr string) []int {
	if keyStr == "" {
		return nil
	}

	parts := strings.Split(keyStr, ",")
	var key []int

	for _, part := range parts {
		if val, err := strconv.Atoi(strings.TrimSpace(part)); err == nil {
			key = append(key, val)
		}
	}

	return key
}
