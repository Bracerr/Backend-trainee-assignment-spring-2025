FROM golang:1.23-alpine AS builder

WORKDIR /app

RUN apk add --no-cache gcc musl-dev

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=1 GOOS=linux go build -o main ./src/cmd/app/main.go

FROM golang:1.23-alpine

WORKDIR /app

RUN apk add --no-cache ca-certificates tzdata gcc musl-dev

COPY --from=builder /app/main .
COPY --from=builder /app/migrations ./migrations
COPY . .

RUN mkdir -p /app/logs

RUN adduser -D appuser
RUN chown -R appuser:appuser /app
USER appuser

COPY .env .

CMD ["./main"] 