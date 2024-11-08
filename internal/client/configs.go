package client

import (
	"fmt"
	"log"
	"os"
	"os/user"
	"path"

	"github.com/pelletier/go-toml/v2"
)

type Config struct {
	RawMaterial   string `toml:"raw_material"`
	CreatedAt     int64  `toml:"created_ts"`
	ClientId      string `toml:"client_id"`
	Secret        string `toml:"secret"`
	ServerAddress string `toml:"server"`
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

func writeToConfigFile(config Config) {
	path_conf, err := getConfFolderPath()
	if err != nil {
		log.Fatalf("erro: %s", err.Error())
	}

	if err := os.MkdirAll(path_conf, 0760); err != nil {
		log.Fatalf("erro ao escrever no ficheiro de configuracao: %s", err.Error())
	}

	file, err := os.Create(path.Join(path_conf, "cryptr.toml"))
	if err != nil {
		log.Fatalf("erro ao criar o ficheiro de configuracao: %s", err.Error())
	}

	content, err := toml.Marshal(config)
	if err != nil {
		log.Fatalf("falha ao ler a configuracao do cliente: %s", err.Error())
	}

	_, err = file.Write(content)
	if err != nil {
		log.Fatal(err)
	}
}
