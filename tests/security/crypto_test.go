package security_test

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"strings"
	"testing"

	"github.com/nixwai/go-game-server/app/security"
)

// validMasterKeyB64 返回一个合法的 base64 编码 32 字节主密钥。
func validMasterKeyB64() string {
	key := make([]byte, 32)
	for i := range key {
		key[i] = byte(i + 1)
	}
	return base64.StdEncoding.EncodeToString(key)
}

func TestNewCryptoManagerValidMasterKey(t *testing.T) {
	cm, err := security.NewCryptoManager(validMasterKeyB64())
	if err != nil {
		t.Fatalf("expected success, got error: %v", err)
	}
	if cm == nil {
		t.Fatal("crypto manager should not be nil")
	}
}

func TestNewCryptoManagerInvalidBase64(t *testing.T) {
	_, err := security.NewCryptoManager("not-valid-base64!!!")
	if err == nil {
		t.Fatal("expected error for invalid base64")
	}
}

func TestNewCryptoManagerWrongKeyLength(t *testing.T) {
	shortKey := base64.StdEncoding.EncodeToString([]byte("too-short"))
	_, err := security.NewCryptoManager(shortKey)
	if err == nil {
		t.Fatal("expected error for short master key")
	}
}

func TestPublicKeyPEM(t *testing.T) {
	cm, _ := security.NewCryptoManager(validMasterKeyB64())
	pemStr, err := cm.PublicKeyPEM()
	if err != nil {
		t.Fatalf("expected success, got error: %v", err)
	}
	if !strings.Contains(pemStr, "BEGIN PUBLIC KEY") {
		t.Fatalf("PEM should contain BEGIN PUBLIC KEY, got: %s", pemStr)
	}
}

func TestRSAEncryptDecryptTransport(t *testing.T) {
	cm, _ := security.NewCryptoManager(validMasterKeyB64())
	pemStr, _ := cm.PublicKeyPEM()

	// 解析 PEM 公钥。
	block, _ := pem.Decode([]byte(pemStr))
	if block == nil {
		t.Fatal("failed to decode PEM block")
	}
	pubKey, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		t.Fatalf("parse public key: %v", err)
	}
	rsaPub, ok := pubKey.(*rsa.PublicKey)
	if !ok {
		t.Fatal("not an RSA public key")
	}

	plaintext := "sk-test-api-key-12345"
	// RSA-OAEP 加密。
	ciphertext, err := rsa.EncryptOAEP(sha256.New(), rand.Reader, rsaPub, []byte(plaintext), nil)
	if err != nil {
		t.Fatalf("RSA encrypt: %v", err)
	}
	b64Cipher := base64.StdEncoding.EncodeToString(ciphertext)

	// 服务端解密。
	decrypted, err := cm.DecryptTransport(b64Cipher)
	if err != nil {
		t.Fatalf("decrypt transport: %v", err)
	}
	if decrypted != plaintext {
		t.Fatalf("expected %q, got %q", plaintext, decrypted)
	}
}

func TestDecryptTransportInvalidBase64(t *testing.T) {
	cm, _ := security.NewCryptoManager(validMasterKeyB64())
	_, err := cm.DecryptTransport("not-valid-base64!!!")
	if err == nil {
		t.Fatal("expected error for invalid base64")
	}
}

func TestDecryptTransportInvalidCiphertext(t *testing.T) {
	cm, _ := security.NewCryptoManager(validMasterKeyB64())
	_, err := cm.DecryptTransport(base64.StdEncoding.EncodeToString([]byte("garbage")))
	if err == nil {
		t.Fatal("expected error for invalid ciphertext")
	}
}

func TestAESEncryptDecryptStorage(t *testing.T) {
	cm, _ := security.NewCryptoManager(validMasterKeyB64())
	plaintext := "sk-my-secret-api-key"
	encrypted, err := cm.EncryptForStorage(plaintext)
	if err != nil {
		t.Fatalf("encrypt for storage: %v", err)
	}
	if encrypted == plaintext {
		t.Fatal("ciphertext should differ from plaintext")
	}
	decrypted, err := cm.DecryptFromStorage(encrypted)
	if err != nil {
		t.Fatalf("decrypt from storage: %v", err)
	}
	if decrypted != plaintext {
		t.Fatalf("expected %q, got %q", plaintext, decrypted)
	}
}

func TestAESEncryptProducesDifferentCiphertexts(t *testing.T) {
	cm, _ := security.NewCryptoManager(validMasterKeyB64())
	plaintext := "sk-same-key"
	enc1, _ := cm.EncryptForStorage(plaintext)
	enc2, _ := cm.EncryptForStorage(plaintext)
	if enc1 == enc2 {
		t.Fatal("same plaintext should produce different ciphertexts due to random nonce")
	}
}

func TestDecryptFromStorageInvalidBase64(t *testing.T) {
	cm, _ := security.NewCryptoManager(validMasterKeyB64())
	_, err := cm.DecryptFromStorage("not-valid-base64!!!")
	if err == nil {
		t.Fatal("expected error for invalid base64")
	}
}

func TestDecryptFromStorageTooShort(t *testing.T) {
	cm, _ := security.NewCryptoManager(validMasterKeyB64())
	_, err := cm.DecryptFromStorage(base64.StdEncoding.EncodeToString([]byte("short")))
	if err == nil {
		t.Fatal("expected error for too-short ciphertext")
	}
}

func TestDecryptFromStorageTamperedCiphertext(t *testing.T) {
	cm, _ := security.NewCryptoManager(validMasterKeyB64())
	encrypted, _ := cm.EncryptForStorage("sk-original")
	// 篡改密文。
	data, _ := base64.StdEncoding.DecodeString(encrypted)
	if len(data) > 0 {
		data[0] ^= 0xFF
	}
	tampered := base64.StdEncoding.EncodeToString(data)
	_, err := cm.DecryptFromStorage(tampered)
	if err == nil {
		t.Fatal("expected error for tampered ciphertext")
	}
}
