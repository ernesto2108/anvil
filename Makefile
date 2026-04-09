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

.PHONY: dashboard-dev dashboard-build dashboard-frontend dashboard-install-cli

# Resolve wails CLI location — installed via `go install` lives in GOBIN or ~/go/bin.
# Used only by the dev target; the build target uses plain `go build`.
GOBIN ?= $(shell go env GOBIN)
ifeq ($(GOBIN),)
	GOBIN := $(shell go env GOPATH)/bin
endif
WAILS := $(GOBIN)/wails

DASHBOARD_BINARY := anvil-full

dashboard-install-cli: ## Install the Wails v2 CLI into $GOBIN (only needed for dashboard-dev)
	$(GO_CMD) install github.com/wailsapp/wails/v2/cmd/wails@latest

dashboard-frontend: ## Build the React frontend into frontend/dist
	cd frontend && npm install --ignore-scripts && npm run build

dashboard-build: dashboard-frontend ## Build anvil-full with embedded dashboard (requires CGO, macOS only for now)
	CGO_ENABLED=1 \
	CGO_LDFLAGS="-framework UniformTypeIdentifiers" \
	$(GO_CMD) build -tags "dashboard production" -o $(DASHBOARD_BINARY) ./cmd/anvil/
	@echo ""
	@echo "Built $(DASHBOARD_BINARY). Run it with: ./$(DASHBOARD_BINARY) dashboard"

dashboard-dev: ## Run Vite dev server for the frontend (HMR). Native window requires manual build.
	@echo "Starting Vite dev server on http://localhost:5173"
	@echo "For the native window, run 'make dashboard-build' and then './$(DASHBOARD_BINARY) dashboard'"
	cd frontend && npm run dev
