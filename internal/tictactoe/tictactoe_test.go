package tictactoe

import (
	"testing"

	"github.com/wbarthol/ascii-arcade-2/internal/vector"
)

func TestValidateMove(t *testing.T) {
	game := NewTicTacToeGame()
	coords := vector.Vector{X: 1, Y: 1}
	turn := TicTacToeTurn{Coords: coords}
	game.ExecuteTurn(turn, 1)

	tests := []struct {
		name     string
		coords   vector.Vector
		expected bool
	}{
		{"Valid move top-left", vector.Vector{X: 0, Y: 0}, true},
		{"Invalid move - negative X", vector.Vector{X: -1, Y: 1}, false},
		{"Invalid move - occupied square", vector.Vector{X: 1, Y: 1}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := game.ValidateMove(tt.coords)
			if result != tt.expected {
				t.Errorf("ValidateMove(%v) = %v, expected %v", tt.coords, result, tt.expected)
			}
		})
	}
}

func TestExecuteTurn(t *testing.T) {
	game := NewTicTacToeGame()

	tests := []struct {
		name           string
		coords         vector.Vector
		playerNum      int
		expectedSquare TicTacToeSquare
	}{
		{"Player 1 places X", vector.Vector{X: 0, Y: 0}, 1, SquareX},
		{"Player 2 places O", vector.Vector{X: 1, Y: 1}, 2, SquareO},
		{"Player 1 places another X", vector.Vector{X: 2, Y: 2}, 1, SquareX},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			turn := TicTacToeTurn{Coords: tt.coords}
			game.ExecuteTurn(turn, tt.playerNum)

			actualSquare := game.Board[tt.coords.Y][tt.coords.X]
			if actualSquare != tt.expectedSquare {
				t.Errorf("ExecuteTurn at (%d,%d) for player %d: expected %v, got %v",
					tt.coords.X, tt.coords.Y, tt.playerNum, tt.expectedSquare, actualSquare)
			}
		})
	}
}
