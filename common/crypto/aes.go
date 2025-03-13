package crypto

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"errors"
)

type AES struct {
	key string
}

func (b *AES) i() {}

func NewAES(key string) *AES {
	return &AES{key}
}

// Encrypt 加密
func (b *AES) Encrypt(encryptStr string) string {
	key := []byte(b.key)
	//创建加密实例
	block, err := aes.NewCipher(key)
	if err != nil {
		panic(err)
	}
	//判断加密快的大小
	blockSize := block.BlockSize()
	//填充
	encryptBytes := pkcs7Padding([]byte(encryptStr), blockSize)
	//初始化加密数据接收切片
	crypted := make([]byte, len(encryptBytes))
	//使用cbc加密模式
	blockMode := cipher.NewCBCEncrypter(block, key[:blockSize])
	//执行加密
	blockMode.CryptBlocks(crypted, encryptBytes)
	return base64.StdEncoding.EncodeToString(crypted)
}

func (b *AES) Decrypt(decryptStr string) string {
	key := []byte(b.key)
	crypted, err := base64.StdEncoding.DecodeString(decryptStr)
	if err != nil {
		panic(err)
	}

	//创建实例
	block, err := aes.NewCipher(key)
	if err != nil {
		panic(err)
	}
	//获取块的大小
	blockSize := block.BlockSize()
	//使用cbc
	blockMode := cipher.NewCBCDecrypter(block, key[:blockSize])
	//初始化解密数据接收切片
	original := make([]byte, len(crypted))
	//执行解密
	blockMode.CryptBlocks(original, crypted)
	//去除填充
	original, err = pkcs7UnPadding(original)
	if err != nil {
		panic(err)
	}

	return string(original)
}

func (b *AES) Compare(hash string, pwd string) bool {
	return true
}

//pkcs7Padding 填充
func pkcs7Padding(data []byte, blockSize int) []byte {
	//判断缺少几位长度。最少1，最多 blockSize
	padding := blockSize - len(data)%blockSize
	//补足位数。把切片[]byte{byte(padding)}复制padding个
	padText := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(data, padText...)
}

//pkcs7UnPadding 填充的反向操作
func pkcs7UnPadding(data []byte) ([]byte, error) {
	length := len(data)
	if length == 0 {
		return nil, errors.New("加密字符串错误！")
	}
	//获取填充的个数
	unPadding := int(data[length-1])
	return data[:(length - unPadding)], nil
}
