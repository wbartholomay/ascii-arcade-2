package messages

import "github.com/wbarthol/ascii-arcade-2/internal/game"

type GameResult int

const (
	GameResultPlayerWin GameResult = iota
	GameResultPlayerLose
	GameResultDraw
)

type GameWrapper struct {
	Type      game.GameType       `json:"type"`
	TicTacToe *game.TicTacToeGame `json:"tic_tac_toe,omitempty"`
	Checkers  *game.CheckersGame  `json:"checkers,omitempty"`
}

func (wrapper *GameWrapper) GetGame() game.Game {
	switch wrapper.Type {
	case game.GameTypeTicTacToe:
		return wrapper.TicTacToe
	case game.GameTypeCheckers:
		return wrapper.Checkers
	default:
		return nil
	}
}

func NewGameWrapper(g game.Game) GameWrapper {
	if g == nil {
		return GameWrapper{}
	}

	gameWrapper := GameWrapper{Type: g.GetGameType()}

	switch g.GetGameType() {
	case game.GameTypeTicTacToe:
		gameWrapper.TicTacToe = g.(*game.TicTacToeGame)
	case game.GameTypeCheckers:
		gameWrapper.Checkers = g.(*game.CheckersGame)
	}

	return gameWrapper
}

type ServerMessageType int

const (
	ServerRoomJoined ServerMessageType = iota
	ServerEnteredGameSelection
	ServerGameStarted
	ServerTurnResult
	ServerRoomDisconnected
	ServerGameFinished
	ServerRoomClosed
	ServerRoomUnavailable
	ServerError
)

func (sType ServerMessageType) String() string {
	switch sType {
	case ServerRoomJoined:
		return "Room Joined"
	case ServerEnteredGameSelection:
		return "Entered Game Selection"
	case ServerGameStarted:
		return "Game Started"
	case ServerTurnResult:
		return "Turn Result"
	case ServerRoomDisconnected:
		return "Room Disconnected"
	case ServerGameFinished:
		return "Game Finished"
	case ServerRoomClosed:
		return "Room Closed"
	case ServerRoomUnavailable:
		return "Room Unavailable"
	case ServerError:
		return "Error"
	default:
		return "Unknown"
	}
}

type ServerMessage struct {
	Type              ServerMessageType `json:"type"`
	PlayerNumber      int               `json:"player_number"`
	PlayerTurn        int               `json:"player_turn"`
	Game              GameWrapper       `json:"game"`
	GameResult        GameResult        `json:"game_result"`
	QuittingPlayerNum int               `json:"quitting_player_num"`
	ErrorMessage      string            `json:"error_message"`
}

type GameTurnWrapper struct {
	GameType      game.GameType      `json:"game_type"`
	TicTacToeTurn game.TicTacToeTurn `json:"tictactoe_turn"`
	CheckersTurn  game.CheckersTurn  `json:"checkers_turn"`
}

func (wrapper *GameTurnWrapper) GetGameTurn() game.GameTurn {
	switch wrapper.GameType {
	case game.GameTypeTicTacToe:
		return wrapper.TicTacToeTurn
	case game.GameTypeCheckers:
		return wrapper.CheckersTurn
	default:
		return nil
	}
}

func NewGameTurnWrapper(g game.GameTurn) GameTurnWrapper {
	if g == nil {
		return GameTurnWrapper{}
	}

	gameWrapper := GameTurnWrapper{GameType: g.GetGameType()}

	switch g.GetGameType() {
	case game.GameTypeTicTacToe:
		gameWrapper.TicTacToeTurn = g.(game.TicTacToeTurn)
	case game.GameTypeCheckers:
		gameWrapper.CheckersTurn = g.(game.CheckersTurn)
	}

	return gameWrapper
}

type ClientMessageType int

const (
	ClientJoinRoom ClientMessageType = iota
	ClientSelectGameType
	ClientSendTurn
	ClientQuitRoom
	ClientPlayAgain
)

type ClientMessage struct {
	Type       ClientMessageType `json:"type"`
	RoomCode   string            `json:"room_code"`
	GameType   game.GameType     `json:"game_type"`
	TurnAction GameTurnWrapper   `json:"turn_action"`
}
