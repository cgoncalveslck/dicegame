# Choose whatever you want, version >= 1.16
FROM golang:1.23-alpine

WORKDIR /app

# I'm just using this for convenience, hot-reloading etc
RUN go install github.com/air-verse/air@latest

COPY go.mod ./
RUN go mod download

CMD ["air", "-c", ".air.toml"]