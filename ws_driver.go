package main

import (
	"fmt"
	"log"

	"github.com/gorilla/websocket"
)
type WSDriver struct {
	open bool
	conn *websocket.Conn
	//notifier Notifier

	serverToPlayer chan ServerMessage
}

func StartWS(url string) (*WSDriver, error) {
	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		return nil, fmt.Errorf("error dialing websocket: %w", err)
	}


	serverToPlayer := make(chan ServerMessage)

	ws := WSDriver{
		open:           true,
		conn:           conn,
		serverToPlayer: serverToPlayer,
	}

	go ws.ReadPump()

	return &ws, nil
}

func (wsDriver *WSDriver) ReadPump() {
	defer func() {
		wsDriver.conn.Close()
		//TODO communicate to client that server has closed
	}()

	for {
		var msg ServerMessage
		err := wsDriver.conn.ReadJSON(&msg)
		if err != nil {
			log.Printf("An error has occurred while reading from the server, shutting down: %v\n", err)
			return
		}

		wsDriver.serverToPlayer <- msg
	}
}

func (wsDriver *WSDriver) handleMsg(msg ServerMessage) {

}

func (driver *WSDriver) WriteToServer(msg PlayerMessage) error {
	return driver.conn.WriteJSON(msg)
}
