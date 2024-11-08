package main

import (
	"crypto/rand"
	"fmt"
	"log"
	"os"
)

func main() {
	// Ignore the prog name in the Args
	args := os.Args[1:]

	if len(args) > 3 {
		log.Fatalf("Demasiados argumentos")
	}

	b := make([]byte, 1600)
	_, err := rand.Read(b)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	fmt.Printf("%x\n", b)

	for _, arg := range args {
		fmt.Printf("\nARG: %s\n", arg)
	}
}
