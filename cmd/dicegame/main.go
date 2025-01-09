package main

import (
	"cgoncalveslck/dicegame/cmd/internal/client"
	"cgoncalveslck/dicegame/cmd/internal/handlers"
	"log"
	"net/http"
)

func main() {
	http.HandleFunc("/", handlers.Handler)

	// i did this before manually cleaning up the sessions, i'll leave it for reader context
	go client.SessionCleanup()
	log.Fatal(http.ListenAndServe(":8080", nil))
}
