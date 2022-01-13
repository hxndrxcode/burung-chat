package database

import (
	"context"
	"os"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var clientInstance *mongo.Client
var clientInstanceError error
var mongoOnce = new(sync.Once)

var (
	ctx      = context.TODO()
	mongoURI = ""
	DBNAME   = "db_burung-chat"
)

func getMongoClient() (*mongo.Client, error) {
	mongoOnce.Do(func() {
		mongoURI = os.Getenv("MONGO_URI")
		var goEnv = os.Getenv("GO_ENV")
		if goEnv == "testing" {
			DBNAME = DBNAME + "_test"
		}

		clientOptions := options.Client().ApplyURI(mongoURI)
		client, err := mongo.Connect(ctx, clientOptions)
		if err != nil {
			clientInstanceError = err
		}

		err = client.Ping(ctx, nil)
		if err != nil {
			clientInstanceError = err
		}
		clientInstance = client
		if clientInstanceError == nil && goEnv == "testing" {
			colls, _ := clientInstance.Database(DBNAME).ListCollectionNames(ctx, bson.M{})
			for _, c := range colls {
				clientInstance.Database(DBNAME).Collection(c).Drop(ctx)
			}
		}
	})
	if clientInstanceError != nil {
		mongoOnce = new(sync.Once)
	}
	return clientInstance, clientInstanceError
}

func LoginUser(user *User) error {
	client, err := getMongoClient()
	if err != nil {
		return err
	}

	updater := bson.M{"$set": bson.M{"last_active": time.Now()}}
	filter := bson.M{"username": user.Username}
	userColl := client.Database(DBNAME).Collection("users")
	err = userColl.FindOne(ctx, filter).Decode(&user)
	if err == nil {
		userColl.UpdateByID(ctx, user.ID, updater)
		return nil
	}

	user.ID = primitive.NewObjectID()
	user.Rooms = []primitive.ObjectID{}
	user.LastActive = time.Now()
	_, err = userColl.InsertOne(ctx, user)
	if err != nil {
		return err
	}

	return nil
}

func StoreChat(chat Chat) error {
	client, err := getMongoClient()
	if err != nil {
		return err
	}

	chat.ID = primitive.NewObjectID()
	chatColl := client.Database(DBNAME).Collection("chats")
	_, err = chatColl.InsertOne(ctx, chat)
	if err != nil {
		return err
	}

	updater := bson.M{"$set": bson.M{"updated_at": time.Now(), "active": 1}}
	roomColl := client.Database(DBNAME).Collection("rooms")
	roomColl.UpdateByID(ctx, chat.RoomID, updater)

	updater = bson.M{"$set": bson.M{"last_active": time.Now()}}
	userColl := client.Database(DBNAME).Collection("users")
	userColl.UpdateByID(ctx, chat.From, updater)

	return nil
}

func CreateRoom(room *Room) error {
	client, err := getMongoClient()
	if err != nil {
		return err
	}

	room.ID = primitive.NewObjectID()
	room.Usernames = []string{}
	room.UpdatedAt = time.Now()
	roomColl := client.Database(DBNAME).Collection("rooms")
	_, err = roomColl.InsertOne(ctx, room)
	if err != nil {
		return err
	}

	return nil
}

func FetchRoom(username string) ([]Room, error) {
	rooms := []Room{}
	client, err := getMongoClient()
	if err != nil {
		return rooms, err
	}

	user := User{}
	filter := bson.M{"username": username}
	userColl := client.Database(DBNAME).Collection("users")
	err = userColl.FindOne(ctx, filter).Decode(&user)
	if err != nil {
		return rooms, err
	}

	filter = bson.M{"_id": bson.M{"$in": user.Rooms}}
	findOpt := options.Find().SetSort(bson.M{"updated_at": -1})
	roomColl := client.Database(DBNAME).Collection("rooms")
	cursor, err := roomColl.Find(ctx, filter, findOpt)
	if err != nil {
		return rooms, err
	}

	err = cursor.All(ctx, &rooms)
	return rooms, err
}

func JoinRoom(id primitive.ObjectID, username string) error {
	client, err := getMongoClient()
	if err != nil {
		return err
	}

	room := Room{}
	filter := bson.M{"_id": id}
	roomColl := client.Database(DBNAME).Collection("rooms")
	err = roomColl.FindOne(ctx, filter).Decode(&room)
	if err != nil {
		return err
	}

	user := User{}
	filter = bson.M{"username": username}
	userColl := client.Database(DBNAME).Collection("users")
	err = userColl.FindOne(ctx, filter).Decode(&user)
	if err != nil {
		return err
	}

	user.Rooms = append(user.Rooms, room.ID)
	updater := bson.M{"$set": bson.M{"rooms": user.Rooms}}
	_, err = userColl.UpdateByID(ctx, user.ID, updater)
	if err != nil {
		return err
	}

	room.Usernames = append(room.Usernames, user.Username)
	updater = bson.M{"$set": bson.M{"usernames": room.Usernames}}
	_, err = roomColl.UpdateByID(ctx, room.ID, updater)
	if err != nil {
		return err
	}

	return nil
}

func GetP2PRoom(user1 string, user2 string) (Room, error) {
	label1 := user1 + "|" + user2
	label2 := user2 + "|" + user1
	room := Room{}

	client, err := getMongoClient()
	if err != nil {
		return room, err
	}

	filter := bson.M{"$or": []bson.M{
		{"label": label1},
		{"label": label2},
	}}
	roomColl := client.Database(DBNAME).Collection("rooms")
	err = roomColl.FindOne(ctx, filter).Decode(&room)
	return room, err
}

func FetchChat(room string) ([]Chat, error) {
	chats := []Chat{}
	client, err := getMongoClient()
	if err != nil {
		return chats, err
	}

	roomID, _ := primitive.ObjectIDFromHex(room)
	filter := bson.M{"room_id": roomID}
	findOpt := options.Find().SetLimit(20).SetSort(bson.M{"created_at": -1})
	chatColl := client.Database(DBNAME).Collection("chats")
	cursor, err := chatColl.Find(ctx, filter, findOpt)
	if err != nil {
		return chats, err
	}

	err = cursor.All(ctx, &chats)
	return chats, err
}

func FetchGroup() ([]Room, error) {
	rooms := []Room{}
	client, err := getMongoClient()
	if err != nil {
		return rooms, err
	}

	filter := bson.M{"type": "group"}
	findOpt := options.Find().SetSort(bson.M{"updated_at": -1})
	roomColl := client.Database(DBNAME).Collection("rooms")
	cursor, err := roomColl.Find(ctx, filter, findOpt)
	if err != nil {
		return rooms, err
	}

	cursor.All(ctx, &rooms)
	return rooms, nil
}

func FetchUser() ([]User, error) {
	users := []User{}
	client, err := getMongoClient()
	if err != nil {
		return users, err
	}

	findOpt := options.Find().SetSort(bson.M{"last_active": -1})
	userColl := client.Database(DBNAME).Collection("users")
	cursor, err := userColl.Find(ctx, bson.M{}, findOpt)
	if err != nil {
		return users, err
	}

	cursor.All(ctx, &users)
	return users, nil
}
