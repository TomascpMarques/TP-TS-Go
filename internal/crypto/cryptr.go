/* Cryptr - A packge that supplies functions for algorithmicly sound encryption and decription, followed by secret generation. */
package crypto

import "crypto/rand"

// Encrypt - Takes a byte array, and returns the encrypted version of the same array, using the given secret
func Encrypt(content []byte, secret []byte) (cypher []byte) {
	return
}

// Decrypt - Decrypts a given byte sequence, given the secret used to encrypt it
func Decrypt(cypher []byte, secret []byte) (content []byte) {
	return
}

// GenerateRawRandomBytes - Generates raw random bytes
func GenerateRawRandomBytes() (secret []byte, err error) {
	secret = make([]byte, 16*100)

	_, err = rand.Read(secret)
	if err != nil {
		return nil, err
	}

	return
}
