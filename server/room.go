package main

import (
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
	closeReq chan string
}

func NewRoom(code string, closeReq chan string) *Room {
	room := &Room{
		code: code,
	}
	room.waitingForPlayerOne = RoomStateWaitingForP1{room}
	room.waitingForPlayerTwo = RoomStateWaitingForP2{room}
	room.running = RoomStateRunning{room}
	room.state = &room.waitingForPlayerOne
	room.requests = make(chan RoomRequest)
	room.closeReq = closeReq
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
	defer func(){
		room.closeReq <- room.code
	}()
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
			err := room.state.handlePlayerMessage(msg, 1)
			if err != nil {
				log.Printf("error while handling player message, closing room: %v", err)
				return
			}
		case msg := <-room.playerTwoChans.playerToRoom:
			err := room.state.handlePlayerMessage(msg, 2)
			if err != nil {
				log.Printf("error while handling player message, closing room: %v", err)
				return
			}
		}
	}
}

// onQuit - Sends message to players who did not quit, informing them of game completion.
func (room *Room) endGameOnQuit(quittingPlayerNum int) {
	roomNotRunning := room.state == room.waitingForPlayerOne || room.state == room.waitingForPlayerTwo
	if roomNotRunning {
		return
	}

	if quittingPlayerNum == 1 {
		p2Message := messages.ServerMessage{
			Type:       messages.ServerGameFinished,
			Game:       room.game,
			GameResult: messages.GameResultPlayerWin,
			Message:    "Player 1 quit.",
		}
		room.playerTwoChans.roomToPlayer <- p2Message
	} else {
		p1Message := messages.ServerMessage{
			Type:       messages.ServerGameFinished,
			Game:       room.game,
			GameResult: messages.GameResultPlayerWin,
			Message:    "Player 2 quit.",
		}
		room.playerOneChans.roomToPlayer <- p1Message
	}
}

func (room *Room) endGameOnCompletion() {
	p1Message := messages.ServerMessage{
		Type: messages.ServerGameFinished,
		Game: room.game,
	}
	p2Message := messages.ServerMessage{
		Type: messages.ServerGameFinished,
		Game: room.game,
	}
	switch room.game.GameStatus {
	case tictactoe.GameStatusDraw:
			p1Message.GameResult, p2Message.GameResult = messages.GameResultDraw, messages.GameResultDraw
	case tictactoe.GameStatusPlayer1Win:
		p1Message.GameResult, p2Message.GameResult = messages.GameResultPlayerWin, messages.GameResultPlayerLose
	case tictactoe.GameStatusPlayer2Win:
		p1Message.GameResult, p2Message.GameResult = messages.GameResultPlayerLose, messages.GameResultPlayerWin
	}

	room.playerOneChans.roomToPlayer <- p1Message
	room.playerTwoChans.roomToPlayer <- p2Message

	//shut down EVERTYTHIGN
	// close(room.playerOneChans.roomToPlayer)
	// close(room.playerTwoChans.roomToPlayer)
	// room.playerOneChans.playerToRoom = nil
	// room.playerTwoChans.playerToRoom = nil

	// ch := s.r.close
	// go func(ch chan string) {
	// 	ch <- s.r.code
	// }(ch)
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
	if playerNumber == 2 {
		return fmt.Errorf("should be no messages from player two while waiting for p2")
	}
	switch msg.Type {
	case messages.ClientQuitRoom:
		state.room.endGameOnQuit(playerNumber)
		return fmt.Errorf("player %v quit", playerNumber)
	}
	return nil
}

type RoomStateRunning struct {
	room *Room
}

func (state RoomStateRunning) handleJoinRequest(req RoomRequest) error {
	//TODO: Reject join request
	req.chans.roomToPlayer <- messages.ServerMessage{
		Type: messages.ServerRoomUnavailable,
	}
	return nil
}

func (state RoomStateRunning) handlePlayerMessage(msg messages.ClientMessage, playerNumber int) error {
	if playerNumber != state.room.playerTurn {
		return fmt.Errorf("received a message from player when it is not their turn")
	}

	switch msg.Type {
	case messages.ClientQuitRoom:
		state.room.endGameOnQuit(playerNumber)
		return fmt.Errorf("player %v quit", playerNumber)

	case messages.ClientSendTurn:
		//TODO could make this more readable
		isMoveValid := state.room.game.ValidateMove(msg.TurnAction)
		if isMoveValid {
			state.room.game.ExecuteTurn(msg.TurnAction, playerNumber)
			state.room.advanceTurn()
		}
		if state.room.game.GameStatus != tictactoe.GameStatusOngoing {
			state.room.endGameOnCompletion()
			return fmt.Errorf("game over, sent results to clients")
		}

		serverMsg := messages.ServerMessage{
			Type:       messages.ServerTurnResult,
			Game:       state.room.game,
			PlayerTurn: state.room.playerTurn,
		}
		if !isMoveValid && playerNumber == 1 {
			state.room.playerOneChans.roomToPlayer <- serverMsg
		} else if !isMoveValid && playerNumber == 2 {
			state.room.playerTwoChans.roomToPlayer <- serverMsg
		} else {
			state.room.playerOneChans.roomToPlayer <- serverMsg
			state.room.playerTwoChans.roomToPlayer <- serverMsg
		}
	}

	return nil
}
