NAME := k8s-to-secretsmanager
PKG := github.com/ndigati/k8s-to-secretsmanager
GO ?= go
CGO_ENABLED ?= 1

COMMIT ?= $(shell git describe --dirty --long --always)

.PHONY: build
build:
	CGO_ENABLED=$(CGO_ENABLED) $(GO) build -trimpath -ldflags "-X main.gitCommit=$(COMMIT)" -o build/$(NAME)

.PHONY: build-linux
	GOOS=linux GOARCH=amd64 CGO_ENABLED=$(CGO_ENABLED) $(GO) build -trimpath -ldflags "-X main.gitCommit=$(COMMIT)" -o build/$(NAME)-linux-amd64

.PHONY: test
test:
	$(GO) test -v ./...
