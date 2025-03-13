package crypto

import "golang.org/x/crypto/bcrypt"

type Bcrypt struct{}

func (b *Bcrypt) i() {}

// Encrypt 加密
func (b *Bcrypt) Encrypt(encryptStr string) string {
	hash, _ := bcrypt.GenerateFromPassword([]byte(encryptStr), bcrypt.MinCost)
	return string(hash)
}

func (b *Bcrypt) Decrypt(decryptStr string) string {
	return ""
}

func (b *Bcrypt) Compare(hash string, pwd string) bool {
	if err := bcrypt.CompareHashAndPassword([]byte(pwd), []byte(hash)); err != nil {
		return false
	}
	return true
}
