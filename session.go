package main

import (
	"fmt"

	"github.com/wbarthol/ascii-arcade-2/internal/messages"
	"github.com/wbarthol/ascii-arcade-2/internal/tictactoe"
)

type ValidationError struct {
	errorMsg string
}

func (err ValidationError) Error() string {
	return err.errorMsg
}

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
	driverToSession chan messages.ServerMessage
	wsDriver        *WSDriver
}

func NewSession() *Session {
	//Could move the dialing to StartWS
	session := Session{
		sessionToOutput: make(chan string, 10),
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

	return &session
}

func (session *Session) StartWS(url string) error {
	wsDriver, err := NewWS(url, session)
	if err != nil {
		return fmt.Errorf("error starting WS: %w", err)
	}
	session.wsDriver = wsDriver
	session.driverToSession = wsDriver.driverToSession
	go session.wsDriver.Run()
	go session.Run()
	return nil
}

func (session *Session) Run() {
	//Decoupling this from WSDriver with singleplayer in mind
	for {
		msg, ok := <-session.driverToSession
		if !ok {
			if session.state != session.stateInMenu {
				session.setState(session.stateInMenu)
			}
			return
		}
		session.state.handleServerMessage(msg)
	}
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

func (session *Session) handleGameOver(gameResult messages.GameResult, detailsFromServer string) {
	session.sessionToOutput <- session.game.DisplayBoard()
	if detailsFromServer != "" {
		session.sessionToOutput <- detailsFromServer
	}

	resultStr := ""
	switch gameResult {
	case messages.GameResultPlayerWin:
		resultStr = AnsiGreen + "You won!" + AnsiReset
	case messages.GameResultPlayerLose:
		resultStr = AnsiRed + "You lost :(" + AnsiReset
	case messages.GameResultDraw:
		resultStr = AnsiBlue + "It's a tie." + AnsiReset
	default:
		panic("server error - game status not accounted for")
	}
	session.sessionToOutput <- resultStr
	session.setState(session.stateInMenu)
}

func (session *Session) WriteToServer(msg messages.ClientMessage) error {
	return session.wsDriver.WriteToServer(msg)
}

func (session *Session) HandlePlayerMessage(msg messages.ClientMessage) error {
	return session.state.handlePlayerMessage(msg)
}

func (session *Session) setState(state SessionState) {
	playerMessage := ""
	switch state.(type) {
	case SessionStateInMenu:
		playerMessage = "Exiting to main menu."
		session.wsDriver.CloseWS()
		session.wsDriver = nil
	case SessionStateWaitingRoom:
		playerMessage = "Entering waiting room."
	case SessionStateInGame:
		playerMessage = "Opponent found, joining game!"
	}
	session.sessionToOutput <- playerMessage
	session.state = state
}

type SessionState interface {
	handleServerMessage(msg messages.ServerMessage) error
	handlePlayerMessage(msg messages.ClientMessage) error
}

type SessionStateInMenu struct {
	session *Session
}

func (state SessionStateInMenu) handleServerMessage(msg messages.ServerMessage) error {
	switch msg.Type {
	case messages.ServerRoomJoined:
		state.session.playerNumber = msg.PlayerNumber
		state.session.setState(state.session.stateWaitingRoom)
	default:
		return fmt.Errorf("unexpected server message type whle in menu: %v", msg.Type)
	}
	return nil
}

func (state SessionStateInMenu) handlePlayerMessage(msg messages.ClientMessage) error {
	switch msg.Type {
	case messages.ClientJoinRoom:
		err := state.session.StartWS(URL)
		if err != nil {
			return err
		}
		err = state.session.WriteToServer(msg)
		if err != nil {
			return err
		}
	default:
		return ValidationError{
			errorMsg: fmt.Sprintf("unexpected player message type while in menu: %v", msg.Type),
		}
	}

	return nil
}

type SessionStateWaitingRoom struct {
	session *Session
}

func (state SessionStateWaitingRoom) handleServerMessage(msg messages.ServerMessage) error {
	switch msg.Type {
	case messages.ServerGameStarted:
		state.session.game = msg.Game
		state.session.playerTurn = msg.PlayerTurn
		state.session.setState(state.session.stateInGame)
		state.session.displayBoardToUser()
	default:
		return fmt.Errorf("unexpected server message type whle in waiting room: %v", msg.Type)
	}

	return nil
}

func (state SessionStateWaitingRoom) handlePlayerMessage(msg messages.ClientMessage) error {
	switch msg.Type {
	case messages.ClientQuitRoom:
		err := state.session.WriteToServer(msg)
		if err != nil {
			return err
		}
		state.session.setState(state.session.stateInMenu)
	default:
		return ValidationError{
			errorMsg: fmt.Sprintf("unexpected player message type while in waiting room: %v", msg.Type),
		}
	}

	return nil
}

type SessionStateInGame struct {
	session *Session
}

func (state SessionStateInGame) handleServerMessage(msg messages.ServerMessage) error {
	switch msg.Type {
	case messages.ServerTurnResult:
		moveFailed := msg.PlayerTurn == state.session.playerTurn
		if moveFailed {
			state.session.sessionToOutput <- "Move invalid - please enter a valid move"
			return nil
		}

		state.session.game = msg.Game
		state.session.playerTurn = msg.PlayerTurn
		state.session.displayBoardToUser()
	case messages.ServerGameFinished:
		state.session.game = msg.Game
		state.session.handleGameOver(msg.GameResult, msg.Message)
	default:
		return fmt.Errorf("unexpected server message type whle in game: %v", msg.Type)
	}

	return nil
}

func (state SessionStateInGame) handlePlayerMessage(msg messages.ClientMessage) error {

	switch msg.Type {
	case messages.ClientSendTurn:
		if state.session.playerNumber != state.session.playerTurn {
			return ValidationError{
				errorMsg: "can not send a move on another players turn",
			}
		}
		err := state.session.WriteToServer(msg)
		if err != nil {
			return err
		}
	case messages.ClientQuitRoom:
		err := state.session.WriteToServer(msg)
		if err != nil {
			return err
		}
		state.session.setState(state.session.stateInMenu)

	default:
		return ValidationError{
			errorMsg: fmt.Sprintf("unexpected player message type while in game: %v", msg.Type),
		}
	}

	return nil
}
