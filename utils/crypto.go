package utils

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
)

func BinaryToHexString(buffer []byte, max int) string {
	length := len(buffer) 
	if(length == 0) {
		return ""
	}

	if(max >= 0 && max <= length) {
		length = max
	}

	text := ""
	for i:= 0; i < length; i ++ {
		text += fmt.Sprintf("%02X", buffer[i])
		if(i + 1 < length) {
			text += " "
		}
	}
	return text
}

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
