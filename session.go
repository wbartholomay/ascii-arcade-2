package main

import (
	"fmt"
	"github.com/wbarthol/ascii-arcade-2/internal/tic_tac_toe"
)

type Session struct {
	stateInMenu      SessionStateInMenu
	stateWaitingRoom SessionStateWaitingRoom
	stateInGame      SessionStateInGame
	state            SessionState

	//TODO make this an interface to allow for many game types
	game tictactoe.TicTacToeGame

	wsDriver *WSDriver
}

func StartSession(url string) (*Session, error) {
	//Could move the dialing to StartWS
	wsDriver, err := StartWS(url)
	if err != nil {
		return nil, fmt.Errorf("error starting WS: %w", err)
	}

	session := Session{
		wsDriver: wsDriver,
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

func (session *Session) Run() {
	for {
		serverMsg := <-session.wsDriver.serverToPlayer
		err := session.state.handleServerMessage(serverMsg)
		if err != nil {
			break
		}
	}

	//TODO shutdown WSDriver
}

func (session *Session) HandlePlayerMessage(msg PlayerMessage) error {
	return session.state.handlePlayerMessage(msg)
}

func (session *Session) WriteToServer(msg PlayerMessage) error {
	return session.wsDriver.WriteToServer(msg)
}

func (session *Session) setState(state SessionState) {
	session.state = state
}

type SessionState interface {
	handleServerMessage(msg ServerMessage) error
	handlePlayerMessage(msg PlayerMessage) error
}

type SessionStateInMenu struct {
	session *Session
}

func (state SessionStateInMenu) handleServerMessage(msg ServerMessage) error {
	switch msg.(type) {
	case ServerMessageRoomJoined:
		state.session.setState(&state.session.stateWaitingRoom)
		return nil
	}

	return fmt.Errorf("unexpected server message type whle in menu: %v", msg.GetType())
}

func (state SessionStateInMenu) handlePlayerMessage(msg PlayerMessage) error {
	switch msg.(type) {
	case PlayerMessageJoinRoom:
		return state.session.WriteToServer(msg)
	}

	return fmt.Errorf("unexpected player message type while in menu: %v", msg.GetType())
}

type SessionStateWaitingRoom struct {
	session *Session
}

func (state SessionStateWaitingRoom) handleServerMessage(msg ServerMessage) error {
	switch m := msg.(type) {
	case ServerMessageGameStarted:
		state.session.game = m.Game
		state.session.setState(&state.session.stateInGame)
		return nil
	}

	return fmt.Errorf("unexpected server message type whle in waiting room: %v", msg.GetType())
}

func (state SessionStateWaitingRoom) handlePlayerMessage(msg PlayerMessage) error {
	switch msg.(type) {
	case PlayerMessageQuitRoom:
		//TODO
		return nil
	}

	return fmt.Errorf("unexpected player message type while in waiting room: %v", msg.GetType())
}

type SessionStateInGame struct {
	session *Session
}

func (state SessionStateInGame) handleServerMessage(msg ServerMessage) error {
	switch msg.(type) {
	case ServerMessageTurnResult:
		//TODO
		return nil
	}

	return fmt.Errorf("unexpected server message type whle in game: %v", msg.GetType())
}

func (state SessionStateInGame) handlePlayerMessage(msg PlayerMessage) error {
	switch msg.(type) {
	case PlayerMessageQuitRoom:
		//TODO
		return nil
	case PlayerMessageSendTurn:
		//TODO
		return nil
	}

	return fmt.Errorf("unexpected player message type while in game: %v", msg.GetType())
}
