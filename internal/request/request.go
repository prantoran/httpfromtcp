package request

import (
	"bytes"
	"errors"
	"fmt"
	"io"

	"github.com/prantoran/httpfromtcp/internal/headers"
)

type parserState string

const (
	StateInit    parserState = "init"
	StateHeaders parserState = "headers"
	StateDone    parserState = "done"
	StateError   parserState = "error"
)

type Request struct {
	RequestLine RequestLine
	state       parserState
	Headers     *headers.Headers
}

func (r *Request) parse(data []byte) (int, error) {
	read := 0
outer:
	for {
		curData := data[read:]
		switch r.state {
		case StateError:
			return 0, ErrorRequestInErrorState
		case StateInit:
			rl, n, err := parseRequestLine(curData)
			if err != nil {
				r.state = StateError
				return 0, err
			}
			if n == 0 {
				break outer
			}
			r.RequestLine = *rl
			read += n
			r.state = StateHeaders
		case StateHeaders:
			n, done, err := r.Headers.Parse(curData)
			if err != nil {
				r.state = StateError
				return 0, err
			}
			if n == 0 && !done {
				break outer
			}
			read += n
			if done {
				r.state = StateDone
			}
		case StateDone:
			break outer
		default:
			return 0, fmt.Errorf("unknown state: %s", r.state)
		}
	}
	return read, nil
}

func (r *Request) done() bool {
	return r.state == StateDone
}
func (r *Request) error() bool {
	return r.state == StateError
}

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

func newRequest() *Request {
	return &Request{
		state:   StateInit,
		Headers: headers.NewHeaders(),
	}
}

var ERROR_BAD_START_LINE = fmt.Errorf("bad start line")
var ERROR_INCOMPLETE_START_LINE = fmt.Errorf("incomplete start line")
var ERROR_UNSUPPORTED_METHOD = fmt.Errorf("unsupported method")
var ErrorRequestInErrorState = fmt.Errorf("request in error state")
var SEPERATOR = []byte("\r\n")

func parseRequestLine(b []byte) (*RequestLine, int, error) {
	idx := bytes.Index(b, SEPERATOR)
	if idx == -1 {
		return nil, 0, nil
	}

	startLine := b[:idx]
	read := idx + len(SEPERATOR)

	parts := bytes.Split(startLine, []byte(" "))
	if len(parts) != 3 {
		return nil, 0, ERROR_BAD_START_LINE
	}

	httpParts := bytes.Split(parts[2], []byte("/"))
	if len(httpParts) != 2 || string(httpParts[0]) != "HTTP" || string(httpParts[1]) != "1.1" {
		return nil, read, ERROR_BAD_START_LINE
	}

	rl := &RequestLine{
		Method:        string(parts[0]),
		RequestTarget: string(parts[1]),
		HttpVersion:   string(httpParts[1]),
	}

	return rl, read, nil
}

func RequestFromReader(reader io.Reader) (*Request, error) {
	request := newRequest()

	buf := make([]byte, 4096)
	bufLen := 0

	for !request.done() && !request.error() {
		n, err := reader.Read(buf[bufLen:])
		if err != nil {
			if err == io.EOF {
				break
			} else {
				return nil, errors.Join(fmt.Errorf("unable to read from reader"), err)
			}
		}
		bufLen += n
		readN, err := request.parse(buf[:bufLen])
		if err != nil {
			return nil, errors.Join(fmt.Errorf("unable to parse request"), err)
		}

		copy(buf, buf[readN:bufLen])
		bufLen -= readN

	}

	return request, nil
}
