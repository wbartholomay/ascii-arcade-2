# ASCII Arcade 🎮

A multiplayer terminal-based arcade featuring classic board games with beautiful ASCII art interfaces. Play Tic Tac Toe and Checkers with friends over the internet!

```
 █████╗ ███████╗ ██████╗██╗██╗     █████╗ ██████╗  ██████╗ █████╗ ██████╗ ███████╗
██╔══██╗██╔════╝██╔════╝██║██║    ██╔══██╗██╔══██╗██╔════╝██╔══██╗██╔══██╗██╔════╝
███████║███████╗██║     ██║██║    ███████║██████╔╝██║     ███████║██║  ██║█████╗  
██╔══██║╚════██║██║     ██║██║    ██╔══██║██╔══██╗██║     ██╔══██║██║  ██║██╔══╝  
██║  ██║███████║╚██████╗██║██║    ██║  ██║██║  ██║╚██████╗██║  ██║██████╔╝███████╗
╚═╝  ╚═╝╚══════╝ ╚═════╝╚═╝╚═╝    ╚═╝  ╚═╝╚═╝  ╚═╝ ╚═════╝╚═╝  ╚═╝╚═════╝ ╚══════╝
```

## 🎯 Features

- **🎮 Multiple Games**: Tic Tac Toe and Checkers
- **👥 Multiplayer**: Real-time gameplay with WebSocket connections
- **🎨 Beautiful ASCII Art**: Rich terminal interfaces with emojis and colors
- **🌐 Cross-Platform**: Works on Windows, macOS, and Linux
- **☁️ Cloud Deployed**: Server hosted on Google Cloud Run
- **🔄 Real-time Updates**: Live game state synchronization

## 🎲 Available Games

### Tic Tac Toe
Classic 3x3 grid game with a clean ASCII interface:
```
   0   1   2
0  X |   | O 
  ---|---|---
1    | X |   
  ---|---|---
2  O |   | X 
```

### Checkers
Full 8x8 checkers board with piece tracking:
```
       0       1       2       3       4       5       6       7
  —————————————————————————————————————————————————————————————————
  |       |#######|       |#######|       |#######|       |#######|
a |       |  🔵₁  |       |  🔵₂  |       |  🔵₃  |       |  🔵₄  |
  |       |#######|       |#######|       |#######|       |#######|
  —————————————————————————————————————————————————————————————————
```

## 🚀 Getting Started

### Prerequisites
- Go 1.24.4 or later
- Terminal with Unicode support

### Installation

1. **Download the latest release**:
   ```bash
   # Download from releases page or clone the repository
   git clone https://github.com/wbarthol/ascii-arcade-2.git
   cd ascii-arcade-2
   ```

2. **Build the client**:
   ```bash
   go build -o ascii-arcade .
   ```

3. **Run the game**:
   ```bash
   ./ascii-arcade
   ```

### Quick Start

1. **Join a room**:
   ```
   Main Menu > join myroom123
   ```

2. **Wait for opponent** or share your room code with a friend

3. **Play your game**:
   - **Tic Tac Toe**: `move <row> <col>` (e.g., `move 1 2`)
   - **Checkers**: `move <piece-id> <direction>` (e.g., `move 5 l`)

4. **Get help anytime**:
   ```
   help
   ```

## 🎮 Commands

### Navigation Commands
- `join <room-code>` - Create or join a game room
- `quit` - Leave current room/game
- `help` - Show available commands

### Tic Tac Toe Commands
- `move <row> <col>` - Place your piece (0-2 for both row and col)

### Checkers Commands
- `move <piece-id> <direction>` - Move a piece
  - **Directions**: `l` (left), `r` (right), `bl` (back-left), `br` (back-right)
  - **Piece IDs**: Shown as subscript numbers on pieces (1-12)

## 🏗️ Architecture

### Client-Server Model
```
┌─────────────┐    WebSocket    ┌─────────────┐
│   Client    │ ◄─────────────► │   Server    │
│  (Go CLI)   │   (Real-time)   │ (Cloud Run) │
└─────────────┘                 └─────────────┘
```

### Project Structure
```
ascii-arcade-2/
├── main.go              # Client entry point
├── session.go           # Game session management
├── repl.go             # Command-line interface
├── commands.go         # Command parsing and execution
├── ws_driver.go        # WebSocket communication
├── internal/
│   ├── game/           # Game logic
│   │   ├── game.go     # Game interface
│   │   ├── tic_tac_toe.go
│   │   ├── checkers.go
│   │   └── *_test.go   # Comprehensive tests
│   ├── messages/       # Network protocol
│   └── vector/         # 2D coordinate system
└── server/             # Server implementation
    ├── main.go
    ├── hub.go          # Connection management
    ├── room.go         # Game room logic
    └── player.go       # Player state
```

## 🔧 Development

### Running Tests
```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run specific package tests
go test ./internal/game
```

### Local Development

1. **Set up environment**:
   ```bash
   cp .env.example .env
   # Edit .env with your settings
   ```

2. **Run local server** (optional):
   ```bash
   cd server
   go run .
   ```

3. **Connect to local server**:
   ```bash
   export SERVER_URL="ws://localhost:8000"
   go run .
   ```

### Building for Distribution
```bash
# Build for current platform
go build -o ascii-arcade .

# Cross-compile for different platforms
GOOS=windows GOARCH=amd64 go build -o ascii-arcade-windows.exe .
GOOS=darwin GOARCH=amd64 go build -o ascii-arcade-macos .
GOOS=linux GOARCH=amd64 go build -o ascii-arcade-linux .
```

## 🌐 Deployment

The server is automatically deployed to Google Cloud Run via GitHub Actions:

- **Production URL**: `wss://ascii-arcade-server-714989044760.us-central1.run.app`
- **CI/CD**: Automatic deployment on push to `main` branch
- **Infrastructure**: Docker containerized, horizontally scalable

### Manual Deployment
```bash
# Build and push Docker image
docker build -t ascii-arcade-server .
docker tag ascii-arcade-server gcr.io/YOUR-PROJECT/ascii-arcade-server
docker push gcr.io/YOUR-PROJECT/ascii-arcade-server

# Deploy to Cloud Run
gcloud run deploy ascii-arcade-server \
  --image gcr.io/YOUR-PROJECT/ascii-arcade-server \
  --region us-central1 \
  --allow-unauthenticated
```

## 🎨 Game Features

### State Management
- **Session States**: Menu → Waiting Room → In Game
- **Real-time Sync**: Game state synchronized between players
- **Turn Management**: Enforced turn-based gameplay
- **Game Over Detection**: Automatic win/lose/draw detection

### Visual Features
- **ANSI Colors**: State-dependent colored prompts
- **Unicode Art**: Rich game boards with emojis
- **Player Feedback**: Clear success/error messages
- **Responsive UI**: Adapts to terminal width

### Network Protocol
- **JSON Messages**: Structured client-server communication
- **WebSocket**: Low-latency real-time updates
- **Error Handling**: Graceful connection management
- **Validation**: Server-side move validation

## 🧪 Testing

The project includes comprehensive test coverage:

- **Unit Tests**: All game logic thoroughly tested
- **Integration Tests**: Client-server communication
- **Game Scenarios**: Win conditions, edge cases, invalid moves
- **Error Handling**: Network failures, invalid input

### Test Coverage Areas
- ✅ Game initialization and setup
- ✅ Move validation (legal/illegal moves)
- ✅ Turn execution and state updates
- ✅ Win/lose/draw detection
- ✅ Board display and rendering
- ✅ Network message handling
- ✅ Edge cases and error conditions

## 🤝 Contributing

1. **Fork the repository**
2. **Create a feature branch**: `git checkout -b feature-name`
3. **Write tests** for your changes
4. **Ensure tests pass**: `go test ./...`
5. **Submit a pull request**

### Adding New Games
1. Implement the `Game` interface in `internal/game/`
2. Add message types in `internal/messages/`
3. Update command parsing in `commands.go`
4. Write comprehensive tests
5. Update this README

## 📝 License

This project is open source and available under the [MIT License](LICENSE).

## 🎯 Roadmap

- [ ] **Chess** implementation
- [ ] **Spectator mode** for watching games
- [ ] **Game replay** system
- [ ] **Player statistics** and rankings
- [ ] **Tournament mode** with brackets
- [ ] **Custom themes** and board styles
- [ ] **Mobile-friendly** web interface

## 🙋‍♂️ Support

- **Issues**: Report bugs on [GitHub Issues](https://github.com/wbarthol/ascii-arcade-2/issues)
- **Discussions**: Feature requests and questions
- **Wiki**: Detailed documentation and guides

---

**Enjoy playing ASCII Arcade!** 🎮✨