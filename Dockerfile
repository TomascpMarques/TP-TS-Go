ARG GO_VERSION=1
FROM golang:${GO_VERSION}-bookworm as builder

WORKDIR /usr/src/app
COPY . .
# COPY go.mod go.sum .
RUN go mod download && go mod tidy && go mod verify
RUN go build -v -o /run-app ./cmd/wserver/wserver.go


FROM debian:bookworm

COPY --from=builder /run-app /usr/local/bin/
CMD ["run-app"]
