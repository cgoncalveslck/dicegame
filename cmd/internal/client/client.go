package client

import (
	"encoding/json"
	"fmt"
	"log"
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
	INVALID_UUID
	INVALID_JASON
	ALREADY_LOGGED
	INVALID_BET
	INVALID_CHOICE
	UNKNOWN_KIND
	CLIENT_NOT_FOUND
)

type cError int

const Timeout = 5 // 5min

type EndPlayResultMessage struct {
	Kind   string `json:"kind"`
	Profit int    `json:"result"`
	Wallet int    `json:"wallet"`
}

type PlayMessage struct {
	Kind     string `json:"kind"`
	ClientId string `json:"clientId"`
	Bet      int    `json:"bet"`
	Choice   string `json:"choice"`
}

type PlayResultMessage struct {
	Kind   string `json:"kind"`
	Result string `json:"result"`
	Roll   int    `json:"roll"`
}

type StartSessionResultMessage struct {
	Kind string `json:"kind"`
}

type WalletResultMessage struct {
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
func (s *Store) DisconnectClient(c *Client) {
	s.Mx.Lock()
	defer s.Mx.Unlock()

	if c.Id == "" {
		return
	}

	client, ok := s.Clients[c.Id]
	if ok {
		if client.Conn != nil {
			client.Conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			client.Conn.Close()
		}
		delete(s.Clients, c.Id)
	}

	slog.Debug("Client removed", slog.String("ClientId", c.Id))
}

func (s *Store) AddClient(c *Client) {
	s.Mx.Lock()
	defer s.Mx.Unlock()
	s.Clients[c.Id] = c

	slog.Debug("Client added", slog.String("ClientId", c.Id))
}

func SessionExpire() {
	// i'm not sure if this is good practice
	// i know about channels and contexts which i'm guessing are usually used for this
	// my thought process is that this is supposed to keep running and if it isn't
	// the server isn't either
	// again, i'm still not sure about it but that's how i thought about it
	// i'd gladly be corrected
	timer := time.NewTicker(2 * time.Second)

	for range timer.C {
		// only doing this loop in the context of this project
		// in a real product obviously we wouldn't do this
		for _, c := range St.Clients {
			if time.Now().Unix()-c.Last_seen > Timeout {
				slog.Debug("expiring client")
				St.DisconnectClient(c)
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

func (c *Client) Init() {
	c.Id = uuid.NewString()
	c.Ip = c.Conn.RemoteAddr().String()
	c.Wallet = 100
	c.Last_seen = time.Now().Unix()
}

func (c *Client) Disconnect() {
	St.DisconnectClient(c)
}

func (c *Client) HandleMessageErrors(cErr *ErrorResultMessage, err error, str string) {
	if err != nil {
		log.Printf("%s error: %+v", str, err)
		err := c.SendMessage(err)
		if err != nil {
			log.Printf("SendMessage error: %+v", err)
		}
	}
	if cErr != nil {
		err := c.SendErrorMessage(cErr)
		if err != nil {
			log.Printf("SendErrorMessage error: %+v", err)
			err = c.SendMessage(err)
			if err != nil {
				log.Printf("SendMessage error: %+v", err)
			}
		}
	}
}

func (c *Client) Play(msg *DefaultMessage) (*ErrorResultMessage, error) {
	if msg.Kind != "PLAY" || msg.ClientId == "" {
		cError := &ErrorResultMessage{
			Kind:    "ERROR",
			Message: "Invalid message",
			Code:    INVALID_JASON,
		}

		return cError, nil
	}

	// needs to start a session/round first
	if c.Session == nil {
		cErr := &ErrorResultMessage{
			Kind:    "ERROR",
			Message: "Attemped to play without session",
			Code:    NO_SESSION,
		}

		return cErr, nil
	}

	var p *PlayMessage
	p, cErr := c.ValidatePlay(msg)
	if cErr != nil {
		return cErr, nil
	}

	// should get 0 to 5 so +1
	// implement DDA for fun?
	num := rand.Intn(6) + 1

	var win bool
	switch p.Choice {
	case "ODD":
		// this should be correct and faster but i didn't check
		win = num&1 != 0
	case "EVEN":
		win = num&1 == 0
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

func (c *Client) ValidatePlay(msg *DefaultMessage) (pMsg *PlayMessage, cErr *ErrorResultMessage) {
	var eMessage string
	var code cError

	switch {
	case msg.Bet > c.Wallet:
		eMessage = "Insufficient points"
		code = NO_BALANCE
	case msg.Bet < 1:
		eMessage = "Invalid bet"
		code = INVALID_BET
	case msg.Choice != "ODD" && msg.Choice != "EVEN":
		eMessage = "Invalid choice (ODD or EVEN)"
		code = INVALID_CHOICE
	}

	if code != 0 {
		cErr = &ErrorResultMessage{
			Kind:    "ERROR",
			Message: eMessage,
			Code:    code,
		}
		return
	}

	pMsg = &PlayMessage{
		Bet:    msg.Bet,
		Choice: msg.Choice,
	}
	return
}

func (c *Client) Auth(conn *websocket.Conn) (*Client, error) {
	if c.Id == "" {
		c.Init()
		St.AddClient(c)

		c.SendMessage(&AuthResultMessage{
			Kind:     "AUTH",
			ClientId: c.Id,
		})

		return c, nil
	}

	addr := conn.RemoteAddr().String()
	// this is a bit RAW and ROUGH but you get the point
	// only doing this loop in the context of this project
	// in a real product obviously we wouldn't do this
	for _, c := range St.Clients {
		if c.Ip == addr {
			err := c.SendMessage(&ErrorResultMessage{
				Kind:    "ERROR",
				Message: "already logged",
				Code:    ALREADY_LOGGED,
			})
			if err != nil {
				log.Printf("SendErrorMessage error: %+v", err)
			}
			return c, nil
		}
	}
	return c, nil
}

func (c *Client) GetWallet(msg *DefaultMessage) (*ErrorResultMessage, error) {
	if msg.Kind != "WALLET" || msg.ClientId == "" {
		cError := &ErrorResultMessage{
			Kind:    "ERROR",
			Message: "Invalid message",
			Code:    INVALID_JASON,
		}

		return cError, nil
	}

	wMessage := WalletResultMessage{
		Kind:   "WALLET",
		Wallet: c.Wallet,
	}

	err := c.SendMessage(wMessage)
	if err != nil {
		slog.Error("GetWallet: Failed to send WalletResultMessage")
		return nil, err
	}

	slog.Debug("GetWallet", slog.String("id", c.Id), slog.Int("wallet", c.Wallet))
	return nil, nil
}

func (c *Client) StartSession(msg *DefaultMessage) (*ErrorResultMessage, error) {
	if msg.ClientId == "" || msg.Kind != "STARTPLAY" {
		cError := &ErrorResultMessage{
			Kind:    "ERROR",
			Message: "Invalid message",
			Code:    INVALID_JASON,
		}

		return cError, nil
	}

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

	err := c.SendMessage(&StartSessionResultMessage{
		Kind: "STARTPLAY",
	})
	if err != nil {
		return nil, err
	}

	slog.Debug("Session started", slog.String("id", c.Id))
	return nil, nil
}

func (c *Client) EndSession(msg *DefaultMessage) (*ErrorResultMessage, error) {
	if msg.ClientId == "" || msg.Kind != "ENDPLAY" {
		cError := &ErrorResultMessage{
			Kind:    "ERROR",
			Message: "Invalid message",
			Code:    INVALID_JASON,
		}

		return cError, nil
	}

	if c.Session == nil {
		cErr := &ErrorResultMessage{
			Kind:    "ERROR",
			Message: "Attemped to end session without session",
			Code:    NO_SESSION,
		}

		return cErr, nil
	}

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

func (c *Client) SendMessage(msg interface{}) error {
	data, err := json.Marshal(msg)
	if err != nil {
		fmt.Println("Error marshalling message", err)
		return err
	}

	err = c.Conn.WriteMessage(websocket.TextMessage, data)
	if err != nil {
		fmt.Println("Error sending message", err)
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
	Playing     bool // might be redundant atm
	Profit      int  // should always be 0 if not playing
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

func HandleClientID(conn *websocket.Conn, msg *DefaultMessage) *ErrorResultMessage {
	_, err := uuid.Parse(msg.ClientId)
	if err != nil {
		return &ErrorResultMessage{
			Kind:    "ERROR",
			Message: "invalid client id, must be UUID",
			Code:    INVALID_UUID,
		}
	}

	_, ok := St.Clients[msg.ClientId]
	if !ok {
		return &ErrorResultMessage{
			Kind:    "ERROR",
			Message: "client not found",
			Code:    CLIENT_NOT_FOUND,
		}
	}

	return nil
}
