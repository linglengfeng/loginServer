package crypto

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"io"
	"os"
	"strings"
)

const (
	modcbc = "cbc"
	modcfb = "cfb"

	pub_key = "gZ23qm6YdCOQVLrpHuJ91kABhbP7lDFE"
	pub_iv  = "GZwm4tUJTdXfKeIy"
	pub_mod = modcbc

	v2Prefix = "v2:"
)

func Encrypt(info ...string) (string, error) {
	if len(info) == 0 {
		return "", errors.New("encrypt: missing plaintext")
	}
	argLen := len(info)
	val := info[0]

	// v2：AES-GCM（带认证）+ 随机 nonce
	// key 获取优先级：函数参数 > 环境变量 > 旧默认 key（兼容）
	key := getCryptoKey()
	if argLen > 1 && info[1] != "" {
		key = info[1]
	}
	if len(key) != 16 && len(key) != 24 && len(key) != 32 {
		return "", errors.New("encrypt: invalid key length (need 16/24/32 bytes)")
	}

	enc, err := aesGCMEncrypt([]byte(val), []byte(key))
	if err != nil {
		return "", err
	}
	return v2Prefix + base64.StdEncoding.EncodeToString(enc), nil
	// return string(enval), erren
}

func Decrypt(info ...string) (string, error) {
	if len(info) == 0 {
		return "", errors.New("decrypt: missing ciphertext")
	}
	argLen := len(info)
	raw := info[0]

	// v2：AES-GCM
	if strings.HasPrefix(raw, v2Prefix) {
		b, err := base64.StdEncoding.DecodeString(strings.TrimPrefix(raw, v2Prefix))
		if err != nil {
			return "", err
		}
		key := getCryptoKey()
		if argLen > 1 && info[1] != "" {
			key = info[1]
		}
		if len(key) != 16 && len(key) != 24 && len(key) != 32 {
			return "", errors.New("decrypt: invalid key length (need 16/24/32 bytes)")
		}
		dec, err := aesGCMDecrypt(b, []byte(key))
		if err != nil {
			return "", err
		}
		return string(dec), nil
	}

	// v1：兼容旧格式（base64 + AES-CBC/CFB + 固定 iv）
	val, err := base64.StdEncoding.DecodeString(raw)
	if err != nil {
		return "", err
	}
	key := pub_key
	iv := pub_iv
	mod := pub_mod
	if argLen > 1 {
		key = info[1]
	}
	if argLen > 2 {
		iv = info[2]
	}
	if argLen > 3 {
		mod = info[3]
	}
	deval, erren := AesDecrypt(val, []byte(key), []byte(iv), mod)
	return string(deval), erren
}

func getCryptoKey() string {
	// 部署时建议设置环境变量，避免固定 key
	if v := os.Getenv("LOGIN_SERVER_CRYPTO_KEY"); v != "" {
		return v
	}
	// 兼容旧逻辑：不设置时使用旧默认 key
	return pub_key
}

// aesGCMEncrypt 返回：nonce || ciphertext（ciphertext 含 tag）
func aesGCMEncrypt(plaintext, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}
	ciphertext := gcm.Seal(nil, nonce, plaintext, nil)
	out := make([]byte, 0, len(nonce)+len(ciphertext))
	out = append(out, nonce...)
	out = append(out, ciphertext...)
	return out, nil
}

func aesGCMDecrypt(payload, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	ns := gcm.NonceSize()
	if len(payload) < ns {
		return nil, errors.New("decrypt: invalid payload")
	}
	nonce := payload[:ns]
	ciphertext := payload[ns:]
	return gcm.Open(nil, nonce, ciphertext, nil)
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
	// PKCS#7 padding 校验：1 <= unPadding <= blockSize 且尾部字节都相同
	if unPadding <= 0 || unPadding > length {
		return nil, errors.New("加密字符串错误！")
	}
	for i := 0; i < unPadding; i++ {
		if data[length-1-i] != byte(unPadding) {
			return nil, errors.New("加密字符串错误！")
		}
	}
	return data[:(length - unPadding)], nil
}

func AesEncrypt(data, key, iv []byte, mod string) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	blockSize := block.BlockSize()
	if len(iv) != blockSize {
		return nil, errors.New("invalid iv length")
	}
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
	blockSize := block.BlockSize()
	if len(iv) != blockSize {
		return nil, errors.New("invalid iv length")
	}
	if mod == modcbc && len(data)%blockSize != 0 {
		return nil, errors.New("invalid ciphertext length")
	}
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
