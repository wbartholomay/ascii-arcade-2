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
	userInput := make(chan []string)

	go func() {
		scanner := bufio.NewScanner(os.Stdin)
		for {
			scanner.Scan()
			input := cleanInput(scanner.Text())
			userInput <- input
		}
	}()

	for {
		fmt.Print("> ")
		select {
		case input := <-userInput:
			processUserInput(input, session)
		case output := <-session.sessionToOutput:
			fmt.Println(output)
		}
	}
}

func processUserInput(input []string, session *Session) error {
	commandName := input[0]
	commandArgs := input[1:]
	cmd, ok := GetCommands()[commandName]
	if !ok {
		fmt.Println("Invalid command, enter 'help' to see a list of commands.")
		return nil
	}

	_, isBasicCommand := cmd.(CommandBasic)
	if isBasicCommand {
		cmd.ExecuteCallback(commandArgs)
		return nil
	}

	gameCommand, isGameCommand := cmd.(GameCommand)
	if !isGameCommand {
		return fmt.Errorf("server error: entered a command that was not a basic or game command")
	}

	msg, err := gameCommand.CreatePlayerMessage(commandArgs)
	if err != nil {
		fmt.Printf("error creating player message from command: %v\n", err)
		return nil
	}

	err = session.ValidatePlayerMessage(msg)
	if err != nil {
		fmt.Println(err)
		return nil
	}

	err = session.WriteToServer(msg)
	if err != nil {
		//only return an error when the process should be killed
		return err
	}
	return nil
}
