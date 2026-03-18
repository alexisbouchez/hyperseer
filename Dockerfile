FROM golang:1.26-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o serve ./cmd/serve

FROM alpine:3.21
RUN apk add --no-cache ca-certificates
COPY --from=builder /app/serve /serve
EXPOSE 4318 7777
CMD ["/serve"]
