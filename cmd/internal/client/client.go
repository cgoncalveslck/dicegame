package client

import (
	"encoding/json"
	"log/slog"
	"sync"
	"time"

	"math/rand"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

const (
	_ cError = iota
	NO_BALANCE
	NO_SESSION
	NOT_PLAYING
	ALREADY_PLAYING
	INTERNAL
	INVALID_UUID
	INVALID_JASON
	NOT_AUTHENTICATED
	ALREADY_LOGGED
)

type cError int

const Timeout = 300 // 5min

// ideally i'd not use UUID's for the the ID's and use an agreed list of constant codes for the "kind" of message
type EndPlayMessage struct {
	Kind     string `json:"kind"`
	ClientId string `json:"clientId"`
}

type EndPlayResultMessage struct {
	Kind   string `json:"kind"`
	Profit int    `json:"result"`
	Wallet int    `json:"wallet"`
}

type PlayResultMessage struct {
	Kind   string `json:"kind"`
	Result string `json:"result"`
	Roll   int    `json:"roll"`
}

type PlayMessage struct {
	Bet    int    `json:"bet"`
	Choice string `json:"choice"`
}

type WalletMessage struct {
	Kind   string `json:"kind"`
	Wallet int    `json:"wallet"`
}

type PlayHistoryItem struct {
	Choice string `json:"choice"`
	Bet    int    `json:"bet"`
	Result string `json:"result"`
	Roll   int    `json:"roll"`
}

type InfoResultMessage struct {
	Kind string `json:"kind"`
}

type AuthResultMessage struct {
	Kind     string `json:"kind"`
	ClientId string `json:"clientId"`
}

type ErrorResultMessage struct {
	Kind    string `json:"kind"`
	Message string `json:"message"`
	Code    cError `json:"code"`
}

type DefaultMessage struct {
	ClientId string `json:"clientId"`
	Kind     string `json:"kind"` // change to const maybe
	Wallet   int    `json:"wallet"`
	Bet      int    `json:"bet"`
	Choice   string `json:"choice"`
}

type Store struct {
	// map[clientId]*Client
	Clients map[string]*Client `json:"clients"`
	Mx      *sync.Mutex        `json:"-"`
}

// Disconnects and removes a client from the store
func (s *Store) DisconnectClient(id string) {
	s.Mx.Lock()
	defer s.Mx.Unlock()

	client, ok := s.Clients[id]
	if ok {
		if client.Conn != nil {
			_ = client.Conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
		}
		delete(s.Clients, id)
	}

	slog.Debug("Client removed", slog.String("ClientId", id))
}

func (s *Store) AddClient(c *Client) {
	s.Mx.Lock()
	defer s.Mx.Unlock()
	s.Clients[c.Id] = c

	slog.Debug("Client added", slog.String("ClientId", c.Id))
}

func SessionExpire() {
	timer := time.NewTicker(1 * time.Minute)

	for range timer.C {
		for _, c := range St.Clients {
			if time.Now().Unix()-c.Last_seen > Timeout {
				St.DisconnectClient(c.Id)
			}
		}
	}
}

type Client struct {
	Conn      *websocket.Conn `json:"-"`
	Id        string          `json:"clientId"`
	Wallet    int             `json:"wallet"`
	Last_seen int64           `json:"-"`
	Session   *Session        `json:"-"`
	Ip        string          `json:"-"`
}

func NewClient(c *websocket.Conn) *Client {
	return &Client{
		Conn:      c,
		Id:        uuid.NewString(),
		Wallet:    100, // free money
		Last_seen: time.Now().Unix(),
		Ip:        c.RemoteAddr().String(),
	}
}

func (c *Client) Play(message *DefaultMessage) (*ErrorResultMessage, error) {
	if c.Session == nil {
		cErr := &ErrorResultMessage{
			Kind:    "ERROR",
			Message: "Attemped to play without session",
			Code:    NO_SESSION,
		}

		return cErr, nil
	}

	p := &PlayMessage{
		Bet:    message.Bet,
		Choice: message.Choice,
	}

	if p.Bet > c.Wallet {
		cErr := &ErrorResultMessage{
			Kind:    "ERROR",
			Message: "Insufficient balance",
			Code:    NO_BALANCE,
		}

		return cErr, nil
	}

	// implement DDA for fun?
	num := rand.Intn(6) + 1

	var win bool
	switch p.Choice {
	case "ODD":
		win = num%2 != 0
	case "EVEN":
		win = num%2 == 0
	}

	var res string
	if win {
		c.Session.Profit += p.Bet
		res = "WIN"
	} else {
		c.Session.Profit -= p.Bet
		res = "LOSE"
	}

	pResult := PlayResultMessage{
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
	err := c.SendMessage(pResult)
	if err != nil {
		return nil, err
	}

	slog.Debug("Completed Play", slog.String("id", c.Id), slog.Int("wallet", c.Wallet), slog.String("choice", p.Choice), slog.Int("bet", p.Bet), slog.String("result", res), slog.Int("roll", num))
	return nil, nil
}

func (c *Client) GetWallet() error {
	wMessage := WalletMessage{
		Kind:   "WALLET",
		Wallet: c.Wallet,
	}

	err := c.SendMessage(wMessage)
	if err != nil {
		slog.Error("GetWallet: Failed to send WalletMessage")
		return err
	}

	slog.Debug("GetWallet", slog.String("id", c.Id), slog.Int("wallet", c.Wallet))
	return nil
}

func (c *Client) StartSession() (*ErrorResultMessage, error) {
	if c.Session != nil && c.Session.Playing {
		cError := &ErrorResultMessage{
			Kind:    "ERROR",
			Message: "Already playing",
			Code:    ALREADY_PLAYING,
		}

		return cError, nil
	}

	c.Session = &Session{
		Playing: true,
		Profit:  0,
		PlayHistory: &PlayHistory{
			Items: make([]PlayHistoryItem, 0),
		},
	}

	err := c.SendMessage(&InfoResultMessage{
		Kind: "STARTPLAY",
	})
	if err != nil {
		return nil, err
	}

	slog.Debug("Session started", slog.String("id", c.Id))
	return nil, nil
}

func (c *Client) EndSession() (*ErrorResultMessage, error) {
	if !c.Session.Playing {
		cError := &ErrorResultMessage{
			Kind:    "ERROR",
			Message: "Not playing",
			Code:    NOT_PLAYING,
		}
		return cError, nil
	}

	newBalance := c.Wallet + c.Session.Profit
	c.Wallet = newBalance

	err := c.SendMessage(&EndPlayResultMessage{
		Kind:   "ENDPLAY",
		Profit: c.Session.Profit,
		Wallet: c.Wallet,
	})
	if err != nil {
		return nil, err
	}

	c.Session.Reset()
	slog.Debug("Session ended", slog.String("id", c.Id))
	return nil, nil
}

func SendConnMessage(conn *websocket.Conn, msg interface{}) error {
	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	err = conn.WriteMessage(websocket.TextMessage, data)
	if err != nil {
		return err
	}

	slog.Debug("Sent message", slog.String("message", string(data)))
	return nil
}

func (c *Client) SendMessage(msg interface{}) error {
	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	err = c.Conn.WriteMessage(websocket.TextMessage, data)
	if err != nil {
		return err
	}

	slog.Debug("Sent message", slog.String("client", c.Id), slog.String("message", string(data)))
	return nil
}

func (c *Client) SendErrorMessage(cErr *ErrorResultMessage) error {
	err := c.SendMessage(cErr)
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) SendIErrorMessage(err error) {
	// idk if this only makes sense to me but the logic behind using these 2 erros types
	// is like differentiating between a 400's and 500's
	// 400's i assume the program is actually working and report to the client with a SendIErrorMessage
	// 500's i try reporting to the client but assume it might not work and log the error

	cErr := &ErrorResultMessage{
		Kind:    "ERROR",
		Message: err.Error(),
		Code:    INTERNAL,
	}

	c.SendErrorMessage(cErr)
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

	slog.Debug("PlayHistory: Added item", slog.String("choice", item.Choice), slog.Int("bet", item.Bet), slog.String("result", item.Result), slog.Int("roll", item.Roll))
}

type Session struct {
	Playing     bool
	Profit      int // should always be 0 if not playing
	PlayHistory *PlayHistory
}

func (s *Session) Reset() {
	s.PlayHistory = nil
	s.Profit = 0
	s.Playing = false
}

var St = &Store{
	Clients: make(map[string]*Client),
	Mx:      &sync.Mutex{},
}

func InitC(conn *websocket.Conn, msg *DefaultMessage) (c *Client, cErr *ErrorResultMessage, err error) {
	// implicit named return
	isAuth := msg.Kind == "AUTH"

	if isAuth {
		c = NewClient(conn)
		St.AddClient(c)

		c.SendMessage(&AuthResultMessage{
			Kind:     "AUTH",
			ClientId: c.Id,
		})

		return
	}

	if !isAuth {
		cErr = &ErrorResultMessage{
			Kind:    "ERROR",
			Message: "not authenticated",
			Code:    NOT_AUTHENTICATED,
		}
		return
	}
	return
}
