package game

type GameType int

const (
	GameTypeTicTacToe GameType = iota
	GameTypeCheckers
)

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
	//TODO improve this from empty interface
	ValidateMove(GameTurn, int) (bool, string)
	ExecuteTurn(GameTurn, int)
	DisplayBoard(int) string
}

type GameTurn interface {
	isGameTurn()
}

func NewGame(gameType GameType) Game {
	switch gameType{
	case GameTypeTicTacToe:
		return NewTicTacToeGame()
	case GameTypeCheckers:
		return NewCheckersGame()
	default:
		return nil
	}
}