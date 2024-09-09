# Makefile for wallet application

# Variables
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
BINARY_NAME=wallet
BINARY_UNIX=$(BINARY_NAME)_unix
PKG=./cmd

# Build targets
all: build

build:
	$(GOBUILD) -o $(BINARY_NAME) -v $(PKG)

test:
	$(GOTEST) -v ./...

clean:
	$(GOCLEAN)
	rm -f $(BINARY_NAME)
	rm -f $(BINARY_UNIX)

run:
	$(GOBUILD) -o $(BINARY_NAME) -v $(PKG)
	./$(BINARY_NAME)

deps:
	$(GOGET) github.com/btcsuite/btcd/rpcclient
	$(GOGET) github.com/spf13/cobra
	$(GOGET) github.com/pkg/errors
	$(GOGET) github.com/AlecAivazis/survey/v2
	$(GOGET) github.com/tyler-smith/go-bip39

# Cross compilation
build-linux:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) -o $(BINARY_UNIX) -v $(PKG)

.PHONY: all build test clean run deps build-linux
