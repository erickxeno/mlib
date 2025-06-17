package aes

import (
	"testing"
)

func TestAESEncryptDecrypt(t *testing.T) {
	// 测试不同密钥长度
	keyLengths := []int{16, 24, 32}
	testData := []byte("Hello, AES-GCM!")

	for _, keyLen := range keyLengths {
		// 生成测试密钥
		key := make([]byte, keyLen)
		for i := range key {
			key[i] = byte(i % 256)
		}

		// 测试加密
		ciphertext, err := AESEncrypt(key, testData)
		if err != nil {
			t.Errorf("AESEncrypt failed with key length %d: %v", keyLen, err)
			continue
		}

		// 验证密文长度
		if len(ciphertext) <= len(testData) {
			t.Errorf("Ciphertext length should be greater than plaintext length")
		}

		// 测试解密
		plaintext, err := AESDecrypt(key, ciphertext)
		if err != nil {
			t.Errorf("AESDecrypt failed with key length %d: %v", keyLen, err)
			continue
		}

		// 验证解密结果
		if string(plaintext) != string(testData) {
			t.Errorf("Decrypted data doesn't match original data with key length %d", keyLen)
		}
	}
}

func TestCheckAESKey(t *testing.T) {
	tests := []struct {
		name      string
		key       []byte
		wantError bool
	}{
		{
			name:      "Valid 16-byte key",
			key:       make([]byte, 16),
			wantError: false,
		},
		{
			name:      "Valid 24-byte key",
			key:       make([]byte, 24),
			wantError: false,
		},
		{
			name:      "Valid 32-byte key",
			key:       make([]byte, 32),
			wantError: false,
		},
		{
			name:      "Invalid 8-byte key",
			key:       make([]byte, 8),
			wantError: true,
		},
		{
			name:      "Invalid 64-byte key",
			key:       make([]byte, 64),
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := CheckAESKey(tt.key)
			if (err != nil) != tt.wantError {
				t.Errorf("CheckAESKey() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

func TestAESDecryptErrors(t *testing.T) {
	// 使用有效的密钥
	key := make([]byte, 16)
	for i := range key {
		key[i] = byte(i % 256)
	}

	// 测试空密文
	_, err := AESDecrypt(key, []byte{})
	if err == nil {
		t.Error("AESDecrypt should fail with empty ciphertext")
	}

	// 测试过短的密文
	_, err = AESDecrypt(key, []byte{1, 2, 3})
	if err == nil {
		t.Error("AESDecrypt should fail with too short ciphertext")
	}

	// 测试无效的密文
	invalidCiphertext := make([]byte, 32)
	_, err = AESDecrypt(key, invalidCiphertext)
	if err == nil {
		t.Error("AESDecrypt should fail with invalid ciphertext")
	}
}

func TestAESEncryptErrors(t *testing.T) {
	// 测试无效密钥长度
	invalidKey := make([]byte, 8)
	_, err := AESEncrypt(invalidKey, []byte("test"))
	if err == nil {
		t.Error("AESEncrypt should fail with invalid key length")
	}

	// 测试空明文
	validKey := make([]byte, 16)
	_, err = AESEncrypt(validKey, []byte{})
	if err != nil {
		t.Error("AESEncrypt should succeed with empty plaintext")
	}
}

func TestAESEncryptDecryptWithEmptyData(t *testing.T) {
	key := make([]byte, 16)
	for i := range key {
		key[i] = byte(i % 256)
	}

	// 测试空数据
	emptyData := []byte{}
	ciphertext, err := AESEncrypt(key, emptyData)
	if err != nil {
		t.Errorf("AESEncrypt failed with empty data: %v", err)
	}

	plaintext, err := AESDecrypt(key, ciphertext)
	if err != nil {
		t.Errorf("AESDecrypt failed with empty data: %v", err)
	}

	if len(plaintext) != 0 {
		t.Error("Decrypted empty data should be empty")
	}
}

func TestAESEncryptDecryptWithLargeData(t *testing.T) {
	key := make([]byte, 16)
	for i := range key {
		key[i] = byte(i % 256)
	}

	// 测试大数据
	largeData := make([]byte, 1024*1024) // 1MB
	for i := range largeData {
		largeData[i] = byte(i % 256)
	}

	ciphertext, err := AESEncrypt(key, largeData)
	if err != nil {
		t.Errorf("AESEncrypt failed with large data: %v", err)
	}

	plaintext, err := AESDecrypt(key, ciphertext)
	if err != nil {
		t.Errorf("AESDecrypt failed with large data: %v", err)
	}

	if len(plaintext) != len(largeData) {
		t.Error("Decrypted large data length doesn't match original")
	}

	for i := range plaintext {
		if plaintext[i] != largeData[i] {
			t.Error("Decrypted large data content doesn't match original")
			break
		}
	}
}
