# WebAssembly Build & Testing Guide

This guide covers how to build, test, and deploy the Go HTTP server as a WebAssembly binary that runs entirely in the browser.

## Table of Contents

- [Prerequisites](#prerequisites)
- [Quick Start](#quick-start)
- [Build Instructions](#build-instructions)
  - [Using Makefile](#using-makefile)
  - [Using CMake](#using-cmake)
- [Local Testing](#local-testing)
- [Using the Terminal UI](#using-the-terminal-ui)
- [Sample Curl Commands](#sample-curl-commands)
- [Architecture](#architecture)
- [Deploying to GitHub Pages](#deploying-to-github-pages)
- [Limitations](#limitations)
- [Troubleshooting](#troubleshooting)

---

## Prerequisites

- **Go 1.26+** — the Go compiler includes native WebAssembly support
- **Python 3** — for the local development HTTP server (or any static file server)
- **A modern browser** — Chrome, Firefox, Safari, or Edge with Service Worker support

> **Note:** Emscripten is NOT required. Go has its own built-in WebAssembly
> compilation target (`GOOS=js GOARCH=wasm`). Emscripten is a C/C++ → WASM
> toolchain and is not compatible with Go source code.

---

## Quick Start

```bash
# Build the WASM binary and copy supporting files
make wasm-all

# Start a local server to test in the browser
make wasm-serve

# Open http://localhost:8080 in your browser
```

---

## Build Instructions

### Using Makefile

The Makefile is the primary build system. All WASM targets are prefixed with `wasm-`.

```bash
# Copy wasm_exec.js from your Go installation
# This file bootstraps the Go runtime in the browser
make wasm-setup

# Compile the Go server to WebAssembly
# Output: wasm/main.wasm
make wasm-build

# Or do both in one command
make wasm-all

# Start a local HTTP server for testing
make wasm-serve

# Clean WASM build artifacts
make wasm-clean
```

**What `make wasm-build` does under the hood:**
```bash
GOOS=js GOARCH=wasm go build -o wasm/main.wasm ./cmd/httpserver
```

- `GOOS=js` — tells Go the target OS is JavaScript (browser)
- `GOARCH=wasm` — tells Go the target architecture is WebAssembly
- Go's build tags (`//go:build js && wasm`) select `main_wasm.go` and `wasm_bridge.go` instead of the native `main.go`

### Using CMake

CMake is provided as an alternative build system for teams that prefer it.

```bash
# Create and enter a build directory
mkdir build && cd build

# Configure the project
cmake ..

# Build the WASM binary
cmake --build . --target wasm-all

# Start a local test server
cmake --build . --target wasm-serve

# Clean artifacts
cmake --build . --target wasm-clean
```

---

## Local Testing

After building with `make wasm-all`, start the local server:

```bash
make wasm-serve
```

This starts a Python HTTP server serving the `wasm/` directory on port 8080.

1. Open **http://localhost:8080** in your browser
2. Wait for the boot sequence to complete (you'll see green checkmarks ✓)
3. Type curl commands in the terminal prompt

> **Important:** Service Workers require either `localhost` or HTTPS. If you use
> a different hostname (like an IP address), the Service Worker won't register.

---

## Using the Terminal UI

The browser page provides a terminal-emulator interface that mimics a real
command-line terminal. Here's how to use it:

### Boot Sequence

When you open the page, the terminal automatically:
1. Loads `wasm_exec.js` (Go's JavaScript support file)
2. Fetches and instantiates `main.wasm` (the compiled Go server)
3. Starts the Go runtime (registers the request handler)
4. Registers the Service Worker

Each step shows a status indicator:
- `$` — a step is running
- `✓` — step completed successfully
- `✗` — step failed (see error message)

### Typing Commands

Once the boot completes, you'll see:
```
Server ready. Type a curl command below, or type "help" for usage.
```

Type curl commands at the `$` prompt and press Enter. The response appears
immediately below, showing the full raw HTTP/1.1 response (status line,
headers, and body).

### Special Commands

| Command | Description |
|---------|-------------|
| `help` | Show available routes and sample commands |
| `clear` | Clear the terminal output |
| ↑/↓ arrows | Navigate command history |

---

## Sample Curl Commands

The terminal supports curl-style commands. The URL `localhost:42069` is a
simulated address — no real TCP port is opened. The terminal recognizes it
and routes the request to the Go WASM handler internally.

### Basic Requests

```bash
# GET the root path — returns 200 OK with a success HTML page
curl localhost:42069/

# GET a "bad request" page — returns 400 Bad Request
curl localhost:42069/yourproblem

# GET a "server error" page — returns 500 Internal Server Error
curl localhost:42069/myproblem
```

### With Flags

```bash
# Explicit GET method
curl -X GET localhost:42069/

# Show response headers (enabled by default in the terminal)
curl -i localhost:42069/yourproblem

# Verbose output — shows the outgoing request details
curl -v localhost:42069/

# Full URL with scheme (also works)
curl http://localhost:42069/myproblem

# Shorthand — just the path (also works)
curl /yourproblem
```

### POST Request

```bash
# POST with a body (method is auto-detected from -d flag)
curl -d "hello world" localhost:42069/

# POST with explicit method and custom header
curl -X POST -H "Content-Type: text/plain" -d "hello" localhost:42069/
```

### Example Session

```
$ curl localhost:42069/
HTTP/1.1 200 OK
content-type: text/html
connection: close
content-length: 130

<html>
  <head>
    <title>200 OK</title>
  </head>
  <body>
    <h1>Success!</h1>
    <p>Your request was a masterpiece.</p>
  </body>
</html>

$ curl localhost:42069/yourproblem
HTTP/1.1 400 Bad Request
content-type: text/html
connection: close
content-length: 153

<html>
  <head>
    <title>400 Bad Request</title>
  </head>
  <body>
    <h1>Bad Request</h1>
    <p>Your request honestly kinda sucked.</p>
  </body>
</html>
```

### What Does `localhost:42069` Mean?

The URL `localhost:42069` in the browser terminal is a **simulated address** —
it does not open a real TCP port on your machine. The terminal UI recognizes
`localhost:42069` (or `http://localhost:42069`) as the address of the Go WASM
server and routes the request internally through the Service Worker to the Go
handler. The port 42069 matches the original native server's port to maintain
familiarity.

---

## Architecture

The WASM server uses a **Service Worker bridge pattern** to simulate a TCP
server in the browser:

```
Browser Tab
├── index.html (Terminal UI)
│   ├── Parses curl commands
│   ├── Calls handleHTTPRequest() (Go WASM)
│   └── Displays raw HTTP/1.1 responses
│
├── Service Worker (sw.js)
│   ├── Intercepts fetch() requests to /api/*
│   ├── Routes them to Go WASM handler
│   └── Returns Response objects
│
└── Go WASM Runtime (main.wasm)
    ├── main_wasm.go — entry point (select{} to stay alive)
    ├── wasm_bridge.go — syscall/js bridge
    │   ├── Constructs raw HTTP/1.1 request bytes
    │   ├── Creates io.Pipe() (simulated TCP connection)
    │   ├── Feeds request through Go's HTTP parser
    │   └── Collects response bytes
    └── handler.go — shared request handler
```

### Key Files

| File | Purpose |
|------|---------|
| `cmd/httpserver/handler.go` | Shared HTTP handler (no build tags) |
| `cmd/httpserver/main.go` | Native entry point (`//go:build !(js && wasm)`) |
| `cmd/httpserver/main_wasm.go` | WASM entry point (`//go:build js && wasm`) |
| `cmd/httpserver/wasm_bridge.go` | syscall/js bridge (`//go:build js && wasm`) |
| `wasm/index.html` | Terminal emulator UI |
| `wasm/sw.js` | Service Worker fetch interceptor |
| `wasm/register-sw.js` | Service Worker registration helper |
| `wasm/wasm_exec.js` | Go's JS support file (copied from Go install) |
| `wasm/main.wasm` | Compiled WASM binary (build output) |

---

## Deploying to GitHub Pages

To deploy to `https://prantoran.github.io`:

1. Build the WASM binary:
   ```bash
   make wasm-all
   ```

2. Copy these files to your GitHub Pages repository:
   - `wasm/index.html`
   - `wasm/main.wasm`
   - `wasm/wasm_exec.js`
   - `wasm/sw.js`
   - `wasm/register-sw.js`

3. Push to GitHub. The site will serve the WASM binary automatically.

> **MIME type:** GitHub Pages correctly serves `.wasm` files with
> `application/wasm` content type, so `WebAssembly.instantiateStreaming`
> will work without issues.

---

## Limitations

### 1. No Terminal Curl Access

A WASM server running in the browser **cannot** be reached by terminal `curl`.
The server exists only within the browser's sandbox — no real TCP port is
opened. Only requests made from within the same browser tab (via the terminal
UI or fetch()) can reach the Go handler.

### 2. No Real TCP/UDP Sockets

Browsers do not allow WebAssembly modules to open raw TCP or UDP sockets.
The `net.Listen()` call in the native server is replaced by an `io.Pipe()`
bridge in WASM mode.

### 3. Routes Not Available in WASM

| Route | Reason | Status |
|-------|--------|--------|
| `/video` | Requires `os.ReadFile()` (no filesystem in browser) | Returns 500 |
| `/httpbin/*` | Requires outbound HTTP + streaming (CORS issues) | May fail |

### 4. Binary Size

The WASM binary includes the full Go runtime and is typically 2-5 MB. This
is acceptable for a demo but may be large for production use. Consider:
- Gzip/Brotli compression (reduces by ~70%)
- TinyGo compiler (smaller binaries but limited stdlib)

### 5. Single-Threaded Execution

WASM in the browser runs on a single thread. Go's goroutine scheduler still
works (cooperative multitasking), but there's no true parallelism. The server
processes one request at a time.

### 6. Service Worker Scope

The Service Worker only intercepts requests within its registered scope. If
deployed to a subdirectory (e.g., `prantoran.github.io/httpserver/`), the
scope must be configured accordingly.

---

## Troubleshooting

### "Go class not found" during boot

The `wasm_exec.js` file is missing or not loaded. Run:
```bash
make wasm-setup
```

### "handleHTTPRequest not registered"

The WASM binary failed to initialize. Check the browser console for errors.
Rebuild with:
```bash
make wasm-all
```

### Service Worker not registering

- Ensure you're accessing via `localhost` or HTTPS (not an IP address)
- Check DevTools → Application → Service Workers for error messages
- Try an incognito/private window to avoid cached Service Workers

### Blank or missing response

- Open DevTools → Console for JavaScript errors
- Verify `main.wasm` is being served with `application/wasm` MIME type
- Check that the Go handler doesn't panic (console will show stack trace)
