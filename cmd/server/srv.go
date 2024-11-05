package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
)

const (
	PORT = 8080
	HOST = "0.0.0.0"
)

func main() {
	address := fmt.Sprintf("%s:%d", HOST, PORT)

	tcp_listener, err := net.Listen("tcp", address)
	if err != nil {
		log.Fatalf("ERRO: %s", err.Error())
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTeapot)
		_, _ = w.Write([]byte("I'm a tea pot"))
	})

	log.Printf("Server a correr em http://%s\n", address)
	log.Fatal(http.Serve(tcp_listener, nil))
}
