package main

import (
	"fmt"
	"log"

	"github.com/gorilla/websocket"
	"github.com/wbarthol/ascii-arcade-2/internal/messages"
)

type PlayerMessageType int

const (
	PlayerSendMove PlayerMessageType = iota
)

type RoomChans struct {
	roomToPlayer chan messages.ServerMessage
	playerToRoom chan messages.ClientMessage
}

type Player struct {
	notInRoom   PlayerStateNotInRoom
	waitingRoom PlayerStateWaitingRoom
	inRoom      PlayerStateInRoom
	// waitForClose   PlayerStateWaitForClose
	state PlayerState

	conn     *websocket.Conn
	wsClosed bool

	playerNumber int

	roomRequests chan RoomRequest
	clientRead   chan messages.ClientMessage
	room         RoomChans

	done chan struct{}
}

func NewPlayer(conn *websocket.Conn, roomRequests chan RoomRequest) *Player {
	p := Player{
		conn:         conn,
		roomRequests: roomRequests,
		clientRead:   make(chan messages.ClientMessage),
	}

	p.notInRoom = PlayerStateNotInRoom{&p}
	p.waitingRoom = PlayerStateWaitingRoom{&p}
	p.inRoom = PlayerStateInRoom{&p}

	p.state = p.notInRoom

	return &p
}

func (p *Player) Run() {
	go p.readPump()
	defer fmt.Println("player goroutine exited")
	for {
		select {
		case cm, ok := <-p.clientRead:
			if !ok {
				fmt.Println()
				//error occurred when reading from client
				//TODO: quit room
				return
			}
			err := p.state.handleClientMessage(cm)
			if err != nil {
				fmt.Printf("Error while handling client message: %v\n", err)
				return
			}
		case rm, ok := <-p.room.roomToPlayer:
			if !ok {
				//TODO
				fmt.Println("Room channel closed")
			}
			err := p.state.handleRoomMessage(rm)
			if err != nil {
				fmt.Printf("Error while handling room message: %v\n", err)
				return
			}
		}
	}
}

func (p *Player) readPump() {
	defer func() {
		log.Println("Shutting down player.")
		p.conn.Close()
		close(p.clientRead)
	}()

	for {
		clientMsg := messages.ClientMessage{}
		err := p.conn.ReadJSON(&clientMsg)
		if err != nil {
			log.Printf("Error occurred while reading message: %v\n", err)
			return
		}

		log.Printf("Received message from Client: %v\n", clientMsg)
		p.clientRead <- clientMsg
	}
}

func (player *Player) setState(state PlayerState) {
	player.state = state
}

func (player *Player) WriteToClient(msg messages.ServerMessage) error {
	log.Printf("Sending message to client %v: %v", player.playerNumber, msg)
	return player.conn.WriteJSON(msg)
}

type PlayerState interface {
	handleClientMessage(cm messages.ClientMessage) error
	handleRoomMessage(rm messages.ServerMessage) error
}

type PlayerStateNotInRoom struct {
	player *Player
}

func (state PlayerStateNotInRoom) handleClientMessage(msg messages.ClientMessage) error {
	switch msg.Type {
	case messages.ClientJoinRoom:
		chans := RoomChans{
			roomToPlayer: make(chan messages.ServerMessage),
			playerToRoom: make(chan messages.ClientMessage),
		}

		state.player.room = chans

		go func(chans RoomChans) {
			state.player.roomRequests <- RoomRequest{
				code:  msg.RoomCode,
				chans: chans,
			}
		}(chans)

		log.Printf("Player waiting for room. Room code: %v", msg.RoomCode)
	default:
		return fmt.Errorf("unsupported message type while waiting for room: %v", msg.Type)
	}

	return nil
}

func (state PlayerStateNotInRoom) handleRoomMessage(msg messages.ServerMessage) error {

	switch msg.Type {
	case messages.ServerRoomJoined:
		state.player.playerNumber = msg.PlayerNumber
		err := state.player.WriteToClient(msg)
		if err != nil {
			//TODO handle shutting down clients
			return err
		}

		state.player.setState(state.player.waitingRoom)

	default:
		return fmt.Errorf("unsupported message type while waiting for room: %v", msg.Type)
	}

	return nil
}

type PlayerStateWaitingRoom struct {
	player *Player
}

func (state PlayerStateWaitingRoom) handleClientMessage(msg messages.ClientMessage) error {
	switch msg.Type {
	case messages.ClientQuitRoom:
		//TODO IMPLEMENT QUITTING ROOMS
	}

	return nil
}

func (state PlayerStateWaitingRoom) handleRoomMessage(msg messages.ServerMessage) error {
	switch msg.Type {
	case messages.ServerGameStarted:
		err := state.player.WriteToClient(msg)
		if err != nil {
			return err
		}
	}

	state.player.setState(state.player.inRoom)
	return nil
}

type PlayerStateInRoom struct {
	player *Player
}

func (state PlayerStateInRoom) handleClientMessage(msg messages.ClientMessage) error {
	switch msg.Type {
	case messages.ClientSendTurn:
		state.player.room.playerToRoom <- msg
	default:
		return fmt.Errorf("unsupported message type while in room: %v", msg.Type)

	}

	return nil
}

func (state PlayerStateInRoom) handleRoomMessage(msg messages.ServerMessage) error {
	switch msg.Type {
	case messages.ServerTurnResult:
		err := state.player.WriteToClient(msg)
		if err != nil {
			return err
		}
	default:
		return fmt.Errorf("unsupported message type while in room: %v", msg.Type)
	}

	return nil
}
