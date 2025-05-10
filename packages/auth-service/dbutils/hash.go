package dbutils

import (
	"crypto/sha256"
	"encoding/hex"
)

func HashString(src string) string {
	hash := sha256.New()
	hash.Write([]byte(src))
	return hex.EncodeToString(hash.Sum(nil))
}
