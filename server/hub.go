package main

import (
	"fmt"
	"log"
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
				log.Printf("Creating new room with code: %v", msg.code)
				room = NewRoom(msg.code, closeReq)
				go room.Run()
				rooms[msg.code] = room
			}
			room.requests <- msg
		case code := <-closeReq:
			room, ok := rooms[code]
			if !ok {
				log.Println("Error closing room - room closed closeReq channel.")
				break
			}
			log.Printf("Closing room %v\n", room.code)
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

	log.Println("New connection established, creating player.")
	player := NewPlayer(conn, h.roomRequests)

	go player.Run()
}
