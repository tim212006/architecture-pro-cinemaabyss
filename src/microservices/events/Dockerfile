FROM golang:1.23 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY main.go ./

RUN CGO_ENABLED=0 GOOS=linux go build -o events .

FROM alpine:3.20

WORKDIR /app

COPY --from=builder /app/events /app/events

EXPOSE 8082

CMD ["/app/events"]
