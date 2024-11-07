package server

import (
	"bufio"
	"errors"
	"io"
	"log"
	"net"

	"github.com/vmihailenco/msgpack/v5"

	msgpacktyps "github.com/TP-TS-Go/internal/msgpack_typs"
)

func HandleNewConnection(con net.Conn) {
	log.Println("New Connection!")

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

		// Echo back the file contents
		_, _ = con.Write([]byte("Received\n"))
	}
}
