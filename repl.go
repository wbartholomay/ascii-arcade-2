package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func cleanInput(text string) []string {
	text = strings.ToLower(text)
	substrings := strings.Fields(text)
	return substrings
}

func startRepl(session *Session) error {
	scanner := bufio.NewScanner(os.Stdin)

	for {
		fmt.Print("> ")
		scanner.Scan()
		t := scanner.Text()
		input := cleanInput(t)
		if len(input) == 0 {
			continue
		}
		commandName := input[0]
		commandArgs := input[1:]
		cmd, ok := GetCommands()[commandName]
		if !ok {
			fmt.Println("Invalid command, enter 'help' to see a list of commands.")
			continue
		}

		_, isBasicCommand := cmd.(CommandBasic)
		if isBasicCommand {
			cmd.ExecuteCallback(commandArgs)
			continue
		}

		gameCommand, isGameCommand := cmd.(GameCommand)
		if !isGameCommand {
			return fmt.Errorf("server error: entered a command that was not a basic or game command")
		}

		msg, err := gameCommand.CreatePlayerMessage(commandArgs)
		if err != nil {
			fmt.Printf("error creating player message from command: %v\n", err)
			continue
		}

		err = session.ValidatePlayerMessage(msg)
		if err != nil {
			fmt.Println(err)
			continue
		}

		err = session.WriteToServer(msg)
		if err != nil {
			return err
		}

		<- session.sessionToOutput
	}
}
