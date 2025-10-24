package encrypt

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"errors"
	"fmt"
)

// RSADynamicEncrypt RAS动态加密算法结构体
type RSADynamicEncrypt struct {
	publicKey  *rsa.PublicKey
	privateKey *rsa.PrivateKey
}

// NewRSADynamicEncrypt 创建新的RAS动态加密实例
func NewRSADynamicEncrypt(publicKeyPEM, privateKeyPEM string) (*RSADynamicEncrypt, error) {
	var pubKey *rsa.PublicKey
	var privKey *rsa.PrivateKey
	var err error

	// 解析公钥
	if publicKeyPEM != "" {
		pubKey, err = PublicKeyFromPEM(publicKeyPEM) // 使用公共函数
		if err != nil {
			return nil, fmt.Errorf("failed to parse public key: %v", err)
		}
	}

	// 解析私钥
	if privateKeyPEM != "" {
		privKey, err = PrivateKeyFromPEM(privateKeyPEM) // 使用公共函数
		if err != nil {
			return nil, fmt.Errorf("failed to parse private key: %v", err)
		}
	}

	return &RSADynamicEncrypt{
		publicKey:  pubKey,
		privateKey: privKey,
	}, nil
}

// GenerateRSADynamicKeyPair 生成RSA动态加密密钥对
func GenerateRSADynamicKeyPair(bits int) (string, string, error) {
	return GenerateRSAKeyPairPEM(bits) // 使用公共函数
}

// generateDynamicKeys 生成动态密钥
func generateDynamicKeys() ([]byte, error) {
	// 生成3-6个随机密钥长度
	var lengthByte [1]byte
	if _, err := rand.Read(lengthByte[:]); err != nil {
		return nil, err
	}
	keyLen := 3 + int(lengthByte[0])%4 // 3-6个密钥

	keys := make([]byte, keyLen)
	for i := 0; i < keyLen; i++ {
		var keyByte [1]byte
		if _, err := rand.Read(keyByte[:]); err != nil {
			return nil, err
		}
		// 确保密钥在1-255范围内
		keys[i] = keyByte[0]
		if keys[i] == 0 {
			keys[i] = 1
		}
	}

	return keys, nil
}

// xorEncrypt 使用动态密钥进行XOR加密
func xorEncrypt(data []byte, keys []byte) []byte {
	result := make([]byte, len(data))
	copy(result, data)

	for _, key := range keys {
		for i := range result {
			result[i] ^= key
		}
	}

	return result
}

// xorDecrypt 使用动态密钥进行XOR解密（与加密相同）
func xorDecrypt(data []byte, keys []byte) []byte {
	return xorEncrypt(data, keys) // XOR解密与加密相同
}

// Encrypt RAS动态加密
func (r *RSADynamicEncrypt) Encrypt(plaintext string) (string, error) {
	if r.publicKey == nil {
		return "", errors.New("public key not set")
	}

	// 1. 生成动态密钥
	dynamicKeys, err := generateDynamicKeys()
	if err != nil {
		return "", fmt.Errorf("failed to generate dynamic keys: %v", err)
	}

	// 2. 使用动态密钥对明文进行XOR加密
	plaintextBytes := []byte(plaintext)
	xorEncrypted := xorEncrypt(plaintextBytes, dynamicKeys)

	// 3. 构造最终数据：密钥长度 + 密钥 + 加密数据
	finalData := make([]byte, 0, 1+len(dynamicKeys)+len(xorEncrypted))
	finalData = append(finalData, byte(len(dynamicKeys))) // 密钥长度

	// 按逆序插入密钥（与C++代码保持一致）
	for i := len(dynamicKeys) - 1; i >= 0; i-- {
		finalData = append(finalData, dynamicKeys[i])
	}

	finalData = append(finalData, xorEncrypted...) // 加密数据

	// 4. 使用RSA公钥加密
	rsaEncrypted, err := rsa.EncryptPKCS1v15(rand.Reader, r.publicKey, finalData)
	if err != nil {
		return "", fmt.Errorf("RSA encryption failed: %v", err)
	}

	// 5. Base64编码
	return base64.StdEncoding.EncodeToString(rsaEncrypted), nil
}

// Decrypt RAS动态解密
func (r *RSADynamicEncrypt) Decrypt(ciphertext string) (string, error) {
	if r.privateKey == nil {
		return "", errors.New("private key not set")
	}

	// 1. Base64解码
	rsaEncrypted, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", fmt.Errorf("base64 decode failed: %v", err)
	}

	// 2. 使用RSA私钥解密
	decryptedData, err := rsa.DecryptPKCS1v15(rand.Reader, r.privateKey, rsaEncrypted)
	if err != nil {
		return "", fmt.Errorf("RSA decryption failed: %v", err)
	}

	if len(decryptedData) < 1 {
		return "", errors.New("decrypted data too short")
	}

	// 3. 提取密钥长度
	keyLen := int(decryptedData[0])
	if len(decryptedData) < 1+keyLen {
		return "", errors.New("invalid decrypted data format")
	}

	// 4. 提取动态密钥（按逆序存储的）
	dynamicKeys := make([]byte, keyLen)
	for i := 0; i < keyLen; i++ {
		dynamicKeys[keyLen-1-i] = decryptedData[1+i] // 恢复原始顺序
	}

	// 5. 提取XOR加密的数据
	xorEncryptedData := decryptedData[1+keyLen:]

	// 6. 使用动态密钥进行XOR解密
	plaintext := xorDecrypt(xorEncryptedData, dynamicKeys)

	return string(plaintext), nil
}

// EncryptWithKeys 使用指定的公钥进行RAS动态加密
func EncryptWithKeys(plaintext, publicKeyPEM string) (string, error) {
	rsaEncrypt, err := NewRSADynamicEncrypt(publicKeyPEM, "")
	if err != nil {
		return "", err
	}
	return rsaEncrypt.Encrypt(plaintext)
}

// DecryptWithKeys 使用指定的私钥进行RAS动态解密
func DecryptWithKeys(ciphertext, privateKeyPEM string) (string, error) {
	rsaEncrypt, err := NewRSADynamicEncrypt("", privateKeyPEM)
	if err != nil {
		return "", err
	}
	return rsaEncrypt.Decrypt(ciphertext)
}
