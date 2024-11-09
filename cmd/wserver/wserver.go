package main

import (
	"context"
	"crypto/sha256"
	"fmt"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/TP-TS-Go/internal/crypto"
)

const (
	HOST = "0.0.0.0"
	PORT = 8080
)

type ServerState struct {
	clientsSecrets    map[string][]byte
	publicRawMaterial []byte
}

func NewServerState() *ServerState {
	rawMaterial, err := crypto.GenerateRawRandomBytes(32)
	if err != nil {
		log.Fatalf("falha ao gerar raw material para o public: %s", err.Error())
	}

	return &ServerState{
		clientsSecrets:    make(map[string][]byte),
		publicRawMaterial: rawMaterial,
	}
}

func (ss *ServerState) AddClientSecret(clientId string, clientSecret []byte) {
	ss.clientsSecrets[clientId] = clientSecret
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
	app.GET("/room/connect", func(ctx *gin.Context) {
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
	c.String(http.StatusOK, "Hello")
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
