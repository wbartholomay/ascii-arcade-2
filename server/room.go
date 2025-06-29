package main

import (
	"errors"
	"fmt"
	"log"

	"github.com/wbarthol/ascii-arcade-2/internal/messages"
	"github.com/wbarthol/ascii-arcade-2/internal/tictactoe"
)

type Room struct {
	code string

	game       tictactoe.TicTacToeGame
	playerTurn int

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

func (room *Room) advanceTurn() {
	if room.playerTurn == 1 {
		room.playerTurn = 2
	} else {
		room.playerTurn = 1
	}
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
			room.state.handlePlayerMessage(msg, 1)
		case msg := <-room.playerTwoChans.playerToRoom:
			room.state.handlePlayerMessage(msg, 2)
		}
	}
}

type RoomState interface {
	handleJoinRequest(req RoomRequest) error
	handlePlayerMessage(msg messages.ClientMessage, playerNumber int) error
}

type RoomStateWaitingForP1 struct {
	room *Room
}

func (state RoomStateWaitingForP1) handleJoinRequest(req RoomRequest) error {
	state.room.playerOneChans = req.chans
	state.room.SetState(state.room.waitingForPlayerTwo)

	state.room.playerOneChans.roomToPlayer <- messages.ServerMessage{
		Type:         messages.ServerRoomJoined,
		PlayerNumber: 1,
	}
	return nil
}

func (state RoomStateWaitingForP1) handlePlayerMessage(msg messages.ClientMessage, playerNumber int) error {
	panic("should be no player messages while waiting for p1")
}

type RoomStateWaitingForP2 struct {
	room *Room
}

func (state RoomStateWaitingForP2) handleJoinRequest(req RoomRequest) error {
	state.room.playerTwoChans = req.chans
	state.room.playerTwoChans.roomToPlayer <- messages.ServerMessage{
		Type:         messages.ServerRoomJoined,
		PlayerNumber: 2,
	}

	state.room.game = tictactoe.NewTicTacToeGame()
	state.room.playerTurn = 1

	state.room.playerOneChans.roomToPlayer <- messages.ServerMessage{Type: messages.ServerGameStarted,
		Game:       state.room.game,
		PlayerTurn: 1,
	}
	state.room.playerTwoChans.roomToPlayer <- messages.ServerMessage{
		Type:       messages.ServerGameStarted,
		Game:       state.room.game,
		PlayerTurn: 1,
	}

	state.room.SetState(state.room.running)
	log.Println("Player two joined room, game started!")
	return nil
}

func (state RoomStateWaitingForP2) handlePlayerMessage(msg messages.ClientMessage, playerNumber int) error {
	if playerNumber != 2 {
		//TODO handle this better than panicking
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

func (state RoomStateRunning) handlePlayerMessage(msg messages.ClientMessage, playerNumber int) error {
	if playerNumber != state.room.playerTurn {
		return fmt.Errorf("received a message from player when it is not their turn")
	}

	switch msg.Type {
	case messages.ClientSendTurn:
		//TODO need to maintain player turn in gamestate
		state.room.game.ExecuteTurn(msg.TurnAction, playerNumber)
		state.room.advanceTurn()

		state.room.playerOneChans.roomToPlayer <- messages.ServerMessage{
		Type:       messages.ServerTurnResult,
		Game:       state.room.game,
		PlayerTurn: state.room.playerTurn,
	}
		state.room.playerTwoChans.roomToPlayer <-  messages.ServerMessage{
		Type:       messages.ServerTurnResult,
		Game:       state.room.game,
		PlayerTurn: state.room.playerTurn,
	}
	}

	return nil
}
