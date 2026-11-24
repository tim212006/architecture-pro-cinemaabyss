FROM golang:1.22 AS builder

WORKDIR /app

COPY go.mod ./
COPY main.go ./

RUN CGO_ENABLED=0 GOOS=linux go build -o proxy .

FROM alpine:3.20

WORKDIR /app

COPY --from=builder /app/proxy /app/proxy

EXPOSE 8000

CMD ["/app/proxy"]
