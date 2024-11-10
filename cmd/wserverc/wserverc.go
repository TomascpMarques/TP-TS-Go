package main

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/websocket"

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
	var clientIdCookie *http.Cookie

	for _, x := range resp.Cookies() {
		if x.Name == "client" {
			clientIdCookie = x
			x1, err := hex.DecodeString(x.Value)
			if err != nil {
				log.Fatalf("erro ao decode hex from cookie: %s", err.Error())
			}

			clientId = make([]byte, len(x1))
			copy(clientId, x1)

			clientIdCookie.Value = fmt.Sprintf("%x", clientId)
			log.Printf("ID: %x", clientId)
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

	request, _ := http.NewRequest("GET", "ws://localhost:8080/chat/broadcast/", nil)
	request.AddCookie(clientIdCookie)

	// client := &http.Client{}

	// resp, err = client.Do(request)
	// if err != nil {
	// 	fmt.Println("Error making request:", err)
	// 	return
	// }

	// Make the WebSocket connection
	ws, _, err := websocket.DefaultDialer.Dial(request.URL.String(), request.Header)
	if err != nil {
		fmt.Println("Error dialing:", err)
		return
	}

	go func() {
		for {
			_, msg, err := ws.ReadMessage()
			if err != nil {
				log.Printf("Erro ao ler: %s", err.Error())
			}

			denc, err := crypto.Decrypt(msg, clientSecret)
			if err != nil {
				log.Fatalf("erro ao desencriptar a msg: %s", err.Error())
			}

			log.Printf("Received message: %s", denc)
		}
	}()

	userInputBuffer := bufio.NewReader(os.Stdin)
	for {
		inputBytes, err := userInputBuffer.ReadBytes(0x0a)
		if err != nil && errors.Is(err, io.EOF) {
			ws.Close()
		}
		if err != nil {
			log.Fatalf("erro ao ler user input: %s", err.Error())
		}

		encMsg, err := crypto.Encrypt(inputBytes, clientSecret)
		if err != nil {
			log.Fatalf("erro ao encriptar a msg: %s", err.Error())
		}

		err = ws.WriteMessage(websocket.TextMessage, encMsg)
		if err != nil {
			log.Fatalf("erro ao escrever na conexao: %s", err.Error())
		}
	}
}
