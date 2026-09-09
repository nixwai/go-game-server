// Package security 提供密码哈希和 JWT 相关安全能力。
package security

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
)

// PasswordHasher 封装 Argon2id 的参数和哈希校验逻辑。
type PasswordHasher struct {
	// Time 是 Argon2id 的时间成本。
	Time uint32
	// Memory 是 Argon2id 的内存成本，单位为 KiB。
	Memory uint32
	// Threads 是 Argon2id 的并行度。
	Threads uint8
	// KeyLen 是输出哈希的字节长度。
	KeyLen uint32
	// SaltLen 是随机盐的字节长度。
	SaltLen uint32
}

// Hash 为密码生成随机盐，并返回 PHC 格式的 Argon2id 哈希字符串。
func (h PasswordHasher) Hash(password string) (string, error) {
	// 为每个密码生成独立随机盐。
	salt := make([]byte, h.SaltLen)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("generate password salt: %w", err)
	}
	hash := argon2.IDKey([]byte(password), salt, h.Time, h.Memory, h.Threads, h.KeyLen)
	return fmt.Sprintf("$argon2id$v=19$m=%d,t=%d,p=%d$%s$%s", h.Memory, h.Time, h.Threads, b64(salt), b64(hash)), nil
}

// Compare 根据哈希字符串中的参数重新计算密码，并使用常量时间比较结果。
func (h PasswordHasher) Compare(password, encoded string) bool {
	// 解析 PHC 格式：算法、版本、参数、盐和哈希值共六段。
	parts := strings.Split(encoded, "$")
	if len(parts) != 6 || parts[1] != "argon2id" || parts[2] != "v=19" {
		return false
	}
	var memory, timeCost uint32
	var threads uint8
	if _, err := fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &memory, &timeCost, &threads); err != nil {
		return false
	}
	salt, err1 := base64.RawStdEncoding.DecodeString(parts[4])
	expected, err2 := base64.RawStdEncoding.DecodeString(parts[5])
	if err1 != nil || err2 != nil || len(expected) == 0 {
		return false
	}
	actual := argon2.IDKey([]byte(password), salt, timeCost, memory, threads, uint32(len(expected)))
	return subtle.ConstantTimeCompare(actual, expected) == 1
}

// b64 使用不带填充的 Base64 编码，生成紧凑且可移植的 PHC 字段。
func b64(value []byte) string { return base64.RawStdEncoding.EncodeToString(value) }
