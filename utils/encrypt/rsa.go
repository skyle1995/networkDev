package encrypt

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
)

// GenerateRSAKeyPair 生成RSA密钥对（公共函数）
func GenerateRSAKeyPair(bits int) (*rsa.PublicKey, *rsa.PrivateKey, error) {
	if bits < 1024 {
		bits = 2048 // 默认2048位
	}

	privateKey, err := rsa.GenerateKey(rand.Reader, bits)
	if err != nil {
		return nil, nil, fmt.Errorf("生成RSA密钥对失败: %v", err)
	}

	return &privateKey.PublicKey, privateKey, nil
}

// PublicKeyToPEM 将RSA公钥转换为PEM格式字符串（公共函数）
func PublicKeyToPEM(publicKey *rsa.PublicKey) (string, error) {
	publicKeyBytes, err := x509.MarshalPKIXPublicKey(publicKey)
	if err != nil {
		return "", fmt.Errorf("序列化公钥失败: %v", err)
	}

	publicKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: publicKeyBytes,
	})

	return string(publicKeyPEM), nil
}

// PrivateKeyToPEM 将RSA私钥转换为PEM格式字符串（公共函数）
func PrivateKeyToPEM(privateKey *rsa.PrivateKey) (string, error) {
	privateKeyBytes := x509.MarshalPKCS1PrivateKey(privateKey)

	privateKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: privateKeyBytes,
	})

	return string(privateKeyPEM), nil
}

// PublicKeyFromPEM 从PEM格式字符串解析RSA公钥（公共函数）
func PublicKeyFromPEM(publicKeyPEM string) (*rsa.PublicKey, error) {
	block, _ := pem.Decode([]byte(publicKeyPEM))
	if block == nil {
		return nil, fmt.Errorf("无效的PEM格式公钥")
	}

	publicKey, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("解析公钥失败: %v", err)
	}

	rsaPublicKey, ok := publicKey.(*rsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("不是RSA公钥")
	}

	return rsaPublicKey, nil
}

// PrivateKeyFromPEM 从PEM格式字符串解析RSA私钥（公共函数）
func PrivateKeyFromPEM(privateKeyPEM string) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode([]byte(privateKeyPEM))
	if block == nil {
		return nil, fmt.Errorf("无效的PEM格式私钥")
	}

	privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("解析私钥失败: %v", err)
	}

	return privateKey, nil
}

// GenerateRSAKeyPairPEM 生成RSA密钥对并返回PEM格式字符串（公共函数）
func GenerateRSAKeyPairPEM(bits int) (string, string, error) {
	publicKey, privateKey, err := GenerateRSAKeyPair(bits)
	if err != nil {
		return "", "", err
	}

	publicKeyPEM, err := PublicKeyToPEM(publicKey)
	if err != nil {
		return "", "", err
	}

	privateKeyPEM, err := PrivateKeyToPEM(privateKey)
	if err != nil {
		return "", "", err
	}

	return publicKeyPEM, privateKeyPEM, nil
}
