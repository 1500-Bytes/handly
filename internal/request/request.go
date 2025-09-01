package request

import (
	"bytes"
	"fmt"
	"io"

	"github.com/merge/handly/internal/headers"
)

var (
	SEPARATOR                                = []byte("\r\n")
	StateInit                    parsetState = "initialized"
	StateDone                    parsetState = "done"
	StateBody                    parsetState = "body"
	StateHeader                  parsetState = "header"
	StateError                   parsetState = "error"
	ERR_POISNED_REQUEST                      = fmt.Errorf("invalid request format")
	ERR_UNSUPPORTED_HTTP_VERSION             = fmt.Errorf("unsupported http version")
	ERR_REQ_IN_ERR_STATE                     = fmt.Errorf("request is in err state")
)

type parsetState string

type RequestLine struct {
	HttpMethod    string
	HttpVersion   string
	RequestTarget string
}

type Request struct {
	RequestLine RequestLine
	Headers     *headers.Headers
	Body        string
	bodyBuffer  []byte
	bodyPos     int
	state       parsetState
}

func newRequest() *Request {
	return &Request{
		Headers: headers.NewHeaders(),
		state:   StateInit,
	}
}

func (r *Request) hasBody() bool {
	length := r.Headers.GetInt("content-length", 0)
	return length > 0

}

func (r *Request) Parse(data []byte) (int, error) {
	read := 0

outer:
	for {
		currentData := data[read:]
		if len(currentData) == 0 {
			break outer
		}

		switch r.state {
		case StateError:
			return 0, ERR_REQ_IN_ERR_STATE
		case StateInit:
			rl, n, err := parseRequestLine(currentData)
			if err != nil {
				r.state = StateError
				return 0, err
			}

			if n == 0 {
				break outer
			}

			r.RequestLine = *rl
			read += n
			r.state = StateHeader

		case StateHeader:
			n, done, err := r.Headers.Parse(currentData)

			if err != nil {
				r.state = StateError
				return 0, err
			}

			if n == 0 {
				break outer
			}

			read += n

			if done {
				if r.hasBody() {
					length := r.Headers.GetInt("content-length", 0)

					r.bodyBuffer = make([]byte, length)
					r.bodyPos = 0
					r.state = StateBody
				} else {
					r.state = StateDone
				}
			}

		case StateBody:
			length := r.Headers.GetInt("content-length", 0)
			if length == 0 {
				r.state = StateDone
				break outer
			}

			remaining := min(length-r.bodyPos, len(currentData))
			copy(r.bodyBuffer[r.bodyPos:], currentData[:remaining])

			r.bodyPos += remaining
			read += remaining

			if r.bodyPos == length {
				r.Body = string(r.bodyBuffer) // Single conversion at end
				r.state = StateDone
			}
		case StateDone:
			break outer
		}
	}

	return read, nil
}

func (r *Request) done() bool {
	return r.state == StateDone || r.state == StateError
}

func parseRequestLine(data []byte) (*RequestLine, int, error) {
	idx := bytes.Index(data, SEPARATOR)
	if idx == -1 {
		return nil, 0, nil
	}

	startOfLine := data[:idx]
	read := idx + len(SEPARATOR)

	parts := bytes.Split(startOfLine, []byte(" "))
	if len(parts) != 3 {
		return nil, read, ERR_POISNED_REQUEST
	}

	httpParts := bytes.Split(parts[2], []byte("/"))
	if len(httpParts) != 2 || string(httpParts[1]) != "1.1" {
		return nil, 0, ERR_POISNED_REQUEST
	}

	return &RequestLine{
		HttpMethod:    string(parts[0]),
		RequestTarget: string(parts[1]),
		HttpVersion:   string(httpParts[1]),
	}, read, nil
}

func RequestFromReader(r io.Reader) (*Request, error) {
	request := newRequest()

	buf := make([]byte, 1024)
	bufLen := 0
	for !request.done() {
		n, err := r.Read(buf[bufLen:])
		if err != nil {
			return nil, err
		}

		bufLen += n
		readN, err := request.Parse(buf[:bufLen])
		if err != nil {
			return nil, err
		}

		copy(buf, buf[readN:bufLen])
		bufLen = bufLen - readN
	}

	return request, nil
}
