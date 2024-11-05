package main

import (
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

	for _, arg := range args {
		fmt.Printf("ARG: %s\n", arg)
	}
}
