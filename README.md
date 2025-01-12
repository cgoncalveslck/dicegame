# Dice Game WebSocket Backend

Simple dice game built for a technical challenge. Built with **Golang backend** for handling game logic and WebSocket communication using [gorilla/websocket](https://github.com/gorilla/websocket), and a **Next.js frontend** for the UI.


## Project Structure


    ├── cmd                                       # All Golang backend code
    ├── web                                       # Small Next.js app to interact with backend API
    ├── DockerFile                                # Dockerfile to build Golang backend
    ├── docker-compose.yml
    └── README.md


## Backend

### Deployment

The API is deployed on [Hetzner](https://www.hetzner.com/) in my VPS using [Coolify](https://coolify.io/).<br>Redeploys setup on push/merge to master branch<br>

#### Deployed at `ws://kkc4so8s0g4c4k0kck0kkgcs.188.245.241.81.sslip.io`

#### Repo
Repo is setup with a github action to build the app and run the tests.<br>
Pull requests on `master` require checks to pass. <br>Rebase merging also required.

### Tests

Implemented tests on `cmd/internal/client_test.go` <br>
Coverage at around 70%

- Run tests with `go test -v ./...`

#### Also used [Excalidraw](https://excalidraw.com/) to "draw" a broad/high-level representation of the problem<br>This was to help me visualize the problem and could be used as documentation of sorts.<br> Link [here](https://excalidraw.com/#json=6-G21rvkM22iVunuzzvjs,WrC-wmp-MJd6DvOfiOf9Kw)


### Setup locally

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
#### API will start on `localhost:8181`

  ## Frontend

Bugs are expected here, the UI was made mostly for fun and to help visualize the problem(s) to solve, didn't put too much time and attention into it.

### Setup locally

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

### Deployment

The UI is deployed on [Vercel](https://vercel.com/).<br>Redeploys also setup on push/merge to master branch<br>

### Deployed [here](https://dicegame-rho-seven.vercel.app)
  ---

  Docker (Optional)

  1. Build and run
  ```sh
  docker compose up --build -d
  ```

  #### App will start on `localhost:3000`, API will start on `localhost:8181`
  <br>


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

| Code | Constant           | Description                                                                 |
|------|--------------------|-----------------------------------------------------------------------------|
| 1    | `NO_BALANCE`       | Bet more than the balance                                                   |
| 2    | `NO_SESSION`       | No active session                                                           |
| 3    | `NOT_PLAYING`      | Round didn't start                                                          |
| 4    | `ALREADY_PLAYING`  | Already in a round                                                          |
| 5    | `INVALID_UUID`     | The provided `clientId` is invalid                                          |
| 6    | `INVALID_JASON`    | The request contains invalid JSON syntax                                    |
| 7    | `ALREADY_LOGGED`   | Already got a `clientId`                                                    |
| 8    | `INVALID_BET`      | The bet amount is invalid (e.g., less than 1)                               |
| 9    | `INVALID_CHOICE`   | The choice is invalid (e.g., not "ODD" or "EVEN")                           |
| 10   | `UNKNOWN_KIND`     | The `kind` field in the request is unknown or unsupported                   |
| 11   | `CLIENT_NOT_FOUND` | The `clientId` does not correspond to any active client                     |
