package handlers

import (
	"cgoncalveslck/dicegame/cmd/internal/client"
	"encoding/json"
	"errors"
	"log"
	"log/slog"
	"net/http"
	"time"

	"github.com/google/uuid"
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

	var c *client.Client
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
				if c != nil {
					client.St.DisconnectClient(c.Id)
				}
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
				err := client.SendConnMessage(conn, cErr)
				if err != nil {
					log.Printf("SendConnMessage error: %+v", err)
				}
				continue
			}

			log.Printf("Read error: %+v", err)
			continue
		}

		slog.Debug("Received message", slog.String("message", string(msg.Kind)))

		if c == nil {
			var cErr *client.ErrorResultMessage

			c, cErr, err = client.InitC(conn, msg)
			if err != nil {
				log.Printf("InitC error: %+v", err)
				cErr := &client.ErrorResultMessage{
					Kind:    "ERROR",
					Message: "failed to initialize client",
					Code:    client.INTERNAL,
				}

				err := client.SendConnMessage(conn, cErr)
				if err != nil {
					log.Printf("SendConnMessage error: %+v", err)
				}

				continue
			}

			if cErr != nil {
				err := client.SendConnMessage(conn, cErr)
				if err != nil {
					log.Printf("SendErrorMessage error: %+v", err)
				}

				continue
			}

			continue
		} else {
			if msg.Kind == "AUTH" {
				addr := conn.RemoteAddr().String()

				// this is a bit RAW and ROUGH but you get the point
				for _, c := range client.St.Clients {
					if c.Ip == addr {
						cErr := &client.ErrorResultMessage{
							Kind:    "ERROR",
							Message: "already logged",
							Code:    client.ALREADY_LOGGED,
						}

						err := client.SendConnMessage(conn, cErr)
						if err != nil {
							log.Printf("SendErrorMessage error: %+v", err)
						}
						continue
					}
				}
				continue
			}
		}

		if msg.ClientId != "" {
			_, err := uuid.Parse(msg.ClientId)
			if err != nil {
				cErr := &client.ErrorResultMessage{
					Kind:    "ERROR",
					Message: "invalid client id, must be UUID",
				}

				c.SendErrorMessage(cErr)
				continue
			}

			// check if client exists in store map lookup
			_, ok := client.St.Clients[msg.ClientId]
			if !ok {
				cErr := &client.ErrorResultMessage{
					Kind:    "ERROR",
					Message: "client not found",
				}

				c.SendErrorMessage(cErr)
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
			err := c.GetWallet()
			if err != nil {
				log.Printf("GetWallet error: %+v", err)
				c.SendIErrorMessage(err)
			}
		case "STARTPLAY":
			cErr, err := c.StartSession()
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
			cErr, err := c.EndSession()
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
