package tictactoe

import (
	"fmt"

	"github.com/wbarthol/ascii-arcade-2/internal/vector"
)

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

func (game *TicTacToeGame) ValidateMove(coords vector.Vector) bool {
	rowInBounds := coords.Y >= 0 && coords.Y <= 2
	colInBounds := coords.X >= 0 && coords.Y <= 2
	if !rowInBounds || !colInBounds {
		return false
	}

	return game.Board[coords.Y][coords.X] == SquareEmpty
}

// ExecuteTurn - Takes coordinates and a player number, executes turn.
func (game *TicTacToeGame) ExecuteTurn(turn TicTacToeTurn, playerNum int) {
	coords := turn.Coords
	playerSquare := SquareX
	if playerNum == 2 {
		playerSquare = SquareO
	}

	game.Board[coords.Y][coords.X] = playerSquare
	//TODO check for game over
}

func (game *TicTacToeGame) DisplayBoard() string {
	var result string
	result += "\n   0   1   2\n"

	for i, row := range game.Board {
		result += fmt.Sprintf("%d ", i)
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
				result += fmt.Sprintf(" %s |", symbol)
			} else {
				result += fmt.Sprintf(" %s ", symbol)
			}
		}
		result += "\n"

		if i < len(game.Board)-1 {
			result += "  ---|---|---\n"
		}
	}
	result += "\n"
	return result
}

type TicTacToeTurn struct {
	Coords vector.Vector `json:"coords"`
}
