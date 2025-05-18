package adapter

import (
	"context"
	"errors"
	"time"

	"github.com/Abdurahmanit/GroupProject/user-service/internal/repository"
	"github.com/Abdurahmanit/GroupProject/user-service/internal/usecase"
	user "github.com/Abdurahmanit/GroupProject/user-service/proto"
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
	h.logger.Info("gRPC Register request received", zap.String("email", req.GetEmail()), zap.String("phoneNumber", req.GetPhoneNumber()))
	if req.GetUsername() == "" || req.GetEmail() == "" || req.GetPassword() == "" || req.GetPhoneNumber() == "" {
		h.logger.Warn("InvalidArgument for Register gRPC request",
			zap.Bool("missingUsername", req.GetUsername() == ""),
			zap.Bool("missingEmail", req.GetEmail() == ""),
			zap.Bool("missingPassword", req.GetPassword() == ""),
			zap.Bool("missingPhoneNumber", req.GetPhoneNumber() == ""))
		return nil, status.Error(codes.InvalidArgument, "Username, email, password, and phone number are required")
	}

	userIDHex, err := h.usecase.Register(ctx, req.Username, req.Email, req.Password, req.PhoneNumber)
	if err != nil {
		h.logger.Error("Usecase failed to register user", zap.String("email", req.Email), zap.Error(err))
		if errors.Is(err, usecase.ErrDuplicateEmail) { // Changed to usecase level error
			return nil, status.Error(codes.AlreadyExists, "Email already exists")
		}
		if errors.Is(err, usecase.ErrDuplicatePhoneNumber) { // Changed to usecase level error
			return nil, status.Error(codes.AlreadyExists, "Phone number already exists")
		}
		if errors.Is(err, usecase.ErrInvalidPhoneNumber) {
			return nil, status.Error(codes.InvalidArgument, usecase.ErrInvalidPhoneNumber.Error())
		}
		if errors.Is(err, usecase.ErrPhoneNumberRequired) {
			return nil, status.Error(codes.InvalidArgument, usecase.ErrPhoneNumberRequired.Error())
		}
		return nil, status.Error(codes.Internal, "Failed to register user")
	}
	h.logger.Info("gRPC Register request processed successfully", zap.String("userID", userIDHex))
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
	h.logger.Info("gRPC GetProfile request received", zap.String("userID", req.GetUserId()))
	if req.GetUserId() == "" {
		h.logger.Warn("InvalidArgument for GetProfile gRPC request: User ID is required")
		return nil, status.Error(codes.InvalidArgument, "User ID is required")
	}
	profile, err := h.usecase.GetProfile(ctx, req.UserId)
	if err != nil {
		h.logger.Error("Usecase failed to get profile", zap.String("userID", req.UserId), zap.Error(err))
		if errors.Is(err, repository.ErrUserNotFound) { // Usecase propagates this
			return nil, status.Error(codes.NotFound, "User profile not found")
		}
		// Note: ErrUserInactive is not typically checked in GetProfile itself,
		// as GetProfile should return the current state. Auth middleware handles active status for protected routes.
		return nil, status.Error(codes.Internal, "Failed to get profile")
	}
	h.logger.Info("gRPC GetProfile request processed successfully", zap.String("userID", profile.ID.Hex()))
	return &user.GetProfileResponse{
		UserId:      profile.ID.Hex(),
		Username:    profile.Username,
		Email:       profile.Email,
		PhoneNumber: profile.PhoneNumber, // Added PhoneNumber mapping
		Role:        profile.Role,
		IsActive:    profile.IsActive,
		CreatedAt:   profile.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   profile.UpdatedAt.Format(time.RFC3339),
	}, nil
}

// UpdateProfile updates a user's profile.
func (h *UserHandler) UpdateProfile(ctx context.Context, req *user.UpdateProfileRequest) (*user.UpdateProfileResponse, error) {
	h.logger.Info("gRPC UpdateProfile request received" /* ... */)
	if req.GetUserId() == "" {
		h.logger.Warn("InvalidArgument for UpdateProfile gRPC request: User ID is required")
		return nil, status.Error(codes.InvalidArgument, "User ID is required")
	}

	err := h.usecase.UpdateProfile(ctx, req.UserId, req.Username, req.Email, req.PhoneNumber)
	if err != nil {
		h.logger.Error("Usecase failed to update profile", zap.String("userID", req.UserId), zap.Error(err))
		if errors.Is(err, repository.ErrUserNotFound) {
			return nil, status.Error(codes.NotFound, "User not found for update")
		}
		if errors.Is(err, usecase.ErrUserInactive) {
			return nil, status.Error(codes.FailedPrecondition, usecase.ErrUserInactive.Error())
		}
		if errors.Is(err, repository.ErrDuplicateEmail) || errors.Is(err, usecase.ErrDuplicateEmail) {
			return nil, status.Error(codes.AlreadyExists, "Email already in use")
		}
		if errors.Is(err, usecase.ErrDuplicatePhoneNumber) || errors.Is(err, repository.ErrDuplicatePhoneNumber) {
			return nil, status.Error(codes.AlreadyExists, "Phone number already in use")
		}
		if errors.Is(err, usecase.ErrInvalidPhoneNumber) {
			return nil, status.Error(codes.InvalidArgument, usecase.ErrInvalidPhoneNumber.Error())
		}
		return nil, status.Error(codes.Internal, "Failed to update profile")
	}
	h.logger.Info("gRPC UpdateProfile request processed successfully", zap.String("userID", req.GetUserId()))
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
	err := h.usecase.DeleteUser(ctx, req.UserId)
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
	h.logger.Info("gRPC AdminListUsers request received", zap.String("adminID", req.GetAdminId()))
	if req.GetAdminId() == "" {
		h.logger.Warn("InvalidArgument for AdminListUsers: Admin ID is required")
		return nil, status.Error(codes.InvalidArgument, "Admin ID is required")
	}
	usersList, err := h.usecase.AdminListUsers(ctx, req.AdminId, req.Skip, req.Limit)
	if err != nil {
		h.logger.Error("Usecase failed for AdminListUsers", zap.String("adminID", req.AdminId), zap.Error(err))
		if errors.Is(err, usecase.ErrUnauthorized) {
			return nil, status.Error(codes.PermissionDenied, "Admin unauthorized")
		}
		return nil, status.Error(codes.Internal, "Failed to list users")
	}

	protoUsers := make([]*user.User, len(usersList))
	for i, u := range usersList {
		protoUsers[i] = &user.User{ // This is the user.User message from proto
			UserId:      u.ID.Hex(),
			Username:    u.Username,
			Email:       u.Email,
			PhoneNumber: u.PhoneNumber, // Added PhoneNumber mapping
			Role:        u.Role,
			IsActive:    u.IsActive,
			CreatedAt:   u.CreatedAt.Format(time.RFC3339),
			UpdatedAt:   u.UpdatedAt.Format(time.RFC3339),
		}
	}
	h.logger.Info("gRPC AdminListUsers processed successfully", zap.String("adminID", req.AdminId), zap.Int("count", len(protoUsers)))
	return &user.AdminListUsersResponse{Users: protoUsers}, nil
}

// AdminSearchUsers searches users for admin.
func (h *UserHandler) AdminSearchUsers(ctx context.Context, req *user.AdminSearchUsersRequest) (*user.AdminSearchUsersResponse, error) {
	h.logger.Info("gRPC AdminSearchUsers request received", zap.String("adminID", req.GetAdminId()), zap.String("query", req.GetQuery()))
	if req.GetAdminId() == "" {
		h.logger.Warn("InvalidArgument for AdminSearchUsers: Admin ID is required")
		return nil, status.Error(codes.InvalidArgument, "Admin ID is required")
	}
	usersList, err := h.usecase.AdminSearchUsers(ctx, req.AdminId, req.Query, req.Skip, req.Limit)
	if err != nil {
		h.logger.Error("Usecase failed for AdminSearchUsers", zap.String("adminID", req.AdminId), zap.String("query", req.Query), zap.Error(err))
		if errors.Is(err, usecase.ErrUnauthorized) {
			return nil, status.Error(codes.PermissionDenied, "Admin unauthorized")
		}
		return nil, status.Error(codes.Internal, "Failed to search users")
	}
	protoUsers := make([]*user.User, len(usersList))
	for i, u := range usersList {
		protoUsers[i] = &user.User{ // This is the user.User message from proto
			UserId:      u.ID.Hex(),
			Username:    u.Username,
			Email:       u.Email,
			PhoneNumber: u.PhoneNumber, // Added PhoneNumber mapping
			Role:        u.Role,
			IsActive:    u.IsActive,
			CreatedAt:   u.CreatedAt.Format(time.RFC3339),
			UpdatedAt:   u.UpdatedAt.Format(time.RFC3339),
		}
	}
	h.logger.Info("gRPC AdminSearchUsers processed successfully", zap.String("adminID", req.AdminId), zap.Int("count", len(protoUsers)))
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
func (h *UserHandler) AdminSetUserActiveStatus(ctx context.Context, req *user.AdminSetUserActiveStatusRequest) (*user.AdminSetUserActiveStatusResponse, error) {
	if req.GetAdminId() == "" || req.GetUserId() == "" {
		return nil, status.Error(codes.InvalidArgument, "Admin ID and User ID are required")
	}

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
		return nil, status.Error(codes.Internal, "Failed to update user active status")
	}

	return &user.AdminSetUserActiveStatusResponse{Success: true}, nil
}
