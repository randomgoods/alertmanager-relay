PREFIX?=/usr
BINDIR?=$(PREFIX)/bin
SHAREDIR?=$(PREFIX)/share
BUILD_DIR := bin

BINARY_NAME := alertmanager-relay
VERSION := $(shell git describe --tags --always --dirty)
LDFLAGS := -ldflags "-X main.Version=$(VERSION)"

GO := go
GOBUILD := $(GO) build $(LDFLAGS)
GOINSTALL := $(GO) install $(LDFLAGS)
GOCLEAN := $(GO) clean
GOTEST := $(GO) test
GOGET := $(GO) get
# GOFMT := gofmt -s -w
GOFMT := $(GO) fmt
GOVET := $(GO) vet
GOMOD := $(GO) mod

PLATFORMS := linux/amd64 darwin/amd64 windows/amd64
PACKAGES := $(shell go list ./...)

.PHONY: default all build cross-compile clean deps fmt vet run test man install uninstall help

default: run

all: test build

build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	@$(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME) ./...

cross-compile:
	@echo "Cross-compiling..."
	@for platform in $(PLATFORMS); do \
		GOOS=$$(echo $$platform | cut -d'/' -f1); \
		GOARCH=$$(echo $$platform | cut -d'/' -f2); \
		output=$(BINARY_NAME)-$$GOOS-$$GOARCH; \
		if [ "$$GOOS" = "windows" ]; then \
		    output=$$output.exe; \
		fi; \
		echo "Building $$output"; \
		env GOOS=$$GOOS GOARCH=$$GOARCH $(GOBUILD) -o $(BUILD_DIR)/$$output $(CMD_DIR); \
	done

clean:
	@echo "Cleaning..."
	@$(GOCLEAN)
	@rm -rf $(BUILD_DIR)

deps:
	@echo "Tidying modules..."
	@$(GOMOD) tidy
	@$(GOMOD) download

fmt:
	@echo "Formatting..."
	@$(GOFMT) ./...

vet: fmt
	@echo "Vetting..."
	@$(GOVET) ./...

run:
	@$(GO) run main.go

test:
	@echo "Running tests..."
	@$(GOTEST) -v ./...

man:
	@scdoc < $(BINARY_NAME).1.scd | tail --lines=+8 > $(BINARY_NAME).1
	@gzip -k $(BINARY_NAME).1
	@install -Dm 644 $(BINARY_NAME).1.gz $(DESTDIR)$(SHAREDIR)/man/man1/$(BINARY_NAME).1.gz
	@rm $(BINARY_NAME).1 $(BINARY_NAME).1.gz
	@mandb --quiet

install: build man
	@install -Dm 755 $(BUILD_DIR)/$(BINARY_NAME) $(DESTDIR)$(BINDIR)/$(BINARY_NAME)

uninstall:
	@rm -f $(DESTDIR)$(BINDIR)/$(BINARY_NAME)
	@rm -f $(DESTDIR)$(SHAREDIR)/man/man1/$(BINARY_NAME).1.gz
	@mandb --quiet

help:
	@echo "Available targets:"
	@echo "  help          - Show this help"
	@echo "  all           - Run tests and build binary"
	@echo "  build         - Build binary for current platform"
	@echo "  cross-compile - Build binaries for multiple platforms"
	@echo "  clean         - Remove build artifacts"
	@echo "  deps          - Tidy and download modules"
	@echo "  fmt           - Format code"
	@echo "  vet           - Run go vet"
	@echo "  run           - Run program"
	@echo "  test          - Run tests"
	@echo "  man           - Build and install documentation"
	@echo "  install       - Build and install binary and documentation"
	@echo "  uninstall     - Remove binary and documentation"
