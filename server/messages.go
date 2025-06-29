package main

import tictactoe "github.com/wbarthol/ascii-arcade-2/internal/tic_tac_toe"

type ServerMessageType int

const (
	ServerRoomJoined ServerMessageType = iota
	ServerGameStarted
	ServerTurnResult
)

type ServerMessage interface {
	//using empty function to define server message types
	GetType() ServerMessageType
}

type ServerMessageRoomJoined struct {
	Type ServerMessageType `json:"type"`
}

func (m ServerMessageRoomJoined) GetType() ServerMessageType {
	return m.Type
}

type ServerMessageGameStarted struct {
	Type ServerMessageType       `json:"type"`
	Game tictactoe.TicTacToeGame `json:"game"`
}

func (m ServerMessageGameStarted) GetType() ServerMessageType {
	return m.Type
}

type ServerMessageTurnResult struct {
	Type ServerMessageType       `json:"type"`
	Game tictactoe.TicTacToeGame `json:"game"`
}

func (m ServerMessageTurnResult) GetType() ServerMessageType {
	return m.Type
}

type ClientMessageType int

const (
	ClientJoinRoom ClientMessageType = iota
	ClientSendTurn
	ClientQuitRoom
)

type ClientMessage struct {
	Type     ClientMessageType `json:"type"`
	RoomCode string            `json:"room_code"`
	Action   TurnAction        `json:"turn_action"`
}
type TurnAction struct {
}
