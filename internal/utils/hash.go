package utils

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
)

func CalculateHash(data []byte, key string) string {
	if key == "" {
		return ""
	}
	h := hmac.New(sha256.New, []byte(key))
	h.Write(data)
	return hex.EncodeToString(h.Sum(nil))
}

func VerifyHash(data []byte, key, expectedHash string) bool {
	if key == "" || expectedHash == "" {
		return true
	}
	return hmac.Equal([]byte(CalculateHash(data, key)), []byte(expectedHash))
}
