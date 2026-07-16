/*
	go run cmd/httpserver/main.go

	curl localhost:42069 > /tmp/server.txt
	cat /tmp/server.txt
*/

package main

import (
	"io"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/prantoran/httpfromtcp/internal/request"
	"github.com/prantoran/httpfromtcp/internal/response"
	"github.com/prantoran/httpfromtcp/internal/server"
)

const port = 42069

func main() {
	s, err := server.Serve(port, func(w io.Writer, req *request.Request) *server.HandlerError {
		if req.RequestLine.RequestTarget == "/yourproblem" {
			return &server.HandlerError{
				StatusCode: response.StatusBadRequest,
				Message:    "Your probsie\n",
			}
		} else if req.RequestLine.RequestTarget == "/myproblem" {
			return &server.HandlerError{
				StatusCode: response.StatusInternalServerError,
				Message:    "My baad\n",
			}
		} else {
			w.Write([]byte("All good\n"))
		}

		return nil
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
