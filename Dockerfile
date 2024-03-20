FROM golang:1.22-alpine AS builder

WORKDIR /app

COPY go.mod .
COPY go.sum .

RUN go mod download

COPY . .

RUN go build -o movie_library .

EXPOSE 8080

CMD ["./movie_library"]


