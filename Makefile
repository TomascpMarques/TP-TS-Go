out_dir=./out

# Cria o diretório de saída, se não existir
$(out_dir):
	mkdir -p $(out_dir)

# Web Server & Client -----------------------------------
build_wsrv: $(out_dir)
	go build -o $(out_dir)/wserver ./cmd/wserver/wserver.go

run_wsrv:
	go run ./cmd/wserver/wserver.go

build_wsrvc:
	go build -o $(out_dir)/wserverc ./cmd/wserverc/wserverc.go

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
