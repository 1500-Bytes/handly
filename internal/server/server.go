package server

import (
	"fmt"
	"io"
	"net"

	"github.com/merge/handly/internal/request"
	"github.com/merge/handly/internal/response"
)

type Handler func(w *response.Writer, r *request.Request)

type Server struct {
	close   bool
	handler Handler
}

type HandlerErr struct {
	StatusCode response.StatusCode
	Message    string
}

func (s *Server) HandleConn(conn io.ReadWriteCloser) {
	defer conn.Close()

	responseWriter := response.NewWriter(conn)
	r, err := request.RequestFromReader(conn)
	if err != nil {
		responseWriter.WriteStatusLine(response.StatusBadRequest)
		responseWriter.WriteHeaders(response.GetDefaultHeaders(0))
		return
	}

	s.handler(responseWriter, r)
}

func (s *Server) Listen(listener net.Listener) {
	for {
		conn, err := listener.Accept()
		if s.close {
			return
		}
		if err != nil {
			return
		}

		go s.HandleConn(conn)
	}

}

func ListenAndServe(port uint16, handler Handler) (*Server, error) {
	netListener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return nil, err
	}

	s := &Server{
		close:   false,
		handler: handler,
	}

	go s.Listen(netListener)

	return s, nil
}

func (s *Server) Close() {
	s.close = true
}
