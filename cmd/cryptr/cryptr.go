package main

import (
	"log"
	"os"

	"github.com/TP-TS-Go/internal/crypto"
)

func main() {
	// Ignore the prog name in the Args
	args := os.Args[1:]

	if len(args) > 3 {
		log.Fatalf("Demasiados argumentos")
	}

	rb, _ := crypto.GenerateRawRandomBytes(32)
	secret, _, _ := crypto.GenerateSecret(rb)

	data, _ := os.ReadFile(args[0])

	if len(data) < 1 {
		panic("read nothing of the file")
	}

	cypher, _ := crypto.Encrypt(data, secret)

	decripted, _ := crypto.Decrypt(cypher, secret)

	_ = os.WriteFile(args[1], cypher, 0644)
	_ = os.WriteFile(args[2], decripted, 0644)
}
