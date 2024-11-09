package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/gorilla/websocket"

	"github.com/TP-TS-Go/internal/crypto"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

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

	c, _, err := websocket.DefaultDialer.Dial(socketURL, nil)
	if err != nil {
		log.Fatal("dial:", err)
	}
	defer c.Close()

	done := make(chan struct{})
	go func() {
		defer c.Close()
		defer close(done)
		for {
			_, message, err := c.ReadMessage()
			if err != nil {
				log.Println("read:",
					err)
				return
			}
			log.Printf("recv: %s", message)
		}
	}()

	for {
		err := c.WriteMessage(websocket.TextMessage, []byte("Hello, World!"))
		if err != nil {
			log.Println("write:", err)
			return
		}
		log.Println("sent: Hello, World!")
	}
}
