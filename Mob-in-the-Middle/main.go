package main

import (
	"bufio"
	"log/slog"
	"net"
	"strings"
)

const (
	tonyAddress = "7YWHMfk9JZe0LM0g1ZauHuiSxhI"
)

func isBGAddress(word string) bool {

	if len(word) < 26 || len(word) > 35 {
		return false
	}
	if word[0] != '7' {
		return false
	}

	for _, char := range word {

		if !((char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') || (char >= '0' && char <= '9')) {
			return false
		}
	}
	return true
}
func searchAndReplaceBGAddress(message string) (newMessage string) {

	numLeadingWhitespaces := 0
	numTrailingWhitespaces := 0

	for _, char := range message {
		if char == ' ' {
			numLeadingWhitespaces++
		} else {
			break
		}
	}

	for ind := range len(message) {
		if message[len(message)-ind-1] == ' ' {
			numTrailingWhitespaces++
		} else {
			break
		}
	}
	words := strings.Split(message, " ")

	firstWord := words[0]
	lastWord := words[len(words)-1]

	bogusCoinAddressDetected := false
	if isBGAddress(firstWord) {
		words[0] = tonyAddress
		bogusCoinAddressDetected = true
	}
	if isBGAddress(lastWord) {
		words[len(words)-1] = tonyAddress
		bogusCoinAddressDetected = true
	}

	if !bogusCoinAddressDetected {
		return message + "\n"
	}

	newMessage = strings.Join(words, " ")
	newMessage = strings.Repeat(" ", numLeadingWhitespaces) + newMessage + strings.Repeat(" ", numTrailingWhitespaces) + "\n"

	return newMessage

}

func listenToClientWriteToUpstream(clientConn net.Conn, upstreamConn net.Conn) {

	clientConnReader := bufio.NewReader(clientConn)

	for {
		clientMessage, err := clientConnReader.ReadString('\n')
		if err != nil {
			slog.Error(err.Error(), "msg", "error while reading message from client.")
			return
		}

		slog.Info("received message " + clientMessage + " from client.")
		upstreamMessage := searchAndReplaceBGAddress(clientMessage[:len(clientMessage)-1])

		slog.Info("sending message " + upstreamMessage + " to upstream server.")
		if _, err := upstreamConn.Write([]byte(upstreamMessage)); err != nil {
			slog.Error(err.Error(), "msg", "error while sending message to upstream server.")
			return
		}
	}
}

func listenToUpstreamWriteToClient(clientConn net.Conn, upstreamConn net.Conn) {

	upstreamConnReader := bufio.NewReader(upstreamConn)

	for {
		response, err := upstreamConnReader.ReadString('\n')
		slog.Info("received message " + response + " from upstream server.")
		if err != nil {
			slog.Error(err.Error(), "msg", "error while reading response from upstream server.")
			return
		}

		if _, err := clientConn.Write([]byte(response)); err != nil {
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

	go listenToClientWriteToUpstream(clientConn, upstreamConn)
	go listenToUpstreamWriteToClient(clientConn, upstreamConn)
}
func main() {

	listener, err := net.Listen("tcp", "0.0.0.0:8080")

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
