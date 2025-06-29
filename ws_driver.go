package main

import (
	"log"

	"github.com/gorilla/websocket"
)

type WSDriver struct {
	open bool
	conn *websocket.Conn
	//notifier Notifier

	serverToPlayer chan ServerMessage
	playerToServer chan PlayerMessage
}

func StartWS(conn *websocket.Conn) *WSDriver{
	serverToPlayer := make (chan ServerMessage)
	playerToServer := make (chan PlayerMessage)

	ws := WSDriver{
		open: true,
		conn: conn,
		serverToPlayer: serverToPlayer,
		playerToServer: playerToServer,
	}

	go ws.Run()

	return &ws
}

func (wsDriver *WSDriver) Run() {
	for {
		select {
		case serverMsg := <- wsDriver.serverToPlayer:
			//TODO notify notifier
			//notifier.Notify()...?
			//wsDriver.state.handleMsg()...?


		case playerMsg := <- wsDriver.playerToServer:
			//TODO write to server
		}
	}
}

func (wsDriver *WSDriver) ReadPump() {
	defer func(){
		wsDriver.conn.Close()
		//TODO communicate to client that server has closed
	}()

	for {
		msg := ServerMessage{}
		err := wsDriver.conn.ReadJSON(&msg)
		if err != nil {
			log.Printf("An error has occurred while reading from the server, shutting down: %v\n", err)
			return
		}

		wsDriver.serverToPlayer <- msg
	}
}

// func (driver *WSDriver) WriteToServer(msg ServerMessage) error {
// 	return driver.conn.WriteJSON(serverMsg)
// }

