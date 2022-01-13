package handler

import (
	"context"
	"encoding/json"
	"log"
	"os"

	"github.com/go-redis/redis/v8"
)

const CHANNEL = "ws-message"

var ctx = context.Background()
var redisClient *redis.Client

// Publish Message to Redis
func PublishMessage(payload SocketPayload) error {
	var goEnv = os.Getenv("GO_ENV")
	if goEnv == "testing" {
		var redisHost = os.Getenv("REDIS_HOST")
		redisClient = redis.NewClient(&redis.Options{
			Addr: redisHost,
		})
	}
	payloadByte, _ := json.Marshal(payload)
	if err := redisClient.Publish(ctx, CHANNEL, payloadByte).Err(); err != nil {
		return err
	}

	return nil
}

// Subscribe to redis channel
func InitRedis() {
	var redisHost = os.Getenv("REDIS_HOST")
	redisClient = redis.NewClient(&redis.Options{
		Addr: redisHost,
	})

	subscriber := redisClient.Subscribe(ctx, CHANNEL)
	payload := SocketPayload{}

	for {
		msg, err := subscriber.ReceiveMessage(ctx)
		if err != nil {
			log.Fatal(err)
		}
		if err := json.Unmarshal([]byte(msg.Payload), &payload); err != nil {
			log.Println(err)
		}

		go BroadcastMessage(payload)
	}
}
