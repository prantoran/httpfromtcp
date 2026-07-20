// Build constraint: this file is ONLY compiled for WebAssembly targets.
// It provides the browser-compatible entry point for the Go HTTP server.
//
// When compiled with: GOOS=js GOARCH=wasm go build -o main.wasm ./cmd/httpserver
// this file's main() is used instead of main.go's main().
//
// How it works:
// -------------
// 1. The compiled main.wasm is loaded by index.html in the browser.
// 2. wasm_exec.js (Go's official JS support file) bootstraps the Go runtime.
// 3. This main() registers a JavaScript-callable function "handleHTTPRequest"
//    via syscall/js, which the Service Worker calls when it intercepts fetch().
// 4. select{} keeps the Go runtime alive indefinitely — without it, the WASM
//    module would exit and the registered JS functions would become invalid.
//
// The request flow in the browser:
//
//   Browser UI → fetch("/yourproblem") → Service Worker (sw.js)
//     → sw.js calls handleHTTPRequest("GET", "/yourproblem", ...)
//     → Go WASM constructs raw HTTP/1.1 bytes, pipes through io.Pipe
//     → Go's request parser + handler processes the request
//     → Raw HTTP/1.1 response bytes are returned to sw.js
//     → sw.js parses response and returns a Response object to fetch()
//     → Browser UI displays the response in the terminal

//go:build js && wasm

package main

import (
	"fmt"
	"log"
)

func main() {
	// Register the HTTP request handler as a global JavaScript function.
	// After this call, JavaScript code can invoke:
	//   globalThis.handleHTTPRequest(method, path, headersJSON, body)
	// which returns the raw HTTP/1.1 response as a string.
	registerWasmHandlers()

	log.Printf("Go HTTP server (WASM) initialized — listening on simulated port %d\n", port)
	fmt.Println("Server ready. Waiting for requests via Service Worker...")

	// Block forever to keep the Go WASM runtime alive.
	// Without this, the Go program would exit immediately after main() returns,
	// and all registered JS functions (handleHTTPRequest) would be garbage
	// collected and become uncallable.
	//
	// In native mode, this role is filled by signal.Notify() + <-sigChan.
	// In WASM mode, there are no OS signals, so we use an empty select{}.
	// The Go runtime will keep running until the browser tab is closed.
	select {}
}
