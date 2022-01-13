# Build

FROM golang:1.16-alpine as builder

WORKDIR /app

COPY . .

RUN CGO_ENABLED=0 go mod download && go build -o burung-chat main.go

# Run

FROM alpine:3.14

WORKDIR /app

COPY --from=builder /app/burung-chat .
COPY .env .
COPY static static

ENV REDIS_HOST_DOCKER=rediscache:6379

EXPOSE 8080

CMD [ "/app/burung-chat" ]
