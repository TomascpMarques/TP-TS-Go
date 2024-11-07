/* Cryptr - A packge that supplies functions for algorithmicly sound encryption and decription, followed by secret generation. */
package crypto

// Encrypt - Takes a byte array, and returns the encrypted version of the same array, using the given secret
func Encrypt(content []byte, secret []byte) (cypher []byte) {
	return
}

// Decrypt - Decrypts a given byte sequence, given the secret used to encrypt it
func Decrypt(cypher []byte, secret []byte) (content []byte) {
	return
}

// GenerateSecret - Generates a secret based on the current time, and other factors
func GenerateSecret() (secret []byte, err error) {
	return
}