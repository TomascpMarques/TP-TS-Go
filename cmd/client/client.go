package main

import (
	"log"
	"os"

	"github.com/TP-TS-Go/internal/client"
)

// ./app INIT server_id -> Sets up the required files, and requests the client ID from the server
// ./app SEND target msg -> usa o client_id, o target e o address do server
// ./app CREATE_SECRET -> usa o client_id, o timestamp_atual e avisa o server do mesmo processo com o timestamp_atual

// CLI ARGS
const (
	Init         = "INIT"
	Send         = "SEND"
	CreateSecret = "CREATE_SECRET"
)

func main() {
	args := os.Args[1:]

	if len(args) > 3 {
		log.Fatalf("Demasiados argumentos")
	}

	for i, arg := range args {
		switch arg {
		case Init:
			log.Println("Inicializar a applicacao...")
			// O user pode fazer: ./cli INIT  o que e invalido
			// O experado seria ./cli INIT 192.168.1.254
			// Dai enviarmos uma slice que pode ser vazia, o check e feito em InitClient
			client.InitClient(args[i:])
		case Send:
			log.Println("A enviar uma MSG")
			client.HandleServerComunication()
		case CreateSecret:
			log.Println("A criar o secret")
		}
	}
}
