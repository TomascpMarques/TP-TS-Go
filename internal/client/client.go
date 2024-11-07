package client

import (
	"log"
	"time"

	"github.com/pelletier/go-toml/v2"
)

// InitClient - Cria os ficheiros de configuracao e pede ID ao server
func InitClient(args []string) {
	if len(args) != 1 {
		log.Fatalf("numero de argumentos errado para iniciar a cli")
	}

	// TODO - Nao verifica se o endereco do server contem uma port
	config := Config{
		CreatedAt:     time.Now().UnixMilli(),
		ServerAddress: args[0],
	}

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
