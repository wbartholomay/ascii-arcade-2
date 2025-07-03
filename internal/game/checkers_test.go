package game

import (
	"testing"

	"github.com/wbarthol/ascii-arcade-2/internal/vector"
)

func TestCheckersValidateMove(t *testing.T) {
	game := NewCheckersGame()

	tests := []struct {
		name        string
		turn        CheckersTurn
		playerNum   int
		expectedOK  bool
		expectedMsg string
	}{
		{
			name:        "Valid white piece move",
			turn:        CheckersTurn{PieceID: 1, Direction: CheckersDirectionLeft},
			playerNum:   1,
			expectedOK:  true,
			expectedMsg: "",
		},
		{
			name:        "Valid black piece move",
			turn:        CheckersTurn{PieceID: 12, Direction: CheckersDirectionLeft},
			playerNum:   2,
			expectedOK:  true,
			expectedMsg: "",
		},
		{
			name:        "Invalid piece ID",
			turn:        CheckersTurn{PieceID: 99, Direction: CheckersDirectionLeft},
			playerNum:   1,
			expectedOK:  false,
			expectedMsg: "no piece found with ID 99",
		},
		{
			name:        "Regular piece trying to move backwards",
			turn:        CheckersTurn{PieceID: 1, Direction: CheckersDirectionBackLeft},
			playerNum:   1,
			expectedOK:  false,
			expectedMsg: "only kings can move backwards",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ok, msg := game.ValidateMove(tt.turn, tt.playerNum)
			if ok != tt.expectedOK {
				t.Errorf("ValidateMove() ok = %v, expected %v", ok, tt.expectedOK)
			}
			if msg != tt.expectedMsg {
				t.Errorf("ValidateMove() msg = %q, expected %q", msg, tt.expectedMsg)
			}
		})
	}
}

func TestCheckersExecuteTurn(t *testing.T) {
	game := NewCheckersGame()

	turn := CheckersTurn{PieceID: 1, Direction: CheckersDirectionLeft}
	playerNum := 1
	truePieceID := 101

	originalPos := game.PiecePositions[truePieceID]

	game.ExecuteTurn(turn, playerNum)

	newPos := game.PiecePositions[truePieceID]
	if newPos.X == originalPos.X && newPos.Y == originalPos.Y {
		t.Error("Piece did not move after ExecuteTurn")
	}

	if game.Board[originalPos.Y][originalPos.X].Color != "" {
		t.Error("Original position should be empty after move")
	}

	if game.Board[newPos.Y][newPos.X].ID != truePieceID {
		t.Error("New position should contain the moved piece")
	}
}

func TestValidateMoveCapture(t *testing.T) {
	game := NewCheckersGame()

	whitePieceID := 101
	blackPieceID := 201

	//clear board and piece positions
	game.Board = [8][8]CheckersPiece{}
	game.PiecePositions = make(map[int]vector.Vector)

	game.Board[4][4] = CheckersPiece{ID: whitePieceID, Color: pieceWhite, IsKing: false}
	game.Board[3][3] = CheckersPiece{ID: blackPieceID, Color: pieceBlack, IsKing: false}
	game.PiecePositions[whitePieceID] = vector.Vector{X: 4, Y: 4}
	game.PiecePositions[blackPieceID] = vector.Vector{X: 3, Y: 3}
	
	// Test valid capture move
	turn := CheckersTurn{PieceID: 1, Direction: CheckersDirectionLeft}
	ok, msg := game.ValidateMove(turn, 1)
	if !ok {
		t.Errorf("Valid capture move should be allowed, got error: %s", msg)
	}
}

func TestExecuteTurnWithCapture(t *testing.T) {
	game := NewCheckersGame()

	whitePieceID := 101
	blackPieceID := 201

	//clear board and piece positions
	game.Board = [8][8]CheckersPiece{}
	game.PiecePositions = make(map[int]vector.Vector)

	game.Board[4][4] = CheckersPiece{ID: whitePieceID, Color: pieceWhite, IsKing: false}
	game.Board[3][3] = CheckersPiece{ID: blackPieceID, Color: pieceBlack, IsKing: false}
	game.PiecePositions[whitePieceID] = vector.Vector{X: 4, Y: 4}
	game.PiecePositions[blackPieceID] = vector.Vector{X: 3, Y: 3}

	originalBlackCount := game.blackPieceCount

	// Execute capture move
	turn := CheckersTurn{PieceID: 1, Direction: CheckersDirectionLeft}
	game.ExecuteTurn(turn, 1)

	if game.blackPieceCount != originalBlackCount-1 {
		t.Error("Black piece count should be decremented after capture")
	}

	if game.Board[3][3].Color != "" {
		t.Error("Captured piece should be removed from board")
	}

	if game.Board[2][2].ID != whitePieceID {
		t.Error("White piece should move to position after captured piece")
	}
}
