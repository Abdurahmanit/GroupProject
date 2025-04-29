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

func (h *UserHandler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(string)
	var req user.UpdateProfileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("Failed to decode request", zap.Error(err))
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	req.UserId = userID

	resp, err := h.userClient.UpdateProfile(context.Background(), &req)
	if err != nil {
		h.logger.Error("Failed to update profile", zap.Error(err))
		http.Error(w, "Failed to update profile", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *UserHandler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(string)
	var req user.ChangePasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("Failed to decode request", zap.Error(err))
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	req.UserId = userID

	resp, err := h.userClient.ChangePassword(context.Background(), &req)
	if err != nil {
		h.logger.Error("Failed to change password", zap.Error(err))
		http.Error(w, "Failed to change password", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *UserHandler) VerifyEmail(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(string)
	req := &user.VerifyEmailRequest{UserId: userID}

	resp, err := h.userClient.VerifyEmail(context.Background(), req)
	if err != nil {
		h.logger.Error("Failed to verify email", zap.Error(err))
		http.Error(w, "Failed to verify email", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *UserHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(string)
	req := &user.DeleteUserRequest{UserId: userID}

	resp, err := h.userClient.DeleteUser(context.Background(), req)
	if err != nil {
		h.logger.Error("Failed to delete user", zap.Error(err))
		http.Error(w, "Failed to delete user", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *UserHandler) AdminDeleteUser(w http.ResponseWriter, r *http.Request) {
	adminID := r.Context().Value("user_id").(string)
	var req user.AdminDeleteUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("Failed to decode request", zap.Error(err))
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	req.AdminId = adminID

	resp, err := h.userClient.AdminDeleteUser(context.Background(), &req)
	if err != nil {
		h.logger.Error("Failed to admin delete user", zap.Error(err))
		http.Error(w, "Failed to admin delete user", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *UserHandler) AdminListUsers(w http.ResponseWriter, r *http.Request) {
	adminID := r.Context().Value("user_id").(string)
	var req user.AdminListUsersRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("Failed to decode request", zap.Error(err))
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	req.AdminId = adminID

	resp, err := h.userClient.AdminListUsers(context.Background(), &req)
	if err != nil {
		h.logger.Error("Failed to list users", zap.Error(err))
		http.Error(w, "Failed to list users", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *UserHandler) AdminSearchUsers(w http.ResponseWriter, r *http.Request) {
	adminID := r.Context().Value("user_id").(string)
	var req user.AdminSearchUsersRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("Failed to decode request", zap.Error(err))
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	req.AdminId = adminID

	resp, err := h.userClient.AdminSearchUsers(context.Background(), &req)
	if err != nil {
		h.logger.Error("Failed to search users", zap.Error(err))
		http.Error(w, "Failed to search users", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *UserHandler) AdminUpdateUserRole(w http.ResponseWriter, r *http.Request) {
	adminID := r.Context().Value("user_id").(string)
	var req user.AdminUpdateUserRoleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("Failed to decode request", zap.Error(err))
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	req.AdminId = adminID

	resp, err := h.userClient.AdminUpdateUserRole(context.Background(), &req)
	if err != nil {
		h.logger.Error("Failed to update user role", zap.Error(err))
		http.Error(w, "Failed to update user role", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
