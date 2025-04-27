package adapter

import (
	"context"
	"time"

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

func NewUserHandler(usecase *usecase.UserUsecase, logger *zap.Logger) *UserHandler {
	return &UserHandler{
		usecase: usecase,
		logger:  logger,
	}
}

func (h *UserHandler) Register(ctx context.Context, req *user.RegisterRequest) (*user.RegisterResponse, error) {
	userID, err := h.usecase.Register(ctx, req.Username, req.Email, req.Password)
	if err != nil {
		h.logger.Error("Failed to register user", zap.Error(err))
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &user.RegisterResponse{UserId: userID}, nil
}

func (h *UserHandler) Login(ctx context.Context, req *user.LoginRequest) (*user.LoginResponse, error) {
	token, err := h.usecase.Login(ctx, req.Email, req.Password)
	if err != nil {
		h.logger.Error("Failed to login user", zap.Error(err))
		if err == usecase.ErrInvalidCredentials {
			return nil, status.Error(codes.Unauthenticated, err.Error())
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &user.LoginResponse{Token: token}, nil
}

func (h *UserHandler) Logout(ctx context.Context, req *user.LogoutRequest) (*user.LogoutResponse, error) {
	if err := h.usecase.Logout(ctx, req.UserId); err != nil {
		h.logger.Error("Failed to logout user", zap.Error(err))
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &user.LogoutResponse{Success: true}, nil
}

func (h *UserHandler) GetProfile(ctx context.Context, req *user.GetProfileRequest) (*user.GetProfileResponse, error) {
	profile, err := h.usecase.GetProfile(ctx, req.UserId)
	if err != nil {
		h.logger.Error("Failed to get profile", zap.Error(err))
		if err == usecase.ErrUnauthorized {
			return nil, status.Error(codes.Unauthenticated, err.Error())
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &user.GetProfileResponse{
		Username:        profile.Username,
		Email:           profile.Email,
		Role:            profile.Role,
		IsEmailVerified: profile.IsEmailVerified,
		IsActive:        profile.IsActive,
		CreatedAt:       profile.CreatedAt.Format(time.RFC3339),
		UpdatedAt:       profile.UpdatedAt.Format(time.RFC3339),
	}, nil
}

func (h *UserHandler) UpdateProfile(ctx context.Context, req *user.UpdateProfileRequest) (*user.UpdateProfileResponse, error) {
	if err := h.usecase.UpdateProfile(ctx, req.UserId, req.Username, req.Email); err != nil {
		h.logger.Error("Failed to update profile", zap.Error(err))
		if err == usecase.ErrUnauthorized {
			return nil, status.Error(codes.Unauthenticated, err.Error())
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &user.UpdateProfileResponse{Success: true}, nil
}

func (h *UserHandler) ChangePassword(ctx context.Context, req *user.ChangePasswordRequest) (*user.ChangePasswordResponse, error) {
	if err := h.usecase.ChangePassword(ctx, req.UserId, req.OldPassword, req.NewPassword); err != nil {
		h.logger.Error("Failed to change password", zap.Error(err))
		if err == usecase.ErrInvalidCredentials {
			return nil, status.Error(codes.Unauthenticated, err.Error())
		}
		if err == usecase.ErrUnauthorized {
			return nil, status.Error(codes.Unauthenticated, err.Error())
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &user.ChangePasswordResponse{Success: true}, nil
}

func (h *UserHandler) VerifyEmail(ctx context.Context, req *user.VerifyEmailRequest) (*user.VerifyEmailResponse, error) {
	if err := h.usecase.VerifyEmail(ctx, req.UserId); err != nil {
		h.logger.Error("Failed to verify email", zap.Error(err))
		if err == usecase.ErrUnauthorized {
			return nil, status.Error(codes.Unauthenticated, err.Error())
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &user.VerifyEmailResponse{Success: true}, nil
}

func (h *UserHandler) DeleteUser(ctx context.Context, req *user.DeleteUserRequest) (*user.DeleteUserResponse, error) {
	if err := h.usecase.DeleteUser(ctx, req.UserId); err != nil {
		h.logger.Error("Failed to delete user", zap.Error(err))
		if err == usecase.ErrUnauthorized {
			return nil, status.Error(codes.Unauthenticated, err.Error())
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &user.DeleteUserResponse{Success: true}, nil
}

func (h *UserHandler) AdminDeleteUser(ctx context.Context, req *user.AdminDeleteUserRequest) (*user.AdminDeleteUserResponse, error) {
	if err := h.usecase.AdminDeleteUser(ctx, req.AdminId, req.UserId); err != nil {
		h.logger.Error("Failed to admin delete user", zap.Error(err))
		if err == usecase.ErrUnauthorized {
			return nil, status.Error(codes.Unauthenticated, err.Error())
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &user.AdminDeleteUserResponse{Success: true}, nil
}

func (h *UserHandler) AdminListUsers(ctx context.Context, req *user.AdminListUsersRequest) (*user.AdminListUsersResponse, error) {
	users, err := h.usecase.AdminListUsers(ctx, req.AdminId, req.Skip, req.Limit)
	if err != nil {
		h.logger.Error("Failed to list users", zap.Error(err))
		if err == usecase.ErrUnauthorized {
			return nil, status.Error(codes.Unauthenticated, err.Error())
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	protoUsers := make([]*user.User, len(users))
	for i, u := range users {
		protoUsers[i] = &user.User{
			UserId:          u.ID,
			Username:        u.Username,
			Email:           u.Email,
			Role:            u.Role,
			IsEmailVerified: u.IsEmailVerified,
			IsActive:        u.IsActive,
			CreatedAt:       u.CreatedAt.Format(time.RFC3339),
			UpdatedAt:       u.UpdatedAt.Format(time.RFC3339),
		}
	}

	return &user.AdminListUsersResponse{Users: protoUsers}, nil
}

func (h *UserHandler) AdminSearchUsers(ctx context.Context, req *user.AdminSearchUsersRequest) (*user.AdminSearchUsersResponse, error) {
	users, err := h.usecase.AdminSearchUsers(ctx, req.AdminId, req.Query, req.Skip, req.Limit)
	if err != nil {
		h.logger.Error("Failed to search users", zap.Error(err))
		if err == usecase.ErrUnauthorized {
			return nil, status.Error(codes.Unauthenticated, err.Error())
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	protoUsers := make([]*user.User, len(users))
	for i, u := range users {
		protoUsers[i] = &user.User{
			UserId:          u.ID,
			Username:        u.Username,
			Email:           u.Email,
			Role:            u.Role,
			IsEmailVerified: u.IsEmailVerified,
			IsActive:        u.IsActive,
			CreatedAt:       u.CreatedAt.Format(time.RFC3339),
			UpdatedAt:       u.UpdatedAt.Format(time.RFC3339),
		}
	}

	return &user.AdminSearchUsersResponse{Users: protoUsers}, nil
}

func (h *UserHandler) AdminUpdateUserRole(ctx context.Context, req *user.AdminUpdateUserRoleRequest) (*user.AdminUpdateUserRoleResponse, error) {
	if err := h.usecase.AdminUpdateUserRole(ctx, req.AdminId, req.UserId, req.Role); err != nil {
		h.logger.Error("Failed to update user role", zap.Error(err))
		if err == usecase.ErrUnauthorized {
			return nil, status.Error(codes.Unauthenticated, err.Error())
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &user.AdminUpdateUserRoleResponse{Success: true}, nil
}
