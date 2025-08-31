package headers

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHeaderParse(t *testing.T) {
	// Test: Valid single header
	h := NewHeaders()
	data := []byte("Host: localhost:42069\r\n\r\n")
	n, done, err := h.Parse(data)
	require.NoError(t, err)
	require.NotNil(t, h)
	host, exists := h.Get("Host")
	assert.True(t, exists)
	assert.Equal(t, "localhost:42069", host)
	assert.Equal(t, 25, n)
	assert.True(t, done)

	// Test: Invalid spacing header
	h = NewHeaders()
	data = []byte("       Host : localhost:42069       \r\n\r\n")
	n, done, err = h.Parse(data)
	require.Error(t, err)
	assert.Equal(t, 0, n)
	assert.False(t, done)

	// Test: Invalid field-name
	h = NewHeaders()
	data = []byte("H©st: localhost:42069\r\n\r\n")
	n, done, err = h.Parse(data)
	require.Error(t, err)
	assert.Equal(t, 0, n)
	assert.False(t, done)

	// Test: Multiple indentical field-name's
	h = NewHeaders()
	data = []byte("Host: localhost:42069\r\nHost: localhost:3000\r\n\r\n")
	n, done, err = h.Parse(data)
	require.NoError(t, err)
	require.NotNil(t, h)
	host, exists = h.Get("Host")
	assert.True(t, exists)
	assert.NotEqual(t, "localhost:42069", host)
	assert.Equal(t, "localhost:42069,localhost:3000", host)
	assert.Equal(t, 47, n)
	assert.True(t, done)

}
