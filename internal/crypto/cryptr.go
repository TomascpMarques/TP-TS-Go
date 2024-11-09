/* Cryptr - A package that supplies functions for algorithmically sound encryption and decryption, along with secret generation. */
package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"io"
	"time"
)

// Encrypt encrypts content using AES-GCM with the provided secret.
// It ensures that a nonce is created for each encryption to add randomness.
func Encrypt(content, secret []byte) ([]byte, error) {
	// Validate secret size for AES (must be 16, 24, or 32 bytes)
	if len(secret) != 16 && len(secret) != 24 && len(secret) != 32 {
		return nil, fmt.Errorf("invalid secret size: %d bytes; must be 16, 24, or 32 bytes", len(secret))
	}

	block, err := aes.NewCipher(secret)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher block: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	// Ensure the nonce size is 12 bytes (recommended size for AES-GCM)
	nonce := make([]byte, 12)
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, fmt.Errorf("failed to generate nonce: %w", err)
	}

	ciphertext := gcm.Seal(nonce, nonce, content, nil)
	return ciphertext, nil
}

// Decrypt decrypts content using AES-GCM with the provided secret.
func Decrypt(ciphertext, secret []byte) ([]byte, error) {
	// Validate secret size for AES (must be 16, 24, or 32 bytes)
	if len(secret) != 16 && len(secret) != 24 && len(secret) != 32 {
		return nil, fmt.Errorf("invalid secret size: %d bytes; must be 16, 24, or 32 bytes", len(secret))
	}

	block, err := aes.NewCipher(secret)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher block: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, fmt.Errorf("ciphertext too short")
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt: %w", err)
	}

	return plaintext, nil
}

// singleByteHash generates a simple hash for index calculation using SHA-256.
// This is now a secure operation compared to the original implementation.
func singleByteHash(i byte) byte {
	hash := sha256.New()
	hash.Write([]byte{i})
	return hash.Sum(nil)[0] // Return the first byte of the hash
}

// GenerateSecret creates a derived secret using the rawBytes and optionally a specific timestamp.
// It ensures the derived secret is exactly 32 bytes long.
func GenerateSecret(rawBytes []byte, when ...time.Time) (secret []byte, generatedAt time.Time, err error) {
	if len(when) != 0 {
		generatedAt = when[0]
	} else {
		generatedAt = time.Now()
	}

	// Validate rawBytes size
	if len(rawBytes) < 32 {
		return nil, generatedAt, fmt.Errorf("rawBytes must be at least 32 bytes long")
	}

	secret = make([]byte, 0, 32) // Predefine capacity to avoid reallocation

	// Ensure indices are valid for accessing rawBytes
	indexFromTheMinutes := int(singleByteHash(byte(generatedAt.Minute()))) % (len(rawBytes) - 16)
	indexFromTheSeconds := int(singleByteHash(byte(generatedAt.Second()))) % (len(rawBytes) - 16)

	// Append 16 bytes based on minutes and seconds
	secret = append(secret, rawBytes[indexFromTheMinutes:indexFromTheMinutes+16]...)
	secret = append(secret, rawBytes[indexFromTheSeconds:indexFromTheSeconds+16]...)

	// Ensure secret is exactly 32 bytes
	if len(secret) > 32 {
		secret = secret[:32]
	} else if len(secret) < 32 {
		padding := make([]byte, 32-len(secret))
		secret = append(secret, padding...)
	}

	return secret, generatedAt, nil
}

// GenerateRawRandomBytes generates a large pool of random bytes for use in secret derivation.
func GenerateRawRandomBytes(size int) (rawBytes []byte, err error) {
	// Ensure the requested size is reasonable
	if size <= 0 {
		return nil, fmt.Errorf("size must be greater than zero")
	}

	rawBytes = make([]byte, size)
	_, err = rand.Read(rawBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to generate random bytes: %w", err)
	}

	return rawBytes, nil
}
