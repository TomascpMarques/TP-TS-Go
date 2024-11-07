out_dir=./out

# Cria o diretório de saída, se não existir
$(out_dir):
	mkdir -p $(out_dir)

# Server ------------------------------------------

build_srv: $(out_dir)
	go build -o $(out_dir)/server ./cmd/server/srv.go

run_srv:
	go run ./cmd/server/srv.go

# Encryptr -------------------------------------------

build_cryptr: $(out_dir)
	go build -o $(out_dir)/cryptr ./cmd/cryptr/cryptr.go

# Client ---------------------------------------------

build_client: $(out_dir)
	go build -o $(out_dir)/client ./cmd/client/client.go

# All ------------------------------------------------
all: build_client build_cryptr build_srv
