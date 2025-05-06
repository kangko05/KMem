package utils

import (
	"crypto/md5"
	"encoding/hex"
	"io"
)

func MD5(strs ...string) string {
	h := md5.New()

	for _, s := range strs {
		io.WriteString(h, s)
	}

	return hex.EncodeToString(h.Sum(nil))
}
