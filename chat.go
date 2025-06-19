package main

import (
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

var mu sync.RWMutex

func getNickname(conn *websocket.Conn) (string, error) {
	messageType, buf, err := conn.ReadMessage()
	if err != nil {
		return "", err
	}

	switch messageType {
	case websocket.TextMessage:
		nick := string(buf)
		return nick, nil
	default:
		return "", fmt.Errorf("unexpected message type: %d", messageType)
	}
}

func broadcast(sender string, message []byte, clients Clients) {
	var err error

	mu.RLock()
	for nick, client := range clients {
		now := time.Now().Format(time.Stamp)
		err = client.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("%s %s: %s", now, sender, string(message))))

		if err != nil {
			slog.Error("Error writting message to", "nick", nick, "addr", client.RemoteAddr().String(), "error", err.Error())
		}

	}
	mu.RUnlock()
}

func appendNewClient(nick string, conn *websocket.Conn, clients Clients) error {
	mu.RLock()
	if _, ok := clients[nick]; ok {
		return errors.New("nickname already taken")
	}
	mu.RUnlock()
	mu.Lock()
	clients[nick] = conn
	mu.Unlock()
	return nil
}

func removeClient(nick string, clients Clients) error {
	if _, ok := clients[nick]; !ok {
		return errors.New("client not found")
	}

	delete(clients, nick)
	return nil
}
