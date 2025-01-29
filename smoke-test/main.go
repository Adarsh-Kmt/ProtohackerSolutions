package main

import (
	"io"
	"log/slog"
	"net"
)

func echo(conn net.Conn) {

	_, err := io.Copy(conn, conn)
	if err != nil {
		slog.Warn(err.Error(), "msg", "error while echoing")
	}

	if err := conn.Close(); err != nil {
		slog.Warn(err.Error(), "msg", "error while closing connection")
	}

}
func main() {

	listener, err := net.Listen("tcp", "0.0.0.0:8080")

	if err != nil {
		slog.Warn(err.Error(), "msg", "error while listening on port 8080")
		panic(err)
	}

	for {
		slog.Info("waiting for connections")
		conn, err := listener.Accept()
		slog.Info("connection accepted from source " + conn.RemoteAddr().String())
		if err != nil {
			slog.Warn(err.Error(), "msg", "error while establishing connection")
			panic(err)
		}
		go echo(conn)

	}
}
