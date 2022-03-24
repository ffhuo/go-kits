package crypto

import (
	cryptoMD5 "crypto/md5"
	"encoding/hex"
)

type MD5 struct{}

func (b *MD5) i() {}

// Encrypt 加密
func (b *MD5) Encrypt(encryptStr string) string {
	s := cryptoMD5.New()
	s.Write([]byte(encryptStr))
	return hex.EncodeToString(s.Sum(nil))
}

func (b *MD5) Decrypt(decryptStr string) string {
	return ""
}

func (b *MD5) Compare(hash string, pwd string) bool {
	return true
}
