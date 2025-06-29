package main

import (
	"fmt"

	"github.com/gorilla/websocket"
	"github.com/wbarthol/ascii-arcade-2/internal/messages"
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

func (driver *WSDriver) WriteToServer(msg messages.ClientMessage) error {
	return driver.conn.WriteJSON(msg)
}
