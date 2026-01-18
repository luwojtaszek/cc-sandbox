.PHONY: all images cli clean help base docker bun-full install

VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "none")
DATE ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
REGISTRY ?=
IMAGE_PREFIX ?= cc-sandbox

LDFLAGS := -ldflags "-s -w -X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.date=$(DATE)"

all: images cli

# Docker Images
images: base docker bun-full
	@echo "✅ All images built successfully"

base:
	@echo "🔨 Building $(IMAGE_PREFIX):base..."
	docker build -t $(IMAGE_PREFIX):base ./docker/base

docker: base
	@echo "🔨 Building $(IMAGE_PREFIX):docker..."
	docker build -t $(IMAGE_PREFIX):docker ./docker/docker

bun-full: docker
	@echo "🔨 Building $(IMAGE_PREFIX):bun-full..."
	docker build -t $(IMAGE_PREFIX):bun-full ./docker/bun-full

# CLI Tool
cli: cli/cc-sandbox
	@echo "✅ CLI built successfully"

cli/cc-sandbox: cli/main.go cli/go.mod
	@echo "🔨 Building CLI..."
	cd cli && go build $(LDFLAGS) -o cc-sandbox .

cli-all: cli-linux cli-darwin cli-windows
	@echo "✅ All CLI binaries built"

cli-linux:
	cd cli && GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o cc-sandbox-linux-amd64 .
	cd cli && GOOS=linux GOARCH=arm64 go build $(LDFLAGS) -o cc-sandbox-linux-arm64 .

cli-darwin:
	cd cli && GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o cc-sandbox-darwin-amd64 .
	cd cli && GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o cc-sandbox-darwin-arm64 .

cli-windows:
	cd cli && GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o cc-sandbox-windows-amd64.exe .

# Installation
install: cli
	sudo cp cli/cc-sandbox /usr/local/bin/cc-sandbox
	@echo "✅ Installed to /usr/local/bin/cc-sandbox"

install-user: cli
	mkdir -p $(HOME)/.local/bin
	cp cli/cc-sandbox $(HOME)/.local/bin/cc-sandbox
	@echo "✅ Installed to ~/.local/bin/cc-sandbox"

# Registry Push
push: images
ifndef REGISTRY
	$(error REGISTRY is not set. Use: make push REGISTRY=ghcr.io/luwojtaszek)
endif
	docker tag $(IMAGE_PREFIX):base $(REGISTRY)/$(IMAGE_PREFIX):base
	docker tag $(IMAGE_PREFIX):docker $(REGISTRY)/$(IMAGE_PREFIX):docker
	docker tag $(IMAGE_PREFIX):bun-full $(REGISTRY)/$(IMAGE_PREFIX):bun-full
	docker push $(REGISTRY)/$(IMAGE_PREFIX):base
	docker push $(REGISTRY)/$(IMAGE_PREFIX):docker
	docker push $(REGISTRY)/$(IMAGE_PREFIX):bun-full

# Development
dev: base
	docker run -it --rm -v $(PWD):/workspace -w /workspace --user $(shell id -u):$(shell id -g) $(IMAGE_PREFIX):base bash

# Cleanup
clean:
	rm -f cli/cc-sandbox cli/cc-sandbox-*

clean-images:
	-docker rmi $(IMAGE_PREFIX):base $(IMAGE_PREFIX):docker $(IMAGE_PREFIX):bun-full 2>/dev/null

clean-all: clean clean-images

help:
	@echo "cc-sandbox - Claude Code Docker Sandbox"
	@echo ""
	@echo "Docker Images:"
	@echo "  images      Build all Docker images"
	@echo "  base        Build base image only"
	@echo "  docker      Build docker image"
	@echo "  bun-full    Build bun-full image"
	@echo ""
	@echo "CLI Tool:"
	@echo "  cli         Build CLI for current platform"
	@echo "  cli-all     Build CLI for all platforms"
	@echo "  install     Install CLI to /usr/local/bin"
	@echo "  install-user Install CLI to ~/.local/bin"
	@echo ""
	@echo "Registry:"
	@echo "  push        Push images (requires REGISTRY=...)"
