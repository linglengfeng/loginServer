package crypto

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"errors"
)

const (
	modcbc = "cbc"
	modcfb = "cfb"

	pub_key = "gZ23qm6YdCOQVLrpHuJ91kABhbP7lDFE"
	pub_iv  = "GZwm4tUJTdXfKeIy"
	pub_mod = modcbc
)

func Encrypt(info ...string) (string, error) {
	len := len(info)
	val := info[0]
	key := pub_key
	iv := pub_iv
	mod := pub_mod
	if len > 1 {
		key = info[1]
	}
	if len > 2 {
		iv = info[2]
	}
	if len > 3 {
		mod = info[3]
	}
	enval, erren := AesEncrypt([]byte(val), []byte(key), []byte(iv), mod)
	return base64.StdEncoding.EncodeToString(enval), erren
	// return string(enval), erren
}

func Decrypt(info ...string) (string, error) {
	len := len(info)
	val, errdestr := base64.StdEncoding.DecodeString(info[0])
	if errdestr != nil {
		return "", errdestr
	}
	// val := []byte(info[0])
	key := pub_key
	iv := pub_iv
	mod := pub_mod
	if len > 1 {
		key = info[1]
	}
	if len > 2 {
		iv = info[2]
	}
	if len > 3 {
		mod = info[3]
	}
	deval, erren := AesDecrypt(val, []byte(key), []byte(iv), mod)
	return string(deval), erren
}

func pkcs7Padding(data []byte, blockSize int) []byte {
	padding := blockSize - len(data)%blockSize
	padText := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(data, padText...)
}

func pkcs7UnPadding(data []byte) ([]byte, error) {
	length := len(data)
	if length == 0 {
		return nil, errors.New("加密字符串错误！")
	}
	unPadding := int(data[length-1])
	return data[:(length - unPadding)], nil
}

func AesEncrypt(data, key, iv []byte, mod string) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	blockSize := block.BlockSize()
	encryptBytes := pkcs7Padding(data, blockSize)
	crypted := make([]byte, len(encryptBytes))
	switch mod {
	case modcbc:
		blockMode := cipher.NewCBCEncrypter(block, iv)
		// blockMode := cipher.NewCBCEncrypter(block, key[:blockSize])
		blockMode.CryptBlocks(crypted, encryptBytes)
	case modcfb:
		blockMode := cipher.NewCFBEncrypter(block, iv)
		blockMode.XORKeyStream(crypted, encryptBytes)
	default:
		blockMode := cipher.NewCBCEncrypter(block, iv)
		blockMode.CryptBlocks(crypted, encryptBytes)
	}
	return crypted, nil
}

func AesDecrypt(data, key, iv []byte, mod string) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	// blockSize := block.BlockSize()
	crypted := make([]byte, len(data))
	switch mod {
	case modcbc:
		blockMode := cipher.NewCBCDecrypter(block, iv)
		// blockMode := cipher.NewCBCDecrypter(block, key[:blockSize])
		blockMode.CryptBlocks(crypted, data)
	case modcfb:
		blockMode := cipher.NewCFBDecrypter(block, iv)
		blockMode.XORKeyStream(crypted, data)
	default:
		blockMode := cipher.NewCBCDecrypter(block, iv)
		blockMode.CryptBlocks(crypted, data)
	}
	crypted, err = pkcs7UnPadding(crypted)
	if err != nil {
		return nil, err
	}
	return crypted, nil
}
