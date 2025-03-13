package crypto

type Crypto interface {
	i()
	// Encrypt 加密
	Encrypt(encryptStr string) string
	Decrypt(decryptStr string) string
	Compare(dst, src string) bool
}
