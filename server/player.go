package main

import (
	"fmt"
	"log"

	"github.com/gorilla/websocket"
)

type PlayerMessageType int

const (
	PlayerSendMove PlayerMessageType = iota
)

type RoomChans struct {
	roomToPlayer chan RoomMessage
	playerToRoom chan PlayerMessage
}

type PlayerMessage struct {
	msgType    PlayerMessageType
	turnAction TurnAction
	chans      RoomChans
}

type Player struct {
	notInRoom      PlayerStateNotInRoom
	waitingRoom PlayerStateWaitingRoom
	inRoom         PlayerStateInRoom
	// waitForClose   PlayerStateWaitForClose
	state PlayerState

	conn     *websocket.Conn
	wsClosed bool

	roomRequests chan RoomRequest
	clientRead   chan ClientMessage
	room         RoomChans

	done chan struct{}
}

func NewPlayer(conn *websocket.Conn, roomRequests chan RoomRequest) *Player {
	p := Player{
		conn:         conn,
		roomRequests: roomRequests,
		clientRead:   make(chan ClientMessage),
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
		clientMsg := ClientMessage{}
		err := p.conn.ReadJSON(&clientMsg)
		if err != nil {
			log.Printf("Error occurred while reading message: %v\n", err)
			return
		}

		p.clientRead <- clientMsg
	}
}

func (player *Player) setState(state PlayerState) {
	player.state = state
}

func (player *Player) WriteToClient(msg RoomMessage) error {
	var serverMsg ServerMessage

	switch msg.msgType {
	case RoomJoined:
		serverMsg = ServerMessageRoomJoined{
			Type: ServerRoomJoined,
		}
	case RoomGameStarted:
		serverMsg = ServerMessageGameStarted{
			Type: ServerGameStarted,
			Game: msg.game,
		}
	case RoomTurnResult:
		serverMsg = ServerMessageTurnResult{
			Type: ServerTurnResult,
			Game: msg.game,
		}
	default:
		return fmt.Errorf("unexpected message type: %v", msg.msgType)
	}
	return player.conn.WriteJSON(serverMsg)
}

type PlayerState interface {
	handleClientMessage(cm ClientMessage) error
	handleRoomMessage(rm RoomMessage) error
}

type PlayerStateNotInRoom struct {
	player *Player
}

func (state PlayerStateNotInRoom) handleClientMessage(msg ClientMessage) error {
	switch msg.Type {
	case ClientJoinRoom:
		chans := RoomChans{
			roomToPlayer: make(chan RoomMessage),
			playerToRoom: make(chan PlayerMessage),
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

func (state PlayerStateNotInRoom) handleRoomMessage(msg RoomMessage) error {

	switch msg.msgType {
	case RoomJoined:
		err := state.player.WriteToClient(msg)
		if err != nil {
			//TODO handle shutting down clients
			return err
		}

		state.player.setState(state.player.waitingRoom)

	default:
		return fmt.Errorf("unsupported message type while waiting for room: %v", msg.msgType)
	}

	return nil
}

type PlayerStateWaitingRoom struct {
	player *Player
}

func (state PlayerStateWaitingRoom) handleClientMessage(msg ClientMessage) error {
	switch msg.Type {
	case ClientQuitRoom:
		//TODO IMPLEMENT QUITTING ROOMS
	}

	return nil
}

func (state PlayerStateWaitingRoom) handleRoomMessage(msg RoomMessage) error {
	switch msg.msgType {
	case RoomGameStarted:
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

func (state PlayerStateInRoom) handleClientMessage(msg ClientMessage) error {
	//TODO

	return nil
}

func (state PlayerStateInRoom) handleRoomMessage(msg RoomMessage) error {
	//TODO

	return nil
}
