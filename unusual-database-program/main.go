package main

import (
	"fmt"
	"log"
	"log/slog"
	"net"
	"sync"
)

type Database struct {
	conn          net.UDPConn
	keyValueStore map[string]string
	mutex         *sync.RWMutex
}

func NewDatabase(addr string) (*Database, error) {

	udpAddr, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		return nil, err
	}

	conn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		return nil, err
	}
	return &Database{
		conn:          *conn,
		keyValueStore: make(map[string]string),
		mutex:         &sync.RWMutex{},
	}, nil
}

func (db *Database) handleInsertRequest(key string, value string) {

	db.mutex.Lock()
	db.keyValueStore[key] = value
	log.Println(db.keyValueStore)
	db.mutex.Unlock()
}

func (db *Database) handleRetrieveRequest(key string) (value string) {

	db.mutex.RLock()
	value = db.keyValueStore[key]
	db.mutex.RUnlock()

	return value
}

func parseInsertRequest(request string) (key string, value string) {

	key = ""
	value = ""

	equalToFound := false

	for _, ch := range request {

		if ch == '=' {
			if equalToFound {
				value = value + string(ch)
			} else {
				equalToFound = true
			}

		} else {
			if equalToFound {
				value = value + string(ch)
			} else {
				key = key + string(ch)
			}
		}
	}

	return key, value
}

func (db *Database) HandleRequests() {

	for {

		buf := make([]byte, 4096)
		n, addr, err := db.conn.ReadFromUDP(buf)

		if err != nil {
			slog.Error(err.Error(), "msg", "error while reading UDP packet.")
		}
		request := string(buf[:n])

		equalToFound := false
		for _, ch := range request {

			if ch == '=' {
				equalToFound = true
				break
			}
		}
		if equalToFound {

			key, value := parseInsertRequest(request)
			slog.Info(fmt.Sprintf("received insert request => key = %q value = %q", key, value))
			db.handleInsertRequest(key, value)

		} else {
			slog.Info(fmt.Sprintf("received query request => key =%q!", request))
			value := db.handleRetrieveRequest(request)
			response := request + "=" + value
			if value == "" {
				slog.Info(fmt.Sprintf("didnt find value for key =%s!", request))
			} else {
				slog.Info(fmt.Sprintf("received value = %s for key =%s!", value, request))
			}
			if _, err := db.conn.WriteToUDP([]byte(response), addr); err != nil {
				slog.Error("error while writing value back to client.")
				return
			}
		}

	}
}
func main() {

	db, err := NewDatabase("0.0.0.0:8080")

	if err != nil {
		slog.Error(err.Error(), "msg", "error while assigning port 8080 to listen for UDP messages.")
		return
	}

	db.HandleRequests()

}
