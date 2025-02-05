package main

import (
	"bufio"
	"fmt"
	"log/slog"
	"net"
	"strings"
)

const (
	tonyAddress = "7YWHMfk9JZe0LM0g1ZauHuiSxhI"
)

func isBGAddress(word string) bool {

	if len(word) < 26 || len(word) > 35 {
		//slog.Info("BG address " + word + " was too short/long")
		return false
	}

	if word[0] != '7' {
		//slog.Info("BG address " + word + " didnt start with 7")
		return false
	}

	for _, char := range word {

		if !((char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') || (char >= '0' && char <= '9')) {
			//slog.Info("BG address " + word + " contains " + string(char) + " which is not alphanumeric")
			return false
		}
	}
	return true
}
func searchAndReplaceBGAddress(message string) (newMessage string) {

	message = message[:len(message)-1]
	words := strings.Split(message, " ")

	for i, word := range words {

		if isBGAddress(word) {
			words[i] = tonyAddress
		}
	}

	newMessage = strings.Join(words, " ")
	newMessage = newMessage + "\n"
	return newMessage

}

func forward(source net.Conn, destination net.Conn) {

	sourceReader := bufio.NewReader(source)

	for {
		message, err := sourceReader.ReadString('\n')
		slog.Info(fmt.Sprintf("received message %q source", message))
		//slog.Info("received message " + response + " from upstream server.")
		if err != nil {
			slog.Error(err.Error(), "msg", "error while reading response from upstream server.")
			return
		}

		message = searchAndReplaceBGAddress(message)
		slog.Info(fmt.Sprintf("sending message %q to client.", message))
		if _, err := destination.Write([]byte(message)); err != nil {
			slog.Error(err.Error(), "msg", "error while sending response back to client.")
			return

		}
	}
}
func handleClient(clientConn net.Conn) {

	defer clientConn.Close()

	upstreamConn, err := net.Dial("tcp", "chat.protohackers.com:16963")

	if err != nil {
		slog.Error(err.Error(), "msg", "error while establishing upstream connection with budget-chat server.")
		return
	}
	defer upstreamConn.Close()

	upstreamConnReader := bufio.NewReader(upstreamConn)

	welcomeMessage, err := upstreamConnReader.ReadString('\n')
	if err != nil {
		slog.Error(err.Error(), "msg", "error while reading welcome message from upstream server.")
		return
	}
	slog.Info("received welcome message " + welcomeMessage + " from upstream server.")
	if _, err := clientConn.Write([]byte(welcomeMessage)); err != nil {
		slog.Error(err.Error(), "msg", "error while sending welcome message to client.")
		return
	}

	go forward(clientConn, upstreamConn)
	forward(upstreamConn, clientConn)

}
func main() {

	listener, err := net.Listen("tcp", "0.0.0.0:8080")

	//slog.Info(searchAndReplaceBGAddress("[RedBob926] Send refunds to 7YWHMfk9JZe0LM0g1ZauHuiSxhI please."))
	if err != nil {
		slog.Error(err.Error(), "msg", "error while listening on port 8080")
		return
	}

	for {

		conn, err := listener.Accept()
		if err != nil {
			slog.Error(err.Error(), "msg", "error while accepting connection")
			return
		}

		go handleClient(conn)
	}

}
