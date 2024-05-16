
.DEFAULT_GOAL := build

build:
	CGO_ENABLED=0 go build -o bin/ -ldflags="-s -w" ./cmd/vai

test:
	go test -race -cover -failfast -timeout 3m ./...

clean:
	rm -rf bin/

hello-world:
	echo "Hello, World!"
