package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"
	"net"
	"net/http"
	"strings"
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
	Id           string
	RecvChannel  chan []byte
	WsConnection *websocket.Conn
}

func NewClient(id string, wsc *websocket.Conn) *Client {
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
	clients     map[string]*Client
	sendChannel chan []byte
}

func (r *Room) RegisterNewClient(client *Client) {
	r.clients[client.Id] = client
}

func (r *Room) BroadcastMsg(msgType int, msg []byte) (err error) {
	for _, target := range r.clients {
		err = target.WsConnection.WriteMessage(msgType, msg)
		if err != nil {
			log.Printf("falha ao enviar msg em broadcast: %s", err.Error())
			break
		}
	}
	return
}

func (r *ServerState) SendMsg(msgType int, msg []byte, target string) error {
	msgTarget, exists := r.Room.clients[target]
	if !exists {
		return fmt.Errorf("alvo nao existe")
	}

	usrId, _ := hex.DecodeString(msgTarget.Id)

	encMessage, err := crypto.Encrypt(msg, usrId)
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
	clientsSecrets    map[string][]byte
	publicRawMaterial []byte
	*Room
}

func NewServerState() *ServerState {
	rawMaterial, err := crypto.GenerateRawRandomBytes(32)
	if err != nil {
		log.Fatalf("falha ao gerar raw material para o public: %s", err.Error())
	}
	state := ServerState{
		clientsSecrets:    make(map[string][]byte),
		publicRawMaterial: rawMaterial,
		Room: &Room{
			clients:     make(map[string]*Client),
			sendChannel: make(chan []byte, 256),
		},
	}

	return &state
}

func (ss *ServerState) AddClientSecret(clientId string, clientSecret []byte) {
	ss.clientsSecrets[clientId] = clientSecret
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
	app.GET("/chat/:method/*target", func(ctx *gin.Context) {
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

	ss.AddClientSecret(clientId, clientSecret)

	log.Printf("The secret is: %x", clientSecret)

	c.SetCookie("client", clientId, 3600, "/", "localhost", false, true)
	c.Status(http.StatusOK)
}

func connectToRoom(c *gin.Context, ss *ServerState) {
	writer, request := c.Writer, c.Request

	// Comunication method
	method := c.Params.ByName("method")

	clientWantsToBroadcast := false
	if method == "broadcast" {
		log.Println("On broadcast")
		clientWantsToBroadcast = true
	}
	targetClient := c.Params.ByName("target")
	if strings.Trim(targetClient, " \n\t") == "" {
		c.Status(http.StatusBadRequest)
	}

	// clientId, err := c.Cookie("client")
	// if errors.Is(err, http.ErrNoCookie) {
	// 	log.Printf("no client ID specefied, no cookie found")
	// 	c.Status(http.StatusBadRequest)
	// 	return
	// }

	ws, err := WsUpgrader.Upgrade(writer, request, nil)
	if err != nil {
		log.Printf("erro ao dar upgrad da conexão: %s", err.Error())
	}
	defer ws.Close()

	client := NewClient(time.Now().UTC().GoString(), ws)
	ss.Room.RegisterNewClient(client)

	for {
		mt, message, err := ws.ReadMessage()
		if err != nil {
			log.Printf("Failed to read the message: %s", err.Error())
		}
		log.Printf("received: %s", message)

		if clientWantsToBroadcast {
			_ = ss.Room.BroadcastMsg(mt, message)
			continue
		}

		dencMessage, err := crypto.Decrypt(message, ss.clientsSecrets[client.Id])
		if err != nil {
			fmt.Printf("failed to encrypt the messages")
		}

		_ = ss.SendMsg(mt, dencMessage, targetClient)
	}
}

// generateNewClientData generates a client ID and client specific secret
func generateNewClientData(rawBytes []byte) (string, []byte) {
	clientID, err := crypto.GenerateRawRandomBytes(24)
	if err != nil {
		log.Fatalf("Erro ao gerar ID do cliente: %s", err.Error())
	}

	hash := sha256.New()
	hash.Write(append(rawBytes, clientID...))
	rawBytesForSecret := hash.Sum(nil)

	clientSecret, _, err := crypto.GenerateSecret(rawBytesForSecret)
	if err != nil {
		log.Fatalf("Erro ao gerar raw bytes for secret: %s", err.Error())
	}

	return fmt.Sprintf("%x", clientID), clientSecret
}
