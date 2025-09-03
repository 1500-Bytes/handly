package response

import (
	"fmt"
	"io"

	"github.com/merge/handly/internal/headers"
)

type StatusCode int

const (
	StatusOK                  StatusCode = 200
	StatusBadRequest          StatusCode = 400
	StatusInternalServerError StatusCode = 500
)

type Response struct {
}

func WriteStatusLine(w io.Writer, statusCode StatusCode) error {
	statusLine := []byte{}

	switch statusCode {
	case StatusOK:
		statusLine = []byte("HTTP/1.1 200 OK\r\n")
	case StatusBadRequest:
		statusLine = []byte("HTTP/1.1 400 Bad Request\r\n")
	case StatusInternalServerError:
		statusLine = []byte("HTTP/1.1 500 Internal Server Error\r\n")
	}

	_, err := w.Write(statusLine)
	return err
}

func WriteHeaders(w io.Writer, h *headers.Headers) error {
	b := []byte{}
	h.ForEach(func(n, v string) {
		b = fmt.Appendf(b, "%s: %s\r\n", n, v)
	})
	b = fmt.Append(b, "\r\n")
	_, err := w.Write(b)
	return err
}

func GetDefaultHeaders(contentLength int) *headers.Headers {
	h := headers.NewHeaders()
	h.Set("Content-length", fmt.Sprintf("%d\r\n", contentLength))
	h.Set("Connection", "close")
	h.Set("Content-type", "text/plain")

	return h
}

type Writer struct {
	Writer io.Writer
}

func NewWriter(writer io.Writer) *Writer {
	return &Writer{
		Writer: writer,
	}
}

func (w *Writer) WriteStatusLine(statusCode StatusCode) error {
	statusLine := []byte{}

	switch statusCode {
	case StatusOK:
		statusLine = []byte("HTTP/1.1 200 OK\r\n")
	case StatusBadRequest:
		statusLine = []byte("HTTP/1.1 400 Bad Request\r\n")
	case StatusInternalServerError:
		statusLine = []byte("HTTP/1.1 500 Internal Server Error\r\n")
	}

	_, err := w.Writer.Write(statusLine)
	return err
}

func (w *Writer) WriteHeaders(h *headers.Headers) error {
	b := []byte{}
	h.ForEach(func(n, v string) {
		b = fmt.Appendf(b, "%s: %s\r\n", n, v)
	})
	b = fmt.Append(b, "\r\n")
	_, err := w.Writer.Write(b)
	return err
}

func (w *Writer) WriteBody(p []byte) (int, error) {
	n, err := w.Writer.Write(p)
	if err != nil {
		return 0, err
	}

	return n, nil
}
