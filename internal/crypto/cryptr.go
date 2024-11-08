/* Cryptr - A packge that supplies functions for algorithmicly sound encryption and decription, followed by secret generation. */
package crypto

import (
	"crypto/rand"
	"log"
	"time"
)

// Encrypt - Takes a byte array, and returns the encrypted version of the same array, using the given secret
// Weak encryption - rellies tomuch on the secret
// Vulnerable to timming attacks
func Encrypt(content []byte, secret []byte) []byte {
	/*
		1. Divide the content in 16 byte increments, pad the ones that are not 16 -> usualy the last one
	*/

	blockCount := len(content) / 16
	blocks := make([][]byte, blockCount)

	log.Println("content len: ", len(content))
	log.Println("Block Count: ", blockCount)

	for blockStart := range blockCount {
		block := make([]byte, 16)
		copy(block, content[blockStart*16:blockStart*16+16])
		blocks[blockStart] = block
	}

	if len(content)%16 != 0 {
		// Padd the last block if
		block := make([]byte, 16)
		copy(block, content[blockCount*16:])
		blocks = append(blocks, block)

	}

	encF := make([]byte, 0)

	for _, block := range blocks {
		enc := make([]byte, 16)

		for i, byt := range block {
			if byt|3 == 0 {
				for _, b := range secret[:len(secret)/2] {
					enc[i] = b ^ byt
				}
			} else {
				for _, b := range secret[len(secret)/2:] {
					enc[i] = byt ^ b
				}
			}
		}
		encF = append(encF, enc...)
	}

	return encF
}

// Decrypt - Decrypts a given byte sequence, given the secret used to encrypt it
func Decrypt(cypher []byte, secret []byte) []byte {
	blockCount := len(cypher) / 16
	blocks := make([][]byte, blockCount)

	log.Println("content len: ", len(cypher))
	log.Println("Block Count: ", blockCount)

	for blockStart := range blockCount {
		block := make([]byte, 16)
		copy(block, cypher[blockStart*16:blockStart*16+16])
		blocks[blockStart] = block
	}

	if len(cypher)%16 != 0 {
		// Padd the last block if
		block := make([]byte, 16)
		copy(block, cypher[blockCount*16:])
		blocks = append(blocks, block)

	}

	dencF := make([]byte, 0)
	for _, block := range blocks {

		denc := make([]byte, 16)
		for i, byt := range block {
			if byt|3 == 0 {
				for _, b := range secret[:len(secret)/2] {
					denc[i] = b ^ byt
				}
			} else {
				for _, b := range secret[len(secret)/2:] {
					denc[i] = byt ^ b
				}
			}
		}
		dencF = append(dencF, denc...)
	}

	return dencF
}

func singleByteHash(i byte) byte {
	return (((((i<<5)%0x4e ^ 5) - i) | i) >> (i % 3)) & 0xfe % 0xff
}

func GenerateSecret(rawBytes []byte, when ...time.Time) (secret []byte, generatedAt time.Time) {
	if len(when) != 0 {
		log.Println("Using time")
		generatedAt = when[0]
	} else {
		generatedAt = time.Now()
	}
	secret = make([]byte, 0)

	indexFromTheHours := singleByteHash(byte(generatedAt.Hour()))
	indexFromTheMinutes := singleByteHash(byte(generatedAt.Minute()))
	indexFromTheSeconds := singleByteHash(byte(generatedAt.Second()))

	secret = append(secret, rawBytes[indexFromTheMinutes:int(indexFromTheMinutes)+16]...)
	secret = append(secret, rawBytes[indexFromTheHours:int(indexFromTheHours)+16]...)
	secret = append(secret, rawBytes[indexFromTheSeconds:int(indexFromTheSeconds)+16]...)

	// log.Printf("Size is :%d\n", len(secret))

	// log.Printf("Secret is: \n%x\n", secret)

	return
}

// GenerateRawRandomBytes - Generates raw random bytes
func GenerateRawRandomBytes() (rawBytes []byte, err error) {
	rawBytes = make([]byte, 16*100)

	_, err = rand.Read(rawBytes)
	if err != nil {
		return nil, err
	}

	return
}
