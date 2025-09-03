package main

import (
	"crypto/sha256"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/merge/handly/internal/headers"
	"github.com/merge/handly/internal/request"
	"github.com/merge/handly/internal/response"
	"github.com/merge/handly/internal/server"
)

const port = 4000

func toStr(data []byte) string {
	builder := strings.Builder{}
	for _, b := range data {
		builder.WriteString(fmt.Sprintf("%02x", b))
	}

	return builder.String()
}

func respond400() []byte {
	body := []byte(`
		<html>
		  <head>
		    <title>400 Bad Request</title>
		  </head>
		  <body>
		    <h1>Bad Request</h1>
		    <p>Your request honestly kinda sucked.</p>
		  </body>
		</html>
		`)
	return body
}

func respond500() []byte {
	body := []byte(`
		<html>
		  <head>
		    <title>500 Internal Server Error</title>
		  </head>
		  <body>
		    <h1>Internal Server Error</h1>
		    <p>Okay, you know what? This one is on me.</p>
		  </body>
		</html>
		`)
	return body
}

func respond200() []byte {
	body := []byte(`
		<html>
		  <head>
		    <title>200 OK</title>
		  </head>
		  <body>
		    <h1>Success!</h1>
		    <p>Your request was an absolute banger.</p>
		  </body>
		</html>
		`)
	return body
}

func main() {
	s, err := server.ListenAndServe(
		port,
		func(w *response.Writer, r *request.Request) {
			var (
				target = r.RequestLine.RequestTarget
				h      = response.GetDefaultHeaders(0)
				status = response.StatusOK
				body   = respond200()
			)

			if target == "/cat" {
				body = respond400()
				status = response.StatusBadRequest
			}

			if target == "/dog" {
				body = respond400()
				status = response.StatusBadRequest
			}

			if target == "/video" {
				slog.Info("jandler", "target", target)

				f, _ := os.ReadFile("assets/vim.mp4")
				h.Replace("Content-type", "video/mp4")
				h.Replace("Content-length", fmt.Sprintf("%d", len(f)))
				w.WriteStatusLine(status)
				w.WriteHeaders(h)
				w.WriteBody(f)

				// file, _ := os.Open("./assets/vim.mp4")
				// stat, _ := file.Stat()
				// h.Replace("Content-type", "video/mp4")
				// h.Replace("Content-length", fmt.Sprintf("%d", stat.Size()))
				// w.WriteHeaders(h)
				// io.Copy(w.Writer, file)
			}

			if strings.HasPrefix(target, "/httpbin/") {
				res, err := http.Get("https://httpbin.org/" + target[len("/httpbin/"):])
				if err != nil {
					status = response.StatusInternalServerError
					body = respond500()
				} else {
					w.WriteStatusLine(status)

					h.Delete("Content-length")
					h.Set("Transfer-encoding", "chuncked")
					h.Replace("Content-type", "text/plain")
					w.WriteHeaders(h)

					fullBody := make([]byte, 0)

					for {
						buf := make([]byte, 30)
						n, err := res.Body.Read(buf)
						if err != nil {
							break
						}

						fullBody = append(fullBody, buf[:n]...)
						w.WriteBody(fmt.Appendf(nil, "%x\r\n", n))
						w.WriteBody(buf[:n])
						w.WriteBody([]byte("\r\n"))
					}

					w.WriteBody([]byte("0\r\n"))

					out := sha256.Sum256(fullBody)
					trailers := headers.NewHeaders()
					trailers.Set("X-Content-SHA256", toStr(out[:]))
					trailers.Set("X-Content-length", fmt.Sprintf("%d", len(fullBody)))
					w.WriteHeaders(trailers)
					return
				}
			}

			h.Replace("Content-length", fmt.Sprintf("%d", len(body)))
			h.Replace("Content-type", "text/html")
			w.WriteStatusLine(status)
			w.WriteHeaders(h)
			w.WriteBody(body)
		})

	if err != nil {
		log.Fatalf("Error starting server: %v", err)
	}

	defer s.Close()
	log.Println("Server started on port", port)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	log.Println("Server gracefully stopped")
}
