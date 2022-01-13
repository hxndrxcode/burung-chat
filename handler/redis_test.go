package handler

import (
	"log"
	"os"
	"testing"

	"github.com/joho/godotenv"
)

func init() {
	godotenv.Load("../.env")
	if r := os.Getenv("REDIS_HOST"); r == "" {
		log.Println("REDIS_HOST is not defined")
		os.Setenv("REDIS_HOST", "localhost:6379")
	}
	os.Setenv("GO_ENV", "testing")
}

func TestPublishMessage(t *testing.T) {
	payload := SocketPayload{
		Message:  "hello",
		Action:   "message",
		Username: "testuser",
		Room:     "testroom",
	}
	err := PublishMessage(payload)
	if err != nil {
		t.Error("Error when PublishMessage", err)
	}
}
