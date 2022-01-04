# Build stage
FROM golang:1.17-alpine3.15 AS builder
WORKDIR /app

COPY . .
RUN go build -o main cmd/main.go

# Run stage
FROM alpine:3.15
WORKDIR /app

COPY --from=builder /app/main .
COPY ./.env .

CMD [ "/app/main" ]