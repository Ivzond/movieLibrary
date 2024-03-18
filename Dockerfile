FROM golang:1.21.6 AS builder

WORKDIR /app

COPY go.mod .
COPY go.sum .

RUN go mod download

COPY . .

RUN go build -o movie_library .

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/movie_library .

EXPOSE 8080

CMD ["./movie_library"]


