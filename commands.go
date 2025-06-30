package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/wbarthol/ascii-arcade-2/internal/messages"
	"github.com/wbarthol/ascii-arcade-2/internal/tictactoe"
	"github.com/wbarthol/ascii-arcade-2/internal/vector"
)

func GetCommands() map[string]Command {
	return map[string]Command{
		"help": CommandBasic{
			name:        "help",
			description: "display the list of all commands",
			callback: func(args []string) error {
				for _, cmd := range GetCommands() {
					fmt.Printf("%s: %s\n", cmd.GetName(), cmd.GetDescription())
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
				description: "join a room. Usage: join <room-code>",
				callback: func(args []string) error {
					return nil
				},
			},
		},
		"move": CommandSendTurn{
			CommandBasic: CommandBasic{
				name:        "move",
				description: "select square. Usage: move <row-num> <col-num>",
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
	CreatePlayerMessage([]string) (messages.ClientMessage, error)
}

type CommandJoin struct {
	CommandBasic
}

func (cmd CommandJoin) CreatePlayerMessage(args []string) (messages.ClientMessage, error) {
	if len(args) < 1 {
		return messages.ClientMessage{}, fmt.Errorf("not enough arguments provided. Expecting <room-code>")
	}
	roomCode := args[0]

	return messages.ClientMessage{
		Type:     messages.ClientJoinRoom,
		RoomCode: roomCode,
	}, nil
}

type CommandSendTurn struct {
	CommandBasic
}

func (cmd CommandSendTurn) CreatePlayerMessage(args []string) (messages.ClientMessage, error) {
	if len(args) < 2 {
		return messages.ClientMessage{}, fmt.Errorf("not enough arguments provided. Expecting <row-num> <col-num>")
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
		TurnAction: tictactoe.TicTacToeTurn{
			Coords: coords,
		},
	}, nil
}

type CommandQuit struct {
	CommandBasic
}

func (cmd CommandQuit) CreatePlayerMessage(args []string) (messages.ClientMessage, error) {
	return messages.ClientMessage{
		Type: messages.ClientQuitRoom,
	}, nil
}
