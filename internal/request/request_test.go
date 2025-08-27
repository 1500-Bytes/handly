package request

import (
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type ChunkReader struct {
	data              string
	numOfBytesPerRead int
	i                 int // current reading index
}

// Read reads up to len(p) or numBytesPerRead bytes from the string per call
// its useful for simulating reading a variable number of bytes per chunk from a network connection
func (cr *ChunkReader) Read(p []byte) (int, error) {
	if cr.i >= len(cr.data) {
		return 0, io.EOF
	}

	// endIndex := len(cr.data) + cr.pos // this stimulates reading up to the len(cr.data) and writing to the p
	endIndex := cr.numOfBytesPerRead + cr.i // this stimulates reading cr.numOfBytesPerRead and writing to the p
	if endIndex > len(cr.data) {
		endIndex = len(cr.data)
	}

	n := copy(p, cr.data[cr.i:endIndex])
	cr.i += n
	return n, nil
}

func TestRequest(t *testing.T) {

	reader := &ChunkReader{
		data:              "GET / HTTP/1.1\r\nHost:localhost:4000\r\nUser-Agent: Go-http-client/1.1\r\n",
		numOfBytesPerRead: 3,
	}

	r, err := RequestFromReader(reader)
	require.NoError(t, err)
	require.NotNil(t, r)
	// Test: Check for root path
	assert.Equal(t, "GET", r.RequestLine.HttpMethod)
	assert.Equal(t, "1.1", r.RequestLine.HttpVersion)
	assert.Equal(t, "/", r.RequestLine.RequestTarget)

	// // Test: Check for target path

	// reader = &ChunkReader{
	// 	data:              "GET /cow HTTP/1.1\r\nHost:localhost:4000\r\nUser-Agent: Go-http-client/1.1\r\n",
	// 	numOfBytesPerRead: 1,
	// }

	// r, err = RequestFromReader(reader)
	// require.NoError(t, err)
	// require.NotNil(t, r)
	// assert.Equal(t, "1.1", r.RequestLine.HttpVersion)
	// assert.Equal(t, "GET", r.RequestLine.HttpMethod)
	// assert.Equal(t, "/cow", r.RequestLine.RequestTarget)

	// // Test:Invalid version
	// reader = &ChunkReader{
	// 	data:              "GET /cow HTTP/1.1\r\nHost:localhost:4000\r\nUser-Agent: Go-http-client/1.1\r\n",
	// 	numOfBytesPerRead: 2,
	// }
	// r, err = RequestFromReader(reader)
	// require.NoError(t, err)
	// require.NotNil(t, r)
	// assert.NotEqual(t, "1.0", r.RequestLine.HttpVersion)
	// assert.Equal(t, "GET", r.RequestLine.HttpMethod)
	// assert.Equal(t, "/cow", r.RequestLine.RequestTarget)
}
