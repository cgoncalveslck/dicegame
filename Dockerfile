# Choose whatever you want, version >= 1.16
FROM golang:1.23-alpine

WORKDIR /app


COPY go.mod ./
COPY go.sum ./

RUN go mod download

# Copy the cmd directory (where your Go code is located)
COPY cmd ./cmd

RUN CGO_ENABLED=0 go build -o /app/main ./cmd/dicegame/main.go

CMD ["/app/main"]
