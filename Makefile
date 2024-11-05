out_dir=./out

build_srv:
	go build -o $(out_dir)/server ./cmd/server.go
