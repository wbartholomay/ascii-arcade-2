package game

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
	"github.com/wbarthol/ascii-arcade-2/internal/vector"
)

type TicTacToeSquare int

const (
	TicTacToeSquareEmpty TicTacToeSquare = iota
	TicTacToeSquareX
	TicTacToeSquareO
)

type TicTacToeGame struct {
	GameType   GameType              `json:"game_type"`
	Board      [3][3]TicTacToeSquare `json:"board"`
	GameStatus GameStatus            `json:"game_status"`
}

func NewTicTacToeGame() *TicTacToeGame {
	var board [3][3]TicTacToeSquare

	return &TicTacToeGame{
		Board:      board,
		GameStatus: GameStatusOngoing,
	}
}

func (game *TicTacToeGame) GetGameType() GameType {
	return game.GameType
}

func (game *TicTacToeGame) GetGameStatus() GameStatus {
	return game.GameStatus
}

func (game *TicTacToeGame) GetGameInstructions() string {
	return "when it is your turn, enter \033[33m move <row-num> <col-num>\033[0m."
}

func (game *TicTacToeGame) ValidateMove(gameTurn GameTurn, playerNum int) (bool, string) {
	turn, ok := gameTurn.(TicTacToeTurn)
	if !ok {
		panic("server error - sent a turn not of type tictactoe turn during tictactoe game")
	}

	coords := turn.Coords
	rowInBounds := coords.Y >= 0 && coords.Y <= 2
	colInBounds := coords.X >= 0 && coords.Y <= 2
	if !rowInBounds || !colInBounds {
		return false, "selected square is out of bounds"
	}

	if !(game.Board[coords.Y][coords.X] == TicTacToeSquareEmpty) {
		return false, "square is occupied"
	}

	return true, ""
}

type TicTacToeTurn struct {
	Coords vector.Vector `json:"coords"`
}

func (turn TicTacToeTurn) GetGameType() GameType {
	return GameTypeTicTacToe
}

// ExecuteTurn - Takes coordinates and a player number, executes turn.
func (game *TicTacToeGame) ExecuteTurn(gameTurn GameTurn, playerNum int) string {
	turn, ok := gameTurn.(TicTacToeTurn)
	if !ok {
		panic("server error - sent a turn not of type tictactoe turn during tictactoe game")
	}
	coords := turn.Coords
	playerSquare := TicTacToeSquareX
	if playerNum == 2 {
		playerSquare = TicTacToeSquareO
	}

	game.Board[coords.Y][coords.X] = playerSquare

	game.GameStatus = game.checkGameStatus()
	return ""
}

func (game *TicTacToeGame) DisplayBoard(cursorPosition vector.Vector, _ int) string {
	// Define styles
	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FAFAFA")).
		Background(lipgloss.Color("#6366F1")).
		Padding(0, 1).
		MarginBottom(1)

	boardStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#6366F1")).
		Padding(1)

	xStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#EF4444"))

	oStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#3B82F6"))

	cursorStyle := lipgloss.NewStyle().
		Background(lipgloss.Color("#10B981")).
		Bold(true)

	// Create header
	header := headerStyle.Render("TIC TAC TOE")

	// Column headers
	columnHeaders := "        0       1       2   "

	// Build the grid
	var result string
	result += "  ┌───────┬───────┬───────┐\n"

	for i := 0; i < 3; i++ {
		// Empty row above content for height
		result += "  │"
		for j := 0; j < 3; j++ {
			if cursorPosition.Equals(vector.NewVector(j, i)) {
				result += cursorStyle.Render("       ") + "│"
			} else {
				result += "       │"
			}
		}
		result += "\n"

		// Content row with row number
		result += fmt.Sprintf("%d │", i)
		for j := 0; j < 3; j++ {
			var symbol string
			switch game.Board[i][j] {
			case TicTacToeSquareEmpty:
				symbol = " "
			case TicTacToeSquareX:
				symbol = "X"
			case TicTacToeSquareO:
				symbol = "O"
			}

			if cursorPosition.Equals(vector.NewVector(j, i)) {
				// Apply cursor styling to the entire cell content including symbol
				content := cursorStyle.Render(fmt.Sprintf("   %s   ", symbol))
				result += content + "│"
			} else {
				// Apply individual symbol styling only when not highlighted
				if game.Board[i][j] == TicTacToeSquareX {
					symbol = xStyle.Render("X")
				} else if game.Board[i][j] == TicTacToeSquareO {
					symbol = oStyle.Render("O")
				}
				result += fmt.Sprintf("   %s   │", symbol)
			}
		}
		result += "\n"

		// Empty row below content for height
		result += "  │"
		for j := 0; j < 3; j++ {
			if cursorPosition.Equals(vector.NewVector(j, i)) {
				result += cursorStyle.Render("       ") + "│"
			} else {
				result += "       │"
			}
		}
		result += "\n"

		// Add horizontal separator (except after last row)
		if i < 2 {
			result += "  ├───────┼───────┼───────┤\n"
		}
	}

	result += "  └───────┴───────┴───────┘"

	// Combine everything
	gridWithHeaders := lipgloss.JoinVertical(lipgloss.Left, columnHeaders, result)
	styledBoard := boardStyle.Render(gridWithHeaders)

	return lipgloss.JoinVertical(lipgloss.Center, header, styledBoard)
}

// Could find some more efficient solution (ex: associating different decimal values with each board space)
func (game *TicTacToeGame) checkGameStatus() GameStatus {
	// Check rows for wins
	for i := 0; i < 3; i++ {
		if game.Board[i][0] != TicTacToeSquareEmpty &&
			game.Board[i][0] == game.Board[i][1] &&
			game.Board[i][1] == game.Board[i][2] {
			if game.Board[i][0] == TicTacToeSquareX {
				return GameStatusPlayer1Win
			} else {
				return GameStatusPlayer2Win
			}
		}
	}

	// Check columns for wins
	for j := 0; j < 3; j++ {
		if game.Board[0][j] != TicTacToeSquareEmpty &&
			game.Board[0][j] == game.Board[1][j] &&
			game.Board[1][j] == game.Board[2][j] {
			if game.Board[0][j] == TicTacToeSquareX {
				return GameStatusPlayer1Win
			} else {
				return GameStatusPlayer2Win
			}
		}
	}

	// Check diagonal (top-left to bottom-right)
	if game.Board[0][0] != TicTacToeSquareEmpty &&
		game.Board[0][0] == game.Board[1][1] &&
		game.Board[1][1] == game.Board[2][2] {
		if game.Board[0][0] == TicTacToeSquareX {
			return GameStatusPlayer1Win
		} else {
			return GameStatusPlayer2Win
		}
	}

	// Check diagonal (top-right to bottom-left)
	if game.Board[0][2] != TicTacToeSquareEmpty &&
		game.Board[0][2] == game.Board[1][1] &&
		game.Board[1][1] == game.Board[2][0] {
		if game.Board[0][2] == TicTacToeSquareX {
			return GameStatusPlayer1Win
		} else {
			return GameStatusPlayer2Win
		}
	}

	boardFull := true
	for i := 0; i < 3; i++ {
		for j := 0; j < 3; j++ {
			if game.Board[i][j] == TicTacToeSquareEmpty {
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
