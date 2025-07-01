package main

const URL = "ws://localhost:8000/"

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
//
func main() {
	session := NewSession()

	err := startRepl(session)
	if err != nil {
		panic(err)
	}
}
