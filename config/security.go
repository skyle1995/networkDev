package config

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"fmt"
)

// ============================================================================
// 公共函数
// ============================================================================

// GenerateSecureJWTSecret 生成安全的JWT密钥
// 生成64字节（512位）的随机密钥，使用base64编码
func GenerateSecureJWTSecret() (string, error) {
	// 生成64字节的随机数据
	bytes := make([]byte, 64)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("生成JWT密钥失败: %w", err)
	}

	// 使用base64编码，便于配置文件存储
	return base64.StdEncoding.EncodeToString(bytes), nil
}

// GenerateSecureEncryptionKey 生成安全的加密密钥
// 生成32字节（256位）的随机密钥，使用十六进制编码
func GenerateSecureEncryptionKey() (string, error) {
	// 生成32字节的随机数据（AES-256）
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("生成加密密钥失败: %w", err)
	}

	// 使用十六进制编码
	return hex.EncodeToString(bytes), nil
}

// GenerateSecureKeys 生成所有安全密钥
func GenerateSecureKeys() (jwtSecret, encryptionKey string, err error) {
	jwtSecret, err = GenerateSecureJWTSecret()
	if err != nil {
		return "", "", err
	}

	encryptionKey, err = GenerateSecureEncryptionKey()
	if err != nil {
		return "", "", err
	}

	return jwtSecret, encryptionKey, nil
}
