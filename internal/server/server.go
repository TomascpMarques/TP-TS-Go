package server

import (
	"bufio"
	"crypto/rand"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"strconv"
	"strings"
	"time"

	msgpack "github.com/vmihailenco/msgpack/v5"

	crypto "github.com/TP-TS-Go/internal/crypto"
	msgpacktyps "github.com/TP-TS-Go/internal/msgpack_typs"
)

type ServerState struct {
	rawMaterial []byte
}

// RegisterNewClient - Returns a new cryptographicly seccure generated ID,
// after adding the new client id to the server state, and raw material to build a secret.
func (ss *ServerState) RegisterNewClient(connection net.Conn) (string, string) {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		log.Fatalf("erro ao tentar gerar id: %s", err.Error())
	}

	clientId := fmt.Sprintf("%x", b)

	return clientId, fmt.Sprintf("%x", ss.rawMaterial)
}

func NewServerState() *ServerState {
	s, err := crypto.GenerateRawRandomBytes()
	if err != nil {
		log.Fatalf("erro ao tentar gerar secret raw material: %s", err.Error())
	}

	return &ServerState{
		rawMaterial: s,
	}
}

var (
	connections        = make(map[string]net.Conn)
	connectionsSecrets = make(map[string][]byte)
)

func HandleNewConnection(con net.Conn, serverState *ServerState) {
	log.Printf("New Connection!")

	buf := bufio.NewReader(con)
	var decoder *msgpack.Decoder

	for {
		decoder = msgpack.NewDecoder(buf)
		log.Printf("RECEIVED SOME DATA")

		// The loop pauses here waiting for the decoder to receive any new data
		var msg msgpacktyps.Message

		err := decoder.Decode(&msg)
		log.Printf("DECODED SOME DATA")

		if err != nil && errors.Is(err, io.EOF) {
			log.Println("conexao terminada")
			break
		}

		if err != nil {
			log.Printf("erro ao decodificar o MsgPack packet: %s", err.Error())
		}

		// Deal with the message type and act accordingly
		switch msg.Type {
		case msgpacktyps.RequestId:

			id, secretRawMaterial := serverState.RegisterNewClient(con)

			msg := msgpacktyps.NewMessage(
				msgpacktyps.RequestIdResponse,
				"",
				"",
				[]byte(fmt.Sprintf("%s|%s", id, secretRawMaterial))...,
			)

			data, err := msgpack.Marshal(&msg)
			if err != nil {
				log.Fatalf("erro ao encodificar mensagem: %s", err.Error())
			}

			n, err := con.Write(append(data, 0x0a))
			if err != nil {
				log.Fatal(err)
			}
			log.Printf("Wrote on RequestId: %d", n)

		case msgpacktyps.SendContent:

			log.Println("SEND CONTENT==============================")

			if _, exists := connections[msg.SenderId]; !exists {
				connections[msg.SenderId] = con
			}

			if strings.Contains(string(msg.Content), "#!createSecret") {
				// Criar secret
				clientCreatedAt, err := strconv.ParseInt(
					string(msg.Content[len("#!createSecret")+1:]),
					10,
					64,
				)

				log.Println("Created AT: ", clientCreatedAt)

				if err != nil {
					log.Fatal("Cant parsse the date/time")
				}

				t := time.Unix(clientCreatedAt, 0)
				secret, _ := crypto.GenerateSecret(serverState.rawMaterial, t)

				connectionsSecrets[msg.SenderId] = secret
				log.Printf("SECRET: %x", secret)
				continue
			}

			data, err := msgpack.Marshal(msg)
			if err != nil {
				log.Fatalf("erro ao encodificar mensagem: %s", err.Error())
			}

			for _, connection := range connections {
				_, err := connection.Write(append(data, 0x0a))
				if err != nil {
					delete(connections, msg.SenderId)
					log.Println(err)
				}
			}
		default:
			log.Println("tipo nao implementado, ignorar....")
			continue
		}
	}
}
