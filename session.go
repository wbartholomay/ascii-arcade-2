package main

import (
	"fmt"

	"github.com/wbarthol/ascii-arcade-2/internal/game"
	"github.com/wbarthol/ascii-arcade-2/internal/messages"
)

type ValidationError struct {
	errorMsg string
}

func (err ValidationError) Error() string {
	return err.errorMsg
}

type Session struct {
	stateInMenu          SessionStateInMenu
	stateWaitingRoom     SessionStateWaitingRoom
	stateInGameSelection SessionStateInGameSelection
	stateInGame          SessionStateInGame
	state                SessionState

	playerNumber int
	playerTurn   int

	gameType game.GameType
	game     game.Game

	sessionToOutput chan string
	driverToSession chan messages.ServerMessage
	wsDriver        *WSDriver
	serverUrl       string
}

func NewSession(serverUrl string) *Session {
	//Could move the dialing to StartWS
	session := Session{
		sessionToOutput: make(chan string, 10),
		serverUrl:       serverUrl,
	}

	session.stateInMenu = SessionStateInMenu{
		session: &session,
	}
	session.stateInGameSelection = SessionStateInGameSelection{
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
	session.sessionToOutput <- session.game.DisplayBoard(session.playerNumber)
	if session.isPlayerTurn() {
		session.sessionToOutput <- "Your turn, make a move."
	} else {
		session.sessionToOutput <- "Waiting on opponents move..."
	}
}

func (session *Session) handleRoomClosure(msg messages.ServerMessage) {
	if session.state == session.stateInGame {
		session.handleGameOver(msg.GameResult, msg.Message)
		return
	}

	session.sessionToOutput <- "a player has quit, closing the room."
	session.setState(session.stateInMenu)
}

func (session *Session) handleGameOver(gameResult messages.GameResult, detailsFromServer string) {
	session.sessionToOutput <- session.game.DisplayBoard(session.playerNumber)
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
	switch state.(type) {
	case SessionStateInMenu:
		session.sessionToOutput <- "Exiting to main menu."
		session.sessionToOutput <- "Welcome to ASCII arcade! Enter \033[33mjoin <room-code>\033[0m to create/join a room, or enter \033[33mhelp\033[0m to see a list of commands."
		session.wsDriver.CloseWS()
		session.wsDriver = nil
	case SessionStateWaitingRoom:
		session.sessionToOutput <- "Entering waiting room."
	case SessionStateInGameSelection:
		if session.playerNumber == 1 {
			session.sessionToOutput <- fmt.Sprintf("%vSelect%v a game \n%v1.%v TicTacToe\n%v2%v. Checkers", AnsiYellow, AnsiReset, AnsiYellow, AnsiReset, AnsiYellow, AnsiReset)
		} else {
			session.sessionToOutput <- "Waiting on Player 1 to select a game..."
		}
	case SessionStateInGame:
		session.sessionToOutput <- "Opponent found, joining game!"
		session.sessionToOutput <- session.game.GetGameInstructions()
	}

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
		err := state.session.StartWS(state.session.serverUrl)
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
	case messages.ServerGameFinished:
		state.session.game = msg.Game.GetGame()
		state.session.handleRoomClosure(msg)
	case messages.ServerEnteredGameSelection:
		state.session.setState(state.session.stateInGameSelection)
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

type SessionStateInGameSelection struct {
	session *Session
}

func (state SessionStateInGameSelection) handleServerMessage(msg messages.ServerMessage) error {
	switch msg.Type {
	case messages.ServerGameStarted:
		state.session.game = msg.Game.GetGame()
		state.session.gameType = state.session.game.GetGameType()
		state.session.playerTurn = msg.PlayerTurn
		state.session.setState(state.session.stateInGame)
		state.session.displayBoardToUser()
	case messages.ServerGameFinished:
		state.session.game = msg.Game.GetGame()
		state.session.handleRoomClosure(msg)
	}
	return nil
}

func (state SessionStateInGameSelection) handlePlayerMessage(msg messages.ClientMessage) error {
	switch msg.Type {
	case messages.ClientQuitRoom:
		err := state.session.WriteToServer(msg)
		if err != nil {
			return err
		}
		state.session.setState(state.session.stateInMenu)
	case messages.ClientSelectGameType:
		if state.session.playerNumber != 1 {
			return ValidationError{
				errorMsg: "only player 1 can select a game",
			}
		}
		err := state.session.WriteToServer(msg)
		if err != nil {
			return err
		}
	default:
		return ValidationError{
			errorMsg: fmt.Sprintf("unexpected player message type while in game: %v", msg.Type),
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
			state.session.sessionToOutput <- AnsiRed + "Move invalid - " + msg.Message + AnsiReset
			return nil
		}

		state.session.game = msg.Game.GetGame()
		state.session.playerTurn = msg.PlayerTurn
		if msg.Message != "" {
			state.session.sessionToOutput <- AnsiBlue + msg.Message + AnsiReset
		}
		state.session.displayBoardToUser()
	case messages.ServerGameFinished:
		state.session.game = msg.Game.GetGame()
		state.session.handleRoomClosure(msg)
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
