package entity

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// User entity without email verification fields.
type User struct {
	ID        primitive.ObjectID // ObjectID, no bson tag here
	Username  string
	Email     string
	Password  string // This will be the hashed password
	Role      string // "admin", "customer"
	IsActive  bool   // For soft deletion
	CreatedAt time.Time
	UpdatedAt time.Time
}
