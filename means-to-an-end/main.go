package main

import (
	"bytes"
	"encoding/binary"
	"github.com/zavitax/sortedset-go"
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

	for {

		buf := make([]byte, 9)

		_, err := conn.Read(buf)
		if err != nil {
			slog.Error(err.Error(), "client-id", clientId, "msg", "error while reading from connection")
			return
		}

		var command int
		if err := binary.Read(bytes.NewBuffer(buf[0:1]), binary.BigEndian, &command); err != nil {
			return
		}

		if command == int('I') {
			request, err := parseInsertRequest(buf)
			if err != nil {
				return
			}
			handleInsertRequest(sortedSet, request)

		} else if command == int('Q') {
			request, err := parseQueryRequest(buf)
			if err != nil {
				return
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

	listener, err := net.Listen("tcp", ":8080")

	clientId := 0

	clientIdMutex := &sync.Mutex{}

	if err != nil {
		slog.Warn(err.Error(), "msg", "error while listening on port 8080")
		return
	}

	for {
		conn, err := listener.Accept()
		if err != nil {
			slog.Warn(err.Error(), "msg", "error while accepting connection")
			return
		}

		clientIdMutex.Lock()
		clientId++
		currClientId := clientId
		clientIdMutex.Unlock()

		go handleClient(conn, currClientId)
	}
}
