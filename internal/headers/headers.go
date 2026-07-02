package headers

import (
	"bytes"
	"fmt"
	"strings"
)

func isToken(str []byte) bool {
	if len(str) == 0 {
		return false
	}

	for _, ch := range str {
		found := false
		if ch >= 'A' && ch <= 'Z' ||
			ch >= 'a' && ch <= 'z' ||
			ch >= '0' && ch <= '9' {
			found = true
		}
		switch ch {
		case '!', '#', '$', '%', '&', '\'', '*', '+', '-', '.', '^', '_', '`', '|', '~':
			found = true
		}

		if !found {
			return false
		}
	}

	return true
}

var rn = []byte("\r\n")

func parseHeader(fieldLine []byte) (string, string, error) {
	parts := bytes.SplitN(fieldLine, []byte(": "), 2)
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid header format")
	}
	name := parts[0]
	value := bytes.TrimSpace(parts[1])

	if bytes.HasSuffix(name, []byte(" ")) {
		return "", "", fmt.Errorf("invalid field name: %s: space v", name)
	}
	if bytes.HasPrefix(name, []byte(" ")) {
		return "", "", fmt.Errorf("invalid field name: %s: space ^", name)
	}

	return string(name), string(value), nil
}

type Headers struct {
	headers map[string]string
}

func NewHeaders() *Headers {
	return &Headers{
		headers: make(map[string]string),
	}
}

func (h *Headers) Get(key string) string {
	return h.headers[strings.ToLower(key)]
}

func (h *Headers) Set(key, value string) {
	key = strings.ToLower(key)
	if v, ok := h.headers[key]; ok {
		h.headers[key] = fmt.Sprintf("%s,%s", v, value)
		return
	}
	h.headers[key] = value
}

func (h *Headers) Parse(data []byte) (int, bool, error) {

	read := 0
	done := false

	for {
		idx := bytes.Index(data[read:], rn)

		if idx == -1 {
			break
		}

		// Empty header
		if idx == 0 {
			done = true
			read += len(rn)
			break
		}

		name, value, err := parseHeader(data[read : read+idx])
		if err != nil {
			return 0, false, err
		}

		if !isToken([]byte(name)) {
			return 0, false, fmt.Errorf("malformed header name")
		}
		read += idx + len(rn)
		h.Set(name, value)
	}

	return read, done, nil
}
