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

				log.Printf("JSON syntax error: %+v", err)
				err := c.SendMessage(cErr)
				if err != nil {
					log.Printf("SendConnMessage error: %+v", err)
				}
				continue
			}

			log.Printf("Read error: %+v", err)
			continue
		}

		slog.Debug("Received message", slog.String("message", string(msg.Kind)))
		if msg.ClientId != "" {
			cErr := client.HandleClientID(conn, msg)
			if cErr != nil {
				log.Printf("HandleClientID error: %+v", cErr)
				c.SendMessage(cErr)
				continue
			}
		}

		c.Last_seen = time.Now().Unix()
		switch string(msg.Kind) {
		case "PLAY":
			cErr, err := c.Play(msg)
			if err != nil {
				log.Printf("Play error: %+v", err)
				c.SendIErrorMessage(err)
			}
			if cErr != nil {
				err := c.SendErrorMessage(cErr)
				if err != nil {
					log.Printf("SendErrorMessage error: %+v", err)
					c.SendIErrorMessage(err)
				}
			}
		case "WALLET":
			cErr, err := c.GetWallet(msg)
			if err != nil {
				log.Printf("GetWallet error: %+v", err)
				c.SendIErrorMessage(err)
			}
			if cErr != nil {
				err := c.SendErrorMessage(cErr)
				if err != nil {
					log.Printf("SendErrorMessage error: %+v", err)
					c.SendIErrorMessage(err)
				}
			}
		case "STARTPLAY":
			cErr, err := c.StartSession(msg)
			if err != nil {
				log.Printf("StartSession error: %+v", err)
				c.SendIErrorMessage(err)
			}

			if cErr != nil {
				err := c.SendErrorMessage(cErr)
				if err != nil {
					log.Printf("SendErrorMessage error: %+v", err)
					c.SendIErrorMessage(err)
				}
			}
		case "ENDPLAY":
			cErr, err := c.EndSession(msg)
			if err != nil {
				log.Printf("EndSession error: %+v", err)
				c.SendIErrorMessage(err)
			}
			if cErr != nil {
				err := c.SendErrorMessage(cErr)
				if err != nil {
					log.Printf("SendErrorMessage error: %+v", err)
					c.SendIErrorMessage(err)
				}
			}
		case "AUTH":
			c, err = c.Auth(conn)
			if err != nil {
				log.Printf("Auth error: %+v", err)
				c.SendIErrorMessage(err)
			}
		default:
			msg := &client.InfoResultMessage{
				Kind: "UNKNOWN_KIND",
			}

			err := c.SendMessage(msg)
			if err != nil {
				log.Printf("SendMessage error: %+v", err)
				c.SendIErrorMessage(err)
			}

			log.Printf("Unknown message: %+v", msg)
		}
	}
}
