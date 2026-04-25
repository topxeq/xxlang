package hlbr

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/hex"
	"encoding/pem"
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

// RSAEncryptHex encrypts plaintext using RSA with hex-encoded modulus and exponent.
// This is used by web applications that use RSA encryption for login passwords.
// The modulus and exponent are typically obtained from the server's API.
func RSAEncryptHex(plaintext, hexModulus, hexExponent string) (string, error) {
	// Pad hex strings to even length
	hexModulus = padHex(hexModulus)
	hexExponent = padHex(hexExponent)

	// Decode hex modulus
	modulusBytes, err := hex.DecodeString(hexModulus)
	if err != nil {
		return "", fmt.Errorf("invalid modulus hex: %v", err)
	}

	// Decode hex exponent
	exponentBytes, err := hex.DecodeString(hexExponent)
	if err != nil {
		return "", fmt.Errorf("invalid exponent hex: %v", err)
	}

	// Convert to big.Int
	modulus := new(big.Int).SetBytes(modulusBytes)
	exponent := new(big.Int).SetBytes(exponentBytes)

	// Build RSA public key
	pubKey := &rsa.PublicKey{
		N: modulus,
		E: int(exponent.Int64()),
	}

	// Encrypt using PKCS1v15
	ciphertext, err := rsa.EncryptPKCS1v15(rand.Reader, pubKey, []byte(plaintext))
	if err != nil {
		return "", fmt.Errorf("RSA encrypt failed: %v", err)
	}

	// Return hex-encoded ciphertext
	return hex.EncodeToString(ciphertext), nil
}

// RSAEncryptHexWithID encrypts plaintext and returns the ciphertext along with an rsa_id.
// Some APIs require the rsa_id to be sent back with the encrypted password.
func RSAEncryptHexWithID(plaintext, hexModulus, hexExponent, rsaID string) (string, string, error) {
	ciphertext, err := RSAEncryptHex(plaintext, hexModulus, hexExponent)
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

	// Convert to PKIX DER format
	derBytes, err := x509.MarshalPKIXPublicKey(pubKey)
	if err != nil {
		return "", fmt.Errorf("marshal public key failed: %v", err)
	}

	// Encode to PEM
	pemBlock := &pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: derBytes,
	}

	return string(pem.EncodeToMemory(pemBlock)), nil
}
