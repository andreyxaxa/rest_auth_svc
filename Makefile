.PHONY: build
build:
		go build -v ./cmd/auth

.PHONY: test
test:
		go test -v -race ./...

.DEFAULT_GOAL := build