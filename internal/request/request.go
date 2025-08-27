package request

import (
	"bytes"
	"fmt"
	"io"
)

var (
	SEPARATOR                                = []byte("\r\n")
	StateInit                    parsetState = "initialized"
	StateDone                    parsetState = "done"
	StateError                   parsetState = "error"
	ERR_POISNED_REQUEST                      = fmt.Errorf("invalid request format")
	ERR_UNSUPPORTED_HTTP_VERSION             = fmt.Errorf("unsupported http version")
)

type parsetState string

type RequestLine struct {
	HttpMethod    string
	HttpVersion   string
	RequestTarget string
}

type Request struct {
	RequestLine RequestLine
	state       parsetState
}

func newRequest() *Request {
	return &Request{
		state: StateInit,
	}
}

func (r *Request) parse(data []byte) (int, error) {
	read := 0

outer:
	for {
		switch r.state {
		case StateError:
			r.state = StateError
			break outer
		case StateInit:
			rl, n, err := parseRequestLine(data[read:])
			if err != nil {
				return 0, err
			}

			if n == 0 {
				break outer
			}

			r.RequestLine = *rl
			read += n
			r.state = StateDone

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
	if len(httpParts) != 2 || string(parts[0]) != "GET" || string(httpParts[1]) != "1.1" {
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
		readN, err := request.parse(buf[:bufLen])
		if err != nil {
			return nil, err
		}

		copy(buf, buf[readN:bufLen])
		bufLen = bufLen - readN
	}

	return request, nil
}
