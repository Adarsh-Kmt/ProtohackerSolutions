package main

import (
	"io"
	"log/slog"
	"net"
)

func echo(conn net.Conn) {

	defer conn.Close()
	body := make([]byte, 15*1024)

	for {
		_, err := conn.Read(body)

		if err != nil {

			if err == io.EOF {
				return
			}
			slog.Warn(err.Error(), "msg", "error while reading from connection")
		}
		slog.Info("received message => " + string(body))

		_, err = conn.Write(body)
		if err != nil {
			slog.Warn(err.Error(), "msg", "error while writing to connection")
		}

		if err := conn.Close(); err != nil {
			slog.Warn(err.Error(), "msg", "error while closing connection")

		}
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
