.PHONY: help setup install dev build build-dmg run clean lint

# Extract version from wails.json
VERSION := $(shell jq -r .info.productversion wails.json)

help:
	@echo "YTDown - YouTube Downloader for macOS"
	@echo "======================================"
	@echo ""
	@echo "Commands:"
	@echo "  make setup      - Install dependencies"
	@echo "  make install    - Download Go modules"
	@echo "  make dev        - Start development server (hot reload)"
	@echo "  make build      - Build production app"
	@echo "  make build-dmg  - Build DMG distribution"
	@echo "  make run        - Run the app from build/"
	@echo "  make clean      - Remove build artifacts"
	@echo ""

setup:
	@bash setup.sh

install:
	go mod download
	go mod tidy

dev:
	wails dev

build:
	@bash build.sh

build-dmg: build
	@echo "📦 DMG distribution is already created by build.sh"
	@echo "   Check: dist/YTDown-$(VERSION).dmg"

run: build
	@open dist/YTDown.app

clean:
	rm -rf build/bin
	rm -rf dist
	rm -rf frontend/dist
	rm -rf frontend/node_modules
	go clean

lint:
	gofmt -w .
	go vet ./...

fmt:
	gofmt -w .
