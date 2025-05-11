package router

import (
	"github.com/Abdurahmanit/GroupProject/api-gateway/internal/handler"
	"github.com/Abdurahmanit/GroupProject/api-gateway/internal/middleware"
	"github.com/go-chi/chi/v5"
)

// SetupUserRoutes configures all user and admin-user related routes.
func SetupUserRoutes(r *chi.Mux, userHandler *handler.UserHandler, jwtSecret string) {
	// Public user routes
	r.Post("/api/user/register", userHandler.Register)
	r.Post("/api/user/login", userHandler.Login)
	// Email verification confirmation route removed.

	// Protected user routes (require JWT authentication)
	r.Group(func(authRouter chi.Router) { // Renamed 'r' to 'authRouter' for clarity within the group
		authRouter.Use(middleware.JWTAuth(jwtSecret))

		authRouter.Post("/api/user/logout", userHandler.Logout)
		authRouter.Get("/api/user/profile", userHandler.GetProfile)
		authRouter.Put("/api/user/profile", userHandler.UpdateProfile)
		authRouter.Post("/api/user/change-password", userHandler.ChangePassword)

		// Email verification initiation and resend routes removed.

		authRouter.Delete("/api/user/delete", userHandler.DeleteUser)       // User hard-deletes their own account
		authRouter.Post("/api/user/deactivate", userHandler.DeactivateUser) // User deactivates their own account (soft delete)

		// Admin routes related to users
		// These also fall under JWTAuth. The handlers/usecases perform further role checks.
		authRouter.Post("/api/admin/user/delete", userHandler.AdminDeleteUser) // Admin hard-deletes a user
		authRouter.Post("/api/admin/users/list", userHandler.AdminListUsers)
		authRouter.Post("/api/admin/users/search", userHandler.AdminSearchUsers)
		authRouter.Post("/api/admin/user/update-role", userHandler.AdminUpdateUserRole)
		authRouter.Post("/api/admin/user/set-active", userHandler.AdminSetUserActiveStatus) // Admin activates/deactivates a user
	})
}
