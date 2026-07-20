// Build constraint: this file is ONLY compiled for native (non-WASM) targets.
// When building with GOOS=js GOARCH=wasm, Go skips this file entirely and
// uses main_wasm.go instead, which provides a browser-compatible entry point.
//
// Why the build tag is needed:
// ----------------------------
// Go does not allow two main() functions in the same package. The native server
// uses net.Listen() for real TCP sockets and os/signal for graceful shutdown —
// neither of which work in the browser. The WASM entry point (main_wasm.go)
// replaces these with syscall/js bindings and a Service Worker bridge.
//
// How to run the native server:
//   go run cmd/httpserver/main.go
//   curl localhost:42069 > /tmp/server.txt
//   cat /tmp/server.txt
//
// ============================================================================
// CURRENT LIMITATIONS FOR WEBASSEMBLY COMPATIBILITY
// ============================================================================
//
// This native server uses several OS-level features that are NOT available in
// the browser WebAssembly sandbox:
//
// 1. TCP Sockets (net.Listen / net.Accept):
//    - Browsers cannot open raw TCP/UDP sockets for security reasons.
//    - The WASM build replaces this with an io.Pipe() bridge: the Service Worker
//      serializes incoming HTTP requests into raw bytes, writes them into one
//      end of the pipe, and the Go HTTP parser reads from the other end.
//    - Status: FULLY WORKED AROUND via the Service Worker bridge pattern.
//
// 2. OS Signals (os/signal, syscall.SIGINT/SIGTERM):
//    - The browser has no concept of POSIX signals.
//    - The WASM build uses select{} to keep the Go runtime alive indefinitely.
//      The server stops when the browser tab is closed.
//    - Status: FULLY WORKED AROUND.
//
// 3. File System Access (os.ReadFile for /video route):
//    - The browser WASM sandbox has no local filesystem.
//    - The /video route returns a 500 error in WASM mode.
//    - To fix: embed the video using Go's //go:embed directive, or serve it
//      as a separate static asset from the web server.
//    - Status: NOT AVAILABLE in WASM. Route returns 500.
//
// 4. Outbound HTTP Requests (http.Get for /httpbin/* proxy route):
//    - In WASM (GOOS=js), Go's net/http client uses the browser's fetch() API
//      under the hood. Basic GET requests work, but the chunked streaming
//      pattern used here (reading 32-byte chunks in a loop) may not behave
//      identically to native TCP streaming.
//    - CORS restrictions apply — httpbin.org must allow cross-origin requests.
//    - Status: PARTIALLY AVAILABLE. May work for simple requests but streaming
//      behavior differs from native.
//
// 5. Concurrent Connections (goroutines per connection):
//    - WASM in the browser is single-threaded. Go's goroutine scheduler still
//      works (cooperative multitasking), but true parallelism is not available.
//    - The Service Worker processes one request at a time.
//    - Status: FUNCTIONALLY EQUIVALENT for single-user browser use.
//
// 6. Binary Size:
//    - The WASM binary includes the full Go runtime (~2-5 MB compressed).
//    - Consider using TinyGo for smaller binaries, but TinyGo has limited
//      standard library support (no net/http, limited reflect, etc.).
//    - Status: ACCEPTABLE for demo/portfolio use.
//
// WHAT NEEDS TO BE DONE TO IMPROVE WEBASSEMBLY COMPATIBILITY:
//
//   a) Embed static assets (video.mp4) using //go:embed for the /video route.
//   b) Add WASM-specific error handling for routes that require OS features.
//   c) Consider using compression (gzip/brotli) for the WASM binary to reduce
//      download size on the portfolio site.
//   d) Add WebSocket support for real-time bidirectional communication if
//      needed (Service Workers cannot intercept WebSocket connections).
//   e) Investigate WASI (WebAssembly System Interface) as a future target
//      that could provide real socket access outside the browser.
//   f) Add SharedArrayBuffer + Web Workers for true concurrent request handling.
//
// ============================================================================

//go:build !(js && wasm)

package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/prantoran/httpfromtcp/internal/server"
)

func main() {
	s, err := server.Serve(port, handleRequest)
	if err != nil {
		log.Fatalf("Error starting server: %v\n", err)
	}
	defer s.Close()
	log.Printf("Server listening on port %d\n", port)

	// Block until SIGINT (Ctrl+C) or SIGTERM is received.
	// In WASM mode, this is replaced by select{} since there are no OS signals.
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
}
