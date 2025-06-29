package main

import tictactoe "github.com/wbarthol/ascii-arcade-2/internal/tic_tac_toe"

type ServerMessageType int

const (
	ServerRoomJoined ServerMessageType = iota
	ServerGameStarted
	ServerTurnResult
)

type ServerMessage interface {
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

type PlayerMessageType int

const (
	PlayerJoinRoom PlayerMessageType = iota
	PlayerSentTurn
	PlayerQuitRoom
)

type PlayerMessage interface {
	GetType() PlayerMessageType
}

type PlayerMessageJoinRoom struct {
	Type PlayerMessageType `json:"type"`
	Code string            `json:"code"`
}

func (msg PlayerMessageJoinRoom) GetType() PlayerMessageType {
	return msg.Type
}

type PlayerMessageSendTurn struct {
	Type PlayerMessageType `json:"type"`
	//TODO
}

func (msg PlayerMessageSendTurn) GetType() PlayerMessageType {
	return msg.Type
}

type PlayerMessageQuitRoom struct {
	Type PlayerMessageType `json:"type"`
}

func (msg PlayerMessageQuitRoom) GetType() PlayerMessageType {
	return msg.Type
}
