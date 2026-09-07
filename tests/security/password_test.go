package security_test

import (
	"strings"
	"testing"

	"github.com/nixwai/go-game-server/app/security"
)

func testHasher() security.PasswordHasher {
	return security.PasswordHasher{Time: 1, Memory: 32 * 1024, Threads: 1, KeyLen: 32, SaltLen: 16}
}
func TestPasswordHasherRoundTrip(t *testing.T) {
	h := testHasher()
	encoded, err := h.Hash("SecurePass123")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(encoded, "$argon2id$") {
		t.Fatalf("unexpected encoding: %s", encoded)
	}
	if !h.Compare("SecurePass123", encoded) {
		t.Fatal("password should match")
	}
	if h.Compare("WrongPass123", encoded) {
		t.Fatal("wrong password should not match")
	}
}
func TestPasswordHasherRejectsMalformedHash(t *testing.T) {
	if testHasher().Compare("SecurePass123", "not-a-hash") {
		t.Fatal("malformed hash should fail")
	}
}
