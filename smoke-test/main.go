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

func echo2(conn net.Conn) {

	defer conn.Close()

	for {

		buf := make([]byte, 4096)

		n, err := conn.Read(buf)

		if err != nil {
			slog.Error(err.Error(), "msg", "error while reading from connection")
			if err == io.EOF {
				return
			}
		}

		_, err = conn.Write(buf[:n])
		if err != nil {
			slog.Error(err.Error(), "msg", "error while writing to connection")
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
		//slog.Info("waiting for connections")
		conn, err := listener.Accept()
		//slog.Info("connection accepted from source " + conn.RemoteAddr().String())
		if err != nil {
			slog.Warn(err.Error(), "msg", "error while establishing connection")
			panic(err)
		}
		go echo2(conn)

	}
}
