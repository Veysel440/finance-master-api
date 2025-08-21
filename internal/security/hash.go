package security

import (
	"crypto/sha256"
	"encoding/hex"

	"github.com/alexedwards/argon2id"
)

func ArgonHash(pw string) (string, error)      { return argon2id.CreateHash(pw, argon2id.DefaultParams) }
func ArgonCheck(pw, hash string) (bool, error) { return argon2id.ComparePasswordAndHash(pw, hash) }

func SHA256Hex(s string) string {
	h := sha256.Sum256([]byte(s))
	return hex.EncodeToString(h[:])
}
