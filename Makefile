
.DEFAULT_GOAL := build

build:
	go build -o bin/ -ldflags="-s -w" ./cmd/vai

clean:
	rm -rf bin/
