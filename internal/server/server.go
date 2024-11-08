package server

import (
	"bufio"
	"crypto/rand"
	"errors"
	"fmt"
	"io"
	"log"
	"net"

	msgpack "github.com/vmihailenco/msgpack/v5"

	crypto "github.com/TP-TS-Go/internal/crypto"
	msgpacktyps "github.com/TP-TS-Go/internal/msgpack_typs"
)

type ServerState struct {
	rooms   map[string]chan msgpacktyps.Message
	clients map[string]string
	secret  []byte
}

// RegisterNewClient - Returns a new cryptographicly seccure generated ID,
// after adding the new client id to the server state, and raw material to build a secret.
func (ss *ServerState) RegisterNewClient(clientAddr string) (string, string) {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		log.Fatalf("erro ao tentar gerar id: %s", err.Error())
	}

	s, err := crypto.GenerateRawRandomBytes()
	if err != nil {
		log.Fatalf("erro ao tentar gerar secret raw material: %s", err.Error())
	}

	ss.clients[clientAddr] = fmt.Sprintf("%x", b)

	return fmt.Sprintf("%x", b), fmt.Sprintf("%x", s)
}

func NewServerState() *ServerState {
	return &ServerState{
		rooms:   make(map[string]chan msgpacktyps.Message),
		clients: make(map[string]string),
	}
}

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

			id, secretRawMaterial := serverState.RegisterNewClient(con.LocalAddr().String())
			msg := msgpacktyps.NewMessage(
				msgpacktyps.RequestIdResponse,
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
			log.Printf("Wrote: %d", n)

		case msgpacktyps.SendContent:

			msgg := msgpacktyps.NewMessage(
				msgpacktyps.RequestIdResponse,
				"",
				[]byte("Received")...,
			)

			data, err := msgpack.Marshal(msgg)
			if err != nil {
				log.Fatalf("erro ao encodificar mensagem: %s", err.Error())
			}

			n, err := con.Write(append(data, 0x0a))
			if err != nil {
				log.Fatal(err)
			}

			log.Printf("Wrote: %d", n)

		default:
			log.Println("tipo nao implementado, ignorar....")
			continue
		}
	}
}
