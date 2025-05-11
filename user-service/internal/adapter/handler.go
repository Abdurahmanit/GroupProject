// File: user-service/internal/adapter/handler.go
package adapter

import (
	"context"
	"errors"
	"time"

	"github.com/Abdurahmanit/GroupProject/user-service/internal/repository" // For error types like repository.ErrUserNotFound
	"github.com/Abdurahmanit/GroupProject/user-service/internal/usecase"
	user "github.com/Abdurahmanit/GroupProject/user-service/proto" // Path to your generated proto package
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type UserHandler struct {
	user.UnimplementedUserServiceServer
	usecase *usecase.UserUsecase // This holds the usecase instance
	logger  *zap.Logger
}

func NewUserHandler(ucase *usecase.UserUsecase, logger *zap.Logger) *UserHandler {
	return &UserHandler{
		usecase: ucase,
		logger:  logger,
	}
}

// Register handles user registration.
func (h *UserHandler) Register(ctx context.Context, req *user.RegisterRequest) (*user.RegisterResponse, error) {
	if req.GetUsername() == "" || req.GetEmail() == "" || req.GetPassword() == "" {
		return nil, status.Error(codes.InvalidArgument, "Username, email, and password are required")
	}
	userIDHex, err := h.usecase.Register(ctx, req.Username, req.Email, req.Password)
	if err != nil {
		h.logger.Error("Failed to register user", zap.Error(err))
		if errors.Is(err, repository.ErrDuplicateEmail) {
			return nil, status.Error(codes.AlreadyExists, "Email already exists")
		}
		return nil, status.Error(codes.Internal, "Failed to register user")
	}
	return &user.RegisterResponse{UserId: userIDHex}, nil
}

// Login handles user login.
func (h *UserHandler) Login(ctx context.Context, req *user.LoginRequest) (*user.LoginResponse, error) {
	if req.GetEmail() == "" || req.GetPassword() == "" {
		return nil, status.Error(codes.InvalidArgument, "Email and password are required")
	}
	token, err := h.usecase.Login(ctx, req.Email, req.Password)
	if err != nil {
		h.logger.Warn("Failed to login user", zap.String("email", req.Email), zap.Error(err))
		if errors.Is(err, usecase.ErrInvalidCredentials) || errors.Is(err, usecase.ErrUserInactive) {
			return nil, status.Error(codes.Unauthenticated, err.Error())
		}
		return nil, status.Error(codes.Internal, "Login failed")
	}
	return &user.LoginResponse{Token: token}, nil
}

// Logout handles user logout.
func (h *UserHandler) Logout(ctx context.Context, req *user.LogoutRequest) (*user.LogoutResponse, error) {
	if req.GetUserId() == "" {
		return nil, status.Error(codes.InvalidArgument, "User ID is required")
	}
	if err := h.usecase.Logout(ctx, req.UserId); err != nil {
		h.logger.Error("Failed to logout user", zap.String("userID", req.UserId), zap.Error(err))
		return nil, status.Error(codes.Internal, "Logout failed")
	}
	return &user.LogoutResponse{Success: true}, nil
}

// GetProfile retrieves a user's profile.
func (h *UserHandler) GetProfile(ctx context.Context, req *user.GetProfileRequest) (*user.GetProfileResponse, error) {
	if req.GetUserId() == "" {
		return nil, status.Error(codes.InvalidArgument, "User ID is required")
	}
	profile, err := h.usecase.GetProfile(ctx, req.UserId)
	if err != nil {
		h.logger.Error("Failed to get profile", zap.String("userID", req.UserId), zap.Error(err))
		if errors.Is(err, repository.ErrUserNotFound) {
			return nil, status.Error(codes.NotFound, "User profile not found")
		}
		if errors.Is(err, usecase.ErrUserInactive) {
			return nil, status.Error(codes.PermissionDenied, usecase.ErrUserInactive.Error())
		}
		return nil, status.Error(codes.Internal, "Failed to get profile")
	}
	// Ensure is_email_verified is not accessed as it's removed from proto
	return &user.GetProfileResponse{
		UserId:    profile.ID.Hex(),
		Username:  profile.Username,
		Email:     profile.Email,
		Role:      profile.Role,
		IsActive:  profile.IsActive,
		CreatedAt: profile.CreatedAt.Format(time.RFC3339),
		UpdatedAt: profile.UpdatedAt.Format(time.RFC3339),
	}, nil
}

// UpdateProfile updates a user's profile.
func (h *UserHandler) UpdateProfile(ctx context.Context, req *user.UpdateProfileRequest) (*user.UpdateProfileResponse, error) {
	if req.GetUserId() == "" {
		return nil, status.Error(codes.InvalidArgument, "User ID is required")
	}
	// Add more validation for username and email if needed
	err := h.usecase.UpdateProfile(ctx, req.UserId, req.Username, req.Email)
	if err != nil {
		h.logger.Error("Failed to update profile", zap.String("userID", req.UserId), zap.Error(err))
		if errors.Is(err, repository.ErrUserNotFound) {
			return nil, status.Error(codes.NotFound, "User not found for update")
		}
		if errors.Is(err, usecase.ErrUserInactive) {
			return nil, status.Error(codes.FailedPrecondition, usecase.ErrUserInactive.Error())
		}
		if errors.Is(err, repository.ErrDuplicateEmail) {
			return nil, status.Error(codes.AlreadyExists, "Email already in use")
		}
		return nil, status.Error(codes.Internal, "Failed to update profile")
	}
	return &user.UpdateProfileResponse{Success: true}, nil
}

// ChangePassword changes a user's password.
func (h *UserHandler) ChangePassword(ctx context.Context, req *user.ChangePasswordRequest) (*user.ChangePasswordResponse, error) {
	if req.GetUserId() == "" || req.GetOldPassword() == "" || req.GetNewPassword() == "" {
		return nil, status.Error(codes.InvalidArgument, "User ID, old password, and new password are required")
	}
	err := h.usecase.ChangePassword(ctx, req.UserId, req.OldPassword, req.NewPassword)
	if err != nil {
		h.logger.Error("Failed to change password", zap.String("userID", req.UserId), zap.Error(err))
		if errors.Is(err, usecase.ErrInvalidCredentials) {
			return nil, status.Error(codes.Unauthenticated, "Invalid old password")
		}
		if errors.Is(err, repository.ErrUserNotFound) {
			return nil, status.Error(codes.NotFound, "User not found")
		}
		if errors.Is(err, usecase.ErrUserInactive) {
			return nil, status.Error(codes.FailedPrecondition, usecase.ErrUserInactive.Error())
		}
		return nil, status.Error(codes.Internal, "Failed to change password")
	}
	return &user.ChangePasswordResponse{Success: true}, nil
}

// DeleteUser (Hard Delete by user).
func (h *UserHandler) DeleteUser(ctx context.Context, req *user.DeleteUserRequest) (*user.DeleteUserResponse, error) {
	if req.GetUserId() == "" {
		return nil, status.Error(codes.InvalidArgument, "User ID is required")
	}
	err := h.usecase.DeleteUser(ctx, req.UserId) // This is now hard delete
	if err != nil {
		h.logger.Error("Failed to delete user (hard)", zap.String("userID", req.UserId), zap.Error(err))
		if errors.Is(err, repository.ErrUserNotFound) {
			return nil, status.Error(codes.NotFound, "User not found for deletion")
		}
		return nil, status.Error(codes.Internal, "Failed to delete user")
	}
	return &user.DeleteUserResponse{Success: true}, nil
}

// DeactivateUser (Soft Delete by user).
func (h *UserHandler) DeactivateUser(ctx context.Context, req *user.DeactivateUserRequest) (*user.DeactivateUserResponse, error) {
	if req.GetUserId() == "" {
		return nil, status.Error(codes.InvalidArgument, "User ID is required")
	}
	err := h.usecase.DeactivateUser(ctx, req.UserId)
	if err != nil {
		h.logger.Error("Failed to deactivate user", zap.String("userID", req.UserId), zap.Error(err))
		if errors.Is(err, repository.ErrUserNotFound) {
			return nil, status.Error(codes.NotFound, "User not found for deactivation")
		}
		return nil, status.Error(codes.Internal, "Failed to deactivate user")
	}
	return &user.DeactivateUserResponse{Success: true}, nil
}

// --- Admin Handlers ---

// AdminDeleteUser (Hard Delete by admin).
func (h *UserHandler) AdminDeleteUser(ctx context.Context, req *user.AdminDeleteUserRequest) (*user.AdminDeleteUserResponse, error) {
	if req.GetAdminId() == "" || req.GetUserIdToDelete() == "" {
		return nil, status.Error(codes.InvalidArgument, "Admin ID and User ID to delete are required")
	}
	err := h.usecase.AdminDeleteUser(ctx, req.AdminId, req.UserIdToDelete)
	if err != nil {
		h.logger.Error("Failed to admin delete user (hard)", zap.String("adminID", req.AdminId), zap.String("userID", req.UserIdToDelete), zap.Error(err))
		if errors.Is(err, usecase.ErrUnauthorized) {
			return nil, status.Error(codes.PermissionDenied, "Admin unauthorized")
		}
		if errors.Is(err, repository.ErrUserNotFound) {
			return nil, status.Error(codes.NotFound, "User to delete not found")
		}
		return nil, status.Error(codes.Internal, "Failed to admin delete user")
	}
	return &user.AdminDeleteUserResponse{Success: true}, nil
}

// AdminListUsers lists users for admin.
func (h *UserHandler) AdminListUsers(ctx context.Context, req *user.AdminListUsersRequest) (*user.AdminListUsersResponse, error) {
	if req.GetAdminId() == "" {
		return nil, status.Error(codes.InvalidArgument, "Admin ID is required")
	}
	usersList, err := h.usecase.AdminListUsers(ctx, req.AdminId, req.Skip, req.Limit)
	if err != nil {
		h.logger.Error("Failed to admin list users", zap.String("adminID", req.AdminId), zap.Error(err))
		if errors.Is(err, usecase.ErrUnauthorized) {
			return nil, status.Error(codes.PermissionDenied, "Admin unauthorized")
		}
		return nil, status.Error(codes.Internal, "Failed to list users")
	}

	protoUsers := make([]*user.User, len(usersList))
	for i, u := range usersList {
		protoUsers[i] = &user.User{ // Ensure this mapping matches the proto User message
			UserId:    u.ID.Hex(),
			Username:  u.Username,
			Email:     u.Email,
			Role:      u.Role,
			IsActive:  u.IsActive,
			CreatedAt: u.CreatedAt.Format(time.RFC3339),
			UpdatedAt: u.UpdatedAt.Format(time.RFC3339),
		}
	}
	return &user.AdminListUsersResponse{Users: protoUsers}, nil
}

// AdminSearchUsers searches users for admin.
func (h *UserHandler) AdminSearchUsers(ctx context.Context, req *user.AdminSearchUsersRequest) (*user.AdminSearchUsersResponse, error) {
	if req.GetAdminId() == "" {
		return nil, status.Error(codes.InvalidArgument, "Admin ID is required")
	}
	usersList, err := h.usecase.AdminSearchUsers(ctx, req.AdminId, req.Query, req.Skip, req.Limit)
	if err != nil {
		h.logger.Error("Failed to admin search users", zap.String("adminID", req.AdminId), zap.String("query", req.Query), zap.Error(err))
		if errors.Is(err, usecase.ErrUnauthorized) {
			return nil, status.Error(codes.PermissionDenied, "Admin unauthorized")
		}
		return nil, status.Error(codes.Internal, "Failed to search users")
	}
	protoUsers := make([]*user.User, len(usersList))
	for i, u := range usersList {
		protoUsers[i] = &user.User{ // Ensure this mapping matches the proto User message
			UserId:    u.ID.Hex(),
			Username:  u.Username,
			Email:     u.Email,
			Role:      u.Role,
			IsActive:  u.IsActive,
			CreatedAt: u.CreatedAt.Format(time.RFC3339),
			UpdatedAt: u.UpdatedAt.Format(time.RFC3339),
		}
	}
	return &user.AdminSearchUsersResponse{Users: protoUsers}, nil
}

// AdminUpdateUserRole updates a user's role for admin.
func (h *UserHandler) AdminUpdateUserRole(ctx context.Context, req *user.AdminUpdateUserRoleRequest) (*user.AdminUpdateUserRoleResponse, error) {
	if req.GetAdminId() == "" || req.GetUserIdToUpdate() == "" || req.GetRole() == "" {
		return nil, status.Error(codes.InvalidArgument, "Admin ID, User ID to update, and Role are required")
	}
	err := h.usecase.AdminUpdateUserRole(ctx, req.AdminId, req.UserIdToUpdate, req.Role)
	if err != nil {
		h.logger.Error("Failed to admin update user role", zap.String("adminID", req.AdminId), zap.String("userID", req.UserIdToUpdate), zap.Error(err))
		if errors.Is(err, usecase.ErrUnauthorized) {
			return nil, status.Error(codes.PermissionDenied, "Admin unauthorized")
		}
		if errors.Is(err, repository.ErrUserNotFound) {
			return nil, status.Error(codes.NotFound, "User to update not found")
		}
		return nil, status.Error(codes.Internal, "Failed to update user role")
	}
	return &user.AdminUpdateUserRoleResponse{Success: true}, nil
}

// AdminSetUserActiveStatus allows an admin to activate or deactivate a user.
// This method now correctly calls the public method on the usecase.
func (h *UserHandler) AdminSetUserActiveStatus(ctx context.Context, req *user.AdminSetUserActiveStatusRequest) (*user.AdminSetUserActiveStatusResponse, error) {
	if req.GetAdminId() == "" || req.GetUserId() == "" {
		return nil, status.Error(codes.InvalidArgument, "Admin ID and User ID are required")
	}

	// Call the public method on the usecase instance
	err := h.usecase.AdminSetUserActiveStatus(ctx, req.AdminId, req.UserId, req.IsActive)
	if err != nil {
		h.logger.Error("Failed to set user active status by admin via usecase",
			zap.String("adminID", req.AdminId),
			zap.String("targetUserID", req.UserId),
			zap.Bool("isActive", req.IsActive),
			zap.Error(err))

		if errors.Is(err, usecase.ErrUnauthorized) {
			return nil, status.Error(codes.PermissionDenied, "Admin unauthorized")
		}
		if errors.Is(err, repository.ErrUserNotFound) { // Check for specific error from repository if bubbled up
			return nil, status.Error(codes.NotFound, "Target user not found")
		}
		// Handle other specific errors from usecase if necessary, or a general internal error
		return nil, status.Error(codes.Internal, "Failed to update user active status")
	}

	return &user.AdminSetUserActiveStatusResponse{Success: true}, nil
}
