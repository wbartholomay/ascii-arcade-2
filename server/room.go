package main

import (
	"errors"
	"log"

	tictactoe "github.com/wbarthol/ascii-arcade-2/internal/tic_tac_toe"
)

type RoomMessageType int

const (
	RoomJoined RoomMessageType = iota
	RoomGameStarted
	RoomTurnResult
	RoomGameFinished
)

type RoomMessage struct {
	msgType         RoomMessageType
	game            tictactoe.TicTacToeGame
	isPlayerOneTurn bool
}

type Room struct {
	code string

	game tictactoe.TicTacToeGame

	waitingForPlayerOne RoomStateWaitingForP1
	waitingForPlayerTwo RoomStateWaitingForP2
	running             RoomStateRunning
	state               RoomState

	playerOneChans RoomChans
	playerTwoChans RoomChans

	requests chan RoomRequest
}

func NewRoom(code string) *Room {
	room := &Room{
		code: code,
	}
	room.waitingForPlayerOne = RoomStateWaitingForP1{room}
	room.waitingForPlayerTwo = RoomStateWaitingForP2{room}
	room.running = RoomStateRunning{room}
	room.state = &room.waitingForPlayerOne
	room.requests = make(chan RoomRequest)
	return room
}

func (room *Room) SetState(state RoomState) {
	room.state = state
}

func (room *Room) Run() {
	for {
		select {
		//TODO hanlde close requests
		case joinRequest := <-room.requests:
			err := room.state.handleJoinRequest(joinRequest)
			if err != nil {
				//TODO
				log.Printf("error encountered while handling join request: %v", err)
				return
			}
		case msg := <-room.playerOneChans.playerToRoom:
			room.state.handlePlayerMessage(msg, true)
		case msg := <-room.playerTwoChans.playerToRoom:
			room.state.handlePlayerMessage(msg, false)
		}
	}
}

type RoomState interface {
	handleJoinRequest(req RoomRequest) error
	handlePlayerMessage(msg PlayerMessage, isPlayerOne bool) error
}

type RoomStateWaitingForP1 struct {
	room *Room
}

func (state RoomStateWaitingForP1) handleJoinRequest(req RoomRequest) error {
	state.room.playerOneChans = req.chans
	state.room.SetState(state.room.waitingForPlayerTwo)

	state.room.playerOneChans.roomToPlayer <- RoomMessage{msgType: RoomJoined}
	return nil
}

func (state RoomStateWaitingForP1) handlePlayerMessage(msg PlayerMessage, isPlayerOne bool) error {
	panic("should be no player messages while waiting for p1")
}

type RoomStateWaitingForP2 struct {
	room *Room
}

func (state RoomStateWaitingForP2) handleJoinRequest(req RoomRequest) error {
	state.room.playerTwoChans = req.chans
	state.room.playerTwoChans.roomToPlayer <- RoomMessage{msgType: RoomJoined}

	state.room.game = tictactoe.NewTicTacToeGame()

	state.room.playerOneChans.roomToPlayer <- RoomMessage{msgType: RoomGameStarted, game: state.room.game}
	state.room.playerTwoChans.roomToPlayer <- RoomMessage{msgType: RoomGameStarted, game: state.room.game}

	state.room.SetState(state.room.running)
	log.Println("Player two joined room")
	return nil
}

func (state RoomStateWaitingForP2) handlePlayerMessage(msg PlayerMessage, isPlayerOne bool) error {
	if !isPlayerOne {
		panic("should be no messages from player two while waiting for p2")
	}
	//SWITCH ON PLAYER ONES MESSAGES HERE, NEED TO HANDLE P1 QUITTING (SET STATE TO CLOSING)
	return nil
}

type RoomStateRunning struct {
	room *Room
}

func (state RoomStateRunning) handleJoinRequest(req RoomRequest) error {
	//TODO: Reject join request
	return errors.New("room is already running, can not join")
}

func (state RoomStateRunning) handlePlayerMessage(msg PlayerMessage, isPlayerOne bool) error {
	// if isPlayerOne != state.room.gameState.isPlayerOneTurn {
	// 	return fmt.Errorf("received a message from player when it is not there turn")
	// }

	// switch msg.msgType {
	// case PlayerSendMove:
	// 	//TODO need to maintain player turn in gamestate
	// 	turnSuccess := state.room.gameState.ExecuteTurn(msg.turnAction, isPlayerOne)
	// 	if turnSuccess {
	// 		state.room.gameState.isPlayerOneTurn = !state.room.gameState.isPlayerOneTurn
	// 	}

	// 	cfg1, cfg2 := state.room.gameState.CreateClientConfigs()

	// 	state.room.playerOneChans.read <- RoomMessage{
	// 		msgType:         RoomTurnResult,
	// 		config:          cfg1,
	// 		isPlayerOneTurn: state.room.gameState.isPlayerOneTurn,
	// 	}

	// 	state.room.playerTwoChans.read <- RoomMessage{
	// 		msgType:         RoomTurnResult,
	// 		config:          cfg2,
	// 		isPlayerOneTurn: state.room.gameState.isPlayerOneTurn,
	// 	}
	// }

	return nil
}
