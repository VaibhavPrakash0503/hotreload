# Hotreload Makefile

BINARY  := hotreload
CMD_PKG := ./cmd/hotreload
BIN_DIR := bin
GO      := go

.PHONY: all build linux mac windows test test-race cover run fmt vet tidy clean install help

all: build

build:
	@mkdir -p $(BIN_DIR)
	$(GO) build -o $(BIN_DIR)/$(BINARY) $(CMD_PKG)

linux:
	@mkdir -p $(BIN_DIR)
	GOOS=linux GOARCH=amd64 $(GO) build -o $(BIN_DIR)/$(BINARY)-linux $(CMD_PKG)

mac:
	@mkdir -p $(BIN_DIR)
	GOOS=darwin GOARCH=arm64 $(GO) build -o $(BIN_DIR)/$(BINARY)-mac $(CMD_PKG)


# Test
test:
	$(GO) test ./internal/... -v -count=1

test-race:
	$(GO) test -race ./internal/... -v -count=1

cover:
	$(GO) test ./internal/... -coverprofile=coverage.out -count=1
	$(GO) tool cover -html=coverage.out


run: build
	$(BIN_DIR)/$(BINARY) --root ./testserver \
		--build "$(GO) build -o /tmp/testserver-dev ./testserver" \
		--exec "/tmp/testserver-dev" --debounce 300 --verbose

install:
	$(GO) install $(CMD_PKG)

clean:
	rm -rf $(BIN_DIR) coverage.out

help:
	@grep -E '^[a-zA-Z_-]+:' $(MAKEFILE_LIST) \
		| grep -v '^\.PHONY' \
		| awk -F: '{printf "\033[36m%-12s\033[0m\n", $$1}'
