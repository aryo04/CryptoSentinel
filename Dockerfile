# ===== STAGE 1: BUILD =====
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Copy go.mod & go.sum
COPY go.mod go.sum ./

# Copy SDK lokal (karena pakai replace => ./teneo-agent-sdk)
COPY teneo-agent-sdk ./teneo-agent-sdk

# Download semua module (termasuk SDK lokal)
RUN go mod download

# Copy seluruh source code
COPY . .

# Build binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -o cryptosentinel ./cmd/cryptosentinel

# ===== STAGE 2: RUNTIME =====
FROM alpine:3.20

WORKDIR /app

RUN apk add --no-cache ca-certificates tzdata

COPY --from=builder /app/cryptosentinel .

CMD ["./cryptosentinel"]
