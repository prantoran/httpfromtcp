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
	host, ok := headers.Get("Host")
	require.True(t, ok)
	assert.Equal(t, "localhost:42069", host)
	foo, ok := headers.Get("Foo")
	require.True(t, ok)
	assert.Equal(t, "bar", foo)
	missing, ok := headers.Get("MissingKey")
	require.False(t, ok)
	assert.Equal(t, "", missing)
	assert.Equal(t, 42, n)
	assert.True(t, done)

	headers = NewHeaders()
	data = []byte("       Host: localhost:42069\r\nContent-Length: 0\r\n\r\n")
	n, done, err = headers.Parse(data)
	require.Error(t, err)
	assert.Equal(t, 0, n)
	assert.False(t, done)

	headers = NewHeaders()
	data = []byte("H@st: localhost:42069\r\nContent-Length: 0\r\n\r\n")
	n, done, err = headers.Parse(data)
	require.Error(t, err)
	assert.Equal(t, 0, n)
	assert.False(t, done)

	headers = NewHeaders()
	data = []byte("Host: localhost:42069\r\nHost: localhost:42069\r\n\r\n")
	n, done, err = headers.Parse(data)
	require.NoError(t, err)
	require.NotNil(t, headers)
	host, ok = headers.Get("Host")
	require.True(t, ok)
	assert.Equal(t, "localhost:42069,localhost:42069", host)
	assert.True(t, done)
}
