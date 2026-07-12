/*
	go run cmd/httpserver/main.go

	curl localhost:42069 > /tmp/server.txt
	cat /tmp/server.txt
*/

package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/prantoran/httpfromtcp/internal/server"
)

const port = 42069

func main() {
	s, err := server.Serve(port)
	if err != nil {
		log.Fatalf("Error starting server: %v\n", err)
	}
	defer s.Close()
	log.Printf("Server listening on port %d\n", port)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
}
