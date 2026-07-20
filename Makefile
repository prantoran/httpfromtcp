# ==============================================================================
# Makefile — Build system for the Go HTTP server (native + WebAssembly)
# ==============================================================================
#
# This Makefile provides targets for both the native TCP server and the
# WebAssembly (WASM) browser build.
#
# Native targets:
#   make build        — compile the native binary
#   make run          — run the native TCP server
#   make test         — run all Go tests
#
# WebAssembly targets:
#   make wasm-all     — full WASM build (setup + compile)
#   make wasm-build   — compile Go to WASM binary (wasm/main.wasm)
#   make wasm-setup   — copy wasm_exec.js from Go installation
#   make wasm-serve   — start a local HTTP server to test in browser
#   make wasm-clean   — remove WASM build artifacts
#
# Prerequisites:
#   - Go 1.26+ installed
#   - Python 3 (for wasm-serve local server)
#
# ==============================================================================

# Go compiler settings
GO         := go
GOFLAGS    :=

# Project paths
CMD_DIR    := cmd/httpserver
WASM_DIR   := wasm
BINARY     := httpserver

# WASM-specific settings
# GOOS=js GOARCH=wasm tells the Go compiler to target the browser WASM environment.
# This is Go's built-in WASM support — no external toolchain (like Emscripten) is needed.
# Emscripten is for C/C++ → WASM; Go has its own native compilation target.
WASM_GOOS  := js
WASM_GOARCH := wasm
WASM_OUTPUT := $(WASM_DIR)/main.wasm

# wasm_exec.js is Go's official JavaScript support file that bootstraps the
# Go runtime in the browser. It must match the Go version used to compile.
# Location varies by Go version: misc/wasm/ (≤1.23) or lib/wasm/ (≥1.24).
WASM_EXEC_SRC := $(shell find "$$($(GO) env GOROOT)" -name "wasm_exec.js" -print -quit 2>/dev/null)
WASM_EXEC_DST := $(WASM_DIR)/wasm_exec.js

# Local development server port for testing the WASM build
SERVE_PORT := 8080

# ==============================================================================
# Default target
# ==============================================================================

.PHONY: help
help: ## Show this help message
	@echo "Usage: make [target]"
	@echo ""
	@echo "Native targets:"
	@echo "  build        Compile the native Go binary"
	@echo "  run          Run the native TCP server on port 42069"
	@echo "  test         Run all Go tests"
	@echo ""
	@echo "WebAssembly targets:"
	@echo "  wasm-all     Full WASM build (setup + compile)"
	@echo "  wasm-build   Compile Go to WASM ($(WASM_OUTPUT))"
	@echo "  wasm-setup   Copy wasm_exec.js from Go installation"
	@echo "  wasm-serve   Start local HTTP server for WASM testing"
	@echo "  wasm-clean   Remove WASM build artifacts"
	@echo ""
	@echo "Other:"
	@echo "  clean        Remove all build artifacts"
	@echo "  help         Show this help message"

# ==============================================================================
# Native targets
# ==============================================================================

.PHONY: build
build: ## Compile the native Go binary
	$(GO) build $(GOFLAGS) -o $(BINARY) ./$(CMD_DIR)

.PHONY: run
run: ## Run the native TCP server
	$(GO) run ./$(CMD_DIR)

.PHONY: test
test: ## Run all Go tests
	$(GO) test ./...

# ==============================================================================
# WebAssembly targets
# ==============================================================================

.PHONY: wasm-all
wasm-all: wasm-setup wasm-build ## Full WASM build (setup + compile)
	@echo "WASM build complete. Run 'make wasm-serve' to test in browser."

.PHONY: wasm-setup
wasm-setup: ## Copy wasm_exec.js from Go installation
	@echo "Copying wasm_exec.js from $(WASM_EXEC_SRC)..."
	@mkdir -p $(WASM_DIR)
	@cp "$(WASM_EXEC_SRC)" "$(WASM_EXEC_DST)"
	@echo "✓ wasm_exec.js copied to $(WASM_EXEC_DST)"

.PHONY: wasm-build
wasm-build: ## Compile Go to WASM binary
	@echo "Compiling Go to WebAssembly..."
	@echo "  GOOS=$(WASM_GOOS) GOARCH=$(WASM_GOARCH)"
	@echo "  Output: $(WASM_OUTPUT)"
	@mkdir -p $(WASM_DIR)
	GOOS=$(WASM_GOOS) GOARCH=$(WASM_GOARCH) $(GO) build $(GOFLAGS) -o $(WASM_OUTPUT) ./$(CMD_DIR)
	@echo "✓ WASM binary compiled: $(WASM_OUTPUT) ($$(du -h $(WASM_OUTPUT) | cut -f1))"

.PHONY: wasm-serve
wasm-serve: ## Start a local HTTP server for WASM testing
	@echo "Starting local HTTP server on http://localhost:$(SERVE_PORT)"
	@echo "Open this URL in your browser to test the WASM server."
	@echo "Press Ctrl+C to stop."
	@echo ""
	@cd $(WASM_DIR) && python3 -m http.server $(SERVE_PORT)

.PHONY: wasm-clean
wasm-clean: ## Remove WASM build artifacts
	@echo "Cleaning WASM build artifacts..."
	rm -f $(WASM_OUTPUT)
	rm -f $(WASM_EXEC_DST)
	@echo "✓ WASM artifacts cleaned"

# ==============================================================================
# Clean all build artifacts
# ==============================================================================

.PHONY: clean
clean: wasm-clean ## Remove all build artifacts
	rm -f $(BINARY)
	@echo "✓ All build artifacts cleaned"
