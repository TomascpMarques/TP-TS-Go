package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"

	"github.com/TP-TS-Go/internal/crypto"
)

const (
	HOST = "0.0.0.0"
	PORT = 8080
)

type Client struct {
	Id           []byte
	Secret       []byte
	RecvChannel  chan []byte
	WsConnection *websocket.Conn
}

func NewClient(id []byte, wsc *websocket.Conn) *Client {
	client := &Client{
		Id:           id,
		RecvChannel:  make(chan []byte),
		WsConnection: wsc,
	}

	go func() {
		for {
			msg := <-client.RecvChannel
			err := client.WsConnection.WriteMessage(websocket.TextMessage, msg)
			if err != nil {
				log.Printf("falha ao escrever a mensagem de outro client: %s", err.Error())
			}
		}
	}()

	return client
}

type Room struct {
	clients map[string]*Client
}

func (r *Room) BroadcastMsg(msgType int, msg []byte) (err error) {
	for _, target := range r.clients {

		log.Printf("BROAD TARGET ID: %x", target.Id)
		// log.Printf("BROAD TARGET secret: %x", target.Secret)

		encMessage, err := crypto.Encrypt(msg, target.Secret)
		if err != nil {
			log.Fatalf("BROAD falha ao encrypt msg para o user %x: %s", target.Id, err.Error())
		}
		_ = encMessage

		// log.Printf("ENC MSG: %x", encMessage)

		err = target.WsConnection.WriteMessage(msgType, encMessage)
		if err != nil {
			log.Printf("falha ao enviar msg em broadcast: %s", err.Error())
			break
		}
	}
	return
}

func (r *Room) SendMsg(msgType int, msg []byte, target string) error {
	msgTarget, exists := r.clients[target]
	if !exists {
		return fmt.Errorf("alvo nao existe")
	}

	// usrId, _ := hex.DecodeString(msgTarget.Id)
	// userSecret := r.clients[msgTarget.Id].Secret

	encMessage, err := crypto.Encrypt(msg, msgTarget.Id)
	if err != nil {
		log.Printf("falha ao encrypt msg: %s", err.Error())
	}

	err = msgTarget.WsConnection.WriteMessage(msgType, encMessage)
	if err != nil {
		log.Printf("falha ao enviar msg para um cliente : %s", err.Error())
	}

	// Will send nil if the last if statement does not run
	return err
}

type ServerState struct {
	publicRawMaterial []byte
	Room
}

func NewServerState() *ServerState {
	rawMaterial, err := crypto.GenerateRawRandomBytes(32)
	if err != nil {
		log.Fatalf("falha ao gerar raw material para o public: %s", err.Error())
	}
	state := ServerState{
		publicRawMaterial: rawMaterial,
		Room: Room{
			clients: make(map[string]*Client),
		},
	}

	return &state
}

func (ss *ServerState) ResgisterNewClient(clientId []byte, clientSecret []byte) {
	id := fmt.Sprintf("%x", clientId)

	ss.Room.clients[id] = &Client{
		Id:     clientId,
		Secret: clientSecret,
	}

	log.Printf("Client Secret: %x", ss.Room.clients[id].Secret)
}

var WsUpgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
	EnableCompression: false,
}

func main() {
	address := fmt.Sprintf("%s:%d", HOST, PORT)

	tcp_listener_conf := net.ListenConfig{
		KeepAlive: time.Minute * 5,
		KeepAliveConfig: net.KeepAliveConfig{
			Enable: true,
			Idle:   time.Minute,
		},
	}

	tcp_listener, err := tcp_listener_conf.Listen(context.Background(), "tcp", address)
	if err != nil {
		log.Fatalf("Erro ao iniciar o server: %s", err.Error())
	}

	app := gin.Default()
	serverState := NewServerState()

	app.GET("/public/identity", func(ctx *gin.Context) {
		getServerRawPublicMaterial(ctx, serverState)
	})

	app.POST("/create/client", func(ctx *gin.Context) {
		newClient(ctx, serverState)
	})
	app.GET("/chat", func(ctx *gin.Context) {
		connectToRoom(ctx, serverState)
	})

	log.Fatal(app.RunListener(tcp_listener))
}

func getServerRawPublicMaterial(c *gin.Context, ss *ServerState) {
	c.Data(http.StatusOK, http.DetectContentType(ss.publicRawMaterial), ss.publicRawMaterial)
}

// newClient creates the new client ID and secret, returns only the ID and stores the custom secret.
// The client will then use the public available raw material, and based on its OWN ID and a Sha256 algo, will generate the same secret on its own
func newClient(c *gin.Context, ss *ServerState) {
	clientId, clientSecret := generateNewClientData(ss.publicRawMaterial)

	// log.Printf("The secret is: %x", clientSecret)
	// log.Printf("The clientID is: %x", clientId)
	ss.ResgisterNewClient(clientId, clientSecret)

	c.SetCookie("client", fmt.Sprintf("%x", clientId), 3600, "/", "localhost", false, true)
	c.Status(http.StatusOK)
}

func connectToRoom(c *gin.Context, ss *ServerState) {
	writer, request := c.Writer, c.Request

	clientId, err := c.Request.Cookie("client")
	if err != nil {
		log.Printf("no client ID specefied, no cookie found")
		c.Status(http.StatusBadRequest)
		return
	}

	ws, err := WsUpgrader.Upgrade(writer, request, nil)
	if err != nil {
		log.Printf("erro ao dar upgrad da conexão: %s", err.Error())
	}
	defer ws.Close()

	x1, err := hex.DecodeString(clientId.Value)
	if err != nil {
		log.Fatalf("erro ao decode hex from cookie: %s", err.Error())
	}
	currentClientId := fmt.Sprintf("%x", x1)

	client, clientExists := ss.Room.clients[currentClientId]

	if !clientExists {
		log.Fatalf("erro ao obter o cliente: %s", currentClientId)
	}
	client.WsConnection = ws

	for {
		mt, message, err := ws.ReadMessage()

		if err != nil && mt != -1 {
			log.Printf("Failed to read the message: %s", err.Error())
		}
		if mt == -1 {
			log.Println("closing WsConnection, could be an error on the client, could be a Ctrl-C ...")
			return
		}

		dencMessage, err := crypto.Decrypt(message, ss.clients[fmt.Sprintf("%x", x1)].Secret)
		if err != nil {
			fmt.Printf("failed to encrypt the messages")
		}

		_ = ss.Room.BroadcastMsg(mt, dencMessage)
	}
}

// generateNewClientData generates a client ID and client specific secret
func generateNewClientData(rawBytes []byte) ([]byte, []byte) {
	clientID, err := crypto.GenerateRawRandomBytes(24)
	if err != nil {
		log.Fatalf("Erro ao gerar ID do cliente: %s", err.Error())
	}

	// Rehash the secret with the client id
	hash := sha256.New()
	hash.Write(append(rawBytes, clientID...))
	rawBytesForSecret := hash.Sum(nil)

	clientSecret, _, err := crypto.GenerateSecret(rawBytesForSecret)
	if err != nil {
		log.Fatalf("Erro ao gerar raw bytes for secret: %s", err.Error())
	}

	return clientID, clientSecret
}
