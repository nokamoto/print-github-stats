HASH = $(shell git rev-parse HEAD)
TIMESTAMP = $(shell date)

all: go.mod
	go fmt ./...
	go test ./...
	go build -ldflags "-X \"main.buildTimeHash=$(HASH)\" -X \"main.buildTimeTimestamp=$(TIMESTAMP)\"" .
	go mod tidy

go.mod:
	go mod init github.com/nokamoto/print-github-stats
