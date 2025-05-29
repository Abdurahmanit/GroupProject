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
	usecase *usecase.UserUsecase
	logger  *zap.Logger
}

func NewUserHandler(ucase *usecase.UserUsecase, logger *zap.Logger) *UserHandler {
	return &UserHandler{
		usecase: ucase,
		logger:  logger.Named("UserGRPCHandler"),
	}
}

func (h *UserHandler) Register(ctx context.Context, req *user.RegisterRequest) (*user.RegisterResponse, error) {
	h.logger.Info("gRPC Register request received", zap.String("email", req.GetEmail()), zap.String("phoneNumber", req.GetPhoneNumber()))
	if req.GetUsername() == "" || req.GetEmail() == "" || req.GetPassword() == "" || req.GetPhoneNumber() == "" {
		h.logger.Warn("InvalidArgument for Register gRPC request: missing fields")
		return nil, status.Error(codes.InvalidArgument, "Username, email, password, and phone number are required")
	}

	userIDHex, err := h.usecase.Register(ctx, req.Username, req.Email, req.Password, req.PhoneNumber)
	if err != nil {
		h.logger.Error("Usecase failed to register user", zap.String("email", req.Email), zap.Error(err))
		switch {
		case errors.Is(err, usecase.ErrDuplicateEmail):
			return nil, status.Error(codes.AlreadyExists, "Email already exists")
		case errors.Is(err, usecase.ErrDuplicatePhoneNumber):
			return nil, status.Error(codes.AlreadyExists, "Phone number already exists")
		case errors.Is(err, usecase.ErrInvalidPhoneNumber):
			return nil, status.Error(codes.InvalidArgument, usecase.ErrInvalidPhoneNumber.Error())
		case errors.Is(err, usecase.ErrPhoneNumberRequired):
			return nil, status.Error(codes.InvalidArgument, usecase.ErrPhoneNumberRequired.Error())
		default:
			return nil, status.Error(codes.Internal, "Failed to register user")
		}
	}
	h.logger.Info("gRPC Register request processed successfully", zap.String("userID", userIDHex))
	return &user.RegisterResponse{UserId: userIDHex}, nil
}

func (h *UserHandler) Login(ctx context.Context, req *user.LoginRequest) (*user.LoginResponse, error) {
	h.logger.Info("gRPC Login request received", zap.String("email", req.GetEmail()))
	if req.GetEmail() == "" || req.GetPassword() == "" {
		h.logger.Warn("InvalidArgument for Login gRPC request: missing fields")
		return nil, status.Error(codes.InvalidArgument, "Email and password are required")
	}
	token, err := h.usecase.Login(ctx, req.Email, req.Password)
	if err != nil {
		h.logger.Warn("Usecase failed to login user", zap.String("email", req.Email), zap.Error(err))
		if errors.Is(err, usecase.ErrInvalidCredentials) || errors.Is(err, usecase.ErrUserInactive) {
			return nil, status.Error(codes.Unauthenticated, err.Error())
		}
		// Could also check for email not verified error if login depends on it
		// if errors.Is(err, errors.New("email not verified")) {
		// 	return nil, status.Error(codes.FailedPrecondition, "Email not verified")
		// }
		return nil, status.Error(codes.Internal, "Login failed")
	}
	h.logger.Info("gRPC Login request processed successfully", zap.String("email", req.GetEmail()))
	return &user.LoginResponse{Token: token}, nil
}

func (h *UserHandler) Logout(ctx context.Context, req *user.LogoutRequest) (*user.LogoutResponse, error) {
	h.logger.Info("gRPC Logout request received", zap.String("userID", req.GetUserId()))
	if req.GetUserId() == "" {
		return nil, status.Error(codes.InvalidArgument, "User ID is required")
	}
	if err := h.usecase.Logout(ctx, req.UserId); err != nil {
		h.logger.Error("Usecase failed to logout user", zap.String("userID", req.UserId), zap.Error(err))
		return nil, status.Error(codes.Internal, "Logout failed")
	}
	h.logger.Info("gRPC Logout request processed successfully", zap.String("userID", req.GetUserId()))
	return &user.LogoutResponse{Success: true}, nil
}

func (h *UserHandler) GetProfile(ctx context.Context, req *user.GetProfileRequest) (*user.GetProfileResponse, error) {
	h.logger.Info("gRPC GetProfile request received", zap.String("userID", req.GetUserId()))
	if req.GetUserId() == "" {
		h.logger.Warn("InvalidArgument for GetProfile gRPC request: User ID is required")
		return nil, status.Error(codes.InvalidArgument, "User ID is required")
	}
	profile, err := h.usecase.GetProfile(ctx, req.UserId)
	if err != nil {
		h.logger.Error("Usecase failed to get profile", zap.String("userID", req.UserId), zap.Error(err))
		if errors.Is(err, repository.ErrUserNotFound) || errors.Is(err, usecase.ErrUserNotFound) {
			return nil, status.Error(codes.NotFound, "User profile not found")
		}
		return nil, status.Error(codes.Internal, "Failed to get profile")
	}

	emailVerifiedAtStr := ""
	if profile.EmailVerifiedAt != nil {
		emailVerifiedAtStr = profile.EmailVerifiedAt.Format(time.RFC3339)
	}

	h.logger.Info("gRPC GetProfile request processed successfully", zap.String("userID", profile.ID.Hex()))
	return &user.GetProfileResponse{
		UserId:          profile.ID.Hex(),
		Username:        profile.Username,
		Email:           profile.Email,
		PhoneNumber:     profile.PhoneNumber,
		Role:            profile.Role,
		IsActive:        profile.IsActive,
		CreatedAt:       profile.CreatedAt.Format(time.RFC3339),
		UpdatedAt:       profile.UpdatedAt.Format(time.RFC3339),
		IsEmailVerified: profile.IsEmailVerified,
		EmailVerifiedAt: emailVerifiedAtStr,
	}, nil
}

func (h *UserHandler) UpdateProfile(ctx context.Context, req *user.UpdateProfileRequest) (*user.UpdateProfileResponse, error) {
	h.logger.Info("gRPC UpdateProfile request received", zap.String("userID", req.GetUserId()))
	if req.GetUserId() == "" {
		h.logger.Warn("InvalidArgument for UpdateProfile gRPC request: User ID is required")
		return nil, status.Error(codes.InvalidArgument, "User ID is required")
	}

	err := h.usecase.UpdateProfile(ctx, req.UserId, req.Username, req.Email, req.PhoneNumber)
	if err != nil {
		h.logger.Error("Usecase failed to update profile", zap.String("userID", req.UserId), zap.Error(err))
		switch {
		case errors.Is(err, repository.ErrUserNotFound) || errors.Is(err, usecase.ErrUserNotFound):
			return nil, status.Error(codes.NotFound, "User not found for update")
		case errors.Is(err, usecase.ErrUserInactive):
			return nil, status.Error(codes.FailedPrecondition, usecase.ErrUserInactive.Error())
		case errors.Is(err, usecase.ErrDuplicateEmail) || errors.Is(err, repository.ErrDuplicateEmail):
			return nil, status.Error(codes.AlreadyExists, "Email already in use")
		case errors.Is(err, usecase.ErrDuplicatePhoneNumber) || errors.Is(err, repository.ErrDuplicatePhoneNumber):
			return nil, status.Error(codes.AlreadyExists, "Phone number already in use")
		case errors.Is(err, usecase.ErrInvalidPhoneNumber):
			return nil, status.Error(codes.InvalidArgument, usecase.ErrInvalidPhoneNumber.Error())
		default:
			return nil, status.Error(codes.Internal, "Failed to update profile")
		}
	}
	h.logger.Info("gRPC UpdateProfile request processed successfully", zap.String("userID", req.GetUserId()))
	return &user.UpdateProfileResponse{Success: true}, nil
}

func (h *UserHandler) ChangePassword(ctx context.Context, req *user.ChangePasswordRequest) (*user.ChangePasswordResponse, error) {
	h.logger.Info("gRPC ChangePassword request received", zap.String("userID", req.GetUserId()))
	if req.GetUserId() == "" || req.GetOldPassword() == "" || req.GetNewPassword() == "" {
		return nil, status.Error(codes.InvalidArgument, "User ID, old password, and new password are required")
	}
	err := h.usecase.ChangePassword(ctx, req.UserId, req.OldPassword, req.NewPassword)
	if err != nil {
		h.logger.Error("Usecase failed to change password", zap.String("userID", req.UserId), zap.Error(err))
		switch {
		case errors.Is(err, usecase.ErrInvalidCredentials):
			return nil, status.Error(codes.Unauthenticated, "Invalid old password")
		case errors.Is(err, repository.ErrUserNotFound) || errors.Is(err, usecase.ErrUserNotFound):
			return nil, status.Error(codes.NotFound, "User not found")
		case errors.Is(err, usecase.ErrUserInactive):
			return nil, status.Error(codes.FailedPrecondition, usecase.ErrUserInactive.Error())
		default:
			return nil, status.Error(codes.Internal, "Failed to change password")
		}
	}
	h.logger.Info("gRPC ChangePassword request processed successfully", zap.String("userID", req.GetUserId()))
	return &user.ChangePasswordResponse{Success: true}, nil
}

func (h *UserHandler) DeleteUser(ctx context.Context, req *user.DeleteUserRequest) (*user.DeleteUserResponse, error) {
	h.logger.Info("gRPC DeleteUser request received", zap.String("userID", req.GetUserId()))
	if req.GetUserId() == "" {
		return nil, status.Error(codes.InvalidArgument, "User ID is required")
	}
	err := h.usecase.DeleteUser(ctx, req.UserId)
	if err != nil {
		h.logger.Error("Usecase failed to delete user (hard)", zap.String("userID", req.UserId), zap.Error(err))
		if errors.Is(err, repository.ErrUserNotFound) || errors.Is(err, usecase.ErrUserNotFound) {
			return nil, status.Error(codes.NotFound, "User not found for deletion")
		}
		return nil, status.Error(codes.Internal, "Failed to delete user")
	}
	h.logger.Info("gRPC DeleteUser request processed successfully", zap.String("userID", req.GetUserId()))
	return &user.DeleteUserResponse{Success: true}, nil
}

func (h *UserHandler) DeactivateUser(ctx context.Context, req *user.DeactivateUserRequest) (*user.DeactivateUserResponse, error) {
	h.logger.Info("gRPC DeactivateUser request received", zap.String("userID", req.GetUserId()))
	if req.GetUserId() == "" {
		return nil, status.Error(codes.InvalidArgument, "User ID is required")
	}
	err := h.usecase.DeactivateUser(ctx, req.UserId)
	if err != nil {
		h.logger.Error("Usecase failed to deactivate user", zap.String("userID", req.UserId), zap.Error(err))
		if errors.Is(err, repository.ErrUserNotFound) || errors.Is(err, usecase.ErrUserNotFound) {
			return nil, status.Error(codes.NotFound, "User not found for deactivation")
		}
		return nil, status.Error(codes.Internal, "Failed to deactivate user")
	}
	h.logger.Info("gRPC DeactivateUser request processed successfully", zap.String("userID", req.GetUserId()))
	return &user.DeactivateUserResponse{Success: true}, nil
}

// Email Verification Handlers
func (h *UserHandler) RequestEmailVerification(ctx context.Context, req *user.RequestEmailVerificationRequest) (*user.RequestEmailVerificationResponse, error) {
	h.logger.Info("gRPC RequestEmailVerification request received", zap.String("userID", req.GetUserId()))
	if req.GetUserId() == "" {
		return nil, status.Error(codes.InvalidArgument, "User ID is required")
	}

	err := h.usecase.RequestEmailVerification(ctx, req.UserId)
	if err != nil {
		h.logger.Error("Usecase failed to request email verification", zap.String("userID", req.UserId), zap.Error(err))
		switch {
		case errors.Is(err, usecase.ErrEmailAlreadyVerified):
			return &user.RequestEmailVerificationResponse{Success: false, Message: err.Error()}, nil // Not an error, but a specific state
		case errors.Is(err, repository.ErrUserNotFound) || errors.Is(err, usecase.ErrUserNotFound):
			return nil, status.Error(codes.NotFound, "User not found")
		case errors.Is(err, usecase.ErrMailerFailed):
			return nil, status.Error(codes.Internal, "Failed to send verification email, please try again later.")
		default:
			return nil, status.Error(codes.Internal, "Failed to request email verification")
		}
	}
	h.logger.Info("gRPC RequestEmailVerification processed successfully", zap.String("userID", req.GetUserId()))
	return &user.RequestEmailVerificationResponse{Success: true, Message: "Verification email sent. Please check your inbox."}, nil
}

func (h *UserHandler) VerifyEmail(ctx context.Context, req *user.VerifyEmailRequest) (*user.VerifyEmailResponse, error) {
	h.logger.Info("gRPC VerifyEmail request received", zap.String("userID", req.GetUserId()))
	if req.GetUserId() == "" || req.GetCode() == "" {
		return nil, status.Error(codes.InvalidArgument, "User ID and verification code are required")
	}

	err := h.usecase.VerifyEmail(ctx, req.UserId, req.Code)
	if err != nil {
		h.logger.Error("Usecase failed to verify email", zap.String("userID", req.UserId), zap.Error(err))
		switch {
		case errors.Is(err, usecase.ErrEmailAlreadyVerified):
			return &user.VerifyEmailResponse{Success: false, Message: err.Error()}, nil // Not an error, specific state
		case errors.Is(err, usecase.ErrInvalidVerificationCode):
			return &user.VerifyEmailResponse{Success: false, Message: err.Error()}, nil // Not an error, specific state for client
		case errors.Is(err, repository.ErrUserNotFound) || errors.Is(err, usecase.ErrUserNotFound):
			return nil, status.Error(codes.NotFound, "User not found")
		default:
			return nil, status.Error(codes.Internal, "Failed to verify email")
		}
	}
	h.logger.Info("gRPC VerifyEmail processed successfully", zap.String("userID", req.GetUserId()))
	return &user.VerifyEmailResponse{Success: true, Message: "Email verified successfully."}, nil
}

func (h *UserHandler) CheckEmailVerificationStatus(ctx context.Context, req *user.CheckEmailVerificationStatusRequest) (*user.CheckEmailVerificationStatusResponse, error) {
	h.logger.Info("gRPC CheckEmailVerificationStatus request received", zap.String("userID", req.GetUserId()))
	if req.GetUserId() == "" {
		return nil, status.Error(codes.InvalidArgument, "User ID is required")
	}
	isVerified, err := h.usecase.CheckEmailVerificationStatus(ctx, req.UserId)
	if err != nil {
		h.logger.Error("Usecase failed to check email verification status", zap.String("userID", req.GetUserId()), zap.Error(err))
		if errors.Is(err, repository.ErrUserNotFound) || errors.Is(err, usecase.ErrUserNotFound) {
			return nil, status.Error(codes.NotFound, "User not found")
		}
		return nil, status.Error(codes.Internal, "Failed to check email verification status")
	}
	h.logger.Info("gRPC CheckEmailVerificationStatus processed successfully", zap.String("userID", req.GetUserId()), zap.Bool("isVerified", isVerified))
	return &user.CheckEmailVerificationStatusResponse{IsVerified: isVerified}, nil
}

// --- Admin Handlers ---
func (h *UserHandler) AdminDeleteUser(ctx context.Context, req *user.AdminDeleteUserRequest) (*user.AdminDeleteUserResponse, error) {
	h.logger.Info("gRPC AdminDeleteUser request", zap.String("adminID", req.GetAdminId()), zap.String("targetUserID", req.GetUserIdToDelete()))
	if req.GetAdminId() == "" || req.GetUserIdToDelete() == "" {
		return nil, status.Error(codes.InvalidArgument, "Admin ID and User ID to delete are required")
	}
	err := h.usecase.AdminDeleteUser(ctx, req.AdminId, req.UserIdToDelete)
	if err != nil {
		h.logger.Error("Usecase failed for AdminDeleteUser", zap.Error(err))
		if errors.Is(err, usecase.ErrUnauthorized) {
			return nil, status.Error(codes.PermissionDenied, "Admin unauthorized")
		}
		if errors.Is(err, repository.ErrUserNotFound) || errors.Is(err, usecase.ErrUserNotFound) {
			return nil, status.Error(codes.NotFound, "User to delete not found")
		}
		return nil, status.Error(codes.Internal, "Failed to admin delete user")
	}
	return &user.AdminDeleteUserResponse{Success: true}, nil
}

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
		emailVerifiedAtStr := ""
		if u.EmailVerifiedAt != nil {
			emailVerifiedAtStr = u.EmailVerifiedAt.Format(time.RFC3339)
		}
		protoUsers[i] = &user.User{
			UserId:          u.ID.Hex(),
			Username:        u.Username,
			Email:           u.Email,
			PhoneNumber:     u.PhoneNumber,
			Role:            u.Role,
			IsActive:        u.IsActive,
			CreatedAt:       u.CreatedAt.Format(time.RFC3339),
			UpdatedAt:       u.UpdatedAt.Format(time.RFC3339),
			IsEmailVerified: u.IsEmailVerified,
			EmailVerifiedAt: emailVerifiedAtStr,
		}
	}
	h.logger.Info("gRPC AdminListUsers processed successfully", zap.String("adminID", req.AdminId), zap.Int("count", len(protoUsers)))
	return &user.AdminListUsersResponse{Users: protoUsers}, nil
}

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
		emailVerifiedAtStr := ""
		if u.EmailVerifiedAt != nil {
			emailVerifiedAtStr = u.EmailVerifiedAt.Format(time.RFC3339)
		}
		protoUsers[i] = &user.User{
			UserId:          u.ID.Hex(),
			Username:        u.Username,
			Email:           u.Email,
			PhoneNumber:     u.PhoneNumber,
			Role:            u.Role,
			IsActive:        u.IsActive,
			CreatedAt:       u.CreatedAt.Format(time.RFC3339),
			UpdatedAt:       u.UpdatedAt.Format(time.RFC3339),
			IsEmailVerified: u.IsEmailVerified,
			EmailVerifiedAt: emailVerifiedAtStr,
		}
	}
	h.logger.Info("gRPC AdminSearchUsers processed successfully", zap.String("adminID", req.AdminId), zap.Int("count", len(protoUsers)))
	return &user.AdminSearchUsersResponse{Users: protoUsers}, nil
}

func (h *UserHandler) AdminUpdateUserRole(ctx context.Context, req *user.AdminUpdateUserRoleRequest) (*user.AdminUpdateUserRoleResponse, error) {
	h.logger.Info("gRPC AdminUpdateUserRole request", zap.String("adminID", req.GetAdminId()), zap.String("targetUserID", req.GetUserIdToUpdate()), zap.String("newRole", req.GetRole()))
	if req.GetAdminId() == "" || req.GetUserIdToUpdate() == "" || req.GetRole() == "" {
		return nil, status.Error(codes.InvalidArgument, "Admin ID, User ID to update, and Role are required")
	}
	err := h.usecase.AdminUpdateUserRole(ctx, req.AdminId, req.UserIdToUpdate, req.Role)
	if err != nil {
		h.logger.Error("Usecase failed for AdminUpdateUserRole", zap.Error(err))
		if errors.Is(err, usecase.ErrUnauthorized) {
			return nil, status.Error(codes.PermissionDenied, "Admin unauthorized")
		}
		if errors.Is(err, repository.ErrUserNotFound) || errors.Is(err, usecase.ErrUserNotFound) {
			return nil, status.Error(codes.NotFound, "User to update not found")
		}
		return nil, status.Error(codes.Internal, "Failed to update user role")
	}
	return &user.AdminUpdateUserRoleResponse{Success: true}, nil
}

func (h *UserHandler) AdminSetUserActiveStatus(ctx context.Context, req *user.AdminSetUserActiveStatusRequest) (*user.AdminSetUserActiveStatusResponse, error) {
	h.logger.Info("gRPC AdminSetUserActiveStatus request", zap.String("adminID", req.GetAdminId()), zap.String("targetUserID", req.GetUserId()), zap.Bool("isActive", req.GetIsActive()))
	if req.GetAdminId() == "" || req.GetUserId() == "" {
		return nil, status.Error(codes.InvalidArgument, "Admin ID and User ID are required")
	}
	err := h.usecase.AdminSetUserActiveStatus(ctx, req.AdminId, req.UserId, req.IsActive)
	if err != nil {
		h.logger.Error("Usecase failed for AdminSetUserActiveStatus", zap.Error(err))
		if errors.Is(err, usecase.ErrUnauthorized) {
			return nil, status.Error(codes.PermissionDenied, "Admin unauthorized")
		}
		if errors.Is(err, repository.ErrUserNotFound) || errors.Is(err, usecase.ErrUserNotFound) {
			return nil, status.Error(codes.NotFound, "Target user not found")
		}
		return nil, status.Error(codes.Internal, "Failed to update user active status")
	}
	return &user.AdminSetUserActiveStatusResponse{Success: true}, nil
}
