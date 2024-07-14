# SPDX-License-Identifier: Apache-2.0
# SPDX-FileCopyrightText: 2024-Present Harry Randazzo

.DEFAULT_GOAL := build

build:
	CGO_ENABLED=0 go build -o bin/ -ldflags="-s -w" ./cmd/vai

alias:
	@echo "alias vai='$(PWD)/bin/vai'" >>  ~/.config/fish/config.fish
	@echo "VAI_COMPLETION=true vai completion fish | source" >> ~/.config/fish/config.fish

hello-world:
	echo "Hello, World!"
