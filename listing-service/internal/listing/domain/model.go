package domain

import "time" // Оставим time, т.к. это стандартная библиотека

type ListingStatus string

const (
	StatusActive   ListingStatus = "active"
	StatusSold     ListingStatus = "sold"
	StatusReserved ListingStatus = "reserved" // Добавил из предыдущих обсуждений
	StatusInactive ListingStatus = "inactive" // Добавил из предыдущих обсуждений
)

type Listing struct {
	ID          string // ID обычно генерируется БД или usecase'ом перед сохранением
	UserID      string // <--- ВАЖНО: Добавь это поле, если его еще нет
	CategoryID  string // <--- ВАЖНО: Добавь это поле, если его еще нет
	Title       string
	Description string
	Price       float64
	Status      ListingStatus
	Photos      []string // URLs to photos
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// Photo как доменная сущность может быть не нужна, если это просто URL в Listing.
// Если Photo имеет свою логику или атрибуты, тогда оставляем.
// Пока предполагаем, что это просто строка URL в Listing.Photos.
/*
type Photo struct {
	ID  string
	URL string
}
*/

type Favorite struct {
	ID        string // Может быть опциональным, если композитный ключ UserID+ListingID уникален
	UserID    string
	ListingID string
	CreatedAt time.Time
}

// Filter для поиска, как и раньше
type Filter struct {
	Query      string
	MinPrice   float64
	MaxPrice   float64
	Status     ListingStatus
	CategoryID string
	UserID     string // Для поиска объявлений конкретного пользователя
	Page       int32
	Limit      int32
	SortBy     string
	SortOrder  string
}

// Ошибки доменного уровня, которые могут быть возвращены usecase'ами
// var (
//  ErrListingNotFound = errors.New("listing not found") // Переместим в usecase
//  ErrForbidden       = errors.New("action forbidden") // Переместим в usecase
// )