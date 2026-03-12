FROM golang:1.26-alpine AS builder

RUN apk add --no-cache git
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o main ./cmd/truewatcher

FROM alpine:latest

# This enables users to use TZ environment variable for the container to match the desirerd timezone
RUN apk add tzdata

WORKDIR /root/
COPY --from=builder /app/main .

CMD ["./main"]
