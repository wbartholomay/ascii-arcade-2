package messages

import "github.com/wbarthol/ascii-arcade-2/internal/tictactoe"

type ServerMessageType int

const (
	ServerRoomJoined ServerMessageType = iota
	ServerGameStarted
	ServerTurnResult
)

type ServerMessage struct {
	Type         ServerMessageType       `json:"type"`
	PlayerNumber int                     `json:"player_number"`
	PlayerTurn   int                     `json:"player_turn"`
	Game         tictactoe.TicTacToeGame `json:"game"`
}

type ClientMessageType int

const (
	ClientJoinRoom ClientMessageType = iota
	ClientSendTurn
	ClientQuitRoom
)

type ClientMessage struct {
	Type       ClientMessageType       `json:"type"`
	RoomCode   string                  `json:"room_code"`
	TurnAction tictactoe.TicTacToeTurn `json:"turn_action"`
}
