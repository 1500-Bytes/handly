build:
	go build -o ./bin/tcp_to_http

run: build
	./bin/tcp_to_http

test:
	go test -v ./...
