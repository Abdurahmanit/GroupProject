package entity

import "time"

type News struct {
	ID        string
	Title     string
	Content   string
	AuthorID  string
	ImageURL  string
	Category  string
	CreatedAt time.Time
	UpdatedAt time.Time
}
