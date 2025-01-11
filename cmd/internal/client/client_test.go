package client_test

import (
	"cgoncalveslck/dicegame/cmd/internal/client"
	"cgoncalveslck/dicegame/cmd/internal/handlers"
	"log"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/gorilla/websocket"
)

func TestMain(m *testing.M) {
	slog.SetLogLoggerLevel(slog.LevelDebug)

	http.HandleFunc("/", handlers.Handler)

	go client.SessionExpire()

	slog.Info("Starting server on :8181")
	go func() {
		log.Fatal(http.ListenAndServe(":8181", nil))
	}()

	m.Run()
}

func TestUnauthenticatedClient(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(handlers.Handler))
	defer server.Close()

	u := url.URL{Scheme: "ws", Host: server.Listener.Addr().String(), Path: "/"}
	conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		t.Fatalf("Failed to connect to WebSocket server: %v", err)
	}
	defer conn.Close()

	msg := client.DefaultMessage{
		Kind: "PLAY",
	}

	err = conn.WriteJSON(msg)
	if err != nil {
		t.Fatalf("Failed to write JSON message: %v", err)
	}
}
