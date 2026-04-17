FROM golang:1.26-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o deliberix-server .

FROM alpine:3.23
RUN apk add --no-cache ca-certificates tzdata
COPY --from=builder /app/deliberix-server /usr/local/bin/deliberix-server
COPY --from=builder /app/locales /app/locales

EXPOSE 8080
WORKDIR /app
CMD ["deliberix-server"]
