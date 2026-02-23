.PHONY: all build test lint clean install

BINARY := gtd-cli
MAIN_PATH := ./cmd/gtd-cli

all: build

build:
	go build -o $(BINARY) $(MAIN_PATH)

test:
	go test ./... -v

test-cover:
	go test ./... -coverprofile=coverage.out
	go tool cover -html=coverage.out

lint:
	@which golangci-lint > /dev/null || go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	golangci-lint run ./...

clean:
	rm -f $(BINARY)
	go clean ./...

install: build
	cp $(BINARY) ~/bin/

dev: build
	./$(BINARY) --help
