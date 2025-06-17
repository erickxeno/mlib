package aes

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"errors"
	"io"
)

// AESEncrypt 使用 AES-GCM 模式加密数据
// 参数:
//   - key: 加密密钥，长度必须是 16、24 或 32 字节
//   - plaintext: 需要加密的明文数据
//
// 返回:
//   - []byte: 加密后的密文（包含 nonce）
//   - error: 如果加密过程中发生错误则返回错误信息
func AESEncrypt(key, plaintext []byte) ([]byte, error) {
	c, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(c)
	if err != nil {
		return nil, err
	}

	nonceSize := gcm.NonceSize()
	nonce := make([]byte, nonceSize)
	_, err = io.ReadFull(rand.Reader, nonce)
	if err != nil {
		return nil, err
	}

	ciphertext := gcm.Seal(nonce, nonce, plaintext, nil)
	return ciphertext, nil
}

// AESDecrypt 使用 AES-GCM 模式解密数据
// 参数:
//   - key: 解密密钥，长度必须是 16、24 或 32 字节
//   - ciphertext: 需要解密的密文（包含 nonce）
//
// 返回:
//   - []byte: 解密后的明文
//   - error: 如果解密过程中发生错误则返回错误信息
func AESDecrypt(key, ciphertext []byte) ([]byte, error) {
	c, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(c)
	if err != nil {
		return nil, err
	}

	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, errors.New("length of ciphertext less then nonce size")
	}
	nonce := make([]byte, nonceSize)
	copy(nonce, ciphertext)

	plaintext, err := gcm.Open(nil, nonce, ciphertext[nonceSize:], nil)
	if err != nil {
		return nil, err
	}
	return plaintext, nil
}

// CheckAESKey 验证 AES 密钥长度是否合法
// 参数:
//   - key: 待验证的密钥
//
// 返回:
//   - error: 如果密钥长度不是 16、24 或 32 字节则返回错误信息
func CheckAESKey(key []byte) error {
	switch len(key) {
	case 16, 24, 32:
		return nil
	default:
		return errors.New("invalid AES key length, allow 16|24|32.")
	}
}
