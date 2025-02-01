package main

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"fmt"
	"github.com/zavitax/sortedset-go"
	"io"
	"log/slog"
	"math"
	"net"
	"sync"
)

type InsertRequest struct {
	price     int32
	timestamp int32
}

type QueryRequest struct {
	minTime int32
	maxTime int32
}

func parseInsertRequest(buf []byte) (*InsertRequest, error) {

	var price int32
	var timestamp int32

	if err := binary.Read(bytes.NewBuffer(buf[1:5]), binary.BigEndian, &timestamp); err != nil {
		return nil, err
	}
	if err := binary.Read(bytes.NewBuffer(buf[5:9]), binary.BigEndian, &price); err != nil {
		return nil, err
	}

	return &InsertRequest{price: price, timestamp: timestamp}, nil
}

func handleInsertRequest(sortedSet *sortedset.SortedSet[int32, int32, int32], request *InsertRequest) {

	sortedSet.AddOrUpdate(request.timestamp, request.timestamp, request.price)

}

func handleQueryRequest(sortedSet *sortedset.SortedSet[int32, int32, int32], request *QueryRequest) (mean int32) {

	nodes := sortedSet.GetRangeByRank(int(request.minTime), int(request.maxTime), false)

	var sum int32 = 0
	for _, node := range nodes {
		sum += node.Value
	}

	mean = int32(math.Ceil(float64(sum) / float64(len(nodes))))
	return mean
}
func parseQueryRequest(buf []byte) (*QueryRequest, error) {
	var minTime int32
	var maxTime int32
	if err := binary.Read(bytes.NewBuffer(buf[1:5]), binary.BigEndian, &minTime); err != nil {
		return nil, err
	}
	if err := binary.Read(bytes.NewBuffer(buf[5:9]), binary.BigEndian, &maxTime); err != nil {
		return nil, err
	}
	return &QueryRequest{minTime: minTime, maxTime: maxTime}, nil
}

func handleClient(conn net.Conn, clientId int) {

	defer conn.Close()

	sortedSet := sortedset.New[int32, int32, int32]()

	reader := bufio.NewReader(conn)
	for {

		buf := make([]byte, 0)

		for i := 0; i < 9; i++ {
			b, err := reader.ReadByte()
			slog.Info(fmt.Sprintf("received byte %b", b))
			if err != nil {
				if err != io.EOF && i == 0 {
					return
				}
			}
			buf = append(buf, b)
		}

		slog.Info(fmt.Sprintf("%v", buf))

		if int(buf[0]) == int('I') {
			request, err := parseInsertRequest(buf)
			if err != nil {
				slog.Error(err.Error(), clientId, "client-id", clientId, "msg", "error while parsing insert request")
				return
			} else {
				slog.Info(fmt.Sprintf("received insert request: %v", request), "client-id", clientId)
			}
			handleInsertRequest(sortedSet, request)

		} else if int(buf[0]) == int('Q') {
			request, err := parseQueryRequest(buf)
			if err != nil {
				slog.Error(err.Error(), clientId, "client-id", clientId, "msg", "error while parsing query request")
				return
			} else {
				slog.Info(fmt.Sprintf("received query request: %v", request), "client-id", clientId)
			}
			mean := handleQueryRequest(sortedSet, request)

			if err = binary.Write(conn, binary.BigEndian, mean); err != nil {
				slog.Error(err.Error(), "client-id", clientId, "msg", "error while writing to connection")
				return
			}

		}
	}
}
func main() {

	listener, err := net.Listen("tcp", "0.0.0.0:8080")

	clientId := 0

	clientIdMutex := &sync.Mutex{}

	if err != nil {
		slog.Warn(err.Error(), "msg", "error while listening on port 8080")
		return
	} else {
		slog.Info("Listening on port 8080")
	}

	for {
		conn, err := listener.Accept()
		if err != nil {
			slog.Warn(err.Error(), "msg", "error while accepting connection")
			return
		} else {
			slog.Info("connection established!")
		}

		clientIdMutex.Lock()
		clientId++
		currClientId := clientId
		clientIdMutex.Unlock()

		go handleClient(conn, currClientId)
	}
}
