# Dice Game WebSocket Backend

Simple dice game built for a technical challenge. Built with **Golang backend** for handling game logic and WebSocket communication using [gorilla/websocket](https://github.com/gorilla/websocket), and a **Next.js frontend** for the UI.


## Project Structure


    ├── cmd                                       # All Golang backend code
    ├── web                                       # Small Next.js app to interact with backend API
    ├── DockerFile                                # Lazy dockerfile to build Golang program with hot-reloading using Air**
    ├── docker-compose.yml
    └── README.md

#### ** Air - [Live reload for Go apps](https://github.com/air-verse/air)

---

## How to setup locally

#### Backend (Go)

  1. Navigate to backend directory
  ```sh
  cd cmd/dicegame
  ```

  2. Install dependencies:
  ```sh
  go mod download
  ```

  3. Run the server
  ```sh
go run main.go
  ```
#### API will start on `localhost:81818`


  ---

  Frontend (Next.js)

  1. Navigate to frontend directory
  ```sh
  cd web/dicegame
  ```

  2. Install dependencies
  ```sh
npm install
  ```

  3. Run the dev server
  ```sh
  npm run dev
  ```

  #### App will start on `localhost:3000`

  ---

  Docker (Optional)

  1. Build and run
  ```sh
  docker compose up -d
  ```

  #### App will start on `localhost:3000`, API will start on `localhost:81818`



# API Documentation


## Messages

### 1. **AUTH**
#### Request:
```json
{
    "kind": "AUTH"
}
```
#### Purpose:
- The client sends this message to request a unique `clientId` (UUID) from the server.
- The server generates a UUID and sends it back to the client. The client must use this `clientId` in all subsequent messages to identify itself.

#### Response:
```json
{
    "kind": "AUTH",
    "clientId": "e044e924-f292-427f-b8f4-ef367d75b5ee"
}
```

---

### 2. **STARTPLAY**
#### Request:
```json
{
    "kind": "STARTPLAY",
    "clientId": "e044e924-f292-427f-b8f4-ef367d75b5ee"
}
```
#### Purpose:
- The client sends this message to start a new game round.
- The server initializes tracking for wins, losses, and profit/loss for this round.

#### Response:
```json
{
    "kind": "STARTPLAY",
}
```

---

### 3. **PLAY**
#### Request:
```json
{
    "kind": "PLAY",
    "clientId": "e044e924-f292-427f-b8f4-ef367d75b5ee",
    "bet": 10,
    "choice": "ODD"
}
```
#### Fields:
- `bet`: The amount the client is betting. This cannot exceed the client's current balance.
- `choice`: The client's bet choice. Valid options are `"ODD"` or `"EVEN"`.

#### Purpose:
- The client sends this message to place a bet.
- The server rolls a dice, determines the outcome, and updates the client's profit/loss for the current round.

#### Response:
```json
{
    "kind": "ROLL",
    "roll": 5, // The dice roll result (1-6)
    "result": "WIN", // "WIN" or "LOSE"
}
```

---

### 4. **ENDPLAY**
#### Request:
```json
{
    "kind": "ENDPLAY",
    "clientId": "e044e924-f292-427f-b8f4-ef367d75b5ee"
}
```
#### Purpose:
- The client sends this message to end the current game round.
- The server calculates the net profit/loss for the round and updates the client's balance.

#### Response:
```json
{
    "kind": "ENDPLAY",
    "result": -20, // The net profit/loss for the round
    "wallet": 80, // The updated balance after applying the net profit/loss
}
```

---

### 5. **WALLET**
#### Request:
```json
{
    "kind": "WALLET",
    "clientId": "e044e924-f292-427f-b8f4-ef367d75b5ee"
}
```
#### Purpose:
- The client sends this message to retrieve their current wallet balance.
- The balance returned does not include any profit/loss from the current round (only finalized balances after `ENDPLAY`).

#### Response:
```json
{
    "kind": "WALLET",
    "balance": 100, // The current wallet balance
}
```

---

## Error Handling

If an error occurs, the server will respond with an `ErrorResultMessage`:

```json
{
    "kind": "ERROR",
    "message": "Invalid JSON syntax",
    "code": 7 // Numeric error code
}
```

### Error Codes
The `code` field in the error response corresponds to one of the following constants:

| Code | Constant          | Description                                                                 |
|------|-------------------|-----------------------------------------------------------------------------|
| 1    | `NO_BALANCE`      | Bet more than the balance              |
| 2    | `NO_SESSION`      | No active session                               |
| 3    | `NOT_PLAYING`     | Round didn't start                                |
| 4    | `ALREADY_PLAYING` | Already in a round                             |
| 5    | `INTERNAL`        | Internal/unexpected server error occurred                                          |
| 6    | `INVALID_UUID`    | The provided `clientId` is invalid                       |
| 7    | `INVALID_JASON`   | The request contains invalid JSON syntax.                                   |
| 8    | `NOT_AUTHENTICATED` | The client is not authenticated (missing or invalid `clientId`).          |
| 9    | `ALREADY_LOGGED`  | Already got a `clientId`                       |
