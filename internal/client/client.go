package client

import (
	"bytes"
	"log"
	"sync"
	"time"

	"github.com/pelletier/go-toml/v2"
	"github.com/vmihailenco/msgpack/v5"

	msgpacktyps "github.com/TP-TS-Go/internal/msgpack_typs"
)

// InitClient - Cria os ficheiros de configuracao e pede ID ao server
func InitClient(args []string) {
	if len(args) != 1 {
		log.Fatalf("numero de argumentos errado para iniciar a cli")
	}

	// FIX - Nao verifica se o endereco do server contem uma port
	config := Config{
		CreatedAt:     time.Now().UnixMilli(),
		ServerAddress: args[0],
	}

	// Connect to the server specefied in the config file
	comHandler := NewComHandler("", args[0], config.ServerAddress)

	var wg sync.WaitGroup
	wg.Add(1)

	comHandler.SetOnMsgReceive(func(m msgpacktyps.Message) {
		// Handle config response
		config.ClientId = string(m.Content[0:16])
		config.RawMaterial = string(m.Content[18:])
		comHandler.senderId = config.ClientId

		wg.Done()
	})

	err := comHandler.CreateConnection()
	if err != nil {
		log.Fatalf("erro ao iniciar o comHandler: %s", err.Error())
	}

	var encoderBuffer bytes.Buffer
	encoder := msgpack.NewEncoder(&encoderBuffer)
	// encoder.UseArrayEncodedStructs(true)

	msg := msgpacktyps.NewMessage(msgpacktyps.RequestId, "", "0", 0x0)

	err = encoder.Encode(msg)
	if err != nil {
		log.Fatalf("erro ao encodificar mensagem: %s", err.Error())
	}

	_, err = comHandler.connection.Write(encoderBuffer.Bytes())
	if err != nil {
		log.Fatal(err)
	}

	wg.Wait()

	configMarshaled, err := toml.Marshal(config)
	if err != nil {
		log.Fatalf("falha ao tornar a configuracao em TOML")
	}

	file, err := createConfFile()
	if err != nil {
		log.Fatal(err.Error())
	}

	if _, err := file.Write(configMarshaled); err != nil {
		log.Fatalf("erro ao escrever a nova configuracao")
	}

	log.Println("Client config success!")
}
