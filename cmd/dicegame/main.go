package main

import (
	"cgoncalveslck/dicegame/cmd/internal/client"
	"cgoncalveslck/dicegame/cmd/internal/handlers"
	"log"
	"log/slog"
	"net/http"
)

func main() {
	slog.SetLogLoggerLevel(slog.LevelDebug)

	http.HandleFunc("/", handlers.Handler)

	go client.SessionExpire()

	slog.Info("Starting server on :8181")
	log.Fatal(http.ListenAndServe(":8181", nil))
}
