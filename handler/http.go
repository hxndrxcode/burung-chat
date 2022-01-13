package handler

import (
	"burung-chat/database"
	"log"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Serve static home page
func Index(c *gin.Context) {
	c.File("./static/index.html")
}

// Handle User Login (Create New or Get Existing User)
func Login(c *gin.Context) {
	username := c.PostForm("username")
	user := database.User{
		Username: username,
	}
	err := database.LoginUser(&user)
	if err != nil {
		c.JSON(401, "unauthorized")
		return
	}
	c.JSON(200, user)
}

// Get All Room by User
func FetchRoom(c *gin.Context) {
	username := c.Query("username")
	rooms, err := database.FetchRoom(username)
	if err != nil {
		log.Println(err)
		c.JSON(404, "notfound")
		return
	}

	c.JSON(200, rooms)
}

// Create New Room
func CreateRoom(c *gin.Context) {
	tipe := c.PostForm("type")
	label := c.PostForm("label")
	user1 := c.PostForm("user1")
	user2 := c.PostForm("user2")

	if tipe == "p2p" {
		getRoom, err := database.GetP2PRoom(user1, user2)
		if err == nil {
			c.JSON(200, getRoom)
			return
		}
	}

	room := database.Room{
		Label: label,
		Type:  tipe,
	}
	err := database.CreateRoom(&room)
	if err != nil {
		c.JSON(500, "error")
		return
	}

	database.JoinRoom(room.ID, user1)
	if tipe == "p2p" {
		database.JoinRoom(room.ID, user2)
	}
	c.JSON(200, room)
}

// Join Room
func JoinRoom(c *gin.Context) {
	room := c.PostForm("room")
	username := c.PostForm("username")

	roomID, err := primitive.ObjectIDFromHex(room)
	if err != nil {
		c.JSON(500, "error")
		return
	}

	database.JoinRoom(roomID, username)
	c.JSON(200, room)
}

// Get All Chat of the Room
func FetchChat(c *gin.Context) {
	room := c.Query("room")
	chats, err := database.FetchChat(room)
	if err != nil {
		c.JSON(404, "notfound")
		return
	}

	c.JSON(200, chats)
}

// Get All Room (type group)
func FetchGroup(c *gin.Context) {
	groups, err := database.FetchGroup()
	if err != nil {
		c.JSON(404, "notfound")
		return
	}

	c.JSON(200, groups)
}

// Get All Registered User
func FetchUser(c *gin.Context) {
	users, err := database.FetchUser()
	if err != nil {
		c.JSON(404, "notfound")
		return
	}

	c.JSON(200, users)
}
