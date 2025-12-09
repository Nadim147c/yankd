GO       ?= go
REVIVE   ?= revive
NAME     ?= yankd
VERSION  ?= $(shell git describe --tags 2>/dev/null || echo "dev-build")
PREFIX   ?= /usr/local/

BUILD_DIR ?= build
BUILD_BIN ?= $(BUILD_DIR)/$(NAME)
BUILD_COMPLETION_DIR ?= $(BUILD_DIR)/completions

INSTALL_BIN                 = $(shell realpath -m "$(PREFIX)/bin/$(NAME)")
INSTALL_LICENSE             = $(shell realpath -m "$(PREFIX)/share/licenses/$(NAME)/LICENSE")
INSTALL_BASH_COMPLETION_DIR = $(shell realpath -m "$(PREFIX)/share/bash-completion/completions")
INSTALL_ZSH_COMPLETION_DIR  = $(shell realpath -m "$(PREFIX)/share/zsh/site-functions")
INSTALL_FISH_COMPLETION_DIR = $(shell realpath -m "$(PREFIX)/share/fish/vendor_completions.d")

-include Makefile.local

.PHONY: build install test

build:
	$(GO) build -trimpath -ldflags '-s -w -X main.version=$(VERSION)' -o $(BUILD_BIN)
	mkdir -p "$(BUILD_COMPLETION_DIR)"
	$(BUILD_BIN) _carapace bash > "$(BUILD_COMPLETION_DIR)/$(NAME).bash"
	$(BUILD_BIN) _carapace zsh  > "$(BUILD_COMPLETION_DIR)/$(NAME).zsh"
	$(BUILD_BIN) _carapace fish > "$(BUILD_COMPLETION_DIR)/$(NAME).fish"

install:
	install -Dm755 $(BUILD_BIN) "$(INSTALL_BIN)"
	install -Dm644 LICENSE "$(INSTALL_LICENSE)"
	install -Dm644 "$(BUILD_COMPLETION_DIR)/$(NAME).bash" "$(INSTALL_BASH_COMPLETION_DIR)/$(NAME)"
	install -Dm644 "$(BUILD_COMPLETION_DIR)/$(NAME).zsh"  "$(INSTALL_ZSH_COMPLETION_DIR)/_$(NAME)"
	install -Dm644 "$(BUILD_COMPLETION_DIR)/$(NAME).fish" "$(INSTALL_FISH_COMPLETION_DIR)/$(NAME).fish"

test:
	$(GO) test -v ./...
	$(REVIVE) -config revive.toml -formatter friendly ./...
