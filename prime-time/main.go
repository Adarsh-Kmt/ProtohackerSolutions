package main

import (
	"bytes"
	"encoding/json"
	"log"
	"log/slog"
	"math"
	"net"
)

type Request struct {
	Method string  `json:"method"`
	Number float64 `json:"number"`
}

type Response struct {
	Method string `json:"method"`
	Prime  bool   `json:"prime"`
}

func checkPrime(n int) Response {
	if n <= 3 {
		if n == 1{
			return Response{Method: "isPrime", Prime: false}
		} else {
			return Response{Method: "isPrime", Prime: true}
		}
	} else if n%2 == 0 || n%3 == 0 {
		return Response{Method: "isPrime", Prime: false}
	}

	for i := 5; i*i <= n; i += 6 {
		if n%i == 0 || n%(i+2) == 0 {
			return Response{Method: "isPrime", Prime: false}
		}
	}

	return Response{Method: "isPrime", Prime: true}
}

func ValidateRequest(request Request) bool {

	if request.Method != "isPrime" {
		slog.Error("invalid method signature")
		return false
	}

	if math.Ceil(request.Number) != request.Number {
		slog.Error("float number")
		return false
	}

	return true
}
func handle(conn net.Conn) {
	defer conn.Close()

	for {

		buf := make([]byte, 4096)

		n, err := conn.Read(buf)
		if err != nil {
			slog.Error(err.Error(), "msg", "error while reading from connection")
			return
		}
		log.Println("content sent " + string(buf[:n]))
		var request Request
		if err := json.NewDecoder(bytes.NewBuffer(buf[:n])).Decode(&request); err != nil {
			slog.Error(err.Error(), "msg", "error while decoding from connection")
			if _, err := conn.Write([]byte("malformed")); err != nil {
				slog.Error(err.Error(), "msg", "error while writing malformed request to connection")

			}
			return
		}
		log.Println(request)
		if !ValidateRequest(request) {
			if _, err := conn.Write([]byte("malformed")); err != nil {
				slog.Error(err.Error(), "msg", "error while writing malformed request to connection")

			}
			return
		}

		response := checkPrime(int(request.Number))
		if err := json.NewEncoder(conn).Encode(response); err != nil {
			slog.Error(err.Error(), "msg", "error while encoding response to connection")
			return
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
		conn, err := listener.Accept()
		if err != nil {
			slog.Warn(err.Error(), "msg", "error while accepting connection")
			panic(err)
		}

		go handle(conn)
	}
}
