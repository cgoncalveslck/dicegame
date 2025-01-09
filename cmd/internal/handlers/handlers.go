package handlers

import (
	"cgoncalveslck/dicegame/cmd/internal/client"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

func Handler(w http.ResponseWriter, r *http.Request) {
	var upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true // local dev
		},
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Upgrade failed: %v", err)
		return
	}
	defer func() {
		conn.Close()
	}()

	c := client.NewClient(conn)
	client.St.AddClient(c)

	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			// I get 1005 using Insomnia's "Disconnect" and 1001 on browser refresh
			// I'm guessing 1001 should imply reconnection logic instead of disconnecting like this
			if websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway, websocket.CloseNoStatusReceived) {
				c.Disconnect()
			}
			log.Printf("Read error: %v", err)
			break
		}

		msg := &client.DefaultMessage{}
		err = json.Unmarshal(message, msg)
		if err != nil {
			log.Printf("Unmarshal error: %v", err)
			break
		}

		c.Id = msg.ClientId
		c.Last_seen = time.Now().Unix()

		switch string(msg.Kind) {
		case "PLAY":
			c.Play(&message)
		case "WALLET":
			c.GetWallet()
		case "STARTPLAY":
			c.StartSession()
		case "ENDPLAY":
			c.EndSession()
		default:
			msg := &client.InfoMessage{
				Kind: "UNKNOWN_KIND",
			}

			c.SendMessage(msg)
			log.Printf("Unknown message: %s", message)
		}
	}
}
