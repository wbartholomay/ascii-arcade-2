package main

import (
	"os"

	"github.com/joho/godotenv"
)

const (
	AnsiReset      = "\033[0m"
	AnsiRed        = "\033[31m"
	AnsiGreen      = "\033[32m"
	AnsiYellow     = "\033[33m"
	AnsiBlue       = "\033[34m"
	AnsiLightRed   = "\033[91m"
	AnsiLightGreen = "\033[92m"
	AnsiLightBlue  = "\033[94m"
)

func main() {
	godotenv.Load()
	url := os.Getenv("SERVER_URL")
	if url == "" {
		url = "wss://ascii-arcade-server-714989044760.us-central1.run.app"
	}

	session := NewSession(url)

	err := startRepl(session)
	if err != nil {
		panic(err)
	}
}
