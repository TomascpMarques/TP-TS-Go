package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"time"

	"github.com/TP-TS-Go/internal/server"
)

const (
	PORT = 9000
	HOST = "0.0.0.0"
	// MAX_CONS = 6
)

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
		log.Fatalf("ERRO - TCP LISTENER: %s", err.Error())
	}

	serverSate := server.NewServerState()

	for {
		conn, err := tcp_listener.Accept()
		if err != nil {
			log.Fatalf("ERRO - Con. ACCEPT : %s", err.Error())
		}

		// Handle new TCP Connection
		go server.HandleNewConnection(conn, serverSate)
	}
}
