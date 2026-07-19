/*
	go run cmd/httpserver/main.go

	curl localhost:42069 > /tmp/server.txt
	cat /tmp/server.txt
*/

package main

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/prantoran/httpfromtcp/internal/headers"
	"github.com/prantoran/httpfromtcp/internal/request"
	"github.com/prantoran/httpfromtcp/internal/response"
	"github.com/prantoran/httpfromtcp/internal/server"
)

func toStr(bytes []byte) string {
	out := ""
	for _, b := range bytes {
		out += fmt.Sprintf("%02x", b)
	}
	return out
}

const port = 42069

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

func main() {
	s, err := server.Serve(port, func(w *response.Writer, req *request.Request) {
		h := response.GetDefaultHeaders(0)
		body := respond200()
		status := response.StatusOK
		if req.RequestLine.RequestTarget == "/yourproblem" {
			body = respond400()
			status = response.StatusBadRequest

		} else if req.RequestLine.RequestTarget == "/myproblem" {
			body = respond500()
			status = response.StatusInternalServerError

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
					data := make([]byte, 32)
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
	})
	if err != nil {
		log.Fatalf("Error starting server: %v\n", err)
	}
	defer s.Close()
	log.Printf("Server listening on port %d\n", port)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
}
