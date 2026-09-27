# ---- build stage ----
FROM golang:1.27-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

ARG VERSION=dev
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags "-X main.Version=${VERSION}" -o /app/bin/api ./cmd/api

# ---- final stage ----
FROM alpine:3.20

RUN apk add --no-cache ca-certificates

WORKDIR /app

COPY --from=builder /app/bin/api .

EXPOSE 8080

CMD ["./api"]
