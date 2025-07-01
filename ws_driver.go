package main

import (
	"fmt"

	"github.com/gorilla/websocket"
	"github.com/wbarthol/ascii-arcade-2/internal/messages"
)

type WSDriver struct {
	wsOpen       bool
	conn         *websocket.Conn
	session      *Session
	driverToSession chan messages.ServerMessage
}

func NewWS(url string, session *Session) (*WSDriver, error) {
	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		return &WSDriver{}, fmt.Errorf("error dialing websocket: %w", err)
	}

	ws := WSDriver{
		wsOpen:       true,
		conn:         conn,
		session:      session,
		driverToSession:make(chan messages.ServerMessage),
	}

	return &ws, nil
}

func (driver *WSDriver) WriteToServer(msg messages.ClientMessage) error {
	return driver.conn.WriteJSON(msg)
}

func (driver *WSDriver) Run() {
	//TODO - Find an effective way to close the connection without having the driver failing the read from server
	//should maybe reverse which part is in the go func
	defer func() {
		fmt.Println("Closing server connection.")
	}()

	for {
		var msg messages.ServerMessage
		err := driver.conn.ReadJSON(&msg)
		// log.Printf("Message from server: %v\n", msg)
		if !driver.wsOpen {
			return
		}
		if err != nil {
			fmt.Printf("error reading message from server %v\n", err)
			driver.CloseWS()
			return
		}
		driver.driverToSession <- msg
	}
}

func (driver *WSDriver) CloseWS() {
	driver.wsOpen = false
	driver.conn.Close()
	close(driver.driverToSession)
}
