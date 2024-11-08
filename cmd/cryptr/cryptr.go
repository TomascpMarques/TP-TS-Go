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

	rb, _ := crypto.GenerateRawRandomBytes()
	secret, _ := crypto.GenerateSecret(rb)

	data, _ := os.ReadFile(args[0])

	if len(data) < 1 {
		panic("read nothing of the file")
	}

	cypher := crypto.Encrypt(data, secret)
	decripted := crypto.Decrypt(cypher, secret)

	_ = os.WriteFile(args[1], decripted, 0644)
}
