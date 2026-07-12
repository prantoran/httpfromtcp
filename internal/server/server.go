package server

import (
	"fmt"
	"io"
	"net"
)

type Server struct {
	Port   int
	Closed bool
}

func runConnection(s *Server, conn io.ReadWriteCloser) {
	out := []byte("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\n\r\nHello local world!\r\n")
	conn.Write(out)
	conn.Close()
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

func Serve(port uint16) (*Server, error) {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return nil, err
	}
	s := &Server{Port: int(port), Closed: false}
	go runServer(s, listener)
	return s, nil
}

func (s *Server) Close() error {
	s.Closed = true
	return nil
}
