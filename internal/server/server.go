package server

import (
	"fmt"
	"io"
	"net"

	"github.com/prantoran/httpfromtcp/internal/request"
	"github.com/prantoran/httpfromtcp/internal/response"
)

type HandlerError struct {
	StatusCode response.StatusCode
	Message    string
}

// Handler is the function signature for HTTP request handlers.
// Both the native TCP server and the WASM bridge use this type.
type Handler func(w *response.Writer, req *request.Request)

type Server struct {
	Port    int
	Closed  bool
	handler Handler
}

// RunConnection handles a single HTTP request/response cycle on the given
// ReadWriteCloser. It reads the raw HTTP/1.1 request, parses it, invokes the
// handler, and writes the response.
//
// This function is exported (capitalized) so the WASM bridge can call it
// directly. In native mode, it's called by runServer() for each accepted
// TCP connection. In WASM mode, the bridge creates an io.Pipe-based
// ReadWriteCloser and passes it here.
func RunConnection(handler Handler, conn io.ReadWriteCloser) {
	defer conn.Close()

	responseWriter := response.NewWriter(conn)
	r, err := request.RequestFromReader(conn)
	if err != nil {
		responseWriter.WriteStatusLine(response.StatusOK)
		responseWriter.WriteHeaders(*response.GetDefaultHeaders(0))
		return
	}

	handler(responseWriter, r)
}

// HandleRequest is the public API for processing a single HTTP request on an
// io.ReadWriteCloser. This is the entry point used by the WASM bridge
// (wasm_bridge.go) to process requests without a TCP listener.
//
// In the WASM environment:
//   - rw.Read()  returns raw HTTP/1.1 request bytes from an io.Pipe
//   - rw.Write() captures raw HTTP/1.1 response bytes into a buffer
//   - rw.Close() is a no-op (no real socket to close)
//
// This function exists separately from RunConnection to provide a clean
// public API that doesn't depend on Server state.
func HandleRequest(handler Handler, rw io.ReadWriteCloser) {
	RunConnection(handler, rw)
}

func runServer(s *Server, listener net.Listener) {
	for {
		conn, err := listener.Accept()
		if s.Closed {
			return
		}
		if err != nil {
			fmt.Println("Error accepting connection:", err)
			return
		}

		go RunConnection(s.handler, conn)
	}
}

func Serve(port uint16, handler Handler) (*Server, error) {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return nil, err
	}
	s := &Server{
		Port:    int(port),
		Closed:  false,
		handler: handler,
	}
	go runServer(s, listener)
	return s, nil
}

func (s *Server) Close() error {
	s.Closed = true
	return nil
}
