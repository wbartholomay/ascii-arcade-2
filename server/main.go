package main

import (
	"log"
	"net/http"
	// _ "net/http/pprof"
)

func main() {
	log.Println("Starting server...")
	hub := NewHub()
	go hub.Run()

	http.HandleFunc("/", hub.ServeWs)

	http.ListenAndServe(":8000", nil)
}
