// pkg/objects/builtin_crypto.go
// Encryption helper functions for Charlang compatibility
package objects

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"encoding/hex"
	"fmt"
	"io"
	mathRand "math/rand"
	"strings"
)

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
	if dataLen < 1 {
		return data
	}

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
	if dataLen < 1 {
		return data
	}

	codeT := code
	if codeT == "" {
		codeT = "topxeq"
	}

	codeBytes := []byte(codeT)
	codeLen := len(codeBytes)

	// Calculate dynamic padding length based on code sum
	sum := int(sumBytes(codeBytes))
	addLen := int(sum%5) + 2   // 2-6 bytes
	encIndex := sum % addLen   // Index of the encryption key byte

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
