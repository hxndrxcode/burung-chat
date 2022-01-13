package database

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Room struct {
	ID        primitive.ObjectID `bson:"_id"`
	UpdatedAt time.Time          `bson:"updated_at"`
	Label     string             `bson:"label"`
	Type      string             `bson:"type"`
	Usernames []string           `bson:"usernames"`
	Active    int                `bson:"active"`
}

type User struct {
	ID         primitive.ObjectID   `bson:"_id"`
	Username   string               `bson:"username"`
	Rooms      []primitive.ObjectID `bson:"rooms"`
	LastActive time.Time            `bson:"last_active"`
}

type Chat struct {
	ID        primitive.ObjectID `bson:"_id"`
	RoomID    primitive.ObjectID `bson:"room_id"`
	Message   string             `bson:"message"`
	CreatedAt time.Time          `bson:"created_at"`
	From      string             `bson:"from"`
	Type      string             `bson:"type"`
}
