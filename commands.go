package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/wbarthol/ascii-arcade-2/internal/game"
	"github.com/wbarthol/ascii-arcade-2/internal/messages"
	"github.com/wbarthol/ascii-arcade-2/internal/vector"
)

func GetCommands() map[string]Command {
	return map[string]Command{
		"help": CommandBasic{
			name:        "help",
			description: "display the list of all commands",
			callback: func(args []string) error {
				for _, cmd := range GetCommands() {
					fmt.Printf("%s: %s\n", AnsiYellow+cmd.GetName()+AnsiReset, cmd.GetDescription())
				}
				return nil
			},
		},
		"exit": CommandBasic{
			name:        "exit",
			description: "exit the application",
			callback: func(args []string) error {
				os.Exit(0)
				return nil
			},
		},
		"join": CommandJoin{
			CommandBasic: CommandBasic{
				name:        "join",
				description: "join a room. Usage: \033[33mjoin <room-code>\033[0m",
				callback: func(args []string) error {
					return nil
				},
			},
		},
		"select": CommandSelectGame{
			CommandBasic: CommandBasic{
				name:        "select game",
				description: "select a game. Usage: \033[33mselect <game-number>\033[0m",
				callback: func(args []string) error {
					return nil
				},
			},
		},
		"move": CommandSendTurn{
			CommandBasic: CommandBasic{
				name:        "move",
				description: "select square. Usage: \033[33mmove <row-num> <col-num>\033[0m",
				callback: func(args []string) error {
					return nil
				},
			},
		},
		"quit": CommandQuit{
			CommandBasic: CommandBasic{
				name:        "quit",
				description: "quits current game.",
				callback: func(args []string) error {
					return nil
				},
			},
		},
	}
}

type Command interface {
	GetName() string
	GetDescription() string
	ExecuteCallback(args []string) error
}

type CommandBasic struct {
	name        string
	description string
	callback    func(args []string) error
}

func (cmd CommandBasic) GetName() string {
	return cmd.name
}

func (cmd CommandBasic) GetDescription() string {
	return cmd.description
}

func (cmd CommandBasic) ExecuteCallback(args []string) error {
	return cmd.callback(args)
}

type GameCommand interface {
	CreatePlayerMessage(game.GameType, []string) (messages.ClientMessage, error)
}

type CommandJoin struct {
	CommandBasic
}

func (cmd CommandJoin) CreatePlayerMessage(_ game.GameType, args []string) (messages.ClientMessage, error) {
	if len(args) < 1 {
		return messages.ClientMessage{}, fmt.Errorf("not enough arguments provided. Expecting \033[33m<room-code>\033[0m")
	}
	roomCode := args[0]

	return messages.ClientMessage{
		Type:     messages.ClientJoinRoom,
		RoomCode: roomCode,
	}, nil
}

type CommandSelectGame struct {
	CommandBasic
}

func (cmd CommandSelectGame) CreatePlayerMessage(_ game.GameType, args []string) (messages.ClientMessage, error) {
	if len(args) < 1 {
		return messages.ClientMessage{}, fmt.Errorf("not enough arguments provided. Expecting \033[33m<game-number>\033[0m")
	}
	gameNumber := args[0]

	var gameType game.GameType
	switch gameNumber {
	case "1":
		gameType = game.GameTypeTicTacToe
	case "2":
		gameType = game.GameTypeCheckers
	default:
		return messages.ClientMessage{}, fmt.Errorf("invalid game number. Valid numbers are: " + AnsiYellow + "'1', '2'" + AnsiReset)
	}

	return messages.ClientMessage{
		Type:     messages.ClientSelectGameType,
		GameType: gameType,
	}, nil
}

type CommandSendTurn struct {
	CommandBasic
}

func (cmd CommandSendTurn) CreatePlayerMessage(gameType game.GameType, args []string) (messages.ClientMessage, error) {
	switch gameType {
	case game.GameTypeTicTacToe:
		return createTicTacToeTurn(args)
	case game.GameTypeCheckers:
		return createCheckersTurn(args)
	}

	return messages.ClientMessage{}, fmt.Errorf("unknown game type")
}

func createTicTacToeTurn(args []string) (messages.ClientMessage, error) {
	if len(args) < 2 {
		return messages.ClientMessage{}, fmt.Errorf("not enough arguments provided. Expecting \033[33m<row-num> <col-num>\033[0m")
	}

	moveRow, err := strconv.ParseInt(args[0], 10, 32)
	if err != nil {
		return messages.ClientMessage{}, fmt.Errorf("row number must be a number")
	}

	moveCol, err := strconv.ParseInt(args[1], 10, 32)
	if err != nil {
		return messages.ClientMessage{}, fmt.Errorf("col number must be a number")
	}

	coords := vector.Vector{
		X: int(moveCol),
		Y: int(moveRow),
	}

	return messages.ClientMessage{
		Type: messages.ClientSendTurn,
		TurnAction: messages.NewGameTurnWrapper(game.TicTacToeTurn{
			Coords: coords,
		}),
	}, nil
}

func createCheckersTurn(args []string) (messages.ClientMessage, error) {
	if len(args) < 2 {
		return messages.ClientMessage{}, fmt.Errorf("not enough arguments provided. Expecting \033[33m<piece-num> <move-direction>\033[0m")
	}

	pieceNum, err := strconv.ParseInt(args[0], 10, 32)
	if err != nil {
		return messages.ClientMessage{}, fmt.Errorf("piece number must be a number")
	}

	directionStr := args[1]
	var direction game.CheckersDirection
	switch directionStr {
	case "l":
		direction = game.CheckersDirectionLeft
	case "r":
		direction = game.CheckersDirectionRight
	case "bl":
		direction = game.CheckersDirectionBackLeft
	case "br":
		direction = game.CheckersDirectionBackRight
	default:
		return messages.ClientMessage{}, fmt.Errorf("invalid direction. Valid directions are: " + AnsiYellow + "'l', 'r', 'bl', 'br'" + AnsiReset)
	}

	return messages.ClientMessage{
		Type: messages.ClientSendTurn,
		TurnAction: messages.NewGameTurnWrapper(game.CheckersTurn{
			PieceID:   int(pieceNum),
			Direction: direction,
		}),
	}, nil
}

type CommandQuit struct {
	CommandBasic
}

func (cmd CommandQuit) CreatePlayerMessage(_ game.GameType, args []string) (messages.ClientMessage, error) {
	return messages.ClientMessage{
		Type: messages.ClientQuitRoom,
	}, nil
}
