# Choose whatever you want, version >= 1.16
FROM golang:1.23-alpine

WORKDIR /app

# I'm just using this for convenience, hot-reloading etc
RUN go install github.com/air-verse/air@latest

COPY go.mod ./
COPY go.sum ./

RUN go mod download

# Copy the cmd directory (where your Go code is located)
COPY cmd ./cmd

# Copy the .air.toml file
COPY .air.toml ./

CMD ["air", "-c", ".air.toml"]