package main

const URL = "ws://localhost:8000/"

func main() {
	session, err := StartSession(URL)
	if err != nil {
		panic(err)
	}

	err = startRepl(session)
	if err != nil {
		panic(err)
	}
}