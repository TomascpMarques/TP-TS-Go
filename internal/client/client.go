package client

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"os/user"
	"path"
	"sync"
	"time"

	"github.com/pelletier/go-toml/v2"
)

type Config struct {
	CreatedAt     int64  `toml:"created_ts"`
	ClientId      string `toml:"client_id"`
	Secret        string `toml:"secret"`
	ServerAddress string `toml:"server"`
}

// InitClient - Cria os ficheiros de configuracao e pede ID ao server
func InitClient(args []string) {
	if len(args) != 2 {
		log.Fatalf("numero de argumentos errado para iniciar a cli")
	}

	// TODO - Nao verifica se o endereco do server contem uma port
	config := Config{
		CreatedAt:     time.Now().UnixMilli(),
		ServerAddress: args[1],
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

func getConfFolderPath() (string, error) {
	// Get user info to create a TOML file in $(home)/.config/cryptr.toml
	currentUser, err := user.Current()
	if err != nil {
		return "", fmt.Errorf("erro ao obter o user da maquina: %s", err.Error())
	}

	confFolderPath := ".config"
	path_conf := path.Join(currentUser.HomeDir, confFolderPath)

	return path_conf, nil
}

func createConfFile() (*os.File, error) {
	path_conf, err := getConfFolderPath()
	if err != nil {
		return nil, fmt.Errorf("erro: %s", err.Error())
	}

	if err := os.MkdirAll(path_conf, 0760); err != nil {
		return nil, fmt.Errorf("erro ao criar o ficheiro de configuracao: %s", err.Error())
	}

	file, err := os.Create(path.Join(path_conf, "cryptr.toml"))
	if err != nil {
		return nil, fmt.Errorf("erro ao criar o ficheiro de configuracao: %s", err.Error())
	}

	return file, nil
}

func loadConfingFromFile() (config *Config, err error) {
	path_conf, err := getConfFolderPath()
	if err != nil {
		return nil, fmt.Errorf("erro: %s", err.Error())
	}

	confContents, err := os.ReadFile(path.Join(path_conf, "cryptr.toml"))
	if err != nil {
		return nil, fmt.Errorf("erro ao obter conteudos do ficheiro de conf: %s", err.Error())
	}

	config = &Config{}
	err = toml.Unmarshal(confContents, config)
	if err != nil {
		log.Fatalf("falha ao ler a configuracao do cliente: %s", err.Error())
	}

	return
}

func HandleServerComunication() {
	config, err := loadConfingFromFile()
	if err != nil {
		log.Fatalf("erro a ler configuracao: %s", err.Error())
	}

	// Connect to the server specefied in the config file
	conn, err := net.Dial("tcp", config.ServerAddress)
	if err != nil {
		log.Fatalf("erro ao iniciar conexao ao server: %e", err)
	}

	var wg sync.WaitGroup
	wg.Add(1)

	// Reads the data sent from the server, on another green thread
	go func() {
		responseBuffer := bufio.NewReader(conn)
		for {
			msg, err := responseBuffer.ReadString('\n')
			if err != nil {
				log.Fatal(err)
			}
			log.Printf("RESP: %s\n", msg)
		}
	}()

	// Handles user input and sends it to the server, on another green thread
	go func() {
		userInputBuffer := bufio.NewReader(os.Stdin)
		for {
			line, err := userInputBuffer.ReadString('\n')
			if err != nil {
				log.Fatalf("erro ao ler user input: %s", err.Error())
			}

			_, err = fmt.Fprintf(conn, "%s", line)
			if err != nil {
				log.Fatal(err)
			}

			if line == "#close#\n" {
				_ = conn.Close()
				wg.Done()
			}
		}
	}()

	// Waits for the user input function to be over,
	// this way, the parent function does not exit before the task is spawned
	wg.Wait()
}
