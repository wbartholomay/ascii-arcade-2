package main

import (
	"errors"
	"fmt"
	"slices"

	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/wbarthol/ascii-arcade-2/internal/game"
	"github.com/wbarthol/ascii-arcade-2/internal/messages"
	"github.com/wbarthol/ascii-arcade-2/internal/vector"
)

type ValidationError struct {
	errorMsg string
}

func (err ValidationError) Error() string {
	return err.errorMsg
}

type ServerMsg struct {
	msg          messages.ServerMessage
	serverClosed bool
}

type SentClientMsg struct{}

type ErrMsg struct {
	err error
}

func (msg ErrMsg) Error() string {
	return msg.err.Error()
}

type Session struct {
	state SessionState

	roomCode string

	waitingForServerResponse bool
	errMsg                   string

	playerNumber int
	playerTurn   int

	gameType   game.GameType
	game       game.Game
	gameResult messages.GameResult

	driverToSession chan messages.ServerMessage
	wsDriver        *WSDriver
	serverUrl       string
}

func NewSession(serverUrl string) Session {
	session := Session{
		serverUrl:       serverUrl,
		driverToSession: make(chan messages.ServerMessage),
	}
	session.state = NewSessionStateInMenu()

	return session
}

func (session Session) StartWS(url string) (Session, error) {
	wsDriver, err := NewWS(url, &session)
	if err != nil {
		return session, fmt.Errorf("error starting WS: %w", err)
	}
	session.wsDriver = wsDriver
	session.driverToSession = wsDriver.driverToSession
	go session.wsDriver.Run()
	return session, nil
}

func (session Session) ListenToServer() tea.Cmd {
	//Decoupling this from WSDriver with singleplayer in mind
	return func() tea.Msg {
		msg, ok := <-session.driverToSession
		if !ok {
			if session.state.GetType() != SessionStateTypeInMenu {
				return ServerMsg{msg: msg, serverClosed: true}
			}
		}
		return ServerMsg{msg: msg}
	}
}

func (session Session) Init() tea.Cmd {
	return session.ListenToServer()
}

func (session Session) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		session.errMsg = ""
		switch msg.String() {
		case "ctrl+c":
			return session, tea.Quit
		default:
			//TODO if waitingForServerResponse, input should be rejected and error msg sent to client
			return session.state.HandleUserInput(msg, session)
		}
	case ServerMsg:
		session.waitingForServerResponse = false
		if msg.serverClosed {
			//TODO Notify client the server closed
			session = session.setState(SessionStateTypeInMenu)
		} else {
			var err error
			session, err = session.state.handleServerMessage(session, msg.msg)
			if err != nil {
				session.errMsg = err.Error()
			}
		}
		return session, session.ListenToServer()
	case SentClientMsg:
		session.waitingForServerResponse = true
	case ErrMsg:
		session.errMsg = msg.Error()
	}

	return session, nil
}

func (session Session) View() string {
	content := session.state.GetDisplayString()

	var parts []string
	parts = append(parts, content)
	if session.errMsg != "" {
		errorStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF3B30")).
			Padding(0, 1)
		parts = append(parts, errorStyle.Render(session.errMsg))
	}

	return lipgloss.JoinVertical(lipgloss.Left, parts...)
}

func (session Session) SendMsgToServer(msg messages.ClientMessage) tea.Cmd {
	return func() tea.Msg {
		err := session.WriteToServer(msg)
		//TODO figure out error handling
		if err != nil {
			return ErrMsg{err}
		}
		//TODO what to return?
		return SentClientMsg{}
	}
}

func (session Session) handleRoomClosure(msg messages.ServerMessage) Session {
	//TODO should move thsi outside of thsi function
	// if session.state.GetType() == SessionStateTypeInGame {
	// 	session = session.handleGameOver(msg.GameResult, msg.QuittingPlayerNum)
	// 	return session
	// }

	session.errMsg = "A player has quit, closing the room."
	session = session.setState(SessionStateTypeInMenu)
	return session
}

func (session Session) handleGameOver(gameResult messages.GameResult, quittingPlayerNum int) Session {

	resultStr := ""
	switch gameResult {
	case messages.GameResultPlayerWin:
		resultStr += AnsiGreen + "You won!" + AnsiReset
	case messages.GameResultPlayerLose:
		resultStr += AnsiRed + "You lost :(" + AnsiReset
	case messages.GameResultDraw:
		resultStr += AnsiBlue + "It's a tie." + AnsiReset
	default:
		panic("server error - game status not accounted for")
	}
	session = session.setState(SessionStateTypeInMenu)
	//TODO need other way to display variable information
	session.errMsg = resultStr
	return session
}

func (session Session) WriteToServer(msg messages.ClientMessage) error {
	return session.wsDriver.WriteToServer(msg)
}

func (session Session) setState(state SessionStateType) Session {
	switch state {
	case SessionStateTypeInMenu:
		session.wsDriver.CloseWS()
		session.wsDriver = nil
		session.driverToSession = make(chan messages.ServerMessage)
		session.state = NewSessionStateInMenu()
	case SessionStateTypeWaitingRoom:
		acceptableStates := []SessionStateType{SessionStateTypeInMenu, SessionStateTypeEndGame}
		if !slices.Contains(acceptableStates, session.state.GetType()) {
			panic(fmt.Sprintf("Unexpected state when transitioning to waiting room: %v", session.state.GetType()))
		}
		session.state = NewSessionStateWaitingRoom(session.roomCode)
	case SessionStateTypeGameSelection:
		if session.state.GetType() != SessionStateTypeWaitingRoom {
			panic(fmt.Sprintf("Unexpected state when transitioning to game selection: %v", session.state.GetType()))
		}
		session.state = NewSessionStateInGameSelection(session.playerNumber)
	case SessionStateTypeInGame:
		if session.state.GetType() != SessionStateTypeGameSelection {
			panic(fmt.Sprintf("Unexpected state when transitioning to in game: %v", session.state.GetType()))
		}
		session.state = NewSessionStateInGame(session.playerNumber, session.playerTurn, session.game)
	case SessionStateTypeEndGame:
		if session.state.GetType() != SessionStateTypeInGame {
			panic(fmt.Sprintf("Unexpected state when transitioning to end game: %v", session.state.GetType()))
		}
		session.state = NewSessionStateEndGame(session.game, session.gameResult, session.playerNumber)
	}
	return session
}

type SessionStateType int

const (
	SessionStateTypeInMenu SessionStateType = iota
	SessionStateTypeWaitingRoom
	SessionStateTypeGameSelection
	SessionStateTypeInGame
	SessionStateTypeEndGame
)

func (sType SessionStateType) String() string {
	switch sType {
	case SessionStateTypeInMenu:
		return "In Menu"
	case SessionStateTypeWaitingRoom:
		return "Waiting Room"
	case SessionStateTypeGameSelection:
		return "Game Selection"
	case SessionStateTypeInGame:
		return "In Game"
	case SessionStateTypeEndGame:
		return "End Game"
	default:
		return "Unknown"
	}
}

type SessionState interface {
	GetType() SessionStateType
	GetDisplayString() string
	HandleUserInput(msg tea.KeyMsg, session Session) (tea.Model, tea.Cmd)
	handleServerMessage(session Session, msg messages.ServerMessage) (Session, error)
}

type SessionStateInMenu struct {
	textArea textarea.Model
}

func (SessionState SessionStateInMenu) GetType() SessionStateType {
	return SessionStateTypeInMenu
}

func NewSessionStateInMenu() *SessionStateInMenu {
	textArea := textarea.New()
	textArea.Placeholder = "Enter room code."
	textArea.Focus()
	textArea.Prompt = " "
	textArea.CharLimit = 5
	textArea.SetWidth(30)
	textArea.SetHeight(1)
	textArea.FocusedStyle.CursorLine = lipgloss.NewStyle()
	textArea.ShowLineNumbers = false
	textArea.KeyMap.InsertNewline.SetEnabled(false)
	return &SessionStateInMenu{textArea: textArea}
}

func (state *SessionStateInMenu) HandleUserInput(msg tea.KeyMsg, session Session) (tea.Model, tea.Cmd) {
	var (
		tiCmd     tea.Cmd
		serverCmd tea.Cmd
	)
	state.textArea, tiCmd = state.textArea.Update(msg)

	switch msg.String() {
	case "enter":
		if state.textArea.Value() == "" {
			return session, func() tea.Msg { return ErrMsg{errors.New("Please enter a code.")} }
		}
		session, err := session.StartWS(session.serverUrl)
		if err != nil {
			return session, func() tea.Msg { return ErrMsg{err} }
		}
		session.roomCode = state.textArea.Value()
		joinMsg := messages.ClientMessage{
			Type:     messages.ClientJoinRoom,
			RoomCode: session.roomCode,
		}
		serverCmd = session.SendMsgToServer(joinMsg)
		return session, tea.Batch(tiCmd, serverCmd)
	default:
		return session, tea.Batch(tiCmd, serverCmd)
	}
}

func (state SessionStateInMenu) GetDisplayString() string {
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FAFAFA")).
		Background(lipgloss.Color("#7D56F4")).
		Padding(0, 1).
		MarginBottom(1)

	instructionStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#626262")).
		MarginBottom(2)

	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#874BFD")).
		Padding(1).
		MarginTop(1)

	title := titleStyle.Render("🎮 ASCII ARCADE")
	instruction := instructionStyle.Render("Enter a room code to join or create a game room")
	textBox := boxStyle.Render(state.textArea.View())

	return lipgloss.JoinVertical(lipgloss.Left, title, instruction, textBox)
}

func (state *SessionStateInMenu) handleServerMessage(session Session, msg messages.ServerMessage) (Session, error) {
	switch msg.Type {
	case messages.ServerRoomJoined:
		session.playerNumber = msg.PlayerNumber
		session = session.setState(SessionStateTypeWaitingRoom)
	default:
		return session, fmt.Errorf("unexpected server message type whle in menu: %v", msg.Type)
	}
	return session, nil
}

type SessionStateWaitingRoom struct {
	roomCode string
}

func (SessionState SessionStateWaitingRoom) GetType() SessionStateType {
	return SessionStateTypeWaitingRoom
}

func NewSessionStateWaitingRoom(roomCode string) *SessionStateWaitingRoom {
	return &SessionStateWaitingRoom{
		roomCode: roomCode,
	}
}

func (state *SessionStateWaitingRoom) HandleUserInput(msg tea.KeyMsg, session Session) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q":
		return session, session.SendMsgToServer(messages.ClientMessage{
			Type: messages.ClientQuitRoom,
		})
	default:
		return session, nil
	}
}

func (state SessionStateWaitingRoom) GetDisplayString() string {
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FAFAFA")).
		Background(lipgloss.Color("#FF6B35")).
		Padding(0, 1).
		MarginBottom(1)

	statusStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#626262")).
		Italic(true).
		MarginBottom(1)

	instructionStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#9CA3AF")).
		Border(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("#6B7280")).
		Padding(1).
		MarginTop(1)

	title := titleStyle.Render("WAITING ROOM | ROOM CODE: " + state.roomCode)
	status := statusStyle.Render("Waiting for another player to join...")
	instruction := instructionStyle.Render("Press 'q' to quit and return to main menu")

	return lipgloss.JoinVertical(lipgloss.Left, title, status, instruction)
}

func (state *SessionStateWaitingRoom) handleServerMessage(session Session, msg messages.ServerMessage) (Session, error) {
	switch msg.Type {
	case messages.ServerRoomClosed:
		session = session.handleRoomClosure(msg)
	case messages.ServerEnteredGameSelection:
		session = session.setState(SessionStateTypeGameSelection)
	default:
		return session, fmt.Errorf("unexpected server message type whle in waiting room: %v", msg.Type)
	}

	return session, nil
}

type SessionStateInGameSelection struct {
	cursor    int
	playerNum int
}

func (SessionState SessionStateInGameSelection) GetType() SessionStateType {
	return SessionStateTypeGameSelection
}
func NewSessionStateInGameSelection(playerNum int) *SessionStateInGameSelection {
	return &SessionStateInGameSelection{
		cursor:    0,
		playerNum: playerNum,
	}
}

func (state *SessionStateInGameSelection) HandleUserInput(msg tea.KeyMsg, session Session) (tea.Model, tea.Cmd) {
	if session.playerNumber != 1 {
		switch msg.String() {
		case "q":
			return session, session.SendMsgToServer(messages.ClientMessage{
				Type: messages.ClientQuitRoom,
			})
		default:
			return session, nil
		}
	}

	switch msg.String() {
	case "q":
		return session, session.SendMsgToServer(messages.ClientMessage{
			Type: messages.ClientQuitRoom,
		})
	case "up", "k", "w":
		if state.cursor > 0 {
			state.cursor--
		}
	case "down", "j", "s":
		if state.cursor < len(game.GetGameTypes())-1 {
			state.cursor++
		}
	case "enter", " ":
		playerMsg := messages.ClientMessage{
			Type:     messages.ClientSelectGameType,
			GameType: game.GetGameTypes()[state.cursor],
		}
		return session, session.SendMsgToServer(playerMsg)
	default:
		return session, nil
	}
	return session, nil
}

func (state SessionStateInGameSelection) GetDisplayString() string {
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FAFAFA")).
		Background(lipgloss.Color("#32D74B")).
		Padding(0, 1).
		MarginBottom(1)

	if state.playerNum != 1 {
		waitingStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF9500")).
			Italic(true).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#FF9500")).
			Padding(1).
			MarginTop(1)

		title := titleStyle.Render("🎯 GAME SELECTION")
		waiting := waitingStyle.Render("Waiting for Player 1 to select a game...")

		return lipgloss.JoinVertical(lipgloss.Left, title, waiting)
	}

	instructionStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#626262")).
		MarginBottom(1)

	selectedStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FAFAFA")).
		Background(lipgloss.Color("#32D74B")).
		Padding(0, 1)

	unselectedStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#9CA3AF")).
		Padding(0, 1)

	controlsStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#6B7280")).
		Border(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("#374151")).
		Padding(1).
		MarginTop(1)

	title := titleStyle.Render("🎯 GAME SELECTION")
	instruction := instructionStyle.Render("Choose a game to play:")

	var gameOptions []string
	for i, gameType := range game.GetGameTypes() {
		if i == state.cursor {
			gameOptions = append(gameOptions, selectedStyle.Render("▶ "+gameType.String()))
		} else {
			gameOptions = append(gameOptions, unselectedStyle.Render("  "+gameType.String()))
		}
	}

	games := lipgloss.JoinVertical(lipgloss.Left, gameOptions...)
	controls := controlsStyle.Render("↑/↓ Navigate • Enter/Space Select • q Quit")

	return lipgloss.JoinVertical(lipgloss.Left, title, instruction, games, controls)
}

func (state *SessionStateInGameSelection) handleServerMessage(session Session, msg messages.ServerMessage) (Session, error) {
	switch msg.Type {
	case messages.ServerGameStarted:
		session.game = msg.Game.GetGame()
		session.gameType = session.game.GetGameType()
		session.playerTurn = msg.PlayerTurn
		session = session.setState(SessionStateTypeInGame)
	case messages.ServerRoomClosed:
		session = session.handleRoomClosure(msg)
	default:
		return session, fmt.Errorf("unexpected server message type whle in waiting room: %v", msg.Type)
	}
	return session, nil
}

type SessionStateInGame struct {
	playerNum        int
	isPlayerTurn     bool
	cursor           vector.Vector
	game             game.Game
	selectedSquare   vector.Vector
	inMoveSelectMode bool
}

func (SessionState SessionStateInGame) GetType() SessionStateType {
	return SessionStateTypeInGame
}

func NewSessionStateInGame(playerNum int, playerTurn int, game game.Game) *SessionStateInGame {

	return &SessionStateInGame{
		playerNum:      playerNum,
		game:           game,
		isPlayerTurn:   playerNum == playerTurn,
		selectedSquare: vector.NewVector(-1, -1),
	}
}

func (state *SessionStateInGame) HandleUserInput(msg tea.KeyMsg, session Session) (tea.Model, tea.Cmd) {
	switch state.game.GetGameType() {
	case game.GameTypeTicTacToe:
		return state.handleTicTacToeInput(msg, session)
	case game.GameTypeCheckers:
		return state.handleCheckersInput(msg, session)
	default:
		panic("game type not accounted for")
	}
}

func (state *SessionStateInGame) handleTicTacToeInput(msg tea.KeyMsg, session Session) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "c", "q":
		return session, session.SendMsgToServer(messages.ClientMessage{
			Type: messages.ClientConcede,
		})
	case "up", "k", "w":
		if state.cursor.Y > 0 {
			state.cursor.Y--
		}
	case "down", "j", "s":
		if state.cursor.Y < 2 {
			state.cursor.Y++
		}
	case "left", "h", "a":
		if state.cursor.X > 0 {
			state.cursor.X--
		}
	case "right", "l", "d":
		if state.cursor.X < 2 {
			state.cursor.X++
		}
	case "enter", " ":
		turnMsg := messages.ClientMessage{
			Type: messages.ClientSendTurn,
			TurnAction: messages.NewGameTurnWrapper(game.TicTacToeTurn{
				Coords: state.cursor,
			}),
		}
		return session, session.SendMsgToServer(turnMsg)
	}
	return session, nil
}

func (state *SessionStateInGame) handleCheckersInput(msg tea.KeyMsg, session Session) (tea.Model, tea.Cmd) {
	if !state.inMoveSelectMode {
		switch msg.String() {
		case "c", "q":
			return session, session.SendMsgToServer(messages.ClientMessage{
				Type: messages.ClientConcede,
			})
		case "up", "k", "w":
			if state.cursor.Y > 0 {
				state.cursor.Y--
			}
		case "down", "j", "s":
			if state.cursor.Y < 7 {
				state.cursor.Y++
			}
		case "left", "h", "a":
			if state.cursor.X > 0 {
				state.cursor.X--
			}
		case "right", "l", "d":
			if state.cursor.X < 7 {
				state.cursor.X++
			}
		case "enter", " ":
			if !state.game.(*game.CheckersGame).SquareHasPlayerPiece(state.cursor, state.playerNum) {
				return session, func() tea.Msg {
					return ErrMsg{fmt.Errorf("you do not have a piece at %v, %v", state.cursor.Y, state.cursor.X)}
				}
			}
			state.inMoveSelectMode = true
			return session, nil
		}
	} else {
		var moveDirection game.CheckersDirection
		switch msg.String() {
		case "c", "q":
			return session, session.SendMsgToServer(messages.ClientMessage{
				Type: messages.ClientConcede,
			})
		case "backspace", "escape":
			state.inMoveSelectMode = false
		case "e":
			moveDirection = game.CheckersDirectionLeft
		case "r":
			moveDirection = game.CheckersDirectionRight
		case "d":
			moveDirection = game.CheckersDirectionBackLeft
		case "f":
			moveDirection = game.CheckersDirectionBackRight
		default:
			return session, func() tea.Msg {
				return ErrMsg{fmt.Errorf("invalid input")}
			}
		}
		state.inMoveSelectMode = false
		return session, session.SendMsgToServer(messages.ClientMessage{
			Type: messages.ClientSendTurn,
			TurnAction: messages.NewGameTurnWrapper(game.CheckersTurn{
				PieceCoords: state.cursor,
				Direction:   moveDirection,
			}),
		})
	}
	return session, nil
}

func (state SessionStateInGame) GetDisplayString() string {
	infoStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#626262")).
		Background(lipgloss.Color("#1C1C1E")).
		Padding(0, 1).
		MarginBottom(1)

	controlsStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#6B7280")).
		Border(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("#374151")).
		Padding(1)

	board := state.game.DisplayBoard(state.cursor, state.playerNum)
	playerTurnMsg := ""
	if state.isPlayerTurn {
		playerTurnMsg = "Your turn!"
	} else {
		playerTurnMsg = "Waiting for opponents move..."
	}
	info := infoStyle.Render(fmt.Sprintf("Player: %d | %v", state.playerNum, playerTurnMsg))
	controls := controlsStyle.Render("WASD/Arrow Keys Move • Enter/Space Select • q/c Concede")

	return lipgloss.JoinVertical(lipgloss.Left, board, info, controls)
}

func (state *SessionStateInGame) handleServerMessage(session Session, msg messages.ServerMessage) (Session, error) {
	switch msg.Type {
	case messages.ServerError:
		return session, errors.New(msg.ErrorMessage)
	case messages.ServerTurnResult:
		session.game = msg.Game.GetGame()
		session.playerTurn = msg.PlayerTurn
		state.game = session.game
		state.isPlayerTurn = session.playerTurn == state.playerNum
		session.playerTurn = msg.PlayerTurn
	case messages.ServerGameFinished:
		session.game = msg.Game.GetGame()
		session.gameResult = msg.GameResult
		session = session.setState(SessionStateTypeEndGame)
	case messages.ServerRoomClosed:
		session = session.handleRoomClosure(msg)
		//todo
	default:
		return session, fmt.Errorf("unexpected server message type whle in game: %v", msg.Type)
	}

	return session, nil
}

type SessionStateEndGame struct {
	game       game.Game
	gameResult messages.GameResult
	playerNum  int
}

func NewSessionStateEndGame(game game.Game, gameResult messages.GameResult, playerNum int) *SessionStateEndGame {
	return &SessionStateEndGame{
		game:       game,
		gameResult: gameResult,
		playerNum:  playerNum,
	}
}

func (state SessionStateEndGame) GetType() SessionStateType {
	return SessionStateTypeEndGame
}
func (state SessionStateEndGame) GetDisplayString() string {
	resultStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FAFAFA")).
		Padding(0, 1).
		MarginBottom(1).
		MarginTop(1)

	promptStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#626262")).
		Background(lipgloss.Color("#1C1C1E")).
		Padding(0, 1).
		MarginBottom(1)

	controlsStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#6B7280")).
		Border(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("#374151")).
		Padding(1)

	board := state.game.DisplayBoard(vector.NewVector(-1, -1), state.playerNum)

	var resultStr string
	switch state.gameResult {
	case messages.GameResultPlayerWin:
		resultStyle = resultStyle.Background(lipgloss.Color("#32D74B"))
		resultStr = "You Won!"
	case messages.GameResultPlayerLose:
		resultStyle = resultStyle.Background(lipgloss.Color("#FF3B30"))
		resultStr = "You Lost"
	case messages.GameResultDraw:
		resultStyle = resultStyle.Background(lipgloss.Color("#FF9500"))
		resultStr = "It's a Draw!"
	default:
		resultStr = "Game Over"
	}

	result := resultStyle.Render(resultStr)
	prompt := promptStyle.Render("Play again? (y/n)")
	controls := controlsStyle.Render("y Play Again • n/q Quit to Menu")

	return lipgloss.JoinVertical(lipgloss.Left, board, result, prompt, controls)
}

func (state *SessionStateEndGame) HandleUserInput(msg tea.KeyMsg, session Session) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "y":
		return session, session.SendMsgToServer(messages.ClientMessage{
			Type:     messages.ClientJoinRoom,
			RoomCode: session.roomCode,
		})
	case "n", "q":
		return session, session.SendMsgToServer(messages.ClientMessage{
			Type: messages.ClientQuitRoom,
		})
	default:
		return session, nil
	}
}

func (state *SessionStateEndGame) handleServerMessage(session Session, msg messages.ServerMessage) (Session, error) {
	//TODO
	switch msg.Type {
	case messages.ServerRoomJoined:
		session.playerNumber = msg.PlayerNumber
		session = session.setState(SessionStateTypeWaitingRoom)
	case messages.ServerRoomClosed:
		session = session.handleRoomClosure(msg)
	default:
		return session, fmt.Errorf("unexpected server message type while in game end: %v", msg.Type)
	}
	return session, nil
}
