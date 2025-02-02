package main

import (
	"bufio"
	"fmt"
	"log/slog"
	"net"
	"sync"
)

const (
	welcomeMessageFormat     = "Welcome to budget-chat! what do we call you?\n"
	userMessageFormat        = "[%s] %s"
	userJoinedMessageFormat  = "* %s has joined the chat.\n"
	userExitedMessageFormat  = "* %s has exited the chat.\n"
	onlineUsersMessageFormat = "* online users: "
)

type BudgetChat struct {
	clients      map[string]net.Conn
	clientsMutex *sync.RWMutex

	connMutexMap map[string]*sync.Mutex
}

func NewBudgetChat() *BudgetChat {

	return &BudgetChat{
		clients:      make(map[string]net.Conn),
		clientsMutex: &sync.RWMutex{},
		connMutexMap: make(map[string]*sync.Mutex),
	}
}
func validateName(name string) bool {

	if len(name) == 0 {
		return false
	}
	for _, ch := range name {
		if !((ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || (ch >= '0' && ch <= '9')) {
			return false
		}
	}
	return true
}

func (chat *BudgetChat) setName(conn net.Conn) (name string, err error) {

	reader := bufio.NewReader(conn)

	name, err = reader.ReadString('\n')
	if err != nil {
		slog.Error(err.Error(), "msg", "error while reading client's name")
		return "", err
	}
	name = name[:len(name)-1]
	chat.clientsMutex.Lock()

	if chat.clients[name] != nil {
		slog.Error("user with name " + name + " already exists.")
		return "", fmt.Errorf("user with name " + name + " already exists.")
	}
	if !validateName(name) {
		slog.Error("user with name " + name + " is not a valid name.")
		return "", fmt.Errorf("user with name " + name + " is not a valid name.")
	}

	chat.clients[name] = conn
	chat.connMutexMap[name] = &sync.Mutex{}

	chat.clientsMutex.Unlock()

	return name, nil
}

func (chat *BudgetChat) broadcastData(aboutUser string, data []byte) error {

	chat.clientsMutex.RLock()
	defer chat.clientsMutex.RUnlock()

	for username, conn := range chat.clients {

		mutex := chat.connMutexMap[username]
		mutex.Lock()
		if username != aboutUser {

			if _, err := conn.Write(data); err != nil {
				slog.Error(err.Error(), "sender-username", aboutUser, "username", username, "msg", "error while sending message "+string(data))
				return err
			}
		}
		mutex.Unlock()
	}

	return nil
}
func (chat *BudgetChat) broadcastMessage(senderUsername string, message string) error {

	message = fmt.Sprintf(userMessageFormat, senderUsername, message)

	return chat.broadcastData(senderUsername, []byte(message))

}

func (chat *BudgetChat) broadcastPresenceNotification(name string) error {

	usersOnlineNotification := onlineUsersMessageFormat

	userJoinedMessage := fmt.Sprintf(userJoinedMessageFormat, name)
	slog.Info("broadcasting message " + userJoinedMessage + " to all users")
	chat.clientsMutex.RLock()

	for username, _ := range chat.clients {
		if username != name {
			usersOnlineNotification = usersOnlineNotification + username + ", "
		}
	}
	usersOnlineNotification = usersOnlineNotification + "\n"
	chat.clientsMutex.RUnlock()

	if err := chat.broadcastData(name, []byte(userJoinedMessage)); err != nil {
		return err
	}
	slog.Info("sending message " + usersOnlineNotification + " to user " + name)
	if err := chat.sendData(name, []byte(usersOnlineNotification)); err != nil {
		return err
	}
	return nil
}

func (chat *BudgetChat) sendData(name string, data []byte) error {

	chat.clientsMutex.RLock()

	conn := chat.clients[name]

	mutex := chat.connMutexMap[name]

	chat.clientsMutex.RUnlock()

	mutex.Lock()
	defer mutex.Unlock()

	if _, err := conn.Write(data); err != nil {
		return err
	}

	return nil
}
func (chat *BudgetChat) handleUserExit(name string) error {

	chat.clientsMutex.Lock()

	delete(chat.clients, name)
	delete(chat.connMutexMap, name)

	chat.clientsMutex.Unlock()

	userExitMessage := fmt.Sprintf(userExitedMessageFormat, name)
	slog.Info(userExitMessage)
	if err := chat.broadcastData(name, []byte(userExitMessage)); err != nil {
		return err
	}

	return nil
}
func (chat *BudgetChat) handleClient(conn net.Conn) {

	defer conn.Close()

	_, err := conn.Write([]byte(welcomeMessageFormat))
	if err != nil {
		slog.Error(err.Error(), "msg", "error while writing welcome message to client.")
		return
	} else {
		slog.Info("sent welcome message.")
	}

	name, err := chat.setName(conn)

	if err != nil {
		_, _ = conn.Write([]byte("invalid name \n"))
		return
	} else {
		slog.Info(name + " has connected to budget chat.")
	}

	if err := chat.broadcastPresenceNotification(name); err != nil {
		slog.Error(err.Error(), "msg", "error while writing welcome message to client")
		return
	}

	reader := bufio.NewReader(conn)

	for {

		message, err := reader.ReadString('\n')
		if err != nil {
			slog.Error(err.Error(), "username", name, "msg", "error while reading message from client.")
			_ = chat.handleUserExit(name)
			return
		}

		if err = chat.broadcastMessage(name, message); err != nil {
			return
		}
	}

}
func main() {

	chat := NewBudgetChat()

	listener, err := net.Listen("tcp", "0.0.0.0:8080")

	if err != nil {
		slog.Error(err.Error(), "msg", "error while creating listener")
		return
	}
	for {
		conn, err := listener.Accept()
		if err != nil {
			slog.Error(err.Error(), "msg", "error while accepting connection")
			return
		}
		go chat.handleClient(conn)
	}
}
