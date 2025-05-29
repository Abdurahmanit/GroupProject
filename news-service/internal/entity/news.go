package entity

import "time"

type News struct {
	ID        string
	Title     string
	Content   string
	AuthorID  string
	ImageURL  string `json:"image_url,omitempty"`
	CreatedAt time.Time
	UpdatedAt time.Time
}
