package tictactoe

type TicTacToeSquare int

const (
	SquareEmpty TicTacToeSquare = iota
	SquareX
	SquareO
)

type GameStatus int

const (
	GameOngoing GameStatus = iota
	XWon
	OWon
)

type TicTacToeGame struct {
	Board [3][3]TicTacToeSquare `json:"board"`
	IsXTurn bool `json:"is_x_turn"`
	GameStatus GameStatus `json:"game_status"`
}

func NewTicTacToeGame() TicTacToeGame {
	var board [3][3]TicTacToeSquare

	return TicTacToeGame{
		Board: board,
		IsXTurn: true,
		GameStatus: GameOngoing,
	}
}