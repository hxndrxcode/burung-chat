package handler

import (
	"log"
	"net/http"
	"strings"
	"time"

	"burung-chat/database"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	gubrak "github.com/novalagung/gubrak/v2"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var connections = make([]*WebSocketConnection, 0)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

type SocketPayload struct {
	Message  string
	Action   string
	Username string
	Room     string
}

type WebSocketConnection struct {
	*websocket.Conn
	Username string
	Room     string
}

// Serve WS Connnection
func ServeWS(c *gin.Context) {
	upgrader.CheckOrigin = func(r *http.Request) bool { return true }
	currentGorillaConn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Println(err)
		c.JSON(400, "badrequest")
	}

	username := c.Query("username")
	room := c.Query("room")
	go HandleIncomingWS(currentGorillaConn, username, room)
}

// Handle WS Connection
func HandleIncomingWS(currentGorillaConn *websocket.Conn, username string, room string) {
	currentConn := WebSocketConnection{Conn: currentGorillaConn, Username: username, Room: room}
	connections = append(connections, &currentConn)

	defer func() {
		if r := recover(); r != nil {
			log.Println(r)
		}
	}()

	for {
		payload := SocketPayload{}
		err := currentConn.ReadJSON(&payload)
		if err != nil {
			if strings.Contains(err.Error(), "websocket: close") {
				ejectConnection(&currentConn)
				return
			}

			log.Println(err)
			continue
		}

		payload.Room = room
		payload.Username = username
		go PublishMessage(payload)

		roomID, _ := primitive.ObjectIDFromHex(room)
		chat := database.Chat{
			RoomID:    roomID,
			Message:   payload.Message,
			From:      currentConn.Username,
			Type:      "message",
			CreatedAt: time.Now(),
		}
		go database.StoreChat(chat)
	}
}

// Remove Connection
func ejectConnection(currentConn *WebSocketConnection) error {
	filtered := gubrak.From(connections).Reject(func(each *WebSocketConnection) bool {
		return each == currentConn
	}).Result()
	connections = filtered.([]*WebSocketConnection)
	return nil
}

// Send to all connection with certain Room ID
func BroadcastMessage(payload SocketPayload) {
	filtered := gubrak.From(connections).Filter(func(each *WebSocketConnection) bool {
		return each.Room == payload.Room
	}).Result()

	for _, eachConn := range filtered.([]*WebSocketConnection) {
		roomID, _ := primitive.ObjectIDFromHex(payload.Room)
		response := database.Chat{
			From:      payload.Username,
			Message:   payload.Message,
			Type:      "message",
			CreatedAt: time.Now(),
			RoomID:    roomID,
		}
		eachConn.WriteJSON(response)
	}
}
