package adapter

import (
	"context"

	"github.com/Abdurahmanit/GroupProject/user-service/internal/usecase"
	user "github.com/Abdurahmanit/GroupProject/user-service/proto"
	"go.uber.org/zap"
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
		return nil, err
	}

	return &user.RegisterResponse{UserId: userID}, nil
}

func (h *UserHandler) Login(ctx context.Context, req *user.LoginRequest) (*user.LoginResponse, error) {
	token, err := h.usecase.Login(ctx, req.Email, req.Password)
	if err != nil {
		h.logger.Error("Failed to login user", zap.Error(err))
		return nil, err
	}

	return &user.LoginResponse{Token: token}, nil
}

func (h *UserHandler) Logout(ctx context.Context, req *user.LogoutRequest) (*user.LogoutResponse, error) {
	if err := h.usecase.Logout(ctx, req.UserId); err != nil {
		h.logger.Error("Failed to logout user", zap.Error(err))
		return nil, err
	}

	return &user.LogoutResponse{Success: true}, nil
}

func (h *UserHandler) GetProfile(ctx context.Context, req *user.GetProfileRequest) (*user.GetProfileResponse, error) {
	profile, err := h.usecase.GetProfile(ctx, req.UserId)
	if err != nil {
		h.logger.Error("Failed to get profile", zap.Error(err))
		return nil, err
	}

	return &user.GetProfileResponse{
		Username: profile.Username,
		Email:    profile.Email,
	}, nil
}
