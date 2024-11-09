package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/TP-TS-Go/internal/crypto"
)

func main() {
	materialResp, err := http.Get("http://localhost:8080/public/identity")
	if err != nil {
		log.Fatalf("erro ao ler raw public material: %s", err.Error())
	}

	material, _ := io.ReadAll(materialResp.Body)

	resp, err := http.Post("http://localhost:8080/create/client", "", nil)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	defer resp.Body.Close()

	var clientId []byte

	for _, x := range resp.Cookies() {
		if x.Name == "client" {
			x1, err := hex.DecodeString(x.Value)
			if err != nil {
				log.Fatalf("erro ao decode hex from cookie: %s", err.Error())
			}

			clientId = make([]byte, len(x1))
			copy(clientId, x1)
		}
	}

	hash := sha256.New()
	hash.Write(append(material, clientId...))
	secret := hash.Sum(nil)

	clientSecret, _, err := crypto.GenerateSecret(secret)
	if err != nil {
		log.Fatalf("Erro ao gerar raw bytes for secret: %s", err.Error())
	}

	log.Printf("The secret is: %x", clientSecret)
}
