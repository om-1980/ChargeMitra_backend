FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod go.sum* ./
RUN go mod download

COPY . .

RUN go build -o api-service ./apps/api/main.go

FROM alpine:3.20

WORKDIR /app

COPY --from=builder /app/api-service /app/api-service

EXPOSE 8080

CMD ["/app/api-service"]