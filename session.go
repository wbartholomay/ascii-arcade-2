package main

import (
	"fmt"
	"log"

	"github.com/wbarthol/ascii-arcade-2/internal/messages"
	"github.com/wbarthol/ascii-arcade-2/internal/tictactoe"
)

type Session struct {
	stateInMenu      SessionStateInMenu
	stateWaitingRoom SessionStateWaitingRoom
	stateInGame      SessionStateInGame
	state            SessionState

	playerNumber int
	playerTurn   int

	//TODO make this an interface to allow for many game types
	game tictactoe.TicTacToeGame

	sessionToOutput chan string
	WSDriver
}

func StartSession(url string) (*Session, error) {
	//Could move the dialing to StartWS
	wsDriver, err := NewWS(url)
	if err != nil {
		return nil, fmt.Errorf("error starting WS: %w", err)
	}

	session := Session{
		sessionToOutput: make(chan string),
		WSDriver:        wsDriver,
	}

	session.stateInMenu = SessionStateInMenu{
		session: &session,
	}
	session.stateWaitingRoom = SessionStateWaitingRoom{
		session: &session,
	}
	session.stateInGame = SessionStateInGame{
		session: &session,
	}
	session.state = session.stateInMenu

	go session.Run()
	return &session, nil
}

func (session Session) isPlayerTurn() bool {
	return session.playerNumber == session.playerTurn
}

func (session *Session) displayBoardToUser() {
	session.sessionToOutput <- session.game.DisplayBoard()
	if session.isPlayerTurn() {
		session.sessionToOutput <- "Your turn, make a move."
	} else {
		session.sessionToOutput <- "Waiting on opponents move..."
	}
}

func (session *Session) Run() {
	defer func() {
		session.conn.Close()
		//TODO communicate to client that server has closed
	}()

	for {
		var msg messages.ServerMessage
		err := session.conn.ReadJSON(&msg)
		if err != nil {
			fmt.Printf("An error has occurred while reading from the server, shutting down: %v\n", err)
			return
		}
		log.Printf("Received message from server: %v\n", msg)
		err = session.state.handleServerMessage(msg)
		if err != nil {
			fmt.Printf("An error has ocurred while handling the servers message, shutting down: %v\n", err)
			return
		}
	}
}

func (session *Session) ValidatePlayerMessage(msg messages.ClientMessage) error {
	return session.state.validatePlayerMessage(msg)
}

func (session *Session) setState(state SessionState) {
	session.state = state
}

type SessionState interface {
	handleServerMessage(msg messages.ServerMessage) error
	validatePlayerMessage(msg messages.ClientMessage) error
}

type SessionStateInMenu struct {
	session *Session
}

func (state SessionStateInMenu) handleServerMessage(msg messages.ServerMessage) error {
	switch msg.Type {
	case messages.ServerRoomJoined:
		fmt.Println("Joining waiting room...")
		state.session.playerNumber = msg.PlayerNumber
		state.session.setState(&state.session.stateWaitingRoom)
		return nil
	}

	return fmt.Errorf("unexpected server message type whle in menu: %v", msg.Type)
}

func (state SessionStateInMenu) validatePlayerMessage(msg messages.ClientMessage) error {
	switch msg.Type {
	case messages.ClientJoinRoom:
		return nil
	}

	return fmt.Errorf("unexpected player message type while in menu: %v", msg.Type)
}

type SessionStateWaitingRoom struct {
	session *Session
}

func (state SessionStateWaitingRoom) handleServerMessage(msg messages.ServerMessage) error {
	switch msg.Type {
	case messages.ServerGameStarted:
		fmt.Println("Both players joined, starting game!\n> ")
		state.session.game = msg.Game
		state.session.playerTurn = msg.PlayerTurn
		state.session.displayBoardToUser()
		state.session.setState(&state.session.stateInGame)
		return nil
	}

	return fmt.Errorf("unexpected server message type whle in waiting room: %v", msg.Type)
}

func (state SessionStateWaitingRoom) validatePlayerMessage(msg messages.ClientMessage) error {
	switch msg.Type {
	case messages.ClientQuitRoom:
		//TODO
		return nil
	}

	return fmt.Errorf("unexpected player message type while in waiting room: %v", msg.Type)
}

type SessionStateInGame struct {
	session *Session
}

func (state SessionStateInGame) handleServerMessage(msg messages.ServerMessage) error {
	switch msg.Type {
	case messages.ServerTurnResult:
		state.session.game = msg.Game
		state.session.playerTurn = msg.PlayerTurn
		state.session.displayBoardToUser()
		return nil
	}

	return fmt.Errorf("unexpected server message type whle in game: %v", msg.Type)
}

func (state SessionStateInGame) validatePlayerMessage(msg messages.ClientMessage) error {

	switch msg.Type {
	case messages.ClientSendTurn:
		if state.session.playerNumber != state.session.playerTurn {
			return fmt.Errorf("can not send a move on another players turn")
		}

		return nil
	case messages.ClientQuitRoom:

		return nil
	}

	return fmt.Errorf("unexpected player message type while in game: %v", msg.Type)
}
