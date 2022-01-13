package main

import (
	"burung-chat/handler"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Println("Error loading .env file")
		log.Fatal(err)
	}

	if m := os.Getenv("MONGO_URI"); m == "" {
		log.Println("MONGO_URI is not defined")
		os.Setenv("MONGO_URI", "mongodb://localhost:27020/db_burung-chat")
	}
	if r := os.Getenv("REDIS_HOST_DOCKER"); r != "" {
		os.Setenv("REDIS_HOST", r)
	}
	if r := os.Getenv("REDIS_HOST"); r == "" {
		log.Println("REDIS_HOST is not defined")
		os.Setenv("REDIS_HOST", "localhost:6379")
	}

	go handler.InitRedis()
	r := gin.Default()
	r.Static("/asset", "./static")
	r.GET("/", handler.Index)
	r.POST("/login", handler.Login)
	r.GET("/room", handler.FetchRoom)
	r.POST("/room", handler.CreateRoom)
	r.POST("/join", handler.JoinRoom)
	r.GET("/chat", handler.FetchChat)
	r.GET("/group", handler.FetchGroup)
	r.GET("/user", handler.FetchUser)
	r.Any("/ws", handler.ServeWS)
	r.Run()

}
