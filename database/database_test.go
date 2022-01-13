package database

import (
	"log"
	"os"
	"testing"
	"time"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func init() {
	godotenv.Load("../.env")
	if m := os.Getenv("MONGO_URI"); m == "" {
		log.Println("MONGO_URI is not defined")
		os.Setenv("MONGO_URI", "mongodb://localhost:27020/db_burung-chat")
	}
	os.Setenv("GO_ENV", "testing")
}

func TestGetMongoClient(t *testing.T) {
	_, err := getMongoClient()
	if err != nil {
		t.Error("Error when getMongoClient", err)
	}
}

func TestLoginUser(t *testing.T) {
	user := User{
		Username: "testuser",
	}
	err := LoginUser(&user)
	if err != nil {
		t.Error("Error when LoginUser", err)
	}
	if user.ID == [12]byte{} {
		t.Errorf("Invalid user.ID when LoginUser, %v", user.ID)
	}
}

func TestStoreChat(t *testing.T) {
	chat := Chat{
		RoomID:    primitive.NewObjectID(),
		Message:   "test message",
		CreatedAt: time.Now(),
		From:      "testuser",
		Type:      "message",
	}
	err := StoreChat(chat)
	if err != nil {
		t.Error("Error when StoreChat", err)
	}
}

func TestCreateRoom(t *testing.T) {
	room := Room{
		UpdatedAt: time.Now(),
		Label:     "testgroup",
		Type:      "group",
		Usernames: []string{},
		Active:    0,
	}
	err := CreateRoom(&room)
	if err != nil {
		t.Error("Error when CreateRoom", err)
	}
	if room.ID == [12]byte{} {
		t.Errorf("Invalid room.ID when CreateRoom, %v", room.ID)
	}
}

func TestFetchRoom(t *testing.T) {
	rooms, err := FetchRoom("testuser")
	if err != nil {
		t.Error("Error when FetchRoom", err)
	}
	if len(rooms) > 0 {
		if rooms[0].ID == [12]byte{} {
			t.Errorf("Invalid room.ID when FetchRoom, %v", rooms[0].ID)
		}
	}
}

func TestJoinRoom(t *testing.T) {
	room := Room{
		UpdatedAt: time.Now(),
		Label:     "testroomforjoin",
		Type:      "group",
		Usernames: []string{},
		Active:    0,
	}
	CreateRoom(&room)

	err := JoinRoom(room.ID, "testuser")
	if err != nil {
		t.Error("Error when JoinRoom", err)
	}
}

func TestGetP2PRoom(t *testing.T) {
	user1 := "testuser"
	user2 := "user2"
	room := Room{
		UpdatedAt: time.Now(),
		Label:     user1 + "|" + user2,
		Type:      "p2p",
		Active:    0,
	}
	CreateRoom(&room)
	JoinRoom(room.ID, user1)
	JoinRoom(room.ID, user2)

	getRoom, err := GetP2PRoom(user1, user2)
	if err != nil {
		t.Error("Error when GetP2PRoom", err)
	}
	if getRoom.ID == [12]byte{} {
		t.Errorf("Invalid roomID when GetP2PRoom, %v", getRoom.ID)
	}
}

func TestFetchChat(t *testing.T) {
	room := Room{
		UpdatedAt: time.Now(),
		Label:     "testroomforchat",
		Type:      "group",
		Usernames: []string{},
		Active:    0,
	}
	CreateRoom(&room)

	chats, err := FetchChat(room.ID.Hex())
	if err != nil {
		t.Error("Error when FetchChat", err)
	}
	if len(chats) > 0 {
		if chats[0].ID == [12]byte{} {
			t.Errorf("Invalid chatID when FetchChat, %v", chats[0].ID)
		}
	}
}

func TestFetchGroup(t *testing.T) {
	rooms, err := FetchGroup()
	if err != nil {
		t.Error("Error when FetchGroup", err)
	}
	if len(rooms) > 0 {
		if rooms[0].ID == [12]byte{} {
			t.Errorf("Invalid room.ID when FetchGroup, %v", rooms[0].ID)
		}
		if rooms[0].Type != "group" {
			t.Errorf("Invalid room.Type when FetchGroup, %v", rooms[0].Type)
		}
	}
}

func TestFetchUser(t *testing.T) {
	users, err := FetchUser()
	if err != nil {
		t.Error("Error when FetchUser", err)
	}
	if len(users) > 0 {
		if users[0].ID == [12]byte{} {
			t.Errorf("Invalid room.ID when FetchGroup, %v", users[0].ID)
		}
	}
}
