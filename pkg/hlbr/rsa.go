package hlbr

import (
	"encoding/hex"
	"encoding/pem"
	"crypto/x509"
	"crypto/rsa"
	"fmt"
	"math/big"
)

// padHex pads a hex string to even length by prepending a zero if needed.
func padHex(s string) string {
	if len(s)%2 != 0 {
		return "0" + s
	}
	return s
}

// reverseBytes reverses a byte slice in place and returns it.
// Many JavaScript RSA libraries (e.g. JSEncrypt-based) reverse the plaintext
// byte order before encryption, matching a little-endian interpretation.
func reverseBytes(b []byte) []byte {
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}
	return b
}

// RSAEncryptHex encrypts plaintext using RSA with hex-encoded modulus and exponent.
// This is the standard PKCS1v15 padding mode, suitable for most RSA implementations.
// The modulus and exponent are typically obtained from the server's API.
func RSAEncryptHex(plaintext, hexModulus, hexExponent string) (string, error) {
	hexModulus = padHex(hexModulus)
	hexExponent = padHex(hexExponent)

	modulusBytes, err := hex.DecodeString(hexModulus)
	if err != nil {
		return "", fmt.Errorf("invalid modulus hex: %v", err)
	}

	exponentBytes, err := hex.DecodeString(hexExponent)
	if err != nil {
		return "", fmt.Errorf("invalid exponent hex: %v", err)
	}

	modulus := new(big.Int).SetBytes(modulusBytes)
	exponent := new(big.Int).SetBytes(exponentBytes)

	pubKey := &rsa.PublicKey{
		N: modulus,
		E: int(exponent.Int64()),
	}

	ciphertext, err := rsa.EncryptPKCS1v15(nil, pubKey, []byte(plaintext))
	if err != nil {
		return "", fmt.Errorf("RSA PKCS1v15 encrypt failed: %v", err)
	}

	return hex.EncodeToString(ciphertext), nil
}

// RSAEncryptHexRaw encrypts plaintext using raw RSA (no padding) with hex-encoded
// modulus and exponent. The plaintext bytes are reversed before encryption to match
// the behavior of common JavaScript RSA libraries (e.g. JSEncrypt) that use
// little-endian byte order for the plaintext integer representation.
// This is the mode used by servers whose JS code calls:
//
//	new RSAKey(exponent, "", modulus).encrypt(password)
//
// Returns hex-encoded ciphertext.
func RSAEncryptHexRaw(plaintext, hexModulus, hexExponent string) (string, error) {
	hexModulus = padHex(hexModulus)
	hexExponent = padHex(hexExponent)

	modulusBytes, err := hex.DecodeString(hexModulus)
	if err != nil {
		return "", fmt.Errorf("invalid modulus hex: %v", err)
	}

	exponentBytes, err := hex.DecodeString(hexExponent)
	if err != nil {
		return "", fmt.Errorf("invalid exponent hex: %v", err)
	}

	modulus := new(big.Int).SetBytes(modulusBytes)
	exponent := new(big.Int).SetBytes(exponentBytes)

	// Reverse plaintext bytes to match JS RSA library convention
	ptBytes := reverseBytes([]byte(plaintext))

	// Convert reversed plaintext to big.Int
	pt := new(big.Int).SetBytes(ptBytes)

	// Raw RSA: ciphertext = plaintext^e mod n
	ct := new(big.Int).Exp(pt, exponent, modulus)

	return hex.EncodeToString(ct.Bytes()), nil
}

// RSAEncryptHexWithID encrypts plaintext and returns the ciphertext along with an rsa_id.
// Some APIs require the rsa_id to be sent back with the encrypted password.
// Uses PKCS1v15 padding mode.
func RSAEncryptHexWithID(plaintext, hexModulus, hexExponent, rsaID string) (string, string, error) {
	ciphertext, err := RSAEncryptHex(plaintext, hexModulus, hexExponent)
	if err != nil {
		return "", "", err
	}
	return ciphertext, rsaID, nil
}

// RSAEncryptHexRawWithID encrypts plaintext using raw RSA (no padding, reversed bytes)
// and returns the ciphertext along with an rsa_id.
func RSAEncryptHexRawWithID(plaintext, hexModulus, hexExponent, rsaID string) (string, string, error) {
	ciphertext, err := RSAEncryptHexRaw(plaintext, hexModulus, hexExponent)
	if err != nil {
		return "", "", err
	}
	return ciphertext, rsaID, nil
}

// RSABuildPublicKeyPEM builds a PEM-encoded RSA public key from hex modulus and exponent.
func RSABuildPublicKeyPEM(hexModulus, hexExponent string) (string, error) {
	modulusBytes, err := hex.DecodeString(hexModulus)
	if err != nil {
		return "", fmt.Errorf("invalid modulus hex: %v", err)
	}

	exponentBytes, err := hex.DecodeString(hexExponent)
	if err != nil {
		return "", fmt.Errorf("invalid exponent hex: %v", err)
	}

	modulus := new(big.Int).SetBytes(modulusBytes)
	exponent := new(big.Int).SetBytes(exponentBytes)

	pubKey := &rsa.PublicKey{
		N: modulus,
		E: int(exponent.Int64()),
	}

	derBytes, err := x509.MarshalPKIXPublicKey(pubKey)
	if err != nil {
		return "", fmt.Errorf("marshal public key failed: %v", err)
	}

	pemBlock := &pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: derBytes,
	}

	return string(pem.EncodeToMemory(pemBlock)), nil
}
