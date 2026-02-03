package crypto

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
)

const (
	// PEM block types
	pemBlockTypeRSA          = "RSA PRIVATE KEY"
	pemBlockTypePrivateKey   = "PRIVATE KEY"
	pemBlockTypePublicKey    = "PUBLIC KEY"
	pemBlockTypeRSAPublicKey = "RSA PUBLIC KEY"
)

// LoadPublicKey loads an RSA public key from a PEM-encoded file.
//
// The file can contain any of the following PEM block types:
//   - "PUBLIC KEY" (PKCS#8)
//   - "RSA PUBLIC KEY" (PKCS#1)
//
// Parameters:
//   - path: Path to the file containing the public key
//
// Returns:
//   - *rsa.PublicKey: The loaded public key
//   - error: An error if the file cannot be read or the key cannot be parsed
//
// Example:
//
//	pubKey, err := crypto.LoadPublicKey("public.pem")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	ciphertext, err := crypto.Encrypt(plaintext, pubKey)
func LoadPublicKey(path string) (*rsa.PublicKey, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read public key file: %w", err)
	}

	block, _ := pem.Decode(data)
	if block == nil {
		return nil, fmt.Errorf("failed to decode PEM block from public key file")
	}

	// Try different key formats
	var pubKey interface{}

	switch block.Type {
	case pemBlockTypePublicKey:
		// PKCS#8 format
		pubKey, err = x509.ParsePKIXPublicKey(block.Bytes)
		if err != nil {
			return nil, fmt.Errorf("failed to parse PKIX public key: %w", err)
		}
	case pemBlockTypeRSAPublicKey:
		// PKCS#1 format
		pubKey, err = x509.ParsePKCS1PublicKey(block.Bytes)
		if err != nil {
			return nil, fmt.Errorf("failed to parse PKCS1 public key: %w", err)
		}
	default:
		return nil, fmt.Errorf("unsupported PEM block type: %s", block.Type)
	}

	rsaPubKey, ok := pubKey.(*rsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("not an RSA public key")
	}

	return rsaPubKey, nil
}

// LoadPrivateKey loads an RSA private key from a PEM-encoded file.
//
// The file can contain any of the following PEM block types:
//   - "PRIVATE KEY" (PKCS#8)
//   - "RSA PRIVATE KEY" (PKCS#1)
//
// Parameters:
//   - path: Path to the file containing the private key
//
// Returns:
//   - *rsa.PrivateKey: The loaded private key
//   - error: An error if the file cannot be read or the key cannot be parsed
//
// Example:
//
//	privKey, err := crypto.LoadPrivateKey("private.pem")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	plaintext, err := crypto.Decrypt(ciphertext, privKey)
func LoadPrivateKey(path string) (*rsa.PrivateKey, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read private key file: %w", err)
	}

	block, _ := pem.Decode(data)
	if block == nil {
		return nil, fmt.Errorf("failed to decode PEM block from private key file")
	}

	var privKey interface{}

	switch block.Type {
	case pemBlockTypePrivateKey:
		// PKCS#8 format
		privKey, err = x509.ParsePKCS8PrivateKey(block.Bytes)
		if err != nil {
			return nil, fmt.Errorf("failed to parse PKCS8 private key: %w", err)
		}
	case pemBlockTypeRSA:
		// PKCS#1 format
		privKey, err = x509.ParsePKCS1PrivateKey(block.Bytes)
		if err != nil {
			return nil, fmt.Errorf("failed to parse PKCS1 private key: %w", err)
		}
	default:
		return nil, fmt.Errorf("unsupported PEM block type: %s", block.Type)
	}

	rsaPrivKey, ok := privKey.(*rsa.PrivateKey)
	if !ok {
		return nil, fmt.Errorf("not an RSA private key")
	}

	return rsaPrivKey, nil
}

// Encrypt encrypts data using RSA-OAEP with SHA-256.
//
// RSA can only encrypt data smaller than the key size. For 2048-bit keys,
// the maximum plaintext size is about 190 bytes after OAEP padding.
// For larger data, consider using hybrid encryption (encrypt with symmetric key,
// then encrypt the symmetric key with RSA).
//
// Parameters:
//   - plaintext: The data to encrypt
//   - pubKey: The RSA public key to use for encryption
//
// Returns:
//   - []byte: The encrypted ciphertext
//   - error: An error if encryption fails or data is too large
//
// Example:
//
//	pubKey, _ := crypto.LoadPublicKey("public.pem")
//	ciphertext, err := crypto.Encrypt([]byte("secret message"), pubKey)
func Encrypt(plaintext []byte, pubKey *rsa.PublicKey) ([]byte, error) {
	ciphertext, err := rsa.EncryptOAEP(
		nil, // hash function parameters (nil = SHA-256 by default)
		nil, // random source (nil = crypto/rand.Reader)
		pubKey,
		plaintext,
		nil, // label (optional associated data)
	)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt data: %w", err)
	}

	return ciphertext, nil
}

// Decrypt decrypts data that was encrypted using RSA-OAEP with SHA-256.
//
// Parameters:
//   - ciphertext: The data to decrypt
//   - privKey: The RSA private key to use for decryption
//
// Returns:
//   - []byte: The decrypted plaintext
//   - error: An error if decryption fails (e.g., wrong key, corrupted data)
//
// Example:
//
//	privKey, _ := crypto.LoadPrivateKey("private.pem")
//	plaintext, err := crypto.Decrypt(ciphertext, privKey)
func Decrypt(ciphertext []byte, privKey *rsa.PrivateKey) ([]byte, error) {
	plaintext, err := rsa.DecryptOAEP(
		nil, // hash function parameters (nil = SHA-256 by default)
		nil, // random source (nil = crypto/rand.Reader)
		privKey,
		ciphertext,
		nil, // label (optional associated data)
	)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt data: %w", err)
	}

	return plaintext, nil
}
