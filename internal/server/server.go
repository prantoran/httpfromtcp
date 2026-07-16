package server

import (
	"bufio"
	"bytes"
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

type Handler func(w io.Writer, req *request.Request) *HandlerError

type Server struct {
	Port    int
	Closed  bool
	handler Handler
}

func runConnection(s *Server, conn io.ReadWriteCloser) {
	defer conn.Close()

	headers := response.GetDefaultHeaders(0)
	r, err := request.RequestFromReader(bufio.NewReader(conn))
	if err != nil {
		response.WriteStatusLine(conn, response.StatusOK)
		response.WriteHeaders(conn, headers)
		return
	}

	writer := bytes.NewBuffer([]byte{})
	handlerErr := s.handler(writer, r)

	var body []byte = nil
	var status response.StatusCode = response.StatusOK
	if handlerErr != nil {
		status = handlerErr.StatusCode
		body = []byte(handlerErr.Message)
	} else {
		body = writer.Bytes()
	}

	headers.Replace("Content-length", fmt.Sprintf("%d", len(body)))

	response.WriteStatusLine(conn, status)
	response.WriteHeaders(conn, headers)
	conn.Write(body)
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

		go runConnection(s, conn)
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
