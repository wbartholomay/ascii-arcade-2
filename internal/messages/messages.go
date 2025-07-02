package messages

import "github.com/wbarthol/ascii-arcade-2/internal/game"

type GameResult int

const (
	GameResultPlayerWin GameResult = iota
	GameResultPlayerLose
	GameResultDraw
)

type ServerMessageType int

const (
	ServerRoomJoined ServerMessageType = iota
	ServerEnteredGameSelection
	ServerGameStarted
	ServerTurnResult
	ServerRoomDisconnected
	ServerGameFinished
	ServerRoomUnavailable
)

type ServerMessage struct {
	Type         ServerMessageType `json:"type"`
	PlayerNumber int               `json:"player_number"`
	PlayerTurn   int               `json:"player_turn"`
	Game         game.Game         `json:"game"`
	GameResult   GameResult        `json:"game_result"`
	Message      string            `json:"message"`
}

type ClientMessageType int

const (
	ClientJoinRoom ClientMessageType = iota
	ClientSelectGameType
	ClientSendTurn
	ClientQuitRoom
)

type ClientMessage struct {
	Type       ClientMessageType `json:"type"`
	RoomCode   string            `json:"room_code"`
	GameType   game.GameType     `json:"game_type"`
	TurnAction game.GameTurn     `json:"turn_action"`
}
