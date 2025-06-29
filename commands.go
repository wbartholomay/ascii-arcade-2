package main

import (
	"fmt"
	"os"
)
func GetCommands() map[string]Command {
	return map[string]Command{
		"help": CommandBasic{
			name: "help",
			description: "display the list of all commands",
			callback: func(args []string) error{
				for _, cmd := range GetCommands() {
					fmt.Printf("%s: %s\n", cmd.GetName(), cmd.GetDescription())
				}
				return nil
			},
		},
		"exit": CommandBasic{
			name: "exit",
			description: "exit the application",
			callback: func(args []string) error{
				os.Exit(0)
				return nil
			},
		},
		"join": CommandJoin{
			CommandBasic: CommandBasic{
				name: "join",
				description: "join a room. Usage: join <room-code>",
				callback: func(args []string) error{
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

func (cmd CommandBasic) ExecuteCallback(args []string) error{
	return cmd.callback(args)
}

type GameCommand interface {
	CreatePlayerMessage([]string) (PlayerMessage, error)
}

type CommandJoin struct {
	CommandBasic
}

func (cmd CommandJoin) CreatePlayerMessage(args []string) (PlayerMessage, error) {
	if len(args) < 1 {
		return PlayerMessageJoinRoom{}, fmt.Errorf("not enough arguments provided. Expecting <room-code>")
	}
	roomCode := args[0]

	return PlayerMessageJoinRoom{
		Type: PlayerJoinRoom,
		RoomCode: roomCode,
	}, nil
}
