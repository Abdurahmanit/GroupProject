package router

import (
	// "net/http" // Не нужен для методов chi
	"github.com/Abdurahmanit/GroupProject/api-gateway/internal/handler"
	"github.com/Abdurahmanit/GroupProject/api-gateway/internal/middleware"
	"github.com/go-chi/chi/v5" // Импортируем chi
)

// SetupListingRoutes добавляет маршруты для Listing Service, используя chi.Router
func SetupListingRoutes(mux *chi.Mux, h *handler.ListingHandler, jwtSecret string) {
	// Группа маршрутов, требующих аутентификации
	mux.Group(func(r chi.Router) {
		// Применяем JWTAuth middleware ко всем маршрутам внутри этой группы
		r.Use(middleware.JWTAuth(jwtSecret))

		r.Post("/api/favorites", h.HandleAddFavorite)
		r.Delete("/api/favorites", h.HandleRemoveFavorite)
		r.Get("/api/favorites", h.HandleGetFavorites)

		// Если другие операции с listings тоже требуют авторизации, добавьте их сюда
		// Например:
		// r.Post("/api/listings", h.HandleCreateListing) // Если создание требует авторизации
		// r.Put("/api/listings/{id}", h.HandleUpdateListing) // Если обновление требует авторизации
	})

	// Публичные маршруты (не требуют авторизации)
	// Если какие-то из этих маршрутов должны быть публичными, оставьте их вне группы.
	// Если ВСЕ /api/listings/* требуют авторизации, то можно было бы сделать:
	// mux.Route("/api/listings", func(r chi.Router) {
	// r.Use(middleware.JWTAuth(jwtSecret)) // Для всех /api/listings/*
	// ...
	// })
	// mux.Route("/api/favorites", func(r chi.Router) {
	// r.Use(middleware.JWTAuth(jwtSecret)) // Для всех /api/favorites/*
	// ...
	// })

	// Пока что оставляем большинство маршрутов listings публичными, как было в вашем примере.
	// Если вы хотите, чтобы, например, создание и обновление объявлений тоже требовали JWT,
	// перенесите их в группу выше или создайте новую группу с JWTAuth.

	mux.Post("/api/listings", h.HandleCreateListing) // Если этот должен быть публичным
	mux.Put("/api/listings/{id}", h.HandleUpdateListing)     // Если этот должен быть публичным
	mux.Delete("/api/listings/{id}", h.HandleDeleteListing) // Если этот должен быть публичным
	mux.Get("/api/listings/{id}", h.HandleGetListingByID)
	mux.Get("/api/listings/search", h.HandleSearchListings) // Скорее всего GET, а не POST, если для поиска
	mux.Post("/api/listings/{id}/photos", h.HandleUploadPhoto)
	mux.Get("/api/listings/{id}/status", h.HandleGetListingStatus)
	mux.Get("/api/listings/{id}/photos", h.HandleGetPhotoURLs)
	mux.Patch("/api/listings/{id}/status", h.HandleUpdateListingStatus) // Patch для частичного обновления
}