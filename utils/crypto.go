package utils

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/hex"
	"strings"
)

func MD5(text string) string {
	hash := md5.New()
	_, err := hash.Write([]byte(text))
	if err != nil {
		return ""
	}
	return strings.ToUpper(hex.EncodeToString(hash.Sum(nil)))
}

func SHA1(text string) string {
	hash := sha1.New()
	_, err := hash.Write([]byte(text))
	if err != nil {
		return ""
	}
	return strings.ToUpper(hex.EncodeToString(hash.Sum(nil)))
}

func SHA256(text string) string {
	hash := sha256.New()
	_, err := hash.Write([]byte(text))
	if err != nil {
		return ""
	}
	return strings.ToUpper(hex.EncodeToString(hash.Sum(nil)))
}
