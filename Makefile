
all: go.mod
	go fmt
	go test ./...
	go build .
	go mod tidy

go.mod:
	go mod init github.com/nokamoto/print-github-stats
