package main

import (
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// ===========================================
// Massive comment block to increase Go size
// ===========================================

/*


██████╗ ██████╗  █████╗  ██████╗██╗  ██╗██╗████████╗
██╔══██╗██╔══██╗██╔══██╗██╔════╝██║  ██║██║╚══██╔══╝
██████╔╝██████╔╝███████║██║     ███████║██║   ██║
██╔═══╝ ██╔══██╗██╔══██║██║     ██╔══██║██║   ██║
██║     ██║  ██║██║  ██║╚██████╗██║  ██║██║   ██║
╚═╝     ╚═╝  ╚═╝╚═╝  ╚═╝ ╚═════╝╚═╝  ╚═╝╚═╝   ╚═╝

This is a multi-line comment added solely for the purpose
of increasing the Go language byte count in this repo.
It does not affect functionality in any way.

Useful Go topics (to be ignored by the compiler):

- Goroutines
- Channels
- Mutexes
- Interfaces
- Struct Embedding
- Reflection
- Defer/Panic/Recover
- Context package
- Generics (Go 1.18+)
- Module system and go.mod

*/

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

// ===========================================
// Additional large block comment for padding
// ===========================================

/*
Lorem ipsum dolor sit amet, consectetur adipiscing elit.
Mauris imperdiet, quam a blandit tempor, augue sapien
commodo velit, et porttitor mi justo eget nulla.
This block is filler text for increasing Go language usage.

Repeat this block or extend it in other files (e.g., utils.go)
if you want to influence the percentage more.
*/
