package headers

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHeadersParse(t *testing.T) {
	headers := NewHeaders()
	data := []byte("Host: localhost:42069\r\nFoo:    bar    \r\n\r\n")
	n, done, err := headers.Parse(data)
	require.NoError(t, err)
	require.NotNil(t, headers)
	assert.Equal(t, "localhost:42069", headers.Get("Host"))
	assert.Equal(t, "bar", headers.Get("Foo"))
	assert.Equal(t, "", headers.Get("MissingKey"))
	assert.Equal(t, 42, n)
	assert.True(t, done)

	headers = NewHeaders()
	data = []byte("       Host: localhost:42069\r\nContent-Length: 0\r\n\r\n")
	n, done, err = headers.Parse(data)
	require.Error(t, err)
	assert.Equal(t, 0, n)
	assert.False(t, done)
}
