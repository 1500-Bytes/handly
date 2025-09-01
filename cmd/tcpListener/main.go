package main

import (
	"fmt"
	"log"
	"net"

	"github.com/merge/handly/internal/request"
)

func main() {
	listener, err := net.Listen("tcp", ":4000")
	if err != nil {
		log.Fatal("failed to listen on: ", listener.Addr().String())
	}

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Fatal("failed to accept connection: ", conn.RemoteAddr().String())
		}

		r, err := request.RequestFromReader(conn)
		if err != nil {
			log.Fatal("failed to read from accepted connection", err)
		}

		fmt.Println("Request line: ")
		fmt.Printf("- Method: %s\n", r.RequestLine.HttpMethod)
		fmt.Printf("- Target: %s\n", r.RequestLine.RequestTarget)
		fmt.Printf("- Version: %s\n", r.RequestLine.HttpVersion)
		fmt.Println("Header: ")
		r.Headers.ForEach(func(n, v string) {
			fmt.Printf("%s: %s\n", n, v)
		})
		fmt.Println("Body: ")
		fmt.Printf("%s\n", r.Body)
	}

}
