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
// It ensures that a unique nonce is created for each encryption, adding randomness to the output.
func Encrypt(content, secret []byte) ([]byte, error) {
	// Validate secret size for AES (must be 16, 24, or 32 bytes)
	if len(secret) != 16 && len(secret) != 24 && len(secret) != 32 {
		return nil, fmt.Errorf("invalid secret size: %d bytes; must be 16, 24, or 32 bytes", len(secret))
	}

	// Create a new AES cipher block from the secret
	block, err := aes.NewCipher(secret)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher block: %w", err)
	}

	// Create a new GCM (Galois/Counter Mode) cipher, which provides encryption + integrity/authentication
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	// Generate a 12-byte nonce (recommended size for AES-GCM)
	nonce := make([]byte, 12)
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, fmt.Errorf("failed to generate nonce: %w", err)
	}

	// Encrypt the content and append the nonce at the beginning
	ciphertext := gcm.Seal(nonce, nonce, content, nil)
	return ciphertext, nil
}

// Decrypt decrypts content using AES-GCM with the provided secret.
// It extracts the nonce and uses it to decrypt the ciphertext.
func Decrypt(ciphertext, secret []byte) ([]byte, error) {
	// Validate secret size for AES (must be 16, 24, or 32 bytes)
	if len(secret) != 16 && len(secret) != 24 && len(secret) != 32 {
		return nil, fmt.Errorf("invalid secret size: %d bytes; must be 16, 24, or 32 bytes", len(secret))
	}

	// Create a new AES cipher block from the secret
	block, err := aes.NewCipher(secret)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher block: %w", err)
	}

	// Create a new GCM cipher for decryption
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	// Ensure the ciphertext includes the nonce at the beginning
	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, fmt.Errorf("ciphertext too short")
	}

	// Split the nonce and the ciphertext
	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	// Decrypt the ciphertext using the nonce
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt: %w", err)
	}

	return plaintext, nil
}

// singleByteHash generates a simple hash for index calculation using SHA-256.
// This is a secure operation, providing a consistent way to derive indices.
func singleByteHash(i byte) byte {
	hash := sha256.New()
	hash.Write([]byte{i})
	// Return the first byte of the hash as the result
	return hash.Sum(nil)[0]
}

// GenerateSecret creates a derived secret using the rawBytes and optionally a specific timestamp.
// It ensures that the derived secret is exactly 32 bytes long by using portions of rawBytes.
func GenerateSecret(rawBytes []byte, when ...time.Time) (secret []byte, generatedAt time.Time, err error) {
	// Use the provided timestamp or the current time
	if len(when) != 0 {
		generatedAt = when[0]
	} else {
		generatedAt = time.Now()
	}

	// Ensure rawBytes is at least 32 bytes long
	if len(rawBytes) < 32 {
		return nil, generatedAt, fmt.Errorf("rawBytes must be at least 32 bytes long")
	}

	// Initialize the secret array with a capacity of 32 bytes
	secret = make([]byte, 0, 32)

	// Derive indices from the current minute and second using a hash
	indexFromTheMinutes := int(singleByteHash(byte(generatedAt.Minute()))) % (len(rawBytes) - 16)
	indexFromTheSeconds := int(singleByteHash(byte(generatedAt.Second()))) % (len(rawBytes) - 16)

	// Append 16 bytes based on the derived minute and second indices
	secret = append(secret, rawBytes[indexFromTheMinutes:indexFromTheMinutes+16]...)
	secret = append(secret, rawBytes[indexFromTheSeconds:indexFromTheSeconds+16]...)

	// Ensure the secret is exactly 32 bytes long
	if len(secret) > 32 {
		secret = secret[:32]
	} else if len(secret) < 32 {
		// Pad with zeros if the secret is too short
		padding := make([]byte, 32-len(secret))
		secret = append(secret, padding...)
	}

	return secret, generatedAt, nil
}

// GenerateRawRandomBytes generates a large pool of random bytes for use in secret derivation.
// The size must be greater than zero.
func GenerateRawRandomBytes(size int) (rawBytes []byte, err error) {
	// Ensure the requested size is positive
	if size <= 0 {
		return nil, fmt.Errorf("size must be greater than zero")
	}

	// Generate random bytes of the requested size
	rawBytes = make([]byte, size)
	_, err = rand.Read(rawBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to generate random bytes: %w", err)
	}

	return rawBytes, nil
}
