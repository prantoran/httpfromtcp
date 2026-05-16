package request

import (
	"fmt"
	"io"
)

type Request struct {
	RequestLine RequestLine
}

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

var ERROR_BAD_START_LINE = fmt.Errorf("Bad start line")

func RequestFromReader(reader io.Reader) (*Request, error) {
	return nil, ERROR_BAD_START_LINE
}
