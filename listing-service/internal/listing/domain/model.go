package domain

import "time"

type ListingStatus string

const (
	StatusActive ListingStatus = "active"
	StatusSold   ListingStatus = "sold"
)

type Listing struct {
	ID          string        `bson:"_id,omitempty"`
	Title       string        `bson:"title"`
	Description string        `bson:"description"`
	Price       float64       `bson:"price"`
	Status      ListingStatus `bson:"status"`
	Photos      []string      `bson:"photos"` // URLs to photos in MinIO/S3
	CreatedAt   time.Time     `bson:"created_at"`
	UpdatedAt   time.Time     `bson:"updated_at"`
}

type Photo struct {
	ID  string `bson:"_id,omitempty"`
	URL string `bson:"url"`
}

type Favorite struct {
	ID        string    `bson:"_id,omitempty"`
	UserID    string    `bson:"user_id"`
	ListingID string    `bson:"listing_id"`
	CreatedAt time.Time `bson:"created_at"`
}