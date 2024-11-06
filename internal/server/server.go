package server

import (
	"bufio"
	"log"
	"net"
)

type Message struct {
	Length  byte
	Target  byte
	Content []byte
}

func HandleNewConnection(con net.Conn) {
	log.Println("New Connection!")

	buf := bufio.NewScanner(con)

	for buf.Scan() {
		data := buf.Text()
		log.Printf("RECEIVED DATA: %s", data)

		_, _ = con.Write(append(buf.Bytes(), '\n'))

		if buf.Text() == "#close#\n" {
			break
		}
	}

	if buf.Err() != nil {
		log.Printf("ERRO on buf scanner: %s", buf.Err().Error())
	}

	err := con.Close()
	if err != nil {
		log.Fatalf("ERROR: %s", err.Error())
	}
}
