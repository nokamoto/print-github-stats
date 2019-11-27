
all: go.mod
	go fmt
	go test ./...

go.mod:
	go mod init github.com/nokamoto/print-github-stats
