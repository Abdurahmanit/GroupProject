package entity

import "time"

type User struct {
	ID              string
	Username        string
	Email           string
	Password        string
	Role            string //"admin", "customer"
	IsEmailVerified bool
	IsActive        bool // For soft deletion
	CreatedAt       time.Time
	UpdatedAt       time.Time
}
