package server

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
)

func HandleNewConnection(con net.Conn) {
	log.Println("New Connection!")

	buf := bufio.NewScanner(con)

	for buf.Scan() {
		data := buf.Text()
		log.Printf("RECEIVED DATA: %s", data)

		_, _ = con.Write(append(buf.Bytes(), '\n'))

		fmt.Println(buf.Text())

		if buf.Text() == io.EOF.Error() {
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
