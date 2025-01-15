package handlers

import (
	"cgoncalveslck/dicegame/cmd/internal/client"
	"encoding/json"
	"errors"
	"log"
	"log/slog"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

func Handler(w http.ResponseWriter, r *http.Request) {
	var upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true // CORS local dev
		},
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Upgrade failed: %+v", err)

		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("WebSocket upgrade failed: " + err.Error()))
		return
	}
	defer func() {
		conn.Close()
	}()

	c := &client.Client{
		Conn:      conn,
		Last_seen: time.Now().Unix(),
	}

	for {
		if conn == nil {
			break
		}

		msg := &client.DefaultMessage{}
		err = conn.ReadJSON(msg)
		if err != nil {
			// I get 1005 using Insomnia's "Disconnect" and 1001 on browser refresh
			// I'm guessing 1001 should imply reconnection logic instead of disconnecting like this
			if websocket.IsCloseError(err,
				websocket.CloseNormalClosure,
				websocket.CloseGoingAway,
				websocket.CloseNoStatusReceived,
				websocket.CloseAbnormalClosure,
			) {
				client.St.DisconnectClient(c)
				break
			}

			// Handle JSON syntax errors
			var syntaxError *json.SyntaxError
			var unmarshalTypeError *json.UnmarshalTypeError

			if errors.As(err, &syntaxError) || errors.As(err, &unmarshalTypeError) {
				cErr := &client.ErrorResultMessage{
					Kind:    "ERROR",
					Message: "failed to parse JSON",
					Code:    client.INVALID_JASON,
				}

				err := c.SendMessage(cErr)
				if err != nil {
					log.Printf("SendConnMessage error: %+v", err)
				}
				continue
			}

			// Session expired
			slog.Debug("Read Message error")
			break
		}

		slog.Debug("Received message", slog.String("message", string(msg.Kind)))
		if msg.ClientId != "" {
			cErr := client.HandleClientID(conn, msg)
			if cErr != nil {
				err := c.SendMessage(cErr)
				if err != nil {
					log.Printf("SendConnMessage error: %+v", err)
				}

				continue
			}
		}

		c.Last_seen = time.Now().Unix()
		switch string(msg.Kind) {
		case "PLAY":
			cErr, err := c.Play(msg)
			c.HandleMessageErrors(cErr, err, "Play")
		case "WALLET":
			cErr, err := c.GetWallet(msg)
			c.HandleMessageErrors(cErr, err, "GetWallet")
		case "STARTPLAY":
			cErr, err := c.StartSession(msg)
			c.HandleMessageErrors(cErr, err, "StartSession")
		case "ENDPLAY":
			cErr, err := c.EndSession(msg)
			c.HandleMessageErrors(cErr, err, "EndSession")
		case "AUTH":
			c, err = c.Auth(conn)
			if err != nil {
				log.Printf("Auth error: %+v", err)
				err = c.SendMessage(err)
				if err != nil {
					log.Printf("SendMessage error: %+v", err)
				}
			}
		default:
			msg := &client.ErrorResultMessage{
				Kind:    "ERROR",
				Code:    client.UNKNOWN_KIND,
				Message: "unknown message kind",
			}

			err := c.SendMessage(msg)
			if err != nil {
				log.Printf("SendMessage error: %+v", err)
				c.SendMessage(err)
			}
		}
	}
}
