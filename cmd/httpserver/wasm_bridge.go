// Build constraint: this file is ONLY compiled for WebAssembly targets.
// It provides the bridge between JavaScript (Service Worker) and Go's
// custom HTTP/1.1 request parser.
//
// Architecture:
// =============
//
// The browser cannot open TCP sockets, so we simulate a TCP connection using
// Go's io.Pipe(). The bridge works as follows:
//
//   1. JavaScript calls handleHTTPRequest(method, path, headersJSON, body)
//   2. This function constructs raw HTTP/1.1 request bytes:
//        "GET /yourproblem HTTP/1.1\r\nHost: localhost:42069\r\n\r\n"
//   3. An io.Pipe() is created to simulate a TCP connection:
//        - pipeReader (read end)  → fed to the request parser as if it were a TCP socket
//        - pipeWriter (write end) → the raw HTTP request bytes are written here
//   4. A bytes.Buffer collects the response (simulating the write side of TCP)
//   5. The server's HandleRequest() function is called with a ReadWriteCloser
//      that reads from pipeReader and writes to the response buffer
//   6. Go's custom HTTP parser reads from pipeReader, parses the request,
//      calls handleRequest() (from handler.go), and writes the response
//   7. The raw HTTP/1.1 response bytes are returned to JavaScript as a string
//
// Why io.Pipe instead of bytes.Buffer for the request:
// ---------------------------------------------------
// The request parser (request.RequestFromReader) calls reader.Read() in a loop,
// expecting io.EOF when the request is fully read. A bytes.Buffer would work,
// but io.Pipe gives us explicit control over the write-then-close lifecycle,
// which more accurately simulates a TCP connection being closed after sending.
//
// Why we can't use the server.Serve() path:
// -----------------------------------------
// server.Serve() calls net.Listen("tcp", ":42069") which is not available in
// the browser. Instead, we directly call server.HandleRequest() which accepts
// an io.ReadWriteCloser, bypassing the TCP listener entirely.

//go:build js && wasm

package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"strings"
	"syscall/js"

	"github.com/prantoran/httpfromtcp/internal/server"
)

type readWriteCloser struct {
	io.Reader
	writeFunc func(p []byte) (n int, err error)
}

func (rwc *readWriteCloser) Write(p []byte) (n int, err error) {
	return rwc.writeFunc(p)
}

func (rwc *readWriteCloser) Close() error {
	return nil
}

func registerWasmHandlers() {
	handler := js.FuncOf(func(this js.Value, args []js.Value) any {
		if len(args) < 6 {
			return nil
		}

		method := args[0].String()
		path := args[1].String()
		headersStr := args[2].String()
		body := args[3].String()
		enqueueChunk := args[4]
		closeStream := args[5]

		promiseConstructor := js.Global().Get("Promise")
		return promiseConstructor.New(js.FuncOf(func(this js.Value, pArgs []js.Value) any {
			resolve := pArgs[0]
			reject := pArgs[1]

			go func() {
				err := processHTTPRequest(method, path, headersStr, body, resolve, enqueueChunk, closeStream)
				if err != nil {
					reject.Invoke(err.Error())
				}
			}()
			return nil
		}))
	})

	js.Global().Set("handleHTTPRequest", handler)
	log.Println("WASM: handleHTTPRequest registered on globalThis")
}

func processHTTPRequest(method, path, headersStr, body string, resolve, enqueueChunk, closeStream js.Value) error {
	var rawRequest strings.Builder
	rawRequest.WriteString(fmt.Sprintf("%s %s HTTP/1.1\r\n", method, path))
	rawRequest.WriteString(fmt.Sprintf("Host: localhost:%d\r\n", port))

	if headersStr != "" && headersStr != "{}" {
		for _, line := range strings.Split(headersStr, "\n") {
			line = strings.TrimSpace(line)
			if line != "" {
				rawRequest.WriteString(line + "\r\n")
			}
		}
	}

	if body != "" {
		rawRequest.WriteString(fmt.Sprintf("Content-Length: %d\r\n", len(body)))
	}
	rawRequest.WriteString("\r\n")

	if body != "" {
		rawRequest.WriteString(body)
	}

	pipeReader, pipeWriter := io.Pipe()

	go func() {
		defer pipeWriter.Close()
		pipeWriter.Write([]byte(rawRequest.String()))
	}()

	var headerBuf bytes.Buffer
	headersSent := false

	rwc := &readWriteCloser{
		Reader: pipeReader,
		writeFunc: func(p []byte) (n int, err error) {
			if !headersSent {
				headerBuf.Write(p)
				if bytes.Contains(headerBuf.Bytes(), []byte("\r\n\r\n")) {
					headersSent = true
					fullBytes := headerBuf.Bytes()
					idx := bytes.Index(fullBytes, []byte("\r\n\r\n"))
					
					headersPart := string(fullBytes[:idx+4])
					bodyPart := fullBytes[idx+4:]

					resolve.Invoke(headersPart)

					if len(bodyPart) > 0 {
						uint8Array := js.Global().Get("Uint8Array").New(len(bodyPart))
						js.CopyBytesToJS(uint8Array, bodyPart)
						enqueueChunk.Invoke(uint8Array)
					}
				}
			} else {
				if len(p) > 0 {
					uint8Array := js.Global().Get("Uint8Array").New(len(p))
					js.CopyBytesToJS(uint8Array, p)
					enqueueChunk.Invoke(uint8Array)
				}
			}
			return len(p), nil
		},
	}

	server.HandleRequest(handleRequest, rwc)

	if !headersSent {
		resolve.Invoke(headerBuf.String())
	}
	closeStream.Invoke()

	return nil
}
