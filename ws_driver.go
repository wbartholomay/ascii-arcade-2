package main

import (
	"fmt"

	"github.com/gorilla/websocket"
)
type WSDriver struct {
	open bool
	conn *websocket.Conn
}

func NewWS(url string) (WSDriver, error) {
	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		return WSDriver{}, fmt.Errorf("error dialing websocket: %w", err)
	}

	ws := WSDriver{
		open:           true,
		conn:           conn,
	}

	return ws, nil
}

func (driver *WSDriver) WriteToServer(msg PlayerMessage) error {
	return driver.conn.WriteJSON(msg)
}
