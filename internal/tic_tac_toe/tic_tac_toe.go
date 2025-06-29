package tictactoe

import "fmt"

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
	Board      [3][3]TicTacToeSquare `json:"board"`
	IsXTurn    bool                  `json:"is_x_turn"`
	GameStatus GameStatus            `json:"game_status"`
}

func NewTicTacToeGame() TicTacToeGame {
	var board [3][3]TicTacToeSquare

	return TicTacToeGame{
		Board:      board,
		IsXTurn:    true,
		GameStatus: GameOngoing,
	}
}

func (game TicTacToeGame) DisplayBoard() {
	fmt.Println("\n   0   1   2")
	for i, row := range game.Board {
		fmt.Printf("%d ", i)
		for j, square := range row {
			var symbol string
			switch square {
			case SquareEmpty:
				symbol = " "
			case SquareX:
				symbol = "X"
			case SquareO:
				symbol = "O"
			}

			if j < len(row)-1 {
				fmt.Printf(" %s |", symbol)
			} else {
				fmt.Printf(" %s ", symbol)
			}
		}
		fmt.Println()

		if i < len(game.Board)-1 {
			fmt.Println("  ---|---|---")
		}
	}
	fmt.Println()
}
