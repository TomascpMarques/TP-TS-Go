out_dir=./out

# Server ------------------------------------------

build_srv:
	go build -o $(out_dir)/server ./cmd/server/srv.go

run_srv:
	go run ./cmd/server/srv.go

# Encryptr -------------------------------------------

build_cryptr:
	go build -o $(out_dir)/cryptr ./cmd/cryptr/cryptr.go

# Client ---------------------------------------------

build_client:
	go build -o $(out_dir)/client ./cmd/client/client.go
