// Package security 提供 AI API Key 的非对称加密传输和对称加密存储能力。
package security

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"fmt"
)

// CryptoManager 负责 API Key 的传输层 RSA-OAEP 解密和存储层 AES-256-GCM 加解密。
type CryptoManager struct {
	// rsaPrivateKey 是用于解密传输层密文的 RSA 私钥，仅存内存。
	rsaPrivateKey *rsa.PrivateKey
	// masterKey 是 AES-256-GCM 加解密使用的 32 字节对称主密钥。
	masterKey []byte
}

// NewCryptoManager 解码 base64 编码的 AES 主密钥并生成 RSA-2048 密钥对。
func NewCryptoManager(masterKeyB64 string) (*CryptoManager, error) {
	masterKey, err := base64.StdEncoding.DecodeString(masterKeyB64)
	if err != nil {
		return nil, fmt.Errorf("decode master key: %w", err)
	}
	if len(masterKey) != 32 {
		return nil, fmt.Errorf("master key must be 32 bytes after base64 decode, got %d", len(masterKey))
	}
	rsaPrivateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, fmt.Errorf("generate RSA key pair: %w", err)
	}
	return &CryptoManager{rsaPrivateKey: rsaPrivateKey, masterKey: masterKey}, nil
}

// PublicKeyPEM 返回 RSA 公钥的 PEM 编码字符串，供前端加密 API Key 使用。
func (m *CryptoManager) PublicKeyPEM() (string, error) {
	der, err := x509.MarshalPKIXPublicKey(&m.rsaPrivateKey.PublicKey)
	if err != nil {
		return "", fmt.Errorf("marshal public key: %w", err)
	}
	block := pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: der})
	return string(block), nil
}

// DecryptTransport 使用 RSA-OAEP 和 SHA-256 哈希解密前端提交的 base64 编码密文，返回明文 API Key。
func (m *CryptoManager) DecryptTransport(b64Ciphertext string) (string, error) {
	ciphertext, err := base64.StdEncoding.DecodeString(b64Ciphertext)
	if err != nil {
		return "", fmt.Errorf("base64 decode transport ciphertext: %w", err)
	}
	plaintext, err := rsa.DecryptOAEP(sha256.New(), rand.Reader, m.rsaPrivateKey, ciphertext, nil)
	if err != nil {
		return "", fmt.Errorf("RSA-OAEP decrypt: %w", err)
	}
	return string(plaintext), nil
}

// EncryptForStorage 使用 AES-256-GCM 加密明文 API Key，返回 base64 编码的 nonce+密文。
func (m *CryptoManager) EncryptForStorage(plaintext string) (string, error) {
	block, err := aes.NewCipher(m.masterKey)
	if err != nil {
		return "", fmt.Errorf("create AES cipher: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("create GCM: %w", err)
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return "", fmt.Errorf("generate nonce: %w", err)
	}
	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// DecryptFromStorage 解密 EncryptForStorage 产生的 base64 编码密文，返回明文 API Key。
func (m *CryptoManager) DecryptFromStorage(b64Ciphertext string) (string, error) {
	data, err := base64.StdEncoding.DecodeString(b64Ciphertext)
	if err != nil {
		return "", fmt.Errorf("base64 decode storage ciphertext: %w", err)
	}
	block, err := aes.NewCipher(m.masterKey)
	if err != nil {
		return "", fmt.Errorf("create AES cipher: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("create GCM: %w", err)
	}
	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return "", errors.New("ciphertext too short")
	}
	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", fmt.Errorf("AES-GCM decrypt: %w", err)
	}
	return string(plaintext), nil
}
