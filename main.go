package main

import (
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

const readtimeout = time.Duration(time.Second * 3)

var connectedClients = make(Clients)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  100,
	WriteBufferSize: 100,
}

func home(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)

		buf, _ := os.ReadFile("public/chat.html")
		w.Write(buf)
	} else {
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func chat(w http.ResponseWriter, r *http.Request) {
	slog.Info("Attempt to upgrade", "addr", r.RemoteAddr)

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		var errorMessage string = fmt.Sprintf("Error upgrading http conn to websocket: %s", err.Error())
		slog.Error(errorMessage)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(errorMessage))
		return
	}
	defer conn.Close()

	slog.Info("Getting nick from user", "addr", conn.RemoteAddr().String())
	nick, err := getNickname(conn)

	if err != nil {
		slog.Error("Error getting nickname", "addr", conn.RemoteAddr().String(), "error", err.Error())
		conn.WriteMessage(websocket.TextMessage, []byte(err.Error()))
		return
	}

	if err := appendNewClient(nick, conn, connectedClients); err != nil {
		conn.WriteMessage(websocket.TextMessage, []byte(err.Error()))
	}

	// A user entered the room
	broadcast(nick, []byte("<came into the room>"), connectedClients)

	for {
		messageType, buf, err := conn.ReadMessage()

		if err != nil {
			if websocket.IsCloseError(err, websocket.CloseGoingAway) {
				removeClient(nick, connectedClients)
				broadcast(nick, []byte("<quit>"), connectedClients)
				break
			}

			slog.Error("Error reading client message", "addr", conn.RemoteAddr().String(), "error", err.Error())
			conn.WriteMessage(websocket.TextMessage, []byte("Couldn't send message. Please, try again later."))
			break
		}

		if messageType != websocket.TextMessage {
			slog.Error("Unexpected message type", "type", messageType, "addr", conn.RemoteAddr().String())
		} else {
			broadcast(nick, buf, connectedClients)
		}
	}
}

func main() {
	router := mux.NewRouter()
	router.HandleFunc("/", home)
	router.HandleFunc("/chat", chat)

	server := &http.Server{
		Addr:        "localhost:8000",
		Handler:     router,
		ReadTimeout: readtimeout,
	}

	if err := server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
