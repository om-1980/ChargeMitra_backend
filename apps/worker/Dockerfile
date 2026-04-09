FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod go.sum* ./
RUN go mod download

COPY . .

RUN go build -o worker-service ./apps/worker/main.go

FROM alpine:3.20

WORKDIR /app

COPY --from=builder /app/worker-service /app/worker-service

EXPOSE 8082

CMD ["/app/worker-service"]