# SPDX-License-Identifier: Apache-2.0
# SPDX-FileCopyrightText: 2024-Present Harry Randazzo

.DEFAULT_GOAL := build

build:
	CGO_ENABLED=0 go build -o bin/ -ldflags="-s -w" ./cmd/vai

test:
	go test -race -cover -coverprofile=coverage.out -failfast -timeout 3m ./...

test-short:
	go test -race -cover -coverprofile=coverage.out -failfast -timeout 3m ./... -short

clean:
	rm -rf bin/

hello-world:
	echo "Hello, World!"

benchmark:
	hyperfine './bin/vai hello-world' 'make hello-world' -N --warmup 10
