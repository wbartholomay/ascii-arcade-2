package game

import (
	"fmt"
	"github.com/wbarthol/ascii-arcade-2/internal/vector"
	"log"
)

const pieceWhite = "w"
const pieceBlack = "b"

type CheckersPiece struct {
	ID     int    `json:"id"`
	Color  string `json:"color"`
	IsKing bool   `json:"is_king"`
}

type CheckersGame struct {
	GameType        GameType              `json:"game_type"`
	Board           [8][8]CheckersPiece   `json:"board"`
	PiecePositions  map[int]vector.Vector `json:"piece_positions"`
	GameStatus      GameStatus            `json:"game_status"`
	whitePieceCount int
	blackPieceCount int
}

func NewCheckersGame() *CheckersGame {
	board := [8][8]CheckersPiece{}
	pieces := map[int]vector.Vector{}

	whitePieceID := 101
	blackPieceID := 201

	for row := range board {
		for col := range board[row] {
			hasPiece := ((row % 2) == 0) == ((col % 2) == 0)

			//initialize pieces
			if hasPiece && row < 3 {
				board[row][col] = CheckersPiece{
					ID:     blackPieceID,
					Color:  pieceBlack,
					IsKing: false,
				}
				pieces[blackPieceID] = vector.Vector{
					Y: row,
					X: col,
				}
				blackPieceID++
			} else if hasPiece && row > 4 {
				board[row][col] = CheckersPiece{
					ID:     whitePieceID,
					Color:  pieceWhite,
					IsKing: false,
				}
				pieces[whitePieceID] = vector.Vector{
					Y: row,
					X: col,
				}
				whitePieceID++
			} else {
				//leaving an empty struct here for now
				board[row][col] = CheckersPiece{}
			}
		}
	}

	return &CheckersGame{
		GameType:        GameTypeCheckers,
		Board:           board,
		PiecePositions:  pieces,
		GameStatus:      GameStatusOngoing,
		whitePieceCount: 12,
		blackPieceCount: 12,
	}
}

func (game *CheckersGame) GetGameType() GameType {
	return game.GameType
}

func (game *CheckersGame) GetGameStatus() GameStatus {
	return game.GameStatus
}

func (game *CheckersGame) GetGameInstructions() string {
	return "when it is your turn, enter \033[33m move <piece-num> <direction>\033[0m.\nPossible directions are \033[33m'l', 'r', 'bl', 'br'\033[0m. Note that only kings can move backwards."
}

type CheckersDirection int

const (
	CheckersDirectionLeft CheckersDirection = iota
	CheckersDirectionRight
	CheckersDirectionBackLeft
	CheckersDirectionBackRight
)

type CheckersTurn struct {
	PieceID   int               `json:"piece_id"`
	Direction CheckersDirection `json:"direction"`
}

func (turn CheckersTurn) GetGameType() GameType {
	return GameTypeCheckers
}

// TODO switch this over to returning an error instead of bool + string
func (game *CheckersGame) ValidateMove(gameTurn GameTurn, playerNum int) (bool, string) {
	turn, ok := gameTurn.(CheckersTurn)
	if !ok {
		panic("server error - sent a turn not of type checkers turn during checkers game")
	}

	var truePieceID int
	if playerNum == 1 {
		truePieceID = turn.PieceID + 100
	} else {
		truePieceID = turn.PieceID + 200
	}

	//check selected square
	pieceCoords, ok := game.PiecePositions[truePieceID]
	if !ok {
		return false, fmt.Sprintf("no piece found with ID %v", turn.PieceID)
	}

	piece := game.Board[pieceCoords.Y][pieceCoords.X]
	if !piece.IsKing {
		if turn.Direction == CheckersDirectionBackLeft || turn.Direction == CheckersDirectionBackRight {
			return false, "only kings can move backwards"
		}
	}

	//get absolute direction based on input direction and piece color
	trueDirection := turn.Direction
	if playerNum == 2 {
		trueDirection = convertDirectionFromBlackToWhite(trueDirection)
	}

	targetSquare := applyMove(pieceCoords, trueDirection)
	if game.isSquareOutOfBounds(targetSquare) {
		return false, "destination is out of bounds"
	}
	targetPiece := game.Board[targetSquare.Y][targetSquare.X]
	if targetPiece.Color == piece.Color {
		return false, "destination is occupied"
	}

	//check for capture
	isOpponentPieceOnDest := targetPiece.Color != "" && targetPiece.Color != piece.Color
	if isOpponentPieceOnDest {
		squareBehindTarget := applyMove(targetSquare, trueDirection)
		if game.isSquareOutOfBounds(squareBehindTarget) {
			return false, "destination is out of bounds"
		}
		if !game.isSquareEmpty(squareBehindTarget) {
			return false, "destination is occupied"
		}
	}

	return true, ""
}

func (game *CheckersGame) ExecuteTurn(gameTurn GameTurn, playerNum int) string{
	turn, ok := gameTurn.(CheckersTurn)
	if !ok {
		panic("server error - sent a turn not of type checkers turn during checkers game")
	}

	var truePieceID int
	if playerNum == 1 {
		truePieceID = turn.PieceID + 100
	} else {
		truePieceID = turn.PieceID + 200
	}

	pieceCoords, ok := game.PiecePositions[truePieceID]
	if !ok {
		panic("validation was not called before execution, or it failed")
	}
	piece := game.Board[pieceCoords.Y][pieceCoords.X]

	//get absolute direction based on input direction and piece color
	trueDirection := turn.Direction
	if playerNum == 2 {
		trueDirection = convertDirectionFromBlackToWhite(trueDirection)
	}

	targetSquare := applyMove(pieceCoords, trueDirection)
	targetPiece := game.Board[targetSquare.Y][targetSquare.X]
	isOpponentPieceOnDest := targetPiece.Color != "" && targetPiece.Color != piece.Color

	//assume validation has already run, and destination being occupied by opponent means capture
	if isOpponentPieceOnDest {
		game.capturePiece(targetSquare)
		targetSquare = applyMove(targetSquare, trueDirection)
		if piece.Color == pieceWhite {
			return "captured a black piece!"
		} else {
			return "captured a white piece!"
		}
	}

	game.Board[targetSquare.Y][targetSquare.X] = game.Board[pieceCoords.Y][pieceCoords.X]
	game.Board[pieceCoords.Y][pieceCoords.X] = CheckersPiece{}
	game.PiecePositions[truePieceID] = targetSquare
	game.GameStatus = game.checkGameStatus()
	return ""
}

func (game *CheckersGame) DisplayBoard(playerNum int) string {
	isWhiteTurn := true
	if playerNum == 2 {
		isWhiteTurn = false
	}

	result := "\n"
	board := game.Board
	rowNum := 0
	increment := 1
	checkIndex := func(i int) bool {
		if isWhiteTurn {
			return i < 8
		} else {
			return i >= 0
		}
	}

	if !isWhiteTurn {
		rowNum = 7
		increment = -1
		result += "       7       6       5       4       3       2       1       0    \n"
	} else {
		result += "       0       1       2       3       4       5       6       7    \n"
	}

	for ; checkIndex(rowNum); rowNum += increment {
		result += "   —————————————————————————————————————————————————————————————————\n"
		squareStr := ""
		if (rowNum%2 == 0 && isWhiteTurn) || (rowNum%2 != 0 && !isWhiteTurn) {
			squareStr = "   |       |#######|       |#######|       |#######|       |#######|"
		} else {
			squareStr = "   |#######|       |#######|       |#######|       |#######|       |"
		}
		result += squareStr + "\n"
		rowStr := fmt.Sprintf("%v  |", string(rune('a'+rowNum)))

		colNum := 0
		if !isWhiteTurn {
			colNum = 7
		}

		for ; checkIndex(colNum); colNum += increment {
			piece := board[rowNum][colNum]
			if rowNum%2 == colNum%2 {
				rowStr += fmt.Sprintf("%v|", piece.renderPiece())
			} else {
				rowStr += "#######|"
			}
		}
		result += rowStr + "\n"
		result += squareStr + "\n"
	}
	result += "   —————————————————————————————————————————————————————————————————\n"

	return result
}

func (piece *CheckersPiece) renderPiece() string {
	if piece.Color == "" {
		return "       "
	}

	pieceStr := ""
	if piece.IsKing {
		pieceStr += "👑"
	} else {
		pieceStr += "  "
	}

	if piece.Color == pieceWhite {
		pieceStr += "⚪"
	} else if piece.Color == pieceBlack {
		pieceStr += "🔵"
	}
	pieceStr += toSubscript(piece.getDisplayID())

	if piece.getDisplayID() < 10 {
		pieceStr += " "
	}

	return pieceStr + " "
}

func toSubscript(n int) string {
	subs := []string{"", "₁", "₂", "₃", "₄", "₅", "₆", "₇", "₈", "₉", "₁₀", "₁₁", "₁₂"}
	return subs[n]
}

func (piece *CheckersPiece) getDisplayID() int {
	displayId := 0
	if piece.Color == pieceWhite {
		displayId = piece.ID - 100
	} else {
		displayId = piece.ID - 200
	}
	return displayId
}

func convertDirectionFromBlackToWhite(direction CheckersDirection) CheckersDirection {
	switch direction {
	case CheckersDirectionLeft:
		return CheckersDirectionBackRight
	case CheckersDirectionRight:
		return CheckersDirectionBackLeft
	case CheckersDirectionBackLeft:
		return CheckersDirectionRight
	case CheckersDirectionBackRight:
		return CheckersDirectionLeft
	}
	return CheckersDirectionLeft
}

func applyMove(srcSquare vector.Vector, direction CheckersDirection) vector.Vector {
	directionVector := vector.Vector{}
	switch direction {
	case CheckersDirectionLeft:
		directionVector = vector.Vector{
			X: -1,
			Y: -1,
		}
	case CheckersDirectionRight:
		directionVector = vector.Vector{
			X: 1,
			Y: -1,
		}
	case CheckersDirectionBackLeft:
		directionVector = vector.Vector{
			X: -1,
			Y: 1,
		}
	case CheckersDirectionBackRight:
		directionVector = vector.Vector{
			X: 1,
			Y: 1,
		}
	}
	srcSquare.Add(directionVector)
	return srcSquare
}

func (game *CheckersGame) isSquareEmpty(coords vector.Vector) bool {
	piece := game.Board[coords.Y][coords.X]
	return piece.Color == ""
}

func (game *CheckersGame) isSquareOutOfBounds(targetSquare vector.Vector) bool {
	return targetSquare.X < 0 || targetSquare.X > 7 || targetSquare.Y < 0 || targetSquare.Y > 7
}

func (game *CheckersGame) capturePiece(targetSquare vector.Vector) {
	targetPiece := game.Board[targetSquare.Y][targetSquare.X]
	if targetPiece.ID == 0 {
		panic("no piece on this square - did validation run?")
	}

	if targetPiece.Color == pieceWhite {
		game.whitePieceCount--
	} else {
		game.blackPieceCount--
	}
	delete(game.PiecePositions, targetPiece.ID)
	game.Board[targetSquare.Y][targetSquare.X] = CheckersPiece{}
	log.Printf("Capture a piece at %v, %v", targetSquare.X, targetSquare.Y)
}

func (game *CheckersGame) checkGameStatus() GameStatus {
	if game.whitePieceCount == 0 {
		return GameStatusPlayer2Win
	}

	if game.blackPieceCount == 0 {
		return GameStatusPlayer1Win
	}

	return GameStatusOngoing
}
