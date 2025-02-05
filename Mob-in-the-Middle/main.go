package main

import (
	"bufio"
	"log/slog"
	"net"
	"strings"
	"sync"
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

	//numLeadingWhitespaces := 0
	//numTrailingWhitespaces := 0
	//
	//for _, char := range message {
	//	if char == ' ' {
	//		numLeadingWhitespaces++
	//	} else {
	//		break
	//	}
	//}
	//
	//for ind := range len(message) {
	//	if message[len(message)-ind-1] == ' ' {
	//		numTrailingWhitespaces++
	//	} else {
	//		break
	//	}
	//}
	words := strings.Split(message, " ")

	if strings.Contains(message, "*") {
		slog.Info("user is joining or leaving, message => " + message)
	}
	for i, word := range words {

		if isBGAddress(word) {
			words[i] = tonyAddress
		}
	}

	//for i := range len(words) {
	//
	//	if isBGAddress(words[len(words) - i - 1]) {
	//		bogusCoinAddressDetected = true
	//		words[len(words) - i - 1] = tonyAddress
	//	}
	//}

	newMessage = strings.Join(words, " ")

	return newMessage

}

func listenToClientWriteToUpstream(clientConn net.Conn, upstreamConn net.Conn, wg *sync.WaitGroup) {

	defer wg.Done()
	clientConnReader := bufio.NewReader(clientConn)

	for {
		clientMessage, err := clientConnReader.ReadString('\n')
		if err != nil {
			//slog.Error(err.Error(), "msg", "error while reading message from client.")
			return
		}

		//slog.Info("received message " + clientMessage + " from client.")

		upstreamMessage := searchAndReplaceBGAddress(clientMessage)

		//slog.Info("sending message " + upstreamMessage + " to upstream server.")
		if _, err := upstreamConn.Write([]byte(upstreamMessage)); err != nil {
			slog.Error(err.Error(), "msg", "error while sending message to upstream server.")
			return
		}
	}
}

func listenToUpstreamWriteToClient(clientConn net.Conn, upstreamConn net.Conn, wg *sync.WaitGroup) {

	defer wg.Done()

	upstreamConnReader := bufio.NewReader(upstreamConn)

	for {
		response, err := upstreamConnReader.ReadString('\n')
		//slog.Info("received message " + response + " from upstream server.")
		if err != nil {
			//slog.Error(err.Error(), "msg", "error while reading response from upstream server.")
			return
		}

		clientMessage := searchAndReplaceBGAddress(response)
		if _, err := clientConn.Write([]byte(clientMessage)); err != nil {
			//slog.Error(err.Error(), "msg", "error while sending response back to client.")
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
	//slog.Info("received welcome message " + welcomeMessage + " from upstream server.")
	if _, err := clientConn.Write([]byte(welcomeMessage)); err != nil {
		slog.Error(err.Error(), "msg", "error while sending welcome message to client.")
		return
	}

	wg := &sync.WaitGroup{}
	wg.Add(1)
	wg.Add(1)
	go listenToClientWriteToUpstream(clientConn, upstreamConn, wg)
	go listenToUpstreamWriteToClient(clientConn, upstreamConn, wg)
	wg.Wait()
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
