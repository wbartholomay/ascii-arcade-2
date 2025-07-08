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
	notInRoom       PlayerStateNotInRoom
	waitingRoom     PlayerStateWaitingRoom
	inGameSelection PlayerStateInGameSelection
	inRoom          PlayerStateInRoom
	// waitForClose   PlayerStateWaitForClose
	state PlayerState

	conn     *websocket.Conn
	wsClosed bool

	playerNumber int

	roomRequests chan RoomRequest
	clientRead   chan messages.ClientMessage
	room         RoomChans
}

func NewPlayer(conn *websocket.Conn, roomRequests chan RoomRequest) *Player {
	p := Player{
		conn:         conn,
		roomRequests: roomRequests,
		clientRead:   make(chan messages.ClientMessage),
	}

	p.notInRoom = PlayerStateNotInRoom{&p}
	p.waitingRoom = PlayerStateWaitingRoom{&p}
	p.inGameSelection = PlayerStateInGameSelection{&p}
	p.inRoom = PlayerStateInRoom{&p}

	p.state = p.notInRoom

	return &p
}

func (p *Player) Run() {
	go p.readPump()
	defer log.Printf("Player %v goroutine exited\n", p.playerNumber)
	for {
		select {
		case cm, ok := <-p.clientRead:
			if !ok {
				//client connection closed, handle like client gracefully quitting room
				log.Println("Client connection closed, closing player.")
				p.clientRead = nil
				if p.room.playerToRoom != nil {
					p.room.playerToRoom <- messages.ClientMessage{
						Type: messages.ClientQuitRoom,
					}
				}
				return
			}
			err := p.state.handleClientMessage(cm)
			if err != nil {
				log.Printf("Error while handling client message: %v\n", err)
				return
			}
		case rm, ok := <-p.room.roomToPlayer:
			if !ok {
				//room closed - server error
				//TODO keep player connection alive on game end
				log.Println("Room closed, closing player.")
				p.WriteToClient(messages.ServerMessage{
					Type: messages.ServerRoomDisconnected,
				})
				return
			}
			err := p.state.handleRoomMessage(rm)
			if err != nil {
				log.Printf("Error while handling room message: %v\n", err)
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
	case messages.ServerRoomUnavailable:
		state.player.WriteToClient(msg)
		return fmt.Errorf("player tried to join full room")
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
		state.player.room.playerToRoom <- msg
	default:
		return fmt.Errorf("unsupported message type while waiting for room: %v", msg.Type)
	}

	return nil
}

func (state PlayerStateWaitingRoom) handleRoomMessage(msg messages.ServerMessage) error {
	switch msg.Type {
	case messages.ServerGameFinished:
		state.player.WriteToClient(msg)
		return fmt.Errorf("game ended, closing client. game result: %v", msg.GameResult)
	case messages.ServerEnteredGameSelection:
		err := state.player.WriteToClient(msg)
		if err != nil {
			return err
		}
	}

	state.player.setState(state.player.inGameSelection)
	return nil
}

type PlayerStateInGameSelection struct {
	player *Player
}

func (state PlayerStateInGameSelection) handleClientMessage(msg messages.ClientMessage) error {
	switch msg.Type {
	case messages.ClientQuitRoom:
		state.player.room.playerToRoom <- msg
	case messages.ClientSelectGameType:
		state.player.room.playerToRoom <- msg
	default:
		return fmt.Errorf("unsupported message type while game selection: %v", msg.Type)
	}

	return nil
}

func (state PlayerStateInGameSelection) handleRoomMessage(msg messages.ServerMessage) error {
	switch msg.Type {
	case messages.ServerGameFinished:
		state.player.WriteToClient(msg)
		return fmt.Errorf("game ended, closing client. game result: %v", msg.GameResult)
	case messages.ServerGameStarted:
		err := state.player.WriteToClient(msg)
		if err != nil {
			return err
		}
		state.player.setState(state.player.inRoom)
	default:
		return fmt.Errorf("unsupported message type while in game selection: %v", msg.Type)
	}

	return nil
}

type PlayerStateInRoom struct {
	player *Player
}

func (state PlayerStateInRoom) handleClientMessage(msg messages.ClientMessage) error {
	switch msg.Type {
	case messages.ClientSendTurn:
		state.player.room.playerToRoom <- msg
	case messages.ClientQuitRoom:
		state.player.room.playerToRoom <- msg
	default:
		return fmt.Errorf("unsupported message type while in room: %v", msg.Type)

	}

	return nil
}

func (state PlayerStateInRoom) handleRoomMessage(msg messages.ServerMessage) error {
	switch msg.Type {
	case messages.ServerGameFinished:
		state.player.WriteToClient(msg)
		return fmt.Errorf("game ended, closing client. game result: %v", msg.GameResult)
	case messages.ServerTurnResult, messages.ServerError:
		err := state.player.WriteToClient(msg)
		if err != nil {
			return err
		}
	default:
		return fmt.Errorf("unsupported message type while in room: %v", msg.Type)
	}

	return nil
}
