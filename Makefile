.PHONY: build
build:
		go build -v ./cmd/auth

.DEFAULT_GOAL := build