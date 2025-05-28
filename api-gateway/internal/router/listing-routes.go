package router

import (
	// "net/http" // Не нужен для методов chi
	"github.com/Abdurahmanit/GroupProject/api-gateway/internal/handler"
	"github.com/Abdurahmanit/GroupProject/api-gateway/internal/middleware"
	"github.com/go-chi/chi/v5" // Импортируем chi
)

func SetupListingRoutes(mux *chi.Mux, h *handler.ListingHandler, jwtSecret string) {
	// Группа маршрутов для ИЗБРАННОГО, требующих аутентификации
	mux.Group(func(r chi.Router) {
		r.Use(middleware.JWTAuth(jwtSecret)) // Применяем JWTAuth middleware

		r.Post("/api/favorites", h.HandleAddFavorite)
		r.Delete("/api/favorites", h.HandleRemoveFavorite) // Убедись, что есть способ указать ID, например, в теле запроса
		r.Get("/api/favorites", h.HandleGetFavorites)
	})

	// Группа маршрутов для ОБЪЯВЛЕНИЙ ("/api/listings")
	mux.Route("/api/listings", func(r chi.Router) {
		// Публичные маршруты для объявлений (не требуют авторизации)
		r.Get("/{id}", h.HandleGetListingByID)           // GET /api/listings/{id}
		r.Get("/search", h.HandleSearchListings)        // GET /api/listings/search
		r.Get("/{id}/photos", h.HandleGetPhotoURLs)     // GET /api/listings/{id}/photos
		r.Get("/{id}/status", h.HandleGetListingStatus) // GET /api/listings/{id}/status

		// Маршруты для объявлений, ТРЕБУЮЩИЕ аутентификации
		r.Group(func(authR chi.Router) {
			authR.Use(middleware.JWTAuth(jwtSecret)) // Применяем JWTAuth middleware

			// Обрати внимание, что пути здесь относительны к "/api/listings"
			authR.Post("/", h.HandleCreateListing)                  // POST /api/listings
			authR.Put("/{id}", h.HandleUpdateListing)               // PUT /api/listings/{id}
			authR.Delete("/{id}", h.HandleDeleteListing)           // DELETE /api/listings/{id}
			authR.Post("/{id}/photos", h.HandleUploadPhoto)         // POST /api/listings/{id}/photos
			authR.Patch("/{id}/status", h.HandleUpdateListingStatus) // PATCH /api/listings/{id}/status
		})
	})
}