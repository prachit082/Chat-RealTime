package main

import "github.com/gorilla/websocket"

type Clients = map[string]*websocket.Conn
