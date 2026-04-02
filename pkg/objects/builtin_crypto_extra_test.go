// pkg/objects/builtin_crypto_extra_test.go
package objects

import (
	"testing"
)

func TestObjectsGetRandomByte(t *testing.T) {
	b := getRandomByte()
	if b > 255 {
		t.Errorf("getRandomByte returned %d, expected 0-255", b)
	}
}

func TestObjectsSumBytes(t *testing.T) {
	data := []byte{1, 2, 3, 4, 5}
	sum := sumBytes(data)
	if sum != 15 {
		t.Errorf("sumBytes = %d, expected 15", sum)
	}
}

func TestObjectsPkcs7Padding(t *testing.T) {
	data := []byte("test")
	blockSize := 16
	padded := pkcs7Padding(data, blockSize)
	if len(padded) != blockSize {
		t.Errorf("padded length = %d, expected %d", len(padded), blockSize)
	}
}

func TestObjectsPkcs7UnPadding(t *testing.T) {
	data := []byte("test")
	padded := pkcs7Padding(data, 16)
	unpadded := pkcs7UnPadding(padded)
	if string(unpadded) != "test" {
		t.Errorf("unpadded = %s, expected 'test'", string(unpadded))
	}
}

func TestObjectsEncryptStringByTXTE(t *testing.T) {
	key := "1234567890123456"
	plaintext := "hello world"
	encrypted := encryptStringByTXTE(plaintext, key)
	if encrypted == "" {
		t.Error("encryptStringByTXTE returned empty string")
	}
}

func TestObjectsDecryptStringByTXTE(t *testing.T) {
	key := "1234567890123456"
	plaintext := "hello world"
	encrypted := encryptStringByTXTE(plaintext, key)
	decrypted := decryptStringByTXTE(encrypted, key)
	if decrypted != plaintext {
		t.Errorf("decrypted = %s, expected %s", decrypted, plaintext)
	}
}

func TestObjectsEncryptDataByTXDEE(t *testing.T) {
	key := "1234567890123456"
	data := []byte("hello world")
	encrypted := encryptDataByTXDEE(data, key)
	if len(encrypted) == 0 {
		t.Error("encryptDataByTXDEE returned empty data")
	}
}

func TestObjectsDecryptDataByTXDEE(t *testing.T) {
	key := "1234567890123456"
	data := []byte("hello world")
	encrypted := encryptDataByTXDEE(data, key)
	decrypted := decryptDataByTXDEE(encrypted, key)
	if string(decrypted) != "hello world" {
		t.Errorf("decrypted = %s, expected 'hello world'", string(decrypted))
	}
}

func TestObjectsEncryptStringByTXDEE(t *testing.T) {
	key := "1234567890123456"
	plaintext := "hello world"
	encrypted := encryptStringByTXDEE(plaintext, key)
	if encrypted == "" {
		t.Error("encryptStringByTXDEE returned empty string")
	}
}

func TestObjectsDecryptStringByTXDEE(t *testing.T) {
	key := "1234567890123456"
	plaintext := "hello world"
	encrypted := encryptStringByTXDEE(plaintext, key)
	decrypted := decryptStringByTXDEE(encrypted, key)
	if decrypted != plaintext {
		t.Errorf("decrypted = %s, expected %s", decrypted, plaintext)
	}
}

func TestObjectsEncryptDataByTXDEF(t *testing.T) {
	key := "12345678901234567890123456789012"
	data := []byte("hello world")
	encrypted := encryptDataByTXDEF(data, key)
	if len(encrypted) == 0 {
		t.Error("encryptDataByTXDEF returned empty data")
	}
}

func TestObjectsDecryptDataByTXDEF(t *testing.T) {
	key := "12345678901234567890123456789012"
	data := []byte("hello world")
	encrypted := encryptDataByTXDEF(data, key)
	decrypted := decryptDataByTXDEF(encrypted, key)
	if string(decrypted) != "hello world" {
		t.Errorf("decrypted = %s, expected 'hello world'", string(decrypted))
	}
}

func TestObjectsEncryptStringByTXDEF(t *testing.T) {
	key := "12345678901234567890123456789012"
	plaintext := "hello world"
	encrypted := encryptStringByTXDEF(plaintext, key)
	if encrypted == "" {
		t.Error("encryptStringByTXDEF returned empty string")
	}
}

func TestObjectsDecryptStringByTXDEF(t *testing.T) {
	key := "12345678901234567890123456789012"
	plaintext := "hello world"
	encrypted := encryptStringByTXDEF(plaintext, key)
	decrypted := decryptStringByTXDEF(encrypted, key)
	if decrypted != plaintext {
		t.Errorf("decrypted = %s, expected %s", decrypted, plaintext)
	}
}

func TestObjectsAesEncryptECB(t *testing.T) {
	key := []byte("1234567890123456")
	data := []byte("hello world test!")
	encrypted, err := aesEncryptECB(data, key)
	if err != nil {
		t.Errorf("aesEncryptECB error: %v", err)
	}
	if len(encrypted) == 0 {
		t.Error("aesEncryptECB returned empty data")
	}
}

func TestObjectsAesDecryptECB(t *testing.T) {
	key := []byte("1234567890123456")
	data := []byte("hello world test!")
	encrypted, _ := aesEncryptECB(data, key)
	decrypted, err := aesDecryptECB(encrypted, key)
	if err != nil {
		t.Errorf("aesDecryptECB error: %v", err)
	}
	if string(decrypted) != string(data) {
		t.Errorf("decrypted = %s, expected %s", string(decrypted), string(data))
	}
}

func TestObjectsAesEncryptCBC(t *testing.T) {
	key := []byte("1234567890123456")
	data := []byte("hello world test!")
	encrypted, err := aesEncryptCBC(data, key)
	if err != nil {
		t.Errorf("aesEncryptCBC error: %v", err)
	}
	if len(encrypted) == 0 {
		t.Error("aesEncryptCBC returned empty data")
	}
}

func TestObjectsAesDecryptCBC(t *testing.T) {
	key := []byte("1234567890123456")
	data := []byte("hello world test!")
	encrypted, _ := aesEncryptCBC(data, key)
	decrypted, err := aesDecryptCBC(encrypted, key)
	if err != nil {
		t.Errorf("aesDecryptCBC error: %v", err)
	}
	if string(decrypted) != string(data) {
		t.Errorf("decrypted = %s, expected %s", string(decrypted), string(data))
	}
}
