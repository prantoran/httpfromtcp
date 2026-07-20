// Package main — shared handler logic for the HTTP server.
//
// This file contains the HTTP handler callback and response helper functions
// used by both the native TCP server (main.go) and the WebAssembly browser
// server (main_wasm.go). It has NO build tags, so it compiles on all platforms.
//
// Why this file exists:
// --------------------
// When targeting WebAssembly (GOOS=js GOARCH=wasm), the server cannot use
// real TCP sockets — browsers don't allow raw socket access. Instead, the
// WASM build uses a Service Worker + io.Pipe bridge to simulate connections.
// Both the native and WASM entry points need the same request handling logic,
// so it is extracted here to avoid duplication.
//
// The handler processes these routes:
//   - "/"             → 200 OK with a success HTML page
//   - "/yourproblem"  → 400 Bad Request with an error HTML page
//   - "/myproblem"    → 500 Internal Server Error with an error HTML page
//   - "/video"        → Serves a video file (native only, not available in WASM)
//   - "/httpbin/*"    → Proxies to httpbin.org (native only, not available in WASM)
//
// The /video and /httpbin/* routes require filesystem and outbound network
// access respectively, which are not available in the browser WASM sandbox.
// The WASM entry point (main_wasm.go) uses the same handler but those routes
// will return 500 errors since os.ReadFile and http.Get are not functional
// in the browser environment.

package main

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"strings"

	"github.com/prantoran/httpfromtcp/internal/headers"
	"github.com/prantoran/httpfromtcp/internal/request"
	"github.com/prantoran/httpfromtcp/internal/response"
)

// toStr converts a byte slice to a hexadecimal string representation.
// Used for computing content SHA256 checksums in the /httpbin/* proxy route.
func toStr(bytes []byte) string {
	out := ""
	for _, b := range bytes {
		out += fmt.Sprintf("%02x", b)
	}
	return out
}

// port is the TCP port the native server listens on, and the simulated port
// displayed in the WASM terminal UI (e.g., "curl localhost:42069/").
// In WASM mode, no real port is opened — it's used only for display purposes.
const port = 42069

// respond400 returns the HTML body for a 400 Bad Request response.
func respond400() []byte {
	return []byte(`<html>
	<head>
		<title>400 Bad Request</title>
	</head>
	<body>
		<h1>Bad Request</h1>
		<p>Your request honestly kinda sucked.</p>
	</body>
</html>`)
}

// respond500 returns the HTML body for a 500 Internal Server Error response.
func respond500() []byte {
	return []byte(`<html>
	<head>
		<title>500 Internal Server Error</title>
	</head>
	<body>
		<h1>Internal Server Error</h1>
		<p>Okay, you know what? This one is on me.</p>
	</body>
</html>`)
}

// respond200 returns the HTML body for a 200 OK success response.
func respond200() []byte {
	return []byte(`<html>
	<head>
		<title>200 OK</title>
	</head>
	<body>
		<h1>Success!</h1>
		<p>Your request was a masterpiece.</p>
	</body>
</html>`)
}

// handleRequest is the shared HTTP request handler used by both the native
// TCP server and the WASM browser server. It inspects the request target
// (URL path) and dispatches to the appropriate response.
//
// Routes:
//   - "/"             → 200 OK
//   - "/yourproblem"  → 400 Bad Request
//   - "/myproblem"    → 500 Internal Server Error
//   - "/video"        → Serves assets/video.mp4 (native only)
//   - "/httpbin/*"    → Proxies to httpbin.org with chunked transfer encoding (native only)
//
// In WASM mode, the /video and /httpbin/* routes will fail gracefully because:
//   - os.ReadFile("assets/video.mp4") → file system not available in browser
//   - http.Get("https://httpbin.org/...") → outbound HTTP works via fetch() but
//     the chunked streaming pattern may not work as expected in WASM
func handleRequest(w *response.Writer, req *request.Request) {
	h := response.GetDefaultHeaders(0)
	body := respond200()
	status := response.StatusOK
	if req.RequestLine.RequestTarget == "/yourproblem" {
		body = respond400()
		status = response.StatusBadRequest

	} else if req.RequestLine.RequestTarget == "/myproblem" {
		body = respond500()
		status = response.StatusInternalServerError

	} else if req.RequestLine.RequestTarget == "/video" {
		f, err := os.ReadFile("assets/video.mp4")
		if err != nil {
			body = respond500()
			status = response.StatusInternalServerError
		}
		h.Replace("content-type", "video/mp4")
		h.Replace("content-length", fmt.Sprintf("%d", len(f)))
		w.WriteStatusLine(response.StatusOK)
		w.WriteHeaders(*h)
		w.WriteBody(f)

	} else if strings.HasPrefix(req.RequestLine.RequestTarget, "/httpbin/") {
		target := req.RequestLine.RequestTarget
		res, err := http.Get("https://httpbin.org/" + target[len("/httpbin/"):])
		if err != nil {
			body = respond500()
			status = response.StatusInternalServerError
		} else {
			w.WriteStatusLine(response.StatusOK)
			h.Delete("Content-length")
			h.Set("Transfer-Encoding", "chunked")
			h.Replace("Content-Type", "text/plain")
			h.Set("Trailer", "X-Content-SHA256")
			h.Set("Trailer", "X-Content-Length")
			w.WriteHeaders(*h)

			fullBody := []byte{}
			for {
				data := make([]byte, 64)
				n, err := res.Body.Read(data)
				if err != nil {
					if errors.Is(io.EOF, err) {
						slog.Info("End of chunks reached\n")
						break
					}
					slog.Error(fmt.Sprintf("Error: %v\n", err))
					break
				}

				fullBody = append(fullBody, data[:n]...)
				w.WriteBody([]byte(fmt.Sprintf("%x\r\n", n)))
				w.WriteBody(data[:n])
				w.WriteBody([]byte("\r\n"))
			}
			w.WriteBody([]byte("0\r\n"))
			trailers := headers.NewHeaders()
			out := sha256.Sum256(fullBody)
			trailers.Set("X-Content_SHA256", toStr(out[:]))
			trailers.Set("X-Content-Length", fmt.Sprintf("%d", len(fullBody)))
			w.WriteHeaders(*trailers)
			w.WriteBody([]byte("0\r\n"))

			return
		}
	}

	w.WriteStatusLine(status)
	h.Replace("Content-length", fmt.Sprintf("%d", len(body)))
	h.Replace("Content-type", "text/html")
	w.WriteHeaders(*h)
	w.WriteBody(body)
}
