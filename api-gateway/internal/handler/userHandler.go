package handler

import (
	"encoding/json"
	"net/http"

	user "github.com/Abdurahmanit/GroupProject/user-service/proto" // Ensure this path is correct
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
		logger:     logger.Named("UserHTTPHandler"),
	}
}

func (h *UserHandler) Register(w http.ResponseWriter, r *http.Request) {
	var grpcReq user.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&grpcReq); err != nil {
		h.logger.Error("Failed to decode request body for Register HTTP", zap.Error(err))
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if grpcReq.GetUsername() == "" || grpcReq.GetEmail() == "" || grpcReq.GetPassword() == "" || grpcReq.GetPhoneNumber() == "" {
		h.logger.Warn("Missing required fields for Register HTTP",
			zap.String("username", grpcReq.GetUsername()),
			zap.String("email", grpcReq.GetEmail()),
			zap.Bool("passwordEmpty", grpcReq.GetPassword() == ""),
			zap.String("phoneNumber", grpcReq.GetPhoneNumber()))
		http.Error(w, "Username, email, password, and phone number are required", http.StatusBadRequest)
		return
	}
	h.logger.Info("HTTP Register request received", zap.String("email", grpcReq.GetEmail()))

	resp, err := h.userClient.Register(r.Context(), &grpcReq)
	if err != nil {
		h.logger.Error("Failed to register user via gRPC from API Gateway", zap.String("email", grpcReq.GetEmail()), zap.Error(err))
		s, _ := status.FromError(err)
		http.Error(w, s.Message(), GRPCCodeToHTTPStatus(s.Code()))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)
	h.logger.Info("HTTP Register request processed successfully", zap.String("email", grpcReq.GetEmail()))
}

func (h *UserHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req user.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("Failed to decode request for Login", zap.Error(err))
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	resp, err := h.userClient.Login(r.Context(), &req) // Use r.Context()
	if err != nil {
		h.logger.Error("Failed to login user via gRPC", zap.Error(err))
		s, _ := status.FromError(err)
		http.Error(w, s.Message(), GRPCCodeToHTTPStatus(s.Code()))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *UserHandler) Logout(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("user_id").(string)
	if !ok || userID == "" {
		h.logger.Warn("User ID not found in token for Logout")
		http.Error(w, "User ID not found in token", http.StatusUnauthorized)
		return
	}
	req := &user.LogoutRequest{UserId: userID}
	resp, err := h.userClient.Logout(r.Context(), req)
	if err != nil {
		h.logger.Error("Failed to logout user via gRPC", zap.String("userID", userID), zap.Error(err))
		s, _ := status.FromError(err)
		http.Error(w, s.Message(), GRPCCodeToHTTPStatus(s.Code()))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *UserHandler) GetProfile(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("user_id").(string)
	if !ok || userID == "" {
		h.logger.Warn("User ID not found in token for GetProfile HTTP request")
		http.Error(w, "User ID not found in token", http.StatusUnauthorized)
		return
	}
	h.logger.Info("HTTP GetProfile request received", zap.String("userID", userID))
	grpcReq := &user.GetProfileRequest{UserId: userID}
	resp, err := h.userClient.GetProfile(r.Context(), grpcReq)
	if err != nil {
		h.logger.Error("Failed to get profile via gRPC from API Gateway", zap.String("userID", userID), zap.Error(err))
		s, _ := status.FromError(err)
		http.Error(w, s.Message(), GRPCCodeToHTTPStatus(s.Code()))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
	h.logger.Info("HTTP GetProfile request processed successfully", zap.String("userID", userID))
}

func (h *UserHandler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("user_id").(string)
	if !ok || userID == "" {
		h.logger.Warn("User ID not found in token for UpdateProfile HTTP request")
		http.Error(w, "User ID not found in token", http.StatusUnauthorized)
		return
	}

	var grpcReq user.UpdateProfileRequest
	if err := json.NewDecoder(r.Body).Decode(&grpcReq); err != nil {
		h.logger.Error("Failed to decode request body for UpdateProfile HTTP", zap.String("userID", userID), zap.Error(err))
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	grpcReq.UserId = userID // Ensure UserId from token is used

	h.logger.Info("HTTP UpdateProfile request received", zap.String("userID", userID),
		zap.String("username", grpcReq.GetUsername()),
		zap.String("email", grpcReq.GetEmail()),
		zap.String("phoneNumber", grpcReq.GetPhoneNumber()))

	resp, err := h.userClient.UpdateProfile(r.Context(), &grpcReq)
	if err != nil {
		h.logger.Error("Failed to update profile via gRPC from API Gateway", zap.String("userID", userID), zap.Error(err))
		s, _ := status.FromError(err)
		http.Error(w, s.Message(), GRPCCodeToHTTPStatus(s.Code()))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
	h.logger.Info("HTTP UpdateProfile request processed successfully", zap.String("userID", userID))
}

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

	resp, err := h.userClient.ChangePassword(r.Context(), &reqBody)
	if err != nil {
		h.logger.Error("Failed to change password via gRPC", zap.String("userID", userID), zap.Error(err))
		s, _ := status.FromError(err)
		http.Error(w, s.Message(), GRPCCodeToHTTPStatus(s.Code()))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// Email Verification Handlers
func (h *UserHandler) RequestEmailVerification(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("user_id").(string)
	if !ok || userID == "" {
		h.logger.Warn("User ID not found in token for RequestEmailVerification")
		http.Error(w, "User ID not found in token", http.StatusUnauthorized)
		return
	}
	h.logger.Info("HTTP RequestEmailVerification request received", zap.String("userID", userID))

	grpcReq := &user.RequestEmailVerificationRequest{UserId: userID}
	resp, err := h.userClient.RequestEmailVerification(r.Context(), grpcReq)
	if err != nil {
		h.logger.Error("gRPC RequestEmailVerification call failed", zap.String("userID", userID), zap.Error(err))
		s, _ := status.FromError(err)
		http.Error(w, s.Message(), GRPCCodeToHTTPStatus(s.Code()))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	// If gRPC call was successful (err == nil), but operation wasn't (Success: false),
	// we return HTTP 200 with the specific message, or a relevant client error code like 409.
	if !resp.GetSuccess() {
		// Example: "email already verified" or "invalid code" are business logic failures, not server errors.
		// You can choose to return 200 OK and let client parse .Success and .Message,
		// or map common messages to specific HTTP status codes.
		// For "email already verified", 409 Conflict might be suitable.
		// For now, simple logic: if not success, return 200 with the message, or 409 if message indicates specific conflict.
		// This part was simplified by removing direct usecase.Error comparison.
		// We can check resp.Message for specific strings if absolutely needed, but generally prefer relying on Success field.
		// The previous logic was:
		// if !resp.GetSuccess() && resp.Message == usecase.ErrEmailAlreadyVerified.Error()
		// For now, let's use a general approach:
		if resp.GetMessage() == "email is already verified" { // Example of specific message checking
			w.WriteHeader(http.StatusConflict) // Or http.StatusOK
		} else {
			w.WriteHeader(http.StatusOK) // Or http.StatusBadRequest for other logical failures
		}
	} else {
		w.WriteHeader(http.StatusOK)
	}
	json.NewEncoder(w).Encode(resp)
	h.logger.Info("HTTP RequestEmailVerification request processed", zap.String("userID", userID), zap.Bool("success", resp.GetSuccess()))
}

func (h *UserHandler) VerifyEmail(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("user_id").(string)
	if !ok || userID == "" {
		h.logger.Warn("User ID not found in token for VerifyEmail")
		http.Error(w, "User ID not found in token", http.StatusUnauthorized)
		return
	}

	var reqBody struct {
		Code string `json:"code"`
	}
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		http.Error(w, "Invalid request body: missing code", http.StatusBadRequest)
		return
	}
	if reqBody.Code == "" {
		http.Error(w, "Verification code is required", http.StatusBadRequest)
		return
	}
	h.logger.Info("HTTP VerifyEmail request received", zap.String("userID", userID))

	grpcReq := &user.VerifyEmailRequest{UserId: userID, Code: reqBody.Code}
	resp, err := h.userClient.VerifyEmail(r.Context(), grpcReq)
	if err != nil {
		h.logger.Error("gRPC VerifyEmail call failed", zap.String("userID", userID), zap.Error(err))
		s, _ := status.FromError(err)
		http.Error(w, s.Message(), GRPCCodeToHTTPStatus(s.Code()))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	// Similar logic for VerifyEmail: check resp.GetSuccess()
	if !resp.GetSuccess() {
		// If message is "email is already verified" or "invalid or expired verification code"
		if resp.GetMessage() == "email is already verified" || resp.GetMessage() == "invalid or expired verification code" {
			w.WriteHeader(http.StatusConflict) // Or http.StatusOK / http.StatusBadRequest
		} else {
			w.WriteHeader(http.StatusOK) // Or another appropriate client error
		}
	} else {
		w.WriteHeader(http.StatusOK)
	}
	json.NewEncoder(w).Encode(resp)
	h.logger.Info("HTTP VerifyEmail request processed", zap.String("userID", userID), zap.Bool("success", resp.GetSuccess()))
}

func (h *UserHandler) CheckEmailVerificationStatus(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("user_id").(string)
	if !ok || userID == "" {
		h.logger.Warn("User ID not found in token for CheckEmailVerificationStatus")
		http.Error(w, "User ID not found in token", http.StatusUnauthorized)
		return
	}
	h.logger.Info("HTTP CheckEmailVerificationStatus request received", zap.String("userID", userID))

	grpcReq := &user.CheckEmailVerificationStatusRequest{UserId: userID}
	resp, err := h.userClient.CheckEmailVerificationStatus(r.Context(), grpcReq)
	if err != nil {
		h.logger.Error("gRPC CheckEmailVerificationStatus call failed", zap.String("userID", userID), zap.Error(err))
		s, _ := status.FromError(err)
		http.Error(w, s.Message(), GRPCCodeToHTTPStatus(s.Code()))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
	h.logger.Info("HTTP CheckEmailVerificationStatus request processed", zap.String("userID", userID), zap.Bool("is_verified", resp.GetIsVerified()))
}

func (h *UserHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("user_id").(string)
	if !ok || userID == "" {
		h.logger.Warn("User ID not found in token for DeleteUser")
		http.Error(w, "User ID not found in token", http.StatusUnauthorized)
		return
	}
	req := &user.DeleteUserRequest{UserId: userID}
	resp, err := h.userClient.DeleteUser(r.Context(), req)
	if err != nil {
		h.logger.Error("Failed to delete user (hard) via gRPC", zap.String("userID", userID), zap.Error(err))
		s, _ := status.FromError(err)
		http.Error(w, s.Message(), GRPCCodeToHTTPStatus(s.Code()))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *UserHandler) DeactivateUser(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("user_id").(string)
	if !ok || userID == "" {
		h.logger.Warn("User ID not found in token for DeactivateUser")
		http.Error(w, "User ID not found in token", http.StatusUnauthorized)
		return
	}
	req := &user.DeactivateUserRequest{UserId: userID}
	resp, err := h.userClient.DeactivateUser(r.Context(), req)
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
	resp, err := h.userClient.AdminDeleteUser(r.Context(), grpcReq)
	if err != nil {
		h.logger.Error("Failed to admin delete user (hard) via gRPC", zap.String("adminID", adminID), zap.String("targetUserID", reqBody.UserIDToDelete), zap.Error(err))
		s, _ := status.FromError(err)
		http.Error(w, s.Message(), GRPCCodeToHTTPStatus(s.Code()))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *UserHandler) AdminListUsers(w http.ResponseWriter, r *http.Request) {
	adminID, ok := r.Context().Value("user_id").(string)
	if !ok || adminID == "" {
		h.logger.Warn("Admin ID not found in token for AdminListUsers")
		http.Error(w, "Admin ID not found in token", http.StatusUnauthorized)
		return
	}
	var reqBody user.AdminListUsersRequest
	_ = json.NewDecoder(r.Body).Decode(&reqBody)
	reqBody.AdminId = adminID

	resp, err := h.userClient.AdminListUsers(r.Context(), &reqBody)
	if err != nil {
		h.logger.Error("Failed to list users by admin via gRPC", zap.String("adminID", adminID), zap.Error(err))
		s, _ := status.FromError(err)
		http.Error(w, s.Message(), GRPCCodeToHTTPStatus(s.Code()))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *UserHandler) AdminSearchUsers(w http.ResponseWriter, r *http.Request) {
	adminID, ok := r.Context().Value("user_id").(string)
	if !ok || adminID == "" {
		h.logger.Warn("Admin ID not found in token for AdminSearchUsers")
		http.Error(w, "Admin ID not found in token", http.StatusUnauthorized)
		return
	}
	var reqBody user.AdminSearchUsersRequest
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		http.Error(w, "Invalid request body for AdminSearchUsers", http.StatusBadRequest)
		return
	}
	reqBody.AdminId = adminID

	resp, err := h.userClient.AdminSearchUsers(r.Context(), &reqBody)
	if err != nil {
		h.logger.Error("Failed to search users by admin via gRPC", zap.String("adminID", adminID), zap.String("query", reqBody.Query), zap.Error(err))
		s, _ := status.FromError(err)
		http.Error(w, s.Message(), GRPCCodeToHTTPStatus(s.Code()))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

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
	resp, err := h.userClient.AdminUpdateUserRole(r.Context(), grpcReq)
	if err != nil {
		h.logger.Error("Failed to update user role by admin via gRPC", zap.String("adminID", adminID), zap.String("targetUserID", reqBody.UserIDToUpdate), zap.Error(err))
		s, _ := status.FromError(err)
		http.Error(w, s.Message(), GRPCCodeToHTTPStatus(s.Code()))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

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
	resp, err := h.userClient.AdminSetUserActiveStatus(r.Context(), grpcReq)
	if err != nil {
		h.logger.Error("Failed to set user active status by admin via gRPC", zap.String("adminID", adminID), zap.String("targetUserID", reqBody.UserID), zap.Error(err))
		s, _ := status.FromError(err)
		http.Error(w, s.Message(), GRPCCodeToHTTPStatus(s.Code()))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// GRPCCodeToHTTPStatus maps gRPC status codes to HTTP status codes.
func GRPCCodeToHTTPStatus(code codes.Code) int {
	switch code {
	case codes.OK:
		return http.StatusOK
	case codes.Canceled:
		return 499 // Client Closed Request
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
		return http.StatusBadRequest
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
