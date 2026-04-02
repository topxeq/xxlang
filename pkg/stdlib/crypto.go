// pkg/stdlib/crypto.go
// Cryptography utilities for the Xxlang standard library.
// Pure Go implementation using standard library - no CGO required.
// Includes Charlang-compatible encryption functions (TXTE, TXDEE, TXDEF, AES).
package stdlib

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/md5"
	"crypto/rand"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"hash"
	"io"
	mathRand "math/rand"
	"strings"

	"github.com/topxeq/xxlang/pkg/objects"
)

// ============================================
// Helper functions for TXDEE/TXDEF encryption
// ============================================

// getRandomByte returns a random byte
func getRandomByte() byte {
	return byte(mathRand.Intn(256))
}

// sumBytes returns the sum of all bytes in a slice
func sumBytes(data []byte) byte {
	var sum byte
	for _, b := range data {
		sum += b
	}
	return sum
}

// pkcs7Padding pads the data to the block size using PKCS7
func pkcs7Padding(data []byte, blockSize int) []byte {
	padding := blockSize - len(data)%blockSize
	padText := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(data, padText...)
}

// pkcs7UnPadding removes PKCS7 padding
func pkcs7UnPadding(data []byte) []byte {
	length := len(data)
	if length == 0 {
		return data
	}
	unPadding := int(data[length-1])
	if unPadding > length || unPadding == 0 {
		return data
	}
	// Verify padding
	for i := 0; i < unPadding; i++ {
		if data[length-1-i] != byte(unPadding) {
			return data
		}
	}
	return data[:length-unPadding]
}

// ============================================
// TXTE - Simple Text Encryption (Charlang compatible)
// ============================================

// encryptStringByTXTE encrypts a string with a simple XOR-based algorithm
func encryptStringByTXTE(text, code string) string {
	if text == "" {
		return ""
	}

	codeT := code
	if codeT == "" {
		codeT = "topxeq"
	}

	srcBytes := []byte(text)
	codeBytes := []byte(codeT)
	srcLen := len(srcBytes)
	codeLen := len(codeBytes)

	result := make([]byte, srcLen)
	for i := 0; i < srcLen; i++ {
		result[i] = srcBytes[i] + codeBytes[i%codeLen] + byte(i+1)
	}

	return strings.ToUpper(hex.EncodeToString(result))
}

// decryptStringByTXTE decrypts a hex string encrypted by TXTE
func decryptStringByTXTE(hexStr, code string) string {
	if hexStr == "" {
		return ""
	}

	codeT := code
	if codeT == "" {
		codeT = "topxeq"
	}

	srcBytes, err := hex.DecodeString(hexStr)
	if err != nil {
		return ""
	}

	codeBytes := []byte(codeT)
	srcLen := len(srcBytes)
	codeLen := len(codeBytes)

	result := make([]byte, srcLen)
	for i := 0; i < srcLen; i++ {
		result[i] = srcBytes[i] - codeBytes[i%codeLen] - byte(i+1)
	}

	return string(result)
}

// ============================================
// TXDEE - Data Encryption Enhanced (Charlang compatible)
// ============================================

// encryptDataByTXDEE encrypts data with random prefix/suffix bytes
func encryptDataByTXDEE(data []byte, code string) []byte {
	if data == nil {
		return nil
	}

	dataLen := len(data)

	codeT := code
	if codeT == "" {
		codeT = "topxeq"
	}

	codeBytes := []byte(codeT)
	codeLen := len(codeBytes)

	// Result has 4 extra bytes: 2 random prefix + encrypted data + 2 random suffix
	result := make([]byte, dataLen+4)

	// Random prefix bytes
	result[0] = getRandomByte()
	result[1] = getRandomByte()

	// Encrypted data
	for i := 0; i < dataLen; i++ {
		result[2+i] = data[i] + codeBytes[i%codeLen] + byte(i+1) + result[1]
	}

	// Random suffix bytes
	result[dataLen+2] = getRandomByte()
	result[dataLen+3] = getRandomByte()

	return result
}

// decryptDataByTXDEE decrypts data encrypted by TXDEE
func decryptDataByTXDEE(data []byte, code string) []byte {
	if data == nil {
		return nil
	}

	dataLen := len(data)
	if dataLen < 4 {
		return nil
	}

	dataLen -= 4 // Remove the 4 extra bytes

	codeT := code
	if codeT == "" {
		codeT = "topxeq"
	}

	codeBytes := []byte(codeT)
	codeLen := len(codeBytes)

	result := make([]byte, dataLen)
	for i := 0; i < dataLen; i++ {
		result[i] = data[2+i] - codeBytes[i%codeLen] - byte(i+1) - data[1]
	}

	return result
}

// encryptStringByTXDEE encrypts a string and returns hex
func encryptStringByTXDEE(text, code string) string {
	if text == "" {
		return ""
	}

	encrypted := encryptDataByTXDEE([]byte(text), code)
	if encrypted == nil {
		return "[ERROR: encrypting failed]"
	}

	return strings.ToUpper(hex.EncodeToString(encrypted))
}

// decryptStringByTXDEE decrypts a hex string encrypted by TXDEE
func decryptStringByTXDEE(hexStr, code string) string {
	if hexStr == "" {
		return ""
	}

	var srcBytes []byte
	var err error

	// Check for legacy prefix "740404"
	if strings.HasPrefix(hexStr, "740404") {
		srcBytes, err = hex.DecodeString(hexStr[6:])
	} else {
		srcBytes, err = hex.DecodeString(hexStr)
	}

	if err != nil {
		return "[ERROR: decrypting failed]"
	}

	decrypted := decryptDataByTXDEE(srcBytes, code)
	if decrypted == nil {
		return "[ERROR: decrypting failed]"
	}

	return string(decrypted)
}

// ============================================
// TXDEF - Data Encryption Flexible (Charlang compatible)
// ============================================

// encryptDataByTXDEF encrypts data with dynamic padding based on code
func encryptDataByTXDEF(data []byte, code string) []byte {
	if data == nil {
		return nil
	}

	dataLen := len(data)

	codeT := code
	if codeT == "" {
		codeT = "topxeq"
	}

	codeBytes := []byte(codeT)
	codeLen := len(codeBytes)

	// Calculate dynamic padding length based on code sum
	sum := int(sumBytes(codeBytes))
	addLen := int(sum%5) + 2 // 2-6 bytes
	encIndex := sum % addLen // Index of the encryption key byte

	result := make([]byte, dataLen+addLen)

	// Random padding bytes
	for i := 0; i < addLen; i++ {
		result[i] = getRandomByte()
	}

	// Encrypted data
	for i := 0; i < dataLen; i++ {
		result[addLen+i] = data[i] + codeBytes[i%codeLen] + byte(i+1) + result[encIndex]
	}

	return result
}

// decryptDataByTXDEF decrypts data encrypted by TXDEF
func decryptDataByTXDEF(data []byte, code string) []byte {
	if data == nil {
		return nil
	}

	codeT := code
	if codeT == "" {
		codeT = "topxeq"
	}

	codeBytes := []byte(codeT)
	codeLen := len(codeBytes)

	// Calculate dynamic padding length
	sum := int(sumBytes(codeBytes))
	addLen := int(sum%5) + 2
	encIndex := sum % addLen

	// Check for header prefix
	if bytes.HasPrefix(data, []byte("//TXDEF#")) {
		data = data[8:]
	}

	dataLen := len(data)
	if dataLen < addLen {
		return nil
	}

	dataLen -= addLen

	result := make([]byte, dataLen)
	for i := 0; i < dataLen; i++ {
		result[i] = data[addLen+i] - codeBytes[i%codeLen] - byte(i+1) - data[encIndex]
	}

	return result
}

// encryptStringByTXDEF encrypts a string and returns hex
func encryptStringByTXDEF(text string, code string) string {
	if text == "" {
		return ""
	}

	encrypted := encryptDataByTXDEF([]byte(text), code)
	if encrypted == nil {
		return "[ERROR: encrypting failed]"
	}

	return strings.ToUpper(hex.EncodeToString(encrypted))
}

// decryptStringByTXDEF decrypts a hex string encrypted by TXDEF
func decryptStringByTXDEF(hexStr string, code string) string {
	if hexStr == "" {
		return ""
	}

	var srcBytes []byte
	var err error

	// Check for various prefixes
	if strings.HasPrefix(hexStr, "740404") {
		srcBytes, err = hex.DecodeString(hexStr[6:])
	} else if strings.HasPrefix(hexStr, "//TXDEF#") {
		srcBytes, err = hex.DecodeString(hexStr[8:])
	} else {
		srcBytes, err = hex.DecodeString(hexStr)
	}

	if err != nil {
		return fmt.Sprintf("[ERROR: decrypting failed: %v]", err)
	}

	decrypted := decryptDataByTXDEF(srcBytes, code)
	if decrypted == nil {
		return "[ERROR: decrypting failed]"
	}

	return string(decrypted)
}

// encryptStreamByTXDEF encrypts a stream
func encryptStreamByTXDEF(reader io.Reader, code string, writer io.Writer) error {
	if reader == nil || writer == nil {
		return fmt.Errorf("nil reader or writer")
	}

	codeT := code
	if codeT == "" {
		codeT = "topxeq"
	}

	codeBytes := []byte(codeT)
	codeLen := len(codeBytes)

	sum := int(sumBytes(codeBytes))
	addLen := int(sum%5) + 2
	encIndex := sum % addLen

	// Write random padding
	addBuf := make([]byte, addLen)
	for i := 0; i < addLen; i++ {
		addBuf[i] = getRandomByte()
	}

	_, err := writer.Write(addBuf)
	if err != nil {
		return err
	}

	idxByte := addBuf[encIndex]

	buf := make([]byte, 4096)
	pos := 0

	for {
		n, err := reader.Read(buf)
		if err != nil && err != io.EOF {
			return err
		}

		// Encrypt each byte
		for i := 0; i < n; i++ {
			buf[i] = buf[i] + codeBytes[pos%codeLen] + byte(pos+1) + idxByte
			pos++
		}

		if n > 0 {
			_, errW := writer.Write(buf[:n])
			if errW != nil {
				return errW
			}
		}

		if err == io.EOF {
			break
		}
	}

	return nil
}

// decryptStreamByTXDEF decrypts a stream
func decryptStreamByTXDEF(reader io.Reader, code string, writer io.Writer) error {
	if reader == nil || writer == nil {
		return fmt.Errorf("nil reader or writer")
	}

	codeT := code
	if codeT == "" {
		codeT = "topxeq"
	}

	codeBytes := []byte(codeT)
	codeLen := len(codeBytes)

	sum := int(sumBytes(codeBytes))
	addLen := int(sum%5) + 2
	encIndex := sum % addLen

	// Read padding
	addBuf := make([]byte, addLen)
	n, err := reader.Read(addBuf)
	if err != nil || n != addLen {
		return fmt.Errorf("failed to read header")
	}

	idxByte := addBuf[encIndex]

	buf := make([]byte, 4096)
	pos := 0

	for {
		n, err := reader.Read(buf)
		if err != nil && err != io.EOF {
			return err
		}

		// Decrypt each byte
		for i := 0; i < n; i++ {
			buf[i] = buf[i] - codeBytes[pos%codeLen] - byte(pos+1) - idxByte
			pos++
		}

		if n > 0 {
			_, errW := writer.Write(buf[:n])
			if errW != nil {
				return errW
			}
		}

		if err == io.EOF {
			break
		}
	}

	return nil
}

// ============================================
// AES Encryption (Charlang compatible)
// ============================================

// aesEncryptECB encrypts data using AES ECB-like mode (CBC with zero IV)
func aesEncryptECB(src, key []byte) ([]byte, error) {
	// Truncate key to 16 bytes if longer
	if len(key) > 16 {
		key = key[:16]
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	blockSize := block.BlockSize()
	src = pkcs7Padding(src, blockSize)

	if len(src)%blockSize != 0 {
		return nil, fmt.Errorf("data is not a multiple of block size")
	}

	out := make([]byte, len(src))
	iv := make([]byte, blockSize) // Zero IV for ECB-like behavior

	dst := out
	for len(src) > 0 {
		// XOR with IV (CBC mode with zero IV)
		for j := 0; j < blockSize; j++ {
			src[j] ^= iv[j]
		}
		block.Encrypt(dst[:blockSize], src[:blockSize])

		// Update IV for next block (CBC chaining)
		for j := 0; j < blockSize; j++ {
			iv[j] = dst[j]
		}

		src = src[blockSize:]
		dst = dst[blockSize:]
	}

	return out, nil
}

// aesDecryptECB decrypts data using AES ECB-like mode
func aesDecryptECB(src, key []byte) ([]byte, error) {
	if len(key) > 16 {
		key = key[:16]
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	blockSize := block.BlockSize()
	if len(src)%blockSize != 0 {
		return nil, fmt.Errorf("data is not a multiple of block size")
	}

	out := make([]byte, len(src))
	iv := make([]byte, blockSize) // Zero IV

	dst := out
	for len(src) > 0 {
		block.Decrypt(dst[:blockSize], src[:blockSize])

		// XOR with IV
		for j := 0; j < blockSize; j++ {
			dst[j] ^= iv[j]
		}

		// Update IV
		for j := 0; j < blockSize; j++ {
			iv[j] = src[j]
		}

		src = src[blockSize:]
		dst = dst[blockSize:]
	}

	return pkcs7UnPadding(out), nil
}

// aesEncryptCBC encrypts data using AES CBC mode
func aesEncryptCBC(src, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	blockSize := block.BlockSize()
	src = pkcs7Padding(src, blockSize)

	// Use key prefix as IV (same as Charlang)
	iv := make([]byte, blockSize)
	copy(iv, key)
	if len(key) < blockSize {
		// Pad with zeros if key is shorter
		for i := len(key); i < blockSize; i++ {
			iv[i] = 0
		}
	}

	encrypted := make([]byte, len(src))
	mode := cipher.NewCBCEncrypter(block, iv)
	mode.CryptBlocks(encrypted, src)

	return encrypted, nil
}

// aesDecryptCBC decrypts data using AES CBC mode
func aesDecryptCBC(src, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	blockSize := block.BlockSize()
	if len(src)%blockSize != 0 {
		return nil, fmt.Errorf("data is not a multiple of block size")
	}

	// Use key prefix as IV
	iv := make([]byte, blockSize)
	copy(iv, key)
	if len(key) < blockSize {
		for i := len(key); i < blockSize; i++ {
			iv[i] = 0
		}
	}

	decrypted := make([]byte, len(src))
	mode := cipher.NewCBCDecrypter(block, iv)
	mode.CryptBlocks(decrypted, src)

	return pkcs7UnPadding(decrypted), nil
}

// ============================================
// Module Registration
// ============================================

func init() {
	Register(&Module{
		Name: "crypto",
		Exports: map[string]objects.Object{
			// Hash functions
			"md5": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("md5() takes exactly 1 argument")
				}
				s, ok := args[0].(*objects.String)
				if !ok {
					return Error("md5() requires a string argument")
				}
				hash := md5.Sum([]byte(s.Value))
				return String(hex.EncodeToString(hash[:]))
			}),

			"sha1": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("sha1() takes exactly 1 argument")
				}
				s, ok := args[0].(*objects.String)
				if !ok {
					return Error("sha1() requires a string argument")
				}
				hash := sha1.Sum([]byte(s.Value))
				return String(hex.EncodeToString(hash[:]))
			}),

			"sha256": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("sha256() takes exactly 1 argument")
				}
				s, ok := args[0].(*objects.String)
				if !ok {
					return Error("sha256() requires a string argument")
				}
				hash := sha256.Sum256([]byte(s.Value))
				return String(hex.EncodeToString(hash[:]))
			}),

			"sha512": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("sha512() takes exactly 1 argument")
				}
				s, ok := args[0].(*objects.String)
				if !ok {
					return Error("sha512() requires a string argument")
				}
				hash := sha512.Sum512([]byte(s.Value))
				return String(hex.EncodeToString(hash[:]))
			}),

			// HMAC functions
			"hmacMd5": BuiltinFunc(func(args ...objects.Object) objects.Object {
				return hmacHash(args, md5.New, "hmacMd5")
			}),

			"hmacSha1": BuiltinFunc(func(args ...objects.Object) objects.Object {
				return hmacHash(args, sha1.New, "hmacSha1")
			}),

			"hmacSha256": BuiltinFunc(func(args ...objects.Object) objects.Object {
				return hmacHash(args, sha256.New, "hmacSha256")
			}),

			"hmacSha512": BuiltinFunc(func(args ...objects.Object) objects.Object {
				return hmacHash(args, sha512.New, "hmacSha512")
			}),

			// Encoding functions
			"base64Encode": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("base64Encode() takes exactly 1 argument")
				}
				s, ok := args[0].(*objects.String)
				if !ok {
					return Error("base64Encode() requires a string argument")
				}
				encoded := base64.StdEncoding.EncodeToString([]byte(s.Value))
				return String(encoded)
			}),

			"base64Decode": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("base64Decode() takes exactly 1 argument")
				}
				s, ok := args[0].(*objects.String)
				if !ok {
					return Error("base64Decode() requires a string argument")
				}
				decoded, err := base64.StdEncoding.DecodeString(s.Value)
				if err != nil {
					return Error(fmt.Sprintf("base64Decode() failed: %s", err.Error()))
				}
				return String(string(decoded))
			}),

			"base64URLEncode": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("base64URLEncode() takes exactly 1 argument")
				}
				s, ok := args[0].(*objects.String)
				if !ok {
					return Error("base64URLEncode() requires a string argument")
				}
				encoded := base64.URLEncoding.EncodeToString([]byte(s.Value))
				return String(encoded)
			}),

			"base64URLDecode": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("base64URLDecode() takes exactly 1 argument")
				}
				s, ok := args[0].(*objects.String)
				if !ok {
					return Error("base64URLDecode() requires a string argument")
				}
				decoded, err := base64.URLEncoding.DecodeString(s.Value)
				if err != nil {
					return Error(fmt.Sprintf("base64URLDecode() failed: %s", err.Error()))
				}
				return String(string(decoded))
			}),

			"hexEncode": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("hexEncode() takes exactly 1 argument")
				}
				s, ok := args[0].(*objects.String)
				if !ok {
					return Error("hexEncode() requires a string argument")
				}
				encoded := hex.EncodeToString([]byte(s.Value))
				return String(encoded)
			}),

			"hexDecode": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("hexDecode() takes exactly 1 argument")
				}
				s, ok := args[0].(*objects.String)
				if !ok {
					return Error("hexDecode() requires a string argument")
				}
				decoded, err := hex.DecodeString(s.Value)
				if err != nil {
					return Error(fmt.Sprintf("hexDecode() failed: %s", err.Error()))
				}
				return String(string(decoded))
			}),

			// Random functions
			"randomBytes": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("randomBytes() takes exactly 1 argument")
				}
				n, ok := args[0].(*objects.Int)
				if !ok {
					return Error("randomBytes() requires an integer argument")
				}
				if n.Value < 0 {
					return Error("randomBytes() requires a non-negative integer")
				}
				if n.Value > 1024*1024 { // 1MB limit
					return Error("randomBytes() size too large (max 1MB)")
				}

				bytes := make([]byte, n.Value)
				_, err := rand.Read(bytes)
				if err != nil {
					return Error(fmt.Sprintf("randomBytes() failed: %s", err.Error()))
				}

				// Return as hex-encoded string for convenience
				return String(hex.EncodeToString(bytes))
			}),

			"randomHex": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("randomHex() takes exactly 1 argument")
				}
				n, ok := args[0].(*objects.Int)
				if !ok {
					return Error("randomHex() requires an integer argument")
				}
				if n.Value < 0 {
					return Error("randomHex() requires a non-negative integer")
				}

				bytes := make([]byte, n.Value)
				_, err := rand.Read(bytes)
				if err != nil {
					return Error(fmt.Sprintf("randomHex() failed: %s", err.Error()))
				}

				return String(hex.EncodeToString(bytes))
			}),

			"randomBase64": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("randomBase64() takes exactly 1 argument")
				}
				n, ok := args[0].(*objects.Int)
				if !ok {
					return Error("randomBase64() requires an integer argument")
				}
				if n.Value < 0 {
					return Error("randomBase64() requires a non-negative integer")
				}

				bytes := make([]byte, n.Value)
				_, err := rand.Read(bytes)
				if err != nil {
					return Error(fmt.Sprintf("randomBase64() failed: %s", err.Error()))
				}

				return String(base64.StdEncoding.EncodeToString(bytes))
			}),

			// UUID generation
			"uuid": BuiltinFunc(func(args ...objects.Object) objects.Object {
				uuid := make([]byte, 16)
				_, err := rand.Read(uuid)
				if err != nil {
					return Error(fmt.Sprintf("uuid() failed: %s", err.Error()))
				}

				// Set version (4) and variant bits
				uuid[6] = (uuid[6] & 0x0f) | 0x40 // Version 4
				uuid[8] = (uuid[8] & 0x3f) | 0x80 // Variant RFC 4122

				return String(fmt.Sprintf("%x-%x-%x-%x-%x",
					uuid[0:4], uuid[4:6], uuid[6:8], uuid[8:10], uuid[10:16]))
			}),

			// ============================================
			// TXTE - Simple Text Encryption (Charlang compatible)
			// ============================================
			"encryptTextByTXTE": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) < 1 {
					return Error("encryptTextByTXTE requires at least 1 argument")
				}
				text := args[0].Inspect()
				code := ""
				if len(args) > 1 {
					code = args[1].Inspect()
				}
				return String(encryptStringByTXTE(text, code))
			}),

			"decryptTextByTXTE": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) < 1 {
					return Error("decryptTextByTXTE requires at least 1 argument")
				}
				hexStr := args[0].Inspect()
				code := ""
				if len(args) > 1 {
					code = args[1].Inspect()
				}
				return String(decryptStringByTXTE(hexStr, code))
			}),

			// ============================================
			// TXDEE - Enhanced Data Encryption (Charlang compatible)
			// ============================================
			"encryptDataByTXDEE": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) < 1 {
					return Error("encryptDataByTXDEE requires at least 1 argument")
				}
				var data []byte
				switch v := args[0].(type) {
				case *objects.Bytes:
					data = v.Value
				default:
					data = []byte(args[0].Inspect())
				}
				code := ""
				if len(args) > 1 {
					code = args[1].Inspect()
				}
				result := encryptDataByTXDEE(data, code)
				if result == nil {
					return Error("encryption failed")
				}
				return &objects.Bytes{Value: result}
			}),

			"decryptDataByTXDEE": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) < 1 {
					return Error("decryptDataByTXDEE requires at least 1 argument")
				}
				var data []byte
				switch v := args[0].(type) {
				case *objects.Bytes:
					data = v.Value
				default:
					data = []byte(args[0].Inspect())
				}
				code := ""
				if len(args) > 1 {
					code = args[1].Inspect()
				}
				result := decryptDataByTXDEE(data, code)
				if result == nil {
					return Error("decryption failed")
				}
				return &objects.Bytes{Value: result}
			}),

			"encryptTextByTXDEE": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) < 1 {
					return Error("encryptTextByTXDEE requires at least 1 argument")
				}
				text := args[0].Inspect()
				code := ""
				if len(args) > 1 {
					code = args[1].Inspect()
				}
				return String(encryptStringByTXDEE(text, code))
			}),

			"decryptTextByTXDEE": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) < 1 {
					return Error("decryptTextByTXDEE requires at least 1 argument")
				}
				hexStr := args[0].Inspect()
				code := ""
				if len(args) > 1 {
					code = args[1].Inspect()
				}
				return String(decryptStringByTXDEE(hexStr, code))
			}),

			// ============================================
			// TXDEF - Flexible Data Encryption (Charlang compatible, default)
			// ============================================
			"encryptData": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) < 1 {
					return Error("encryptData requires at least 1 argument")
				}
				var data []byte
				switch v := args[0].(type) {
				case *objects.Bytes:
					data = v.Value
				default:
					data = []byte(args[0].Inspect())
				}
				code := ""
				if len(args) > 1 {
					code = args[1].Inspect()
				}
				result := encryptDataByTXDEF(data, code)
				if result == nil {
					return Error("encryption failed")
				}
				return &objects.Bytes{Value: result}
			}),

			"encryptBytes": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) < 1 {
					return Error("encryptBytes requires at least 1 argument")
				}
				var data []byte
				switch v := args[0].(type) {
				case *objects.Bytes:
					data = v.Value
				default:
					data = []byte(args[0].Inspect())
				}
				code := ""
				if len(args) > 1 {
					code = args[1].Inspect()
				}
				result := encryptDataByTXDEF(data, code)
				if result == nil {
					return Error("encryption failed")
				}
				return &objects.Bytes{Value: result}
			}),

			"decryptData": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) < 1 {
					return Error("decryptData requires at least 1 argument")
				}
				var data []byte
				switch v := args[0].(type) {
				case *objects.Bytes:
					data = v.Value
				default:
					data = []byte(args[0].Inspect())
				}
				code := ""
				if len(args) > 1 {
					code = args[1].Inspect()
				}
				result := decryptDataByTXDEF(data, code)
				if result == nil {
					return Error("decryption failed")
				}
				return &objects.Bytes{Value: result}
			}),

			"decryptBytes": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) < 1 {
					return Error("decryptBytes requires at least 1 argument")
				}
				var data []byte
				switch v := args[0].(type) {
				case *objects.Bytes:
					data = v.Value
				default:
					data = []byte(args[0].Inspect())
				}
				code := ""
				if len(args) > 1 {
					code = args[1].Inspect()
				}
				result := decryptDataByTXDEF(data, code)
				if result == nil {
					return Error("decryption failed")
				}
				return &objects.Bytes{Value: result}
			}),

			"encryptText": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) < 1 {
					return Error("encryptText requires at least 1 argument")
				}
				text := args[0].Inspect()
				code := ""
				if len(args) > 1 {
					code = args[1].Inspect()
				}
				return String(encryptStringByTXDEF(text, code))
			}),

			"encryptStr": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) < 1 {
					return Error("encryptStr requires at least 1 argument")
				}
				text := args[0].Inspect()
				code := ""
				if len(args) > 1 {
					code = args[1].Inspect()
				}
				return String(encryptStringByTXDEF(text, code))
			}),

			"decryptText": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) < 1 {
					return Error("decryptText requires at least 1 argument")
				}
				hexStr := args[0].Inspect()
				code := ""
				if len(args) > 1 {
					code = args[1].Inspect()
				}
				return String(decryptStringByTXDEF(hexStr, code))
			}),

			"decryptStr": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) < 1 {
					return Error("decryptStr requires at least 1 argument")
				}
				hexStr := args[0].Inspect()
				code := ""
				if len(args) > 1 {
					code = args[1].Inspect()
				}
				return String(decryptStringByTXDEF(hexStr, code))
			}),

			"encryptStream": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) < 3 {
					return Error("encryptStream requires 3 arguments: reader, code, writer")
				}
				reader, ok := args[0].(io.Reader)
				if !ok {
					return Error("first argument must be a reader")
				}
				code := args[1].Inspect()
				writer, ok := args[2].(io.Writer)
				if !ok {
					return Error("third argument must be a writer")
				}
				err := encryptStreamByTXDEF(reader, code, writer)
				if err != nil {
					return Error(fmt.Sprintf("encryption failed: %v", err))
				}
				return objects.NULL
			}),

			"decryptStream": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) < 3 {
					return Error("decryptStream requires 3 arguments: reader, code, writer")
				}
				reader, ok := args[0].(io.Reader)
				if !ok {
					return Error("first argument must be a reader")
				}
				code := args[1].Inspect()
				writer, ok := args[2].(io.Writer)
				if !ok {
					return Error("third argument must be a writer")
				}
				err := decryptStreamByTXDEF(reader, code, writer)
				if err != nil {
					return Error(fmt.Sprintf("decryption failed: %v", err))
				}
				return objects.NULL
			}),

			// ============================================
			// AES Encryption (Charlang compatible)
			// ============================================
			"aesEncrypt": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) < 2 {
					return Error("aesEncrypt requires at least 2 arguments: data, key")
				}
				var data []byte
				switch v := args[0].(type) {
				case *objects.Bytes:
					data = v.Value
				default:
					data = []byte(args[0].Inspect())
				}
				key := []byte(args[1].Inspect())

				mode := ""
				if len(args) > 2 {
					mode = args[2].Inspect()
				}

				var result []byte
				var err error

				if strings.Contains(mode, "cbc") || strings.Contains(mode, "-cbc") {
					result, err = aesEncryptCBC(data, key)
				} else {
					result, err = aesEncryptECB(data, key)
				}

				if err != nil {
					return Error(fmt.Sprintf("AES encryption failed: %v", err))
				}
				return String(string(result))
			}),

			"aesDecrypt": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) < 2 {
					return Error("aesDecrypt requires at least 2 arguments: data, key")
				}
				var data []byte
				switch v := args[0].(type) {
				case *objects.Bytes:
					data = v.Value
				default:
					data = []byte(args[0].Inspect())
				}
				key := []byte(args[1].Inspect())

				mode := ""
				if len(args) > 2 {
					mode = args[2].Inspect()
				}

				var result []byte
				var err error

				if strings.Contains(mode, "cbc") || strings.Contains(mode, "-cbc") {
					result, err = aesDecryptCBC(data, key)
				} else {
					result, err = aesDecryptECB(data, key)
				}

				if err != nil {
					return Error(fmt.Sprintf("AES decryption failed: %v", err))
				}
				return String(string(result))
			}),

			// ============================================
			// JWT Functions
			// ============================================
			"genJwtToken": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) < 2 {
					return Error("genJwtToken requires at least 2 arguments: payload, secret")
				}

				// Get payload (must be a Map)
				payload, ok := args[0].(*objects.Map)
				if !ok {
					return Error("first argument to genJwtToken must be MAP")
				}

				// Get secret
				var secret []byte
				switch s := args[1].(type) {
				case *objects.String:
					secret = []byte(s.Value)
				case *objects.Bytes:
					secret = s.Value
				default:
					return Error("second argument to genJwtToken must be STRING or BYTES")
				}

				// Parse options
				withType := true
				base64Secret := false
				expireSeconds := int64(0)

				if len(args) > 2 {
					for _, opt := range args[2:] {
						if optStr, ok := opt.(*objects.String); ok {
							switch {
							case optStr.Value == "-noType":
								withType = false
							case optStr.Value == "-base64Secret":
								base64Secret = true
							case strings.HasPrefix(optStr.Value, "-expire="):
								expStr := strings.TrimPrefix(optStr.Value, "-expire=")
								expireSeconds = parseIntFromString(expStr)
							}
						}
					}
				}

				token, err := objects.GenJwtToken(payload, secret, withType, base64Secret, expireSeconds)
				if err != nil {
					return Error(fmt.Sprintf("genJwtToken failed: %v", err))
				}
				return String(token)
			}),

			"parseJwtToken": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) < 2 {
					return Error("parseJwtToken requires at least 2 arguments: token, secret")
				}

				// Get token string
				tokenStr, ok := args[0].(*objects.String)
				if !ok {
					return Error("first argument to parseJwtToken must be STRING")
				}

				// Get secret
				var secret []byte
				switch s := args[1].(type) {
				case *objects.String:
					secret = []byte(s.Value)
				case *objects.Bytes:
					secret = s.Value
				default:
					return Error("second argument to parseJwtToken must be STRING or BYTES")
				}

				// Parse options
				base64Secret := false
				noValidate := false

				if len(args) > 2 {
					for _, opt := range args[2:] {
						if optStr, ok := opt.(*objects.String); ok {
							switch optStr.Value {
							case "-base64Secret":
								base64Secret = true
							case "-noValidate":
								noValidate = true
							}
						}
					}
				}

				claims, err := objects.ParseJwtToken(tokenStr.Value, secret, base64Secret, noValidate)
				if err != nil {
					return Error(fmt.Sprintf("parseJwtToken failed: %v", err))
				}

				// Convert claims to Map
				result := make(map[objects.HashKey]objects.MapPair)
				for k, v := range claims {
					keyObj := objects.NewString(k)
					hashKey := keyObj.HashKey()
					result[hashKey] = objects.MapPair{
						Key:   keyObj,
						Value: objects.InterfaceToObject(v),
					}
				}
				return objects.NewMap(result)
			}),
		},
	})
}

// parseIntFromString parses integer from string
func parseIntFromString(s string) int64 {
	var result int64
	for _, c := range s {
		if c >= '0' && c <= '9' {
			result = result*10 + int64(c-'0')
		}
	}
	return result
}

// hmacHash is a helper function for HMAC operations
func hmacHash(args []objects.Object, hashFunc func() hash.Hash, name string) objects.Object {
	if len(args) != 2 {
		return Error(fmt.Sprintf("%s() takes exactly 2 arguments", name))
	}

	key, ok := args[0].(*objects.String)
	if !ok {
		return Error(fmt.Sprintf("%s() requires a string key", name))
	}

	data, ok := args[1].(*objects.String)
	if !ok {
		return Error(fmt.Sprintf("%s() requires a string data", name))
	}

	h := hmac.New(hashFunc, []byte(key.Value))
	h.Write([]byte(data.Value))
	return String(hex.EncodeToString(h.Sum(nil)))
}
