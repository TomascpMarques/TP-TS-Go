package main

import (
	"context"
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
	clientsSecrets map[string][]byte
}

func NewServerState() *ServerState {
	return &ServerState{
		clientsSecrets: make(map[string][]byte),
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

	app.POST("/create/client", func(ctx *gin.Context) {
		newClient(ctx, serverState)
	})
	app.GET("/room/connect", func(ctx *gin.Context) {
		connectToRoom(ctx, serverState)
	})

	log.Fatal(app.RunListener(tcp_listener))
}

func newClient(c *gin.Context, ss *ServerState) {
	clientId, clientSecret := generateNewClientData()

	ss.AddClientSecret(clientId, clientSecret)

	c.SetCookie("client", clientId, 3600, "/", "localhost", false, true)
	c.Data(http.StatusOK, http.DetectContentType(clientSecret), clientSecret)
}

func connectToRoom(c *gin.Context, ss *ServerState) {
	c.String(http.StatusOK, "Hello")
}

// generateNewClientData generates a client ID and client specific secret
func generateNewClientData() (string, []byte) {
	clientID, err := crypto.GenerateRawRandomBytes(24)
	if err != nil {
		log.Fatalf("Erro ao gerar ID do cliente: %s", err.Error())
	}

	rawBytesForSecret, err := crypto.GenerateRawRandomBytes(32)
	if err != nil {
		log.Fatalf("Erro ao gerar raw bytes for secret: %s", err.Error())
	}

	clientSecret, _, err := crypto.GenerateSecret(rawBytesForSecret, time.Now())
	if err != nil {
		log.Fatalf("Erro ao gerar raw bytes for secret: %s", err.Error())
	}

	return fmt.Sprintf("%x", clientID), clientSecret
}
