# syntax=docker/dockerfile:1
FROM golang:1.21 AS builder
WORKDIR /app
COPY go.mod ./
RUN go mod download
COPY tcp_server.go .
RUN go mod tidy && CGO_ENABLED=0 go build -o tcp_server .

FROM alpine:3.20.2
WORKDIR /root/
COPY --from=builder /app/tcp_server .
EXPOSE 3333
CMD ["./tcp_server"]