package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/joho/godotenv"
	"github.com/wbarthol/ascii-arcade-2/internal/logging"
)

const (
	AnsiReset       = "\033[0m"
	AnsiRed         = "\033[31m"
	AnsiGreen       = "\033[32m"
	AnsiYellow      = "\033[33m"
	AnsiBlue        = "\033[34m"
	AnsiLightRed    = "\033[91m"
	AnsiLightGreen  = "\033[92m"
	AnsiLightYellow = "\033[93m"
	AnsiLightBlue   = "\033[94m"
)

var logger logging.Logger

func main() {
	godotenv.Load()
	url := os.Getenv("SERVER_URL")
	if url == "" {
		url = "wss://ascii-arcade-server-714989044760.us-central1.run.app"
	}

	session := NewSession(url)

	if len(os.Getenv("DEBUG")) > 0 {
		logger = logging.NewLogger(logging.LoggerDebug)
		f, err := tea.LogToFile("debug.log", "debug")
		if err != nil {
			fmt.Println("fatal:", err)
			os.Exit(1)
		}
		defer f.Close()
	}
	if _, err := tea.NewProgram(session).Run(); err != nil {
		panic(err)
	}
}
