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
#### API will start on `localhost:8080`


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

  #### App will start on `localhost:3000`, API will start on `localhost:8080`

