BINARY_NAME := anvil
BUILD_DIR := .
GO_CMD := go

.PHONY: build install clean test vet

build:
	$(GO_CMD) build -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/anvil/

install:
	$(GO_CMD) install ./cmd/anvil/

clean:
	rm -f $(BUILD_DIR)/anvil

test:
	$(GO_CMD) test -race ./...

vet:
	$(GO_CMD) vet ./...
