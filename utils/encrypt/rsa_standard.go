package encrypt

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
)

// RSAEncrypt 普通RSA加密算法结构体
type RSAEncrypt struct {
	publicKey  *rsa.PublicKey
	privateKey *rsa.PrivateKey
}

// NewRSAEncrypt 创建新的RSA加密实例
func NewRSAEncrypt(publicKey *rsa.PublicKey, privateKey *rsa.PrivateKey) *RSAEncrypt {
	return &RSAEncrypt{
		publicKey:  publicKey,
		privateKey: privateKey,
	}
}

// Encrypt RSA公钥加密
func (r *RSAEncrypt) Encrypt(plaintext string) (string, error) {
	if r.publicKey == nil {
		return "", fmt.Errorf("RSA公钥不能为空")
	}

	data := []byte(plaintext)

	// 使用OAEP填充进行加密
	encrypted, err := rsa.EncryptOAEP(sha256.New(), rand.Reader, r.publicKey, data, nil)
	if err != nil {
		return "", fmt.Errorf("RSA加密失败: %v", err)
	}

	// Base64编码
	return base64.StdEncoding.EncodeToString(encrypted), nil
}

// Decrypt RSA私钥解密
func (r *RSAEncrypt) Decrypt(ciphertext string) (string, error) {
	if r.privateKey == nil {
		return "", fmt.Errorf("RSA私钥不能为空")
	}

	// Base64解码
	data, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", fmt.Errorf("Base64解码失败: %v", err)
	}

	// 使用OAEP填充进行解密
	decrypted, err := rsa.DecryptOAEP(sha256.New(), rand.Reader, r.privateKey, data, nil)
	if err != nil {
		return "", fmt.Errorf("RSA解密失败: %v", err)
	}

	return string(decrypted), nil
}

// EncryptLargeData RSA分块加密大数据
func (r *RSAEncrypt) EncryptLargeData(plaintext string) (string, error) {
	if r.publicKey == nil {
		return "", fmt.Errorf("RSA公钥不能为空")
	}

	data := []byte(plaintext)
	keySize := r.publicKey.Size()
	blockSize := keySize - 2*sha256.Size - 2 // OAEP填充的最大明文长度

	var encrypted []byte

	for i := 0; i < len(data); i += blockSize {
		end := i + blockSize
		if end > len(data) {
			end = len(data)
		}

		block := data[i:end]
		encryptedBlock, err := rsa.EncryptOAEP(sha256.New(), rand.Reader, r.publicKey, block, nil)
		if err != nil {
			return "", fmt.Errorf("RSA分块加密失败: %v", err)
		}

		encrypted = append(encrypted, encryptedBlock...)
	}

	return base64.StdEncoding.EncodeToString(encrypted), nil
}

// DecryptLargeData RSA分块解密大数据
func (r *RSAEncrypt) DecryptLargeData(ciphertext string) (string, error) {
	if r.privateKey == nil {
		return "", fmt.Errorf("RSA私钥不能为空")
	}

	data, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", fmt.Errorf("Base64解码失败: %v", err)
	}

	keySize := r.privateKey.Size()
	var decrypted []byte

	for i := 0; i < len(data); i += keySize {
		end := i + keySize
		if end > len(data) {
			end = len(data)
		}

		block := data[i:end]
		decryptedBlock, err := rsa.DecryptOAEP(sha256.New(), rand.Reader, r.privateKey, block, nil)
		if err != nil {
			return "", fmt.Errorf("RSA分块解密失败: %v", err)
		}

		decrypted = append(decrypted, decryptedBlock...)
	}

	return string(decrypted), nil
}
