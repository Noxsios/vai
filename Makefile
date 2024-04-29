
.DEFAULT_GOAL := build

build:
	CGO_ENABLED=0 go build -o bin/ -ldflags="-s -w" ./cmd/vai

clean:
	rm -rf bin/

hello-world:
	echo "Hello, World!"
