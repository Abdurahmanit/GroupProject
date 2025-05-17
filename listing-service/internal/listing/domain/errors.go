package domain

import "errors"

var (
	ErrListingNotFound     = errors.New("listing not found")
	ErrFavoriteNotFound    = errors.New("favorite not found")
	ErrInvalidListingData  = errors.New("invalid listing data")
	ErrInvalidFilter       = errors.New("invalid filter parameters")
	ErrDuplicateFavorite   = errors.New("favorite already exists")
)