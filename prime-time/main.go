package main

import (
	"encoding/json"
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

func checkPrime(number int) Response {

	for i := 2; i <= int(math.Sqrt(float64(number))); i++ {
		if number%i == 0 {
			return Response{Method: "isPrime", Prime: false}
		}
	}
	return Response{Method: "isPrime", Prime: true}
}

func ValidateRequest(request Request) bool {

	if request.Method != "isPrime" {
		return false
	}
	if request.Number < 0 {
		return false
	}
	if math.Ceil(request.Number) != request.Number {
		return false
	}

	return true
}
func handle(conn net.Conn) {
	defer conn.Close()

	for {

		var request Request
		if err := json.NewDecoder(conn).Decode(&request); err != nil {
			slog.Error(err.Error(), "msg", "error while decoding from connection")
			return
		}

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
