package main

import (
	"log"
	"os"

	"github.com/TP-TS-Go/internal/crypto"
)

func main() {
	// Retrieve command-line arguments (ignoring the program name)
	args := os.Args[1:]

	// Ensure no more than 3 arguments are provided
	if len(args) > 3 {
		log.Fatalf("Too many arguments. Expected 3 arguments: input file, output encrypted file, output decrypted file.")
	}

	// Generate raw random bytes to use as a base for secret generation
	rb, err := crypto.GenerateRawRandomBytes(32)
	if err != nil {
		log.Fatalf("Failed to generate random bytes: %v", err)
	}

	// Generate a derived secret based on the random bytes
	secret, _, err := crypto.GenerateSecret(rb)
	if err != nil {
		log.Fatalf("Failed to generate secret: %v", err)
	}

	// Read the input file specified in the arguments
	data, err := os.ReadFile(args[0])
	if err != nil {
		log.Fatalf("Failed to read input file: %v", err)
	}

	// Check if the file is empty
	if len(data) < 1 {
		log.Fatalf("Input file is empty.")
	}

	// Encrypt the data using the generated secret
	ciphertext, err := crypto.Encrypt(data, secret)
	if err != nil {
		log.Fatalf("Failed to encrypt data: %v", err)
	}

	// Decrypt the ciphertext back to original data
	decrypted, err := crypto.Decrypt(ciphertext, secret)
	if err != nil {
		log.Fatalf("Failed to decrypt data: %v", err)
	}

	// Write the encrypted data to the specified output file
	if err := os.WriteFile(args[1], ciphertext, 0644); err != nil {
		log.Fatalf("Failed to write encrypted file: %v", err)
	}

	// Write the decrypted data to the specified output file
	if err := os.WriteFile(args[2], decrypted, 0644); err != nil {
		log.Fatalf("Failed to write decrypted file: %v", err)
	}

	log.Println("Encryption and decryption processes completed successfully.")
}
