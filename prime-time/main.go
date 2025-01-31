package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"log/slog"
	"math"
	"net"
	"sync"
)

type Request struct {
	Method string   `json:"method"`
	Number *float64 `json:"number"`
}

type Response struct {
	Method string `json:"method"`
	Prime  bool   `json:"prime"`
}

func checkPrime(n float64) Response {

	if math.Ceil(n) != n {
		return Response{Method: "isPrime", Prime: false}
	}

	m := int64(n)
	if m <= 1 {
		return Response{Method: "isPrime", Prime: false}
	}

	var i int64
	for i = 2; i*i <= m; i++ {

		if m%i == 0 {
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

	return true
}
func handle(conn net.Conn, mutex *sync.Mutex, clientId int) {
	defer conn.Close()

	reader := bufio.NewReader(conn)
	for {

		requestString, err := reader.ReadString('\n')
		if err != nil {
			slog.Error(err.Error(), "client-id", clientId, "msg", "error while reading from connection")
			return
		}

		mutex.Lock()
		slog.Info("content sent "+requestString, "client-id", clientId)
		mutex.Unlock()

		var request Request
		if err := json.NewDecoder(bytes.NewBuffer([]byte(requestString))).Decode(&request); err != nil {
			slog.Error(err.Error(), "msg", "error while decoding from connection")
			if _, err := conn.Write([]byte("malformed")); err != nil {
				slog.Error(err.Error(), "msg", "error while writing malformed request to connection")

			}
			return
		} else if request.Number == nil {
			if _, err := conn.Write([]byte("malformed")); err != nil {
				slog.Error(err.Error(), "msg", "error while writing malformed request to connection")

			}
			return
		}
		//mutex.Lock()
		//slog.Info(fmt.Sprintf("%v", request), "client-id", clientId)
		//mutex.Unlock()

		if !ValidateRequest(request) {
			if _, err := conn.Write([]byte("malformed")); err != nil {
				slog.Error(err.Error(), "msg", "error while writing malformed request to connection")
			}
			return
		}

		response := checkPrime(*request.Number)
		if err := json.NewEncoder(conn).Encode(response); err != nil {
			slog.Error(err.Error(), "msg", "error while encoding response to connection")
			return
		}

	}
}
func main() {

	listener, err := net.Listen("tcp", "0.0.0.0:8080")

	mutex := &sync.Mutex{}
	idMutex := &sync.Mutex{}

	id := 0
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

		idMutex.Lock()
		id++
		localId := id
		idMutex.Unlock()

		go handle(conn, mutex, localId)
	}
}
