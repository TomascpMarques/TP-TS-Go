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
	"net/url"
	"os"

	"github.com/gorilla/websocket"

	"github.com/TP-TS-Go/internal/crypto"
)

type GivenClientInformation struct {
	IdBytes  []byte
	IdCookie *http.Cookie
}

// getArgsNoProg returns the args the program was called with, and ignores the first one (the program name)
func getArgsNoProg() []string {
	return os.Args[1:]
}

// Usage wserverc 0.0.0.0 8080
func main() {
	args := getArgsNoProg()

	if len(args) != 1 {
		log.Fatal("Numero invalido de argumentos!")
	}

	serverHostname := args[0]

	baseServerUrl, err := url.Parse(fmt.Sprintf("https://%s", serverHostname))
	if err != nil {
		log.Fatalf("erro ao criar server URL: %s", err)
	}

	log.Printf("Connecting to <%s> ...\n", baseServerUrl)

	serverPublicMaterialUrl := baseServerUrl.JoinPath("public", "identity")

	materialResp, err := http.Get(serverPublicMaterialUrl.String())
	if err != nil {
		log.Fatalf("erro ao ler raw public material: %s", err.Error())
	}

	material, err := io.ReadAll(materialResp.Body)
	if err != nil {
		log.Fatalf("error reading the response body: %s", err.Error())
	}

	serverCreateNewClientUrl := baseServerUrl.JoinPath("create", "client")

	resp, err := http.Post(serverCreateNewClientUrl.String(), "", nil)
	if err != nil {
		log.Fatalf("erro ao criar user: %s", err.Error())
	}
	defer resp.Body.Close()

	clientInfo := GivenClientInformation{
		IdBytes:  make([]byte, 0),
		IdCookie: nil,
	}

	for _, x := range resp.Cookies() {
		if x.Name == "client" {
			clientInfo.IdCookie = x

			clientInfo.IdBytes, err = hex.DecodeString(x.Value)
			if err != nil {
				log.Fatalf("erro ao decode hex from cookie: %s", err.Error())
			}

			clientInfo.IdCookie.Value = fmt.Sprintf("%x", clientInfo.IdBytes)
			log.Printf("Given ID: %x", clientInfo.IdBytes)
		}
	}

	secret := hashMaterialWithClientId(material, clientInfo)

	clientSecret, _, err := crypto.GenerateSecret(secret)
	if err != nil {
		log.Fatalf("Erro ao gerar raw bytes for secret: %s", err.Error())
	}

	log.Printf("The secret is: %x", clientSecret)

	serverEnterChatRoomUrl := baseServerUrl.JoinPath("chat")
	serverEnterChatRoomUrl.Scheme = "wss"

	request, _ := http.NewRequest("GET", serverEnterChatRoomUrl.String(), nil)
	request.AddCookie(clientInfo.IdCookie)

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

func hashMaterialWithClientId(material []byte, clientInfo GivenClientInformation) []byte {
	hash := sha256.New()
	hash.Write(append(material, clientInfo.IdBytes...))
	secret := hash.Sum(nil)
	return secret
}
