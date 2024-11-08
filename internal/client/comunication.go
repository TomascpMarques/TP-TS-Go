package client

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"sync"
	"time"

	"github.com/vmihailenco/msgpack/v5"

	msgpacktyps "github.com/TP-TS-Go/internal/msgpack_typs"
)

type ComHandler struct {
	srvAddress string
	target     string
	connection net.Conn
	//---
	onMsgReceive func(msgpacktyps.Message)
	// ---
	listenConCloseChn   chan bool
	listenUsrIoCloseChn chan bool
}

func NewComHandler(target string, srvAddress string) *ComHandler {
	handler := ComHandler{
		target:     target,
		srvAddress: srvAddress,
	}

	return &handler
}

func (ch *ComHandler) spawnUserIoListenerRoutine() {
	userInputBuffer := bufio.NewReader(os.Stdin)

	go func() {
		for {
			inputBytes, err := userInputBuffer.ReadBytes('\n')
			if err != nil && errors.Is(err, io.EOF) {
				ch.connection.Close()
			}
			if err != nil {
				log.Fatalf("erro ao ler user input: %s", err.Error())
			}

			// Create and encode the message into the MsgPack Format
			msg := msgpacktyps.NewMessage(msgpacktyps.SendContent, ch.target, inputBytes...)

			b, err := msgpack.Marshal(&msg)
			if err != nil {
				log.Fatalf("erro no marshaling: %s", err.Error())
			}

			_, err = ch.connection.Write(b)
			if err != nil {
				log.Fatalf("erro ao escrever na conexao: %s", err.Error())
			}

			// if nn < len(inputBytes) {
			// 	log.Printf("[WARN] nao escreveu os bytes todos: %s", err.Error())
			// }
			// TODO - Handle connection close from server
		}
	}()

	<-ch.listenUsrIoCloseChn
	log.Println("QUITTING usrListener")
}

func (ch *ComHandler) spawnConnectionListenerRoutine() {
	connectionRespBuff := bufio.NewReader(ch.connection)

	go func() {
		for {
			data, err := connectionRespBuff.ReadBytes('\n')
			if err != nil {
				log.Fatal(err)
			}
			// Invoke the given handler

			var msgM msgpacktyps.Message
			err = msgpack.Unmarshal(data, &msgM)
			if err != nil {
				log.Fatalf("erro ao descodificar a msg: %s", err.Error())
			}

			ch.onMsgReceive(msgM)
		}
	}()

	<-ch.listenConCloseChn

	// Quit the listening routine
	log.Println("!CLOSING CONNECTION!")
	ch.connection.Close()
}

func (ch *ComHandler) CreateConnection() error {
	if ch.onMsgReceive == nil {
		return fmt.Errorf("impossivel criar conexao sem onResponseHandler")
	}

	conn, err := net.Dial("tcp", ch.srvAddress)
	if err != nil {
		return fmt.Errorf("falha ao iciar a conexao: %s", err.Error())
	}

	ch.connection = conn

	// Reads the data sent from the server, on a coroutine
	go ch.spawnConnectionListenerRoutine()

	// Reads the users input from the os.StdIo, on a coroutine
	go ch.spawnUserIoListenerRoutine()

	return nil
}

func (ch *ComHandler) SetOnMsgReceive(function func(msgpacktyps.Message)) {
	ch.onMsgReceive = function
}

func (ch *ComHandler) ShutDown() {
	ch.listenConCloseChn <- true
	ch.listenUsrIoCloseChn <- true
}

func testExample(msg msgpacktyps.Message) {
	log.Printf("MSG DATA: %s\n", msg.Content)
}

func HandleServerComunication(args []string) {
	if len(args) != 1 {
		log.Fatalf("demasiados argumentos para a funcao")
	}

	// Load configurations
	config, err := loadConfingFromFile()
	if err != nil {
		log.Fatalf("erro a ler configuracao: %s", err.Error())
	}

	// Connect to the server specefied in the config file
	comHandler := NewComHandler(args[0], config.ServerAddress)

	comHandler.SetOnMsgReceive(testExample)

	err = comHandler.CreateConnection()
	if err != nil {
		log.Fatalf("erro ao iniciar o comHandler: %s", err.Error())
	}

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		time.Sleep(time.Second * 5)
		comHandler.ShutDown()
		wg.Done()
		log.Println("Quitting")
	}()

	// TODO - Tell the server that I want to talk to the client with id == target

	wg.Wait()
}
