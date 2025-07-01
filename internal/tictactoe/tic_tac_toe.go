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
	GameStatusOngoing GameStatus = iota
	GameStatusPlayer1Win
	GameStatusPlayer2Win
	GameStatusDraw
)

type TicTacToeGame struct {
	Board      [3][3]TicTacToeSquare `json:"board"`
	GameStatus GameStatus            `json:"game_status"`
}

func NewTicTacToeGame() TicTacToeGame {
	var board [3][3]TicTacToeSquare

	return TicTacToeGame{
		Board:      board,
		GameStatus: GameStatusOngoing,
	}
}

func (game *TicTacToeGame) ValidateMove(turn TicTacToeTurn) bool {
	coords := turn.Coords
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
	
	game.GameStatus = game.CheckGameStatus()
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
				symbol = "\033[31mX\033[0m"
			case SquareO:
				symbol = "\033[34mO\033[0m"
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

// Could find some more efficient solution (ex: associating different decimal values with each board space)
func (game *TicTacToeGame) CheckGameStatus() GameStatus {
	// Check rows for wins
	for i := 0; i < 3; i++ {
		if game.Board[i][0] != SquareEmpty &&
			game.Board[i][0] == game.Board[i][1] &&
			game.Board[i][1] == game.Board[i][2] {
			if game.Board[i][0] == SquareX {
				return GameStatusPlayer1Win
			} else {
				return GameStatusPlayer2Win
			}
		}
	}

	// Check columns for wins
	for j := 0; j < 3; j++ {
		if game.Board[0][j] != SquareEmpty &&
			game.Board[0][j] == game.Board[1][j] &&
			game.Board[1][j] == game.Board[2][j] {
			if game.Board[0][j] == SquareX {
				return GameStatusPlayer1Win
			} else {
				return GameStatusPlayer2Win
			}
		}
	}

	// Check diagonal (top-left to bottom-right)
	if game.Board[0][0] != SquareEmpty &&
		game.Board[0][0] == game.Board[1][1] &&
		game.Board[1][1] == game.Board[2][2] {
		if game.Board[0][0] == SquareX {
			return GameStatusPlayer1Win
		} else {
			return GameStatusPlayer2Win
		}
	}

	// Check diagonal (top-right to bottom-left)
	if game.Board[0][2] != SquareEmpty &&
		game.Board[0][2] == game.Board[1][1] &&
		game.Board[1][1] == game.Board[2][0] {
		if game.Board[0][2] == SquareX {
			return GameStatusPlayer1Win
		} else {
			return GameStatusPlayer2Win
		}
	}

	boardFull := true
	for i := 0; i < 3; i++ {
		for j := 0; j < 3; j++ {
			if game.Board[i][j] == SquareEmpty {
				boardFull = false
				break
			}
		}
		if !boardFull {
			break
		}
	}

	if boardFull {
		return GameStatusDraw
	}

	return GameStatusOngoing
}

type TicTacToeTurn struct {
	Coords vector.Vector `json:"coords"`
}
