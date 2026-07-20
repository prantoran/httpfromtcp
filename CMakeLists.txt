# ==============================================================================
# CMakeLists.txt — Alternative build system for Go HTTP Server WebAssembly
# ==============================================================================
#
# This CMake file provides an alternative way to build the WebAssembly binary.
# It wraps the same Go compiler commands that the Makefile uses, providing
# integration with CMake-based workflows and IDEs.
#
# Why CMake for a Go project?
# ---------------------------
# While Go projects typically use Makefiles or go build directly, CMake is
# included here as requested to support teams that use CMake as their primary
# build system. CMake calls the Go compiler via custom commands — it does NOT
# use Emscripten's CMake toolchain file, because Go has its own WASM target.
#
# Usage:
#   mkdir build && cd build
#   cmake ..
#   cmake --build . --target wasm          # Build WASM binary
#   cmake --build . --target wasm-setup    # Copy wasm_exec.js
#   cmake --build . --target wasm-all      # Setup + build
#   cmake --build . --target wasm-serve    # Start local test server
#   cmake --build . --target wasm-clean    # Clean WASM artifacts
#
# ==============================================================================

cmake_minimum_required(VERSION 3.16)

project(httpfromtcp-wasm
  VERSION 1.0.0
  DESCRIPTION "Go HTTP Server compiled to WebAssembly"
  LANGUAGES NONE  # No C/C++ — this project uses Go
)

# ==============================================================================
# Configuration
# ==============================================================================

# Find the Go compiler
find_program(GO_EXECUTABLE go REQUIRED)

# Get GOROOT for locating wasm_exec.js
execute_process(
  COMMAND ${GO_EXECUTABLE} env GOROOT
  OUTPUT_VARIABLE GOROOT
  OUTPUT_STRIP_TRAILING_WHITESPACE
)

# Paths
set(CMD_DIR "${CMAKE_SOURCE_DIR}/cmd/httpserver")
set(WASM_DIR "${CMAKE_SOURCE_DIR}/wasm")
set(WASM_OUTPUT "${WASM_DIR}/main.wasm")
set(WASM_EXEC_DST "${WASM_DIR}/wasm_exec.js")
set(SERVE_PORT 8080)

# Find wasm_exec.js — location varies by Go version:
#   misc/wasm/wasm_exec.js (≤1.23) or lib/wasm/wasm_exec.js (≥1.24)
execute_process(
  COMMAND find "${GOROOT}" -name "wasm_exec.js" -print -quit
  OUTPUT_VARIABLE WASM_EXEC_SRC
  OUTPUT_STRIP_TRAILING_WHITESPACE
)

# ==============================================================================
# WASM Build Target
# ==============================================================================

# cmake --build . --target wasm
# Compiles the Go HTTP server to a WebAssembly binary using Go's native
# WASM compilation target (GOOS=js GOARCH=wasm). No Emscripten needed.
add_custom_target(wasm
  COMMAND ${CMAKE_COMMAND} -E make_directory "${WASM_DIR}"
  COMMAND ${CMAKE_COMMAND} -E env
    "GOOS=js"
    "GOARCH=wasm"
    ${GO_EXECUTABLE} build -o "${WASM_OUTPUT}" "./cmd/httpserver"
  WORKING_DIRECTORY "${CMAKE_SOURCE_DIR}"
  COMMENT "Compiling Go to WebAssembly: ${WASM_OUTPUT}"
  VERBATIM
)

# ==============================================================================
# WASM Setup Target
# ==============================================================================

# cmake --build . --target wasm-setup
# Copies Go's official wasm_exec.js support file to the WASM output directory.
# This file must match the Go version used to compile the WASM binary.
add_custom_target(wasm-setup
  COMMAND ${CMAKE_COMMAND} -E make_directory "${WASM_DIR}"
  COMMAND ${CMAKE_COMMAND} -E copy "${WASM_EXEC_SRC}" "${WASM_EXEC_DST}"
  COMMENT "Copying wasm_exec.js from ${WASM_EXEC_SRC}"
  VERBATIM
)

# ==============================================================================
# Combined Target
# ==============================================================================

# cmake --build . --target wasm-all
# Runs setup (copy wasm_exec.js) then builds the WASM binary.
add_custom_target(wasm-all
  DEPENDS wasm-setup wasm
  COMMENT "Full WASM build complete"
)

# ==============================================================================
# Local Test Server
# ==============================================================================

# cmake --build . --target wasm-serve
# Starts a local HTTP server on port 8080 for testing the WASM build in a browser.
# Uses Python 3's built-in HTTP server module.
add_custom_target(wasm-serve
  COMMAND python3 -m http.server ${SERVE_PORT}
  WORKING_DIRECTORY "${WASM_DIR}"
  COMMENT "Starting local HTTP server on http://localhost:${SERVE_PORT}"
  USES_TERMINAL
)

# ==============================================================================
# Clean Target
# ==============================================================================

# cmake --build . --target wasm-clean
# Removes the compiled WASM binary and copied wasm_exec.js.
add_custom_target(wasm-clean
  COMMAND ${CMAKE_COMMAND} -E rm -f "${WASM_OUTPUT}"
  COMMAND ${CMAKE_COMMAND} -E rm -f "${WASM_EXEC_DST}"
  COMMENT "Cleaning WASM build artifacts"
)

# ==============================================================================
# Status Message
# ==============================================================================

message(STATUS "Go HTTP Server WASM Build System")
message(STATUS "  Go compiler: ${GO_EXECUTABLE}")
message(STATUS "  GOROOT: ${GOROOT}")
message(STATUS "  WASM output: ${WASM_OUTPUT}")
message(STATUS "  wasm_exec.js: ${WASM_EXEC_SRC}")
message(STATUS "")
message(STATUS "Available targets:")
message(STATUS "  wasm         - Compile Go to WebAssembly")
message(STATUS "  wasm-setup   - Copy wasm_exec.js")
message(STATUS "  wasm-all     - Setup + build")
message(STATUS "  wasm-serve   - Start local test server")
message(STATUS "  wasm-clean   - Clean WASM artifacts")
