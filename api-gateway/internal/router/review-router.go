package router

import (
	"github.com/Abdurahmanit/GroupProject/api-gateway/internal/handler"
	"github.com/Abdurahmanit/GroupProject/api-gateway/internal/middleware"
	"github.com/go-chi/chi/v5"
)

// SetupReviewRoutes configures routes for the Review service.
func SetupReviewRoutes(mux *chi.Mux, h *handler.ReviewHandler, jwtSecret string) {
	// Public routes for reviews (mostly read operations)
	mux.Get("/api/reviews/{reviewId}", h.HandleGetReview)
	mux.Get("/api/products/{productId}/reviews", h.HandleListReviewsByProduct) // Example: list reviews for a product
	mux.Get("/api/products/{productId}/reviews/rating", h.HandleGetProductAverageRating)

	// Protected routes for reviews (require JWT authentication)
	mux.Group(func(r chi.Router) {
		r.Use(middleware.JWTAuth(jwtSecret)) // Apply JWT authentication

		r.Post("/api/reviews", h.HandleCreateReview)
		r.Put("/api/reviews/{reviewId}", h.HandleUpdateReview)
		r.Delete("/api/reviews/{reviewId}", h.HandleDeleteReview)
		r.Get("/api/reviews/my", h.HandleListReviewsByUser)

		r.Patch("/api/admin/reviews/{reviewId}/moderate", h.HandleModerateReview)
	})
}
