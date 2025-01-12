package client_test

import (
	"cgoncalveslck/dicegame/cmd/internal/client"
	"cgoncalveslck/dicegame/cmd/internal/handlers"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

type AuthMessage struct {
	Kind string `json:"kind"`
}

type WalletMessage struct {
	Kind     string `json:"kind"`
	ClientId string `json:"clientId"`
}

type EndPlayMessage struct {
	Kind     string `json:"kind"`
	ClientId string `json:"clientId"`
}

type StartSessionMessage struct {
	Kind     string `json:"kind"`
	ClientId string `json:"clientId"`
}

var validUUID string
var ws *websocket.Conn
var s *httptest.Server

func TestMain(m *testing.M) {
	var err error

	// server
	s = httptest.NewServer(http.HandlerFunc(handlers.Handler))

	// http to ws
	u := "ws" + strings.TrimPrefix(s.URL, "http")

	// client
	ws, _, err = websocket.DefaultDialer.Dial(u, nil)
	if err != nil {
		fmt.Printf("Error: %+v\n", err)
	}

	m.Run()
}

func TestUnknowKind(t *testing.T) {
	authM := &AuthMessage{
		Kind: "NOTAUTH",
	}

	err := ws.WriteJSON(authM)
	if err != nil {
		t.Errorf("Error: %+v", err)
	}

	cErr := &client.ErrorResultMessage{}
	err = ws.ReadJSON(cErr)
	if err != nil {
		t.Errorf("Error: %+v", err)
	}

	if cErr.Kind != "ERROR" {
		t.Errorf("Expected ERROR as Kind but got %s", cErr.Kind)
	}

	if cErr.Message == "" {
		t.Errorf("Expected Message but got empty string")
	}

	if cErr.Code != client.UNKNOWN_KIND {
		t.Errorf("Expected UNKNOWN_KIND as Code but got %d", cErr.Code)
	}
}

func TestInvalidJSON(t *testing.T) {
	invalidJSON := `{"kind": "AUTH" "clientId": "1234"}`

	cErr := &client.ErrorResultMessage{}
	err := ws.WriteMessage(websocket.TextMessage, []byte(invalidJSON))
	if err != nil {
		t.Errorf("Error: %+v", err)
	}

	err = ws.ReadJSON(cErr)
	if err != nil {
		t.Errorf("Error: %+v", err)
	}

	if cErr.Kind != "ERROR" {
		t.Errorf("Expected ERROR as Kind but got %s", cErr.Kind)
	}

	if cErr.Message == "" {
		t.Errorf("Expected Message but got empty string")
	}

	if cErr.Code != client.INVALID_JASON {
		t.Errorf("Expected INVALID_JSON as Code but got %d", cErr.Code)
	}
}

// writing these is boring but i found a bug while writing this one
func TestInvalidClientId(t *testing.T) {
	authM := &WalletMessage{
		Kind:     "WALLET",
		ClientId: "NOTUUID",
	}

	err := ws.WriteJSON(authM)
	if err != nil {
		t.Errorf("Error: %+v", err)
	}

	cErr := &client.ErrorResultMessage{}
	err = ws.ReadJSON(cErr)
	if err != nil {
		t.Errorf("Error: %+v", err)
	}

	if cErr.Kind != "ERROR" {
		t.Errorf("Expected ERROR as Kind but got %s", cErr.Kind)
	}

	if cErr.Message == "" {
		t.Errorf("Expected Message but got empty string")
	}

	if cErr.Code != client.INVALID_UUID {
		t.Errorf("Expected INVALID_UUID as Code but got %d", cErr.Code)
	}
}

func TestValidClientId(t *testing.T) {
	rndUUID := uuid.New().String()

	authM := &WalletMessage{
		Kind:     "WALLET",
		ClientId: rndUUID,
	}

	err := ws.WriteJSON(authM)
	if err != nil {
		t.Errorf("Error: %+v", err)
	}

	cErr := &client.ErrorResultMessage{}
	err = ws.ReadJSON(cErr)
	if err != nil {
		t.Errorf("Error: %+v", err)
	}

	if cErr.Kind != "ERROR" {
		t.Errorf("Expected ERROR as Kind but got %s", cErr.Kind)
	}

	if cErr.Message == "" {
		t.Errorf("Expected Message but got empty string")
	}

	if cErr.Code != client.CLIENT_NOT_FOUND {
		t.Errorf("Expected CLIENT_NOT_FOUND as Code but got %d", cErr.Code)
	}
}

func TestValidAuth(t *testing.T) {
	authM := &AuthMessage{
		Kind: "AUTH",
	}

	err := ws.WriteJSON(authM)
	if err != nil {
		t.Errorf("Error: %+v", err)
	}

	authRM := &client.AuthResultMessage{}
	err = ws.ReadJSON(authRM)
	if err != nil {
		t.Errorf("Error: %+v", err)
	}

	if authRM.Kind != "AUTH" {
		t.Errorf("Expected AUTH as Kind but got %s", authRM.Kind)
	}

	if authRM.ClientId == "" {
		t.Errorf("Expected ClientId but got empty string")
	}

	_, err = uuid.Parse(authRM.ClientId)
	if err != nil {
		t.Errorf("Expected valid UUID: Error: %+v", err)
	}

	validUUID = authRM.ClientId
}

func TestNoSession(t *testing.T) {
	pMsg := &client.PlayMessage{
		Kind:     "PLAY",
		Bet:      10,
		Choice:   "ODD",
		ClientId: validUUID,
	}

	err := ws.WriteJSON(pMsg)
	if err != nil {
		t.Errorf("Error: %+v", err)
	}

	cErr := &client.ErrorResultMessage{}
	err = ws.ReadJSON(cErr)
	if err != nil {
		t.Errorf("Error: %+v", err)
	}

	if cErr.Kind != "ERROR" {
		t.Errorf("Expected ERROR as Kind but got %s", cErr.Kind)
	}

	if cErr.Message == "" {
		t.Errorf("Expected Message but got empty string")
	}

	if cErr.Code != client.NO_SESSION {
		t.Errorf("Expected NO_SESSION as Code but got %d", cErr.Code)
	}
}

func TestEndSessionWithoutSession(t *testing.T) {
	eMsg := &EndPlayMessage{
		Kind:     "ENDPLAY",
		ClientId: validUUID,
	}

	err := ws.WriteJSON(eMsg)
	if err != nil {
		t.Errorf("Error: %+v", err)
	}

	cErr := &client.ErrorResultMessage{}
	err = ws.ReadJSON(cErr)
	if err != nil {
		t.Errorf("Error: %+v", err)
	}

	if cErr.Kind != "ERROR" {
		t.Errorf("Expected ERROR as Kind but got %s", cErr.Kind)
	}

	if cErr.Message == "" {
		t.Errorf("Expected Message but got empty string")
	}

	if cErr.Code != client.NO_SESSION {
		t.Errorf("Expected NO_SESSION as Code but got %d", cErr.Code)
	}
}

func TestStartSession(t *testing.T) {
	sMsg := &StartSessionMessage{
		Kind:     "STARTPLAY",
		ClientId: validUUID,
	}

	err := ws.WriteJSON(sMsg)
	if err != nil {
		t.Errorf("Error: %+v", err)
	}

	msg := &client.ErrorResultMessage{}
	err = ws.ReadJSON(msg)
	if err != nil {
		t.Errorf("Error: %+v", err)
	}

	if msg.Kind != "STARTPLAY" {
		t.Errorf("Expected STARTPLAY as Kind but got %s", msg.Kind)
	}
}

func TestNoBalance(t *testing.T) {
	pMsg := &client.PlayMessage{
		Kind:     "PLAY",
		Bet:      1000,
		Choice:   "ODD",
		ClientId: validUUID,
	}

	err := ws.WriteJSON(pMsg)
	if err != nil {
		t.Errorf("Error: %+v", err)
	}

	cErr := &client.ErrorResultMessage{}
	err = ws.ReadJSON(cErr)
	if err != nil {
		t.Errorf("Error: %+v", err)
	}

	if cErr.Kind != "ERROR" {
		t.Errorf("Expected ERROR as Kind but got %s", cErr.Kind)
	}

	if cErr.Message == "" {
		t.Errorf("Expected Message but got empty string")
	}

	if cErr.Code != client.NO_BALANCE {
		t.Errorf("Expected NO_BALANCE as Code but got %d", cErr.Code)
	}
}

func TestInvalidBet(t *testing.T) {
	pMsg := &client.PlayMessage{
		Kind:     "PLAY",
		Bet:      0,
		Choice:   "ODD",
		ClientId: validUUID,
	}

	err := ws.WriteJSON(pMsg)
	if err != nil {
		t.Errorf("Error: %+v", err)
	}

	cErr := &client.ErrorResultMessage{}
	err = ws.ReadJSON(cErr)
	if err != nil {
		t.Errorf("Error: %+v", err)
	}

	if cErr.Kind != "ERROR" {
		t.Errorf("Expected ERROR as Kind but got %s", cErr.Kind)
	}

	if cErr.Message == "" {
		t.Errorf("Expected Message but got empty string")
	}

	if cErr.Code != client.INVALID_BET {
		t.Errorf("Expected INVALID_BET as Code but got %d", cErr.Code)
	}
}

func TestInvalidChoice(t *testing.T) {
	pMsg := &client.PlayMessage{
		Kind:     "PLAY",
		Bet:      10,
		Choice:   "INVALID",
		ClientId: validUUID,
	}

	err := ws.WriteJSON(pMsg)
	if err != nil {
		t.Errorf("Error: %+v", err)
	}

	cErr := &client.ErrorResultMessage{}
	err = ws.ReadJSON(cErr)
	if err != nil {
		t.Errorf("Error: %+v", err)
	}

	if cErr.Kind != "ERROR" {
		t.Errorf("Expected ERROR as Kind but got %s", cErr.Kind)
	}

	if cErr.Message == "" {
		t.Errorf("Expected Message but got empty string")
	}

	if cErr.Code != client.INVALID_CHOICE {
		t.Errorf("Expected INVALID_CHOICE as Code but got %d", cErr.Code)
	}
}

func TestClientNotFound(t *testing.T) {
	pMsg := &client.PlayMessage{
		Kind:     "PLAY",
		Bet:      10,
		Choice:   "ODD",
		ClientId: uuid.NewString(),
	}

	err := ws.WriteJSON(pMsg)
	if err != nil {
		t.Errorf("Error: %+v", err)
	}

	cErr := &client.ErrorResultMessage{}
	err = ws.ReadJSON(cErr)
	if err != nil {
		t.Errorf("Error: %+v", err)
	}

	if cErr.Kind != "ERROR" {
		t.Errorf("Expected ERROR as Kind but got %s", cErr.Kind)
	}

	if cErr.Message == "" {
		t.Errorf("Expected Message but got empty string")
	}

	if cErr.Code != client.CLIENT_NOT_FOUND {
		t.Errorf("Expected CLIENT_NOT_FOUND as Code but got %d", cErr.Code)
	}
}
func TestAlreadyPlaying(t *testing.T) {
	sMsg := &StartSessionMessage{
		Kind:     "STARTPLAY",
		ClientId: validUUID,
	}

	err := ws.WriteJSON(sMsg)
	if err != nil {
		t.Errorf("Error: %+v", err)
	}

	msg := &client.ErrorResultMessage{}
	err = ws.ReadJSON(msg)
	if err != nil {
		t.Errorf("Error: %+v", err)
	}

	if msg.Kind != "ERROR" {
		t.Errorf("Expected ERROR as Kind but got %s", msg.Kind)
	}

	if msg.Code != client.ALREADY_PLAYING {
		t.Errorf("Expected ALREADY_PLAYING as Code but got %d", msg.Code)
	}
}

func TestValidBet(t *testing.T) {
	vBet := &client.PlayMessage{
		Kind:     "PLAY",
		Bet:      10,
		Choice:   "ODD",
		ClientId: validUUID,
	}

	err := ws.WriteJSON(vBet)
	if err != nil {
		t.Errorf("Error: %+v", err)
	}

	prM := &client.PlayResultMessage{}
	err = ws.ReadJSON(prM)
	if err != nil {
		t.Errorf("Error: %+v", err)
	}

	if prM.Kind != "ROLL" {
		t.Errorf("Expected ROLL as Kind but got %s", prM.Kind)
	}

	if prM.Result != "WIN" && prM.Result != "LOSE" {
		t.Errorf("Expected WIN or LOSE but got %s", prM.Result)
	}

	if prM.Roll < 1 || prM.Roll > 6 {
		t.Errorf("Expected Roll between 1 and 6 but got %d", prM.Roll)
	}
}

func TestInvalidEndSession(t *testing.T) {
	eMsg := &EndPlayMessage{
		Kind:     "ENDPLAY",
		ClientId: "",
	}

	err := ws.WriteJSON(eMsg)
	if err != nil {
		t.Errorf("Error: %+v", err)
	}

	cErr := &client.ErrorResultMessage{}
	err = ws.ReadJSON(cErr)
	if err != nil {
		t.Errorf("Error: %+v", err)
	}

	if cErr.Kind != "ERROR" {
		t.Errorf("Expected ERROR as Kind but got %s", cErr.Kind)
	}

	if cErr.Code != client.INVALID_JASON {
		t.Errorf("Expected INVALID_JASON as Code but got %d", cErr.Code)
	}
}

func TestEndSession(t *testing.T) {
	eMsg := &EndPlayMessage{
		Kind:     "ENDPLAY",
		ClientId: validUUID,
	}

	err := ws.WriteJSON(eMsg)
	if err != nil {
		t.Errorf("Error: %+v", err)
	}

	erM := &client.EndPlayResultMessage{}
	err = ws.ReadJSON(erM)
	if err != nil {
		t.Errorf("Error: %+v", err)
	}

	if erM.Kind != "ENDPLAY" {
		t.Errorf("Expected ENDPLAY as Kind but got %s", erM.Kind)
	}

	if erM.Profit == 0 {
		t.Errorf("Expected changed Profit but got 0")
	}

	if erM.Wallet == 100 {
		t.Errorf("Expected changed Wallet but got 100")
	}
}

func TestNotPlaying(t *testing.T) {
	eMsg := &EndPlayMessage{
		Kind:     "ENDPLAY",
		ClientId: validUUID,
	}

	err := ws.WriteJSON(eMsg)
	if err != nil {
		t.Errorf("Error: %+v", err)
	}

	cErr := &client.ErrorResultMessage{}
	err = ws.ReadJSON(cErr)
	if err != nil {
		t.Errorf("Error: %+v", err)
	}

	if cErr.Kind != "ERROR" {
		t.Errorf("Expected ERROR as Kind but got %s", cErr.Kind)
	}

	if cErr.Message == "" {
		t.Errorf("Expected Message but got empty string")
	}

	if cErr.Code != client.NOT_PLAYING {
		t.Errorf("Expected NOT_PLAYING as Code but got %d", cErr.Code)
	}
}

func TestAlreadyLogged(t *testing.T) {
	authM := &AuthMessage{
		Kind: "AUTH",
	}

	err := ws.WriteJSON(authM)
	if err != nil {
		t.Errorf("Error: %+v", err)
	}

	cErr := &client.ErrorResultMessage{}
	err = ws.ReadJSON(cErr)
	if err != nil {
		t.Errorf("Error: %+v", err)
	}

	if cErr.Kind != "ERROR" {
		t.Errorf("Expected AUTH as Kind but got %s", cErr.Kind)
	}

	if cErr.Code != client.ALREADY_LOGGED {
		t.Errorf("Expected %v code: Error: %v", client.ALREADY_LOGGED, cErr.Code)
	}
}

func TestGetWallet(t *testing.T) {
	gWallet := &WalletMessage{
		Kind:     "WALLET",
		ClientId: validUUID,
	}

	err := ws.WriteJSON(gWallet)
	if err != nil {
		t.Errorf("Error: %+v", err)
	}

	wrMsg := &client.WalletResultMessage{}
	err = ws.ReadJSON(wrMsg)
	if err != nil {
		t.Errorf("Error: %+v", err)
	}

	if wrMsg.Kind != "WALLET" {
		t.Errorf("Expected WALLET as Kind but got %s", wrMsg.Kind)
	}

	if wrMsg.Wallet == 0 {
		t.Errorf("Expected Balance but got 0")
	}
}
