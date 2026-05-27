FROM golang:1.22 AS builder
WORKDIR /app

RUN apt-get update && apt-get install -y gcc

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=1 GOOS=linux go build -o endpoint-parser .

FROM debian:bookworm-slim
RUN apt-get update && apt-get install -y ca-certificates && rm -rf /var/lib/apt/lists/*
WORKDIR /app
COPY --from=builder /app/endpoint-parser .
EXPOSE 8080
CMD ["./endpoint-parser"]
