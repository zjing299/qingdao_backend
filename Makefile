BINARY_NAME := qingdao_backend
.DEFAULT_GOAL := help
.PHONY: help build build-linux run test fmt tidy
help:
	@echo "Targets: build build-linux run test fmt tidy"
build:
	go build -trimpath -o dist/$(BINARY_NAME) .
build-linux:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -o dist/$(BINARY_NAME) .
run:
	go run .
test:
	go test ./...
fmt:
	go fmt ./...
tidy:
	go mod tidy
