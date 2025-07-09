package game

import "github.com/wbarthol/ascii-arcade-2/internal/vector"

type GameType int

const (
	GameTypeTicTacToe GameType = iota
	GameTypeCheckers
)

func GetGameTypes() []GameType {
	return []GameType{GameTypeTicTacToe}
}

func (gt GameType) String() string {
	switch gt {
	case GameTypeTicTacToe:
		return "TicTacToe"
	case GameTypeCheckers:
		return "Checkers"
	default:
		return "Unknown"
	}
}

type GameStatus int

const (
	GameStatusOngoing GameStatus = iota
	GameStatusPlayer1Win
	GameStatusPlayer2Win
	GameStatusDraw
)

type Game interface {
	GetGameType() GameType
	GetGameStatus() GameStatus
	OverrideGameStatus(GameStatus)
	GetGameInstructions() string
	ValidateMove(GameTurn, int) (bool, string)
	ExecuteTurn(GameTurn, int) string
	DisplayBoard(vector.Vector, int) string
}

type GameTurn interface {
	GetGameType() GameType
}

func NewGame(gameType GameType) Game {
	switch gameType {
	case GameTypeTicTacToe:
		return NewTicTacToeGame()
	case GameTypeCheckers:
		return NewCheckersGame()
	default:
		return nil
	}
}
