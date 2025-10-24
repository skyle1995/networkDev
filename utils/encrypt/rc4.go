package encrypt

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
)

// RC4Encrypt RC4加密算法结构体
type RC4Encrypt struct {
	key []byte
}

// NewRC4Encrypt 创建新的RC4加密实例
func NewRC4Encrypt(key []byte) *RC4Encrypt {
	return &RC4Encrypt{
		key: key,
	}
}

// GenerateRC4Key 生成RC4密钥
func GenerateRC4Key(length int) ([]byte, error) {
	if length <= 0 || length > 256 {
		length = 16 // 默认16字节密钥
	}

	key := make([]byte, length)
	_, err := rand.Read(key)
	if err != nil {
		return nil, fmt.Errorf("生成RC4密钥失败: %v", err)
	}

	return key, nil
}

// rc4KeyScheduling RC4密钥调度算法(KSA)
func (r *RC4Encrypt) rc4KeyScheduling() []int {
	s := make([]int, 256)

	// 初始化S盒
	for i := 0; i < 256; i++ {
		s[i] = i
	}

	// 密钥调度
	j := 0
	keyLen := len(r.key)
	for i := 0; i < 256; i++ {
		j = (j + s[i] + int(r.key[i%keyLen])) % 256
		s[i], s[j] = s[j], s[i] // 交换
	}

	return s
}

// rc4PseudoRandomGeneration RC4伪随机生成算法(PRGA)
func (r *RC4Encrypt) rc4PseudoRandomGeneration(s []int, data []byte) []byte {
	result := make([]byte, len(data))
	i, j := 0, 0

	for k, b := range data {
		i = (i + 1) % 256
		j = (j + s[i]) % 256
		s[i], s[j] = s[j], s[i] // 交换

		keystream := s[(s[i]+s[j])%256]
		result[k] = b ^ byte(keystream)
	}

	return result
}

// Encrypt RC4加密
func (r *RC4Encrypt) Encrypt(plaintext string) (string, error) {
	if len(r.key) == 0 {
		return "", fmt.Errorf("RC4密钥不能为空")
	}

	data := []byte(plaintext)
	s := r.rc4KeyScheduling()
	encrypted := r.rc4PseudoRandomGeneration(s, data)

	// Base64编码
	return base64.StdEncoding.EncodeToString(encrypted), nil
}

// Decrypt RC4解密
func (r *RC4Encrypt) Decrypt(ciphertext string) (string, error) {
	if len(r.key) == 0 {
		return "", fmt.Errorf("RC4密钥不能为空")
	}

	// Base64解码
	data, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", fmt.Errorf("Base64解码失败: %v", err)
	}

	s := r.rc4KeyScheduling()
	decrypted := r.rc4PseudoRandomGeneration(s, data)

	return string(decrypted), nil
}

// ParseRC4KeyFromString 从字符串解析RC4密钥
func ParseRC4KeyFromString(keyStr string) ([]byte, error) {
	key, err := base64.StdEncoding.DecodeString(keyStr)
	if err != nil {
		return nil, fmt.Errorf("解析RC4密钥失败: %v", err)
	}
	return key, nil
}
