package main

import (
	"fmt"
	"net/http"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
        return true
    },
}

type RoomRequest struct {
	code  string
	chans RoomChans
}

type Hub struct {
	roomRequests chan RoomRequest
}

func NewHub() *Hub {
	return &Hub{make(chan RoomRequest)}
}

func (h *Hub) Run() {
	rooms := make(map[string]*Room)
	closeReq := make(chan string)

	for {
		select {
		case msg := <-h.roomRequests:
			room, ok := rooms[msg.code]
			if !ok {
				// room = NewRoom(msg.code, closeReq)
				fmt.Println("Creating new room")
				room = NewRoom(msg.code)
				go room.Run()
				rooms[msg.code] = room
			}
			fmt.Println("player joining room")
			room.requests <- msg
			fmt.Println("player joined room")
		case code := <-closeReq:
			room, ok := rooms[code]
			if !ok {
				fmt.Println("illegal")
				break
			}
			close(room.requests)
			delete(rooms, code)
		}
	}
}

func (h *Hub) ServeWs(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println("hub: conn error")
		return
	}

	player := NewPlayer(conn, h.roomRequests)

	fmt.Println("start player goroutine")
	go player.Run()
}
