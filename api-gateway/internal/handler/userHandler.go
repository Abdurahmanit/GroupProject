package handler

import (
	"context"
	"encoding/json"
	"net/http"

	user "github.com/Abdurahmanit/GroupProject/user-service/proto"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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

// Register handles user registration requests.
func (h *UserHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req user.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("Failed to decode request for Register", zap.Error(err))
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	resp, err := h.userClient.Register(context.Background(), &req)
	if err != nil {
		h.logger.Error("Failed to register user via gRPC", zap.Error(err))
		s, _ := status.FromError(err) // Even if not ok, s will be non-nil with codes.Unknown
		http.Error(w, s.Message(), GRPCCodeToHTTPStatus(s.Code()))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)
}

// Login handles user login requests.
func (h *UserHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req user.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("Failed to decode request for Login", zap.Error(err))
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	resp, err := h.userClient.Login(context.Background(), &req)
	if err != nil {
		h.logger.Error("Failed to login user via gRPC", zap.Error(err))
		s, _ := status.FromError(err)
		http.Error(w, s.Message(), GRPCCodeToHTTPStatus(s.Code()))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// Logout handles user logout requests.
func (h *UserHandler) Logout(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("user_id").(string)
	if !ok || userID == "" {
		h.logger.Warn("User ID not found in token for Logout")
		http.Error(w, "User ID not found in token", http.StatusUnauthorized)
		return
	}
	req := &user.LogoutRequest{UserId: userID}
	resp, err := h.userClient.Logout(context.Background(), req)
	if err != nil {
		h.logger.Error("Failed to logout user via gRPC", zap.String("userID", userID), zap.Error(err))
		s, _ := status.FromError(err)
		http.Error(w, s.Message(), GRPCCodeToHTTPStatus(s.Code()))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// GetProfile handles requests to get user profile.
func (h *UserHandler) GetProfile(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("user_id").(string)
	if !ok || userID == "" {
		h.logger.Warn("User ID not found in token for GetProfile")
		http.Error(w, "User ID not found in token", http.StatusUnauthorized)
		return
	}
	req := &user.GetProfileRequest{UserId: userID}
	resp, err := h.userClient.GetProfile(context.Background(), req)
	if err != nil {
		h.logger.Error("Failed to get profile via gRPC", zap.String("userID", userID), zap.Error(err))
		s, _ := status.FromError(err)
		http.Error(w, s.Message(), GRPCCodeToHTTPStatus(s.Code()))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// UpdateProfile handles requests to update user profile.
func (h *UserHandler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("user_id").(string)
	if !ok || userID == "" {
		h.logger.Warn("User ID not found in token for UpdateProfile")
		http.Error(w, "User ID not found in token", http.StatusUnauthorized)
		return
	}
	var reqBody user.UpdateProfileRequest
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	reqBody.UserId = userID

	resp, err := h.userClient.UpdateProfile(context.Background(), &reqBody)
	if err != nil {
		h.logger.Error("Failed to update profile via gRPC", zap.String("userID", userID), zap.Error(err))
		s, _ := status.FromError(err)
		http.Error(w, s.Message(), GRPCCodeToHTTPStatus(s.Code()))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// ChangePassword handles requests to change user password.
func (h *UserHandler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("user_id").(string)
	if !ok || userID == "" {
		h.logger.Warn("User ID not found in token for ChangePassword")
		http.Error(w, "User ID not found in token", http.StatusUnauthorized)
		return
	}
	var reqBody user.ChangePasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	reqBody.UserId = userID

	resp, err := h.userClient.ChangePassword(context.Background(), &reqBody)
	if err != nil {
		h.logger.Error("Failed to change password via gRPC", zap.String("userID", userID), zap.Error(err))
		s, _ := status.FromError(err)
		http.Error(w, s.Message(), GRPCCodeToHTTPStatus(s.Code()))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// DeleteUser handles requests for a user to (hard) delete their own account.
func (h *UserHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("user_id").(string)
	if !ok || userID == "" {
		h.logger.Warn("User ID not found in token for DeleteUser")
		http.Error(w, "User ID not found in token", http.StatusUnauthorized)
		return
	}
	req := &user.DeleteUserRequest{UserId: userID}
	resp, err := h.userClient.DeleteUser(context.Background(), req)
	if err != nil {
		h.logger.Error("Failed to delete user (hard) via gRPC", zap.String("userID", userID), zap.Error(err))
		s, _ := status.FromError(err)
		http.Error(w, s.Message(), GRPCCodeToHTTPStatus(s.Code()))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// DeactivateUser handles requests for a user to deactivate (soft delete) their own account.
func (h *UserHandler) DeactivateUser(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("user_id").(string)
	if !ok || userID == "" {
		h.logger.Warn("User ID not found in token for DeactivateUser")
		http.Error(w, "User ID not found in token", http.StatusUnauthorized)
		return
	}
	req := &user.DeactivateUserRequest{UserId: userID}
	resp, err := h.userClient.DeactivateUser(context.Background(), req)
	if err != nil {
		h.logger.Error("Failed to deactivate user via gRPC", zap.String("userID", userID), zap.Error(err))
		s, _ := status.FromError(err)
		http.Error(w, s.Message(), GRPCCodeToHTTPStatus(s.Code()))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// --- Admin Handlers ---

// AdminDeleteUser handles admin requests to (hard) delete a user.
func (h *UserHandler) AdminDeleteUser(w http.ResponseWriter, r *http.Request) {
	adminID, ok := r.Context().Value("user_id").(string)
	if !ok || adminID == "" {
		h.logger.Warn("Admin ID not found in token for AdminDeleteUser")
		http.Error(w, "Admin ID not found in token", http.StatusUnauthorized)
		return
	}
	var reqBody struct {
		UserIDToDelete string `json:"user_id_to_delete"`
	}
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		http.Error(w, "Invalid request body: missing user_id_to_delete", http.StatusBadRequest)
		return
	}
	if reqBody.UserIDToDelete == "" {
		http.Error(w, "user_id_to_delete is required", http.StatusBadRequest)
		return
	}
	grpcReq := &user.AdminDeleteUserRequest{AdminId: adminID, UserIdToDelete: reqBody.UserIDToDelete}
	resp, err := h.userClient.AdminDeleteUser(context.Background(), grpcReq)
	if err != nil {
		h.logger.Error("Failed to admin delete user (hard) via gRPC", zap.String("adminID", adminID), zap.String("targetUserID", reqBody.UserIDToDelete), zap.Error(err))
		s, _ := status.FromError(err)
		http.Error(w, s.Message(), GRPCCodeToHTTPStatus(s.Code()))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// AdminListUsers handles admin requests to list users.
func (h *UserHandler) AdminListUsers(w http.ResponseWriter, r *http.Request) {
	adminID, ok := r.Context().Value("user_id").(string)
	if !ok || adminID == "" {
		h.logger.Warn("Admin ID not found in token for AdminListUsers")
		http.Error(w, "Admin ID not found in token", http.StatusUnauthorized)
		return
	}
	var reqBody user.AdminListUsersRequest // Skip, Limit from body
	_ = json.NewDecoder(r.Body).Decode(&reqBody)
	reqBody.AdminId = adminID

	resp, err := h.userClient.AdminListUsers(context.Background(), &reqBody)
	if err != nil {
		h.logger.Error("Failed to list users by admin via gRPC", zap.String("adminID", adminID), zap.Error(err))
		s, _ := status.FromError(err)
		http.Error(w, s.Message(), GRPCCodeToHTTPStatus(s.Code()))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// AdminSearchUsers handles admin requests to search users.
func (h *UserHandler) AdminSearchUsers(w http.ResponseWriter, r *http.Request) {
	adminID, ok := r.Context().Value("user_id").(string)
	if !ok || adminID == "" {
		h.logger.Warn("Admin ID not found in token for AdminSearchUsers")
		http.Error(w, "Admin ID not found in token", http.StatusUnauthorized)
		return
	}
	var reqBody user.AdminSearchUsersRequest // Query, Skip, Limit from body
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		http.Error(w, "Invalid request body for AdminSearchUsers", http.StatusBadRequest)
		return
	}
	// Query can be empty, gRPC service should handle this (e.g., list all if query is empty)
	reqBody.AdminId = adminID

	resp, err := h.userClient.AdminSearchUsers(context.Background(), &reqBody)
	if err != nil {
		h.logger.Error("Failed to search users by admin via gRPC", zap.String("adminID", adminID), zap.String("query", reqBody.Query), zap.Error(err))
		s, _ := status.FromError(err)
		http.Error(w, s.Message(), GRPCCodeToHTTPStatus(s.Code()))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// AdminUpdateUserRole handles admin requests to update a user's role.
func (h *UserHandler) AdminUpdateUserRole(w http.ResponseWriter, r *http.Request) {
	adminID, ok := r.Context().Value("user_id").(string)
	if !ok || adminID == "" {
		h.logger.Warn("Admin ID not found in token for AdminUpdateUserRole")
		http.Error(w, "Admin ID not found in token", http.StatusUnauthorized)
		return
	}
	var reqBody struct {
		UserIDToUpdate string `json:"user_id_to_update"`
		Role           string `json:"role"`
	}
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		http.Error(w, "Invalid request body for AdminUpdateUserRole", http.StatusBadRequest)
		return
	}
	if reqBody.UserIDToUpdate == "" || reqBody.Role == "" {
		http.Error(w, "user_id_to_update and role are required", http.StatusBadRequest)
		return
	}
	grpcReq := &user.AdminUpdateUserRoleRequest{
		AdminId:        adminID,
		UserIdToUpdate: reqBody.UserIDToUpdate,
		Role:           reqBody.Role,
	}
	resp, err := h.userClient.AdminUpdateUserRole(context.Background(), grpcReq)
	if err != nil {
		h.logger.Error("Failed to update user role by admin via gRPC", zap.String("adminID", adminID), zap.String("targetUserID", reqBody.UserIDToUpdate), zap.Error(err))
		s, _ := status.FromError(err)
		http.Error(w, s.Message(), GRPCCodeToHTTPStatus(s.Code()))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// AdminSetUserActiveStatus handles admin requests to activate or deactivate a user.
func (h *UserHandler) AdminSetUserActiveStatus(w http.ResponseWriter, r *http.Request) {
	adminID, ok := r.Context().Value("user_id").(string)
	if !ok || adminID == "" {
		h.logger.Warn("Admin ID not found in token for AdminSetUserActiveStatus")
		http.Error(w, "Admin ID not found in token", http.StatusUnauthorized)
		return
	}
	var reqBody struct {
		UserID   string `json:"user_id"`
		IsActive bool   `json:"is_active"`
	}
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		http.Error(w, "Invalid request body for AdminSetUserActiveStatus", http.StatusBadRequest)
		return
	}
	if reqBody.UserID == "" {
		http.Error(w, "user_id is required", http.StatusBadRequest)
		return
	}
	grpcReq := &user.AdminSetUserActiveStatusRequest{
		AdminId:  adminID,
		UserId:   reqBody.UserID,
		IsActive: reqBody.IsActive,
	}
	resp, err := h.userClient.AdminSetUserActiveStatus(context.Background(), grpcReq)
	if err != nil {
		h.logger.Error("Failed to set user active status by admin via gRPC", zap.String("adminID", adminID), zap.String("targetUserID", reqBody.UserID), zap.Error(err))
		s, _ := status.FromError(err)
		http.Error(w, s.Message(), GRPCCodeToHTTPStatus(s.Code()))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// GRPCCodeToHTTPStatus maps gRPC status codes to HTTP status codes for consistent error handling.
func GRPCCodeToHTTPStatus(code codes.Code) int {
	switch code {
	case codes.OK:
		return http.StatusOK
	case codes.Canceled:
		return 499
	case codes.Unknown:
		return http.StatusInternalServerError
	case codes.InvalidArgument:
		return http.StatusBadRequest
	case codes.DeadlineExceeded:
		return http.StatusGatewayTimeout
	case codes.NotFound:
		return http.StatusNotFound
	case codes.AlreadyExists:
		return http.StatusConflict
	case codes.PermissionDenied:
		return http.StatusForbidden
	case codes.ResourceExhausted:
		return http.StatusTooManyRequests
	case codes.FailedPrecondition:
		return http.StatusPreconditionFailed
	case codes.Aborted:
		return http.StatusConflict
	case codes.OutOfRange:
		return http.StatusBadRequest
	case codes.Unimplemented:
		return http.StatusNotImplemented
	case codes.Internal:
		return http.StatusInternalServerError
	case codes.Unavailable:
		return http.StatusServiceUnavailable
	case codes.DataLoss:
		return http.StatusInternalServerError
	case codes.Unauthenticated:
		return http.StatusUnauthorized
	default:
		return http.StatusInternalServerError
	}
}
