package entity

import "time"

type Comment struct {
	ID        string
	NewsID    string
	UserID    string
	Content   string
	CreatedAt time.Time
	UpdatedAt time.Time
}
