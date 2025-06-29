package main

import (
	"fmt"
	"net/http"

	// _ "net/http/pprof"
)

func main() {
	fmt.Println("Starting server...")
	hub := NewHub()
	go hub.Run()

	static := http.Dir("web/dist")

	http.Handle("/", http.FileServer(static))
	http.HandleFunc("/ws", hub.ServeWs)

	http.ListenAndServe("localhost:8000", nil)
}
