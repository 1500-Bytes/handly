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
		data:              "GET / HTTP/1.1\r\nHost:localhost:4000\r\nUser-Agent: Go-http-client/1.1\r\n\r\n ",
		numOfBytesPerRead: 3,
	}

	r, err := RequestFromReader(reader)
	require.NoError(t, err)
	require.NotNil(t, r)
	// Test: Check for root path
	assert.Equal(t, "GET", r.RequestLine.HttpMethod)
	assert.Equal(t, "1.1", r.RequestLine.HttpVersion)
	assert.Equal(t, "/", r.RequestLine.RequestTarget)

	// Test: Check for target path
	reader = &ChunkReader{
		data:              "GET /cow HTTP/1.1\r\nHost:localhost:4000\r\nUser-Agent: Go-http-client/1.1\r\n\r\n  ",
		numOfBytesPerRead: 1,
	}

	r, err = RequestFromReader(reader)
	require.NoError(t, err)
	require.NotNil(t, r)
	assert.Equal(t, "1.1", r.RequestLine.HttpVersion)
	assert.Equal(t, "GET", r.RequestLine.HttpMethod)
	assert.Equal(t, "/cow", r.RequestLine.RequestTarget)

	// Test:Invalid version
	reader = &ChunkReader{
		data:              "GET /cow HTTP/1.1\r\nHost:localhost:4000\r\nUser-Agent: Go-http-client/1.1\r\n\r\n  ",
		numOfBytesPerRead: 2,
	}
	r, err = RequestFromReader(reader)
	require.NoError(t, err)
	require.NotNil(t, r)
	assert.NotEqual(t, "1.0", r.RequestLine.HttpVersion)
	assert.Equal(t, "GET", r.RequestLine.HttpMethod)
	assert.Equal(t, "/cow", r.RequestLine.RequestTarget)
}

func TestParseHeaders(t *testing.T) {
	// Test: Standard Headers
	reader := &ChunkReader{
		data:              "GET / HTTP/1.1\r\nHost: localhost:3000\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n",
		numOfBytesPerRead: 3,
	}
	r, err := RequestFromReader(reader)
	require.NoError(t, err)
	require.NotNil(t, r)
	host, exists := r.Headers.Get("host")
	assert.True(t, exists)
	assert.Equal(t, "localhost:3000", host)
	userAgent, exists := r.Headers.Get("user-agent")
	assert.True(t, exists)
	assert.Equal(t, "curl/7.81.0", userAgent)
	accept, exists := r.Headers.Get("accept")
	assert.True(t, exists)
	assert.Equal(t, "*/*", accept)

	// Test: Malformed Header
	reader = &ChunkReader{
		data:              "GET / HTTP/1.1\r\nHost localhost:3000\r\n\r\n",
		numOfBytesPerRead: 3,
	}
	r, err = RequestFromReader(reader)
	require.Error(t, err)
}

func TestParseBody(t *testing.T) {
	// Test: Standard Body
	reader := &ChunkReader{
		data:              "POST /submit HTTP/1.1\r\nHost: localhost:3000\r\nContent-Length: 13\r\n\r\nhello world!\n",
		numOfBytesPerRead: 3,
	}
	r, err := RequestFromReader(reader)
	require.NoError(t, err)

	require.NotNil(t, r)
	assert.Equal(t, "hello world!\n", string(r.Body))

	// Test: Body shorter than reported content length
	reader = &ChunkReader{
		data: "POST /submit HTTP/1.1\r\n" +
			"Host: localhost:3000\r\n" +
			"Content-Length: 20\r\n" +
			"\r\n" +
			"partial content",
		numOfBytesPerRead: 3,
	}
	r, err = RequestFromReader(reader)
	require.Error(t, err)
}
