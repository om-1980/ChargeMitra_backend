FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod go.sum* ./
RUN go mod download

COPY . .

RUN go build -o ocpp-service ./apps/ocpp/main.go

FROM alpine:3.20

WORKDIR /app

COPY --from=builder /app/ocpp-service /app/ocpp-service

EXPOSE 8081

CMD ["/app/ocpp-service"]