package entity

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type User struct {
	ID        primitive.ObjectID
	Username  string
	Email     string
	Password  string
	Role      string // "admin", "customer"
	IsActive  bool
	CreatedAt time.Time
	UpdatedAt time.Time
}
