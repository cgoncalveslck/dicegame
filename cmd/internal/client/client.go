package client

import (
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"math/rand"

	"github.com/gorilla/websocket"
)

const Timeout = 30 // short for debugging
// ideally i'd not use UUID's for the the ID's and use an agreed list of constant codes for the "kind" of message
type EndPlayMessage struct {
	Kind     string   `json:"kind"`
	ClientId ClientId `json:"clientId"`
}

type EndPlayResultMessage struct {
	Kind   string `json:"kind"`
	Result string `json:"result"`
}

type PlayResultMessage struct {
	Kind   string `json:"kind"`
	Result string `json:"result"`
	Roll   int    `json:"roll"`
}

type PlayMessage struct {
	Kind     string   `json:"kind"`
	ClientId ClientId `json:"clientId"`
	Bet      int      `json:"bet"`
	Choice   string   `json:"choice"`
}

type SessionStartMessage struct {
	Kind     string   `json:"kind"`
	ClientId ClientId `json:"clientId"`
}

type WalletMessage struct {
	Kind     string   `json:"kind"`
	ClientId ClientId `json:"clientId"`
	Wallet   int      `json:"wallet"`
}

type PlayHistoryItem struct {
	Choice string `json:"choice"`
	Bet    int    `json:"bet"`
	Result string `json:"result"`
	Roll   int    `json:"roll"`
}

type InfoMessage struct {
	Kind string `json:"kind"`
}

type DefaultMessage struct {
	ClientId ClientId `json:"clientId"`
	Kind     string   `json:"kind"`
}

// idk if this is very idiomatic or correct but I like giving a type to keys
// to make it easier to understand what I'm doing with Store struct
type ClientId string

type Store struct {
	Clients map[ClientId]*Client `json:"clients"`
	Mx      *sync.Mutex          `json:"-"`
}

func (s *Store) RemoveClient(id ClientId) {
	s.Mx.Lock()
	defer s.Mx.Unlock()

	client, ok := s.Clients[id]
	if !ok {
		return
	}

	if client.Conn != nil {
		_ = client.Conn.Close()
	}
	delete(s.Clients, id)
}

func (s *Store) AddClient(c *Client) {
	s.Mx.Lock()
	defer s.Mx.Unlock()
	s.Clients[c.Id] = c
}

func SessionCleanup() {
	timer := time.NewTicker(30 * time.Second)

	for range timer.C {
		for _, c := range St.Clients {
			if time.Now().Unix()-c.Last_seen > Timeout {
				c.Disconnect()
			}
		}
	}
}

type Client struct {
	Conn      *websocket.Conn `json:"-"`
	Id        ClientId        `json:"clientId"`
	Wallet    int             `json:"wallet"`
	Last_seen int64           `json:"-"`
	Session   *Session        `json:"-"`
}

func NewClient(c *websocket.Conn) *Client {
	return &Client{
		Conn:      c,
		Id:        "e38dc209-4fd2-449b-874c-812ac890ead0", // debug
		Wallet:    100,                                    // free money
		Last_seen: time.Now().Unix(),
		Session:   nil,
	}
}

func (c *Client) Play(message *[]byte) {
	if c.Conn == nil {
		log.Println("No connection")
		return
	}

	if c.Session == nil {
		msg := &InfoMessage{
			Kind: "NO_SESSION",
		}

		c.SendMessage(msg)
		log.Print("Tried to play without a session")
		return
	}

	p := &PlayMessage{}
	err := json.Unmarshal(*message, p)
	if err != nil {
		log.Printf("Unmarshal error: %v", err)
		return
	}

	if p.Bet > c.Wallet {
		message := &InfoMessage{
			Kind: "NO_BALANCE",
		}

		c.SendMessage(message)
		return
	}

	num := rand.Intn(6) + 1

	var win bool
	switch p.Choice {
	case "ODD":
		win = num%2 != 0
	case "EVEN":
		win = num%2 == 0
	}

	if win {
		c.Wallet += p.Bet
	} else {
		c.Wallet -= p.Bet
	}

	var res string
	if win {
		res = "WIN"
	} else {
		res = "LOSE"
	}

	walletMessage := PlayResultMessage{
		Kind:   "ROLL",
		Result: res,
		Roll:   num,
	}

	PlayHistoryItem := PlayHistoryItem{
		Choice: p.Choice,
		Bet:    p.Bet,
		Result: res,
		Roll:   num,
	}

	c.Session.PlayHistory.Add(PlayHistoryItem)
	c.SendMessage(walletMessage)
}

func (c *Client) GetWallet() {
	if c.Conn == nil {
		return
	}

	walletMessage := WalletMessage{
		Kind:     "WALLET",
		ClientId: c.Id,
		Wallet:   c.Wallet,
	}

	c.SendMessage(walletMessage)
}

func (c *Client) StartSession() {
	if c.Session != nil && c.Session.Playing {
		msg := &InfoMessage{
			Kind: "ALREADY_PLAYING",
		}

		c.SendMessage(msg)
		return
	}

	c.Session = &Session{
		Playing: true,
		Profit:  0,
		PlayHistory: &PlayHistory{
			Items: make([]PlayHistoryItem, 0),
		},
	}
}

func (c *Client) EndSession() {
	if !c.Session.Playing {
		c.SendMessage(&InfoMessage{
			Kind: "NOT_PLAYING",
		})
		return
	}

	c.Session.Playing = false
	c.Wallet += c.Session.Profit
	c.Session.Profit = 0
	c.Session.PlayHistory = nil

	c.SendMessage(&InfoMessage{
		Kind: "ENDPLAY",
	})
}

func (c *Client) SendMessage(msg interface{}) {
	data, err := json.Marshal(msg)
	if err != nil {
		log.Printf("Marshal error: %v", err)
		return
	}

	err = c.Conn.WriteMessage(websocket.TextMessage, data)
	if err != nil {
		// handle this properly
		log.Printf("Write error: %v", err)
	}
}

func (c *Client) Disconnect() {
	fmt.Println("Disconnecting client", c.Id) // debug
	if c.Conn != nil {
		_ = c.Conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseGoingAway, "TIMEOUT"))
		c.Conn.Close()
	}

	St.RemoveClient(c.Id)
}

// added this for fun i guess, could be cool to show on UI
type PlayHistory struct {
	Items []PlayHistoryItem
}

func (ph *PlayHistory) Add(item PlayHistoryItem) {
	if len(ph.Items) == 10 {
		ph.Items = ph.Items[1:]
	}
	ph.Items = append(ph.Items, item)
}

type Session struct {
	Playing     bool
	Profit      int // should always be 0 if not playing
	PlayHistory *PlayHistory
}

var St = &Store{
	Clients: make(map[ClientId]*Client),
	Mx:      &sync.Mutex{},
}
