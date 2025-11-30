package common

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
)

const HashHeader = "HashSHA256"

func GetHashData(data []byte, key string) string {
	h := hmac.New(sha256.New, []byte(key))
	h.Write(data)

	return hex.EncodeToString(h.Sum(nil))
}
