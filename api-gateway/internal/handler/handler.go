package handler

import (
	"context"
	"encoding/json"
	"net/http"

	user "github.com/Abdurahmanit/GroupProject/user-service/proto"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type UserHandler struct {
	userClient user.UserServiceClient
	logger     *zap.Logger
}

func NewUserHandler(conn *grpc.ClientConn, logger *zap.Logger) *UserHandler {
	return &UserHandler{
		userClient: user.NewUserServiceClient(conn),
		logger:     logger,
	}
}

func (h *UserHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req user.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("Failed to decode request", zap.Error(err))
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	resp, err := h.userClient.Register(context.Background(), &req)
	if err != nil {
		h.logger.Error("Failed to register user", zap.Error(err))
		http.Error(w, "Failed to register user", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *UserHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req user.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("Failed to decode request", zap.Error(err))
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	resp, err := h.userClient.Login(context.Background(), &req)
	if err != nil {
		h.logger.Error("Failed to login user", zap.Error(err))
		http.Error(w, "Failed to login user", http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *UserHandler) Logout(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(string)
	req := &user.LogoutRequest{UserId: userID}

	resp, err := h.userClient.Logout(context.Background(), req)
	if err != nil {
		h.logger.Error("Failed to logout user", zap.Error(err))
		http.Error(w, "Failed to logout user", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *UserHandler) GetProfile(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(string)
	req := &user.GetProfileRequest{UserId: userID}

	resp, err := h.userClient.GetProfile(context.Background(), req)
	if err != nil {
		h.logger.Error("Failed to get profile", zap.Error(err))
		http.Error(w, "Failed to get profile", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
