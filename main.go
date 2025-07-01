package main

const URL = "ws://localhost:8000/"

func main() {
	session := NewSession()

	err := startRepl(session)
	if err != nil {
		panic(err)
	}
}