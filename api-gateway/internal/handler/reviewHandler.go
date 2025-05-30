package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/Abdurahmanit/GroupProject/api-gateway/internal/middleware" // Для UserIDCtxKey
	// Используем ваш вариант импорта pb, предполагая, что он работает для других сервисов
	pb "github.com/Abdurahmanit/GroupProject/review-service"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
	"google.golang.org/grpc" // Для *grpc.ClientConn
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// ReviewHandler обрабатывает HTTP запросы для Review Service.
type ReviewHandler struct {
	client pb.ReviewServiceClient
	logger *zap.Logger
}

// NewReviewHandler создает новый ReviewHandler.
func NewReviewHandler(conn *grpc.ClientConn, logger *zap.Logger) *ReviewHandler {
	return &ReviewHandler{
		client: pb.NewReviewServiceClient(conn),
		logger: logger.Named("ReviewHTTPHandler"),
	}
}

// --- Начало вспомогательных функций (если их нет в общем месте) ---

func withAuthFromHttpRequest(ctx context.Context, r *http.Request) context.Context {
	token := r.Header.Get("Authorization")
	if token != "" {
		return metadata.NewOutgoingContext(ctx, metadata.Pairs("authorization", token))
	}
	return ctx
}

func parseIntQueryParam(r *http.Request, key string, defaultValue int32) int32 {
	valStr := r.URL.Query().Get(key)
	if valStr == "" {
		return defaultValue
	}
	valInt, err := strconv.ParseInt(valStr, 10, 32)
	if err != nil {
		return defaultValue
	}
	return int32(valInt)
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	if payload != nil {
		if err := json.NewEncoder(w).Encode(payload); err != nil {
			// log.Printf("Error encoding JSON response: %v", err) // Используйте ваш логгер
		}
	}
}

func handleGRPCError(w http.ResponseWriter, err error, defaultMessage string, logger *zap.Logger) {
	st, ok := status.FromError(err)
	if ok {
		httpStatus := GRPCCodeToHTTPStatus(st.Code())
		logger.Warn("gRPC error occurred", zap.String("grpc_code", st.Code().String()), zap.String("grpc_message", st.Message()), zap.Int("http_status", httpStatus))
		http.Error(w, st.Message(), httpStatus)
	} else {
		logger.Error("Non-gRPC error occurred or failed to convert to gRPC status", zap.Error(err), zap.String("default_message", defaultMessage))
		http.Error(w, defaultMessage+": "+err.Error(), http.StatusInternalServerError)
	}
}

// --- Конец вспомогательных функций ---

func (h *ReviewHandler) HandleCreateReview(w http.ResponseWriter, r *http.Request) {
	var req pb.CreateReviewRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("Invalid request body for CreateReview", zap.Error(err))
		http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	userIDFromToken, ok := r.Context().Value(middleware.UserIDCtxKey).(string)
	if !ok || userIDFromToken == "" {
		h.logger.Warn("CreateReview: User ID not found in token context")
		http.Error(w, "Unauthorized: User ID missing from token", http.StatusUnauthorized)
		return
	}
	req.UserId = userIDFromToken

	ctx := withAuthFromHttpRequest(r.Context(), r)
	resp, err := h.client.CreateReview(ctx, &req)
	if err != nil {
		h.logger.Error("gRPC CreateReview call failed", zap.Error(err))
		handleGRPCError(w, err, "Failed to create review", h.logger)
		return
	}
	respondWithJSON(w, http.StatusCreated, resp)
}

func (h *ReviewHandler) HandleGetReview(w http.ResponseWriter, r *http.Request) {
	reviewID := chi.URLParam(r, "reviewId")
	if reviewID == "" {
		http.Error(w, "Missing review ID", http.StatusBadRequest)
		return
	}
	req := &pb.GetReviewRequest{ReviewId: reviewID}
	resp, err := h.client.GetReview(context.Background(), req)
	if err != nil {
		h.logger.Error("gRPC GetReview call failed", zap.String("review_id", reviewID), zap.Error(err))
		handleGRPCError(w, err, "Failed to get review", h.logger)
		return
	}
	respondWithJSON(w, http.StatusOK, resp)
}

func (h *ReviewHandler) HandleUpdateReview(w http.ResponseWriter, r *http.Request) {
	reviewID := chi.URLParam(r, "reviewId")
	userIDFromToken, ok := r.Context().Value(middleware.UserIDCtxKey).(string)
	if !ok || userIDFromToken == "" {
		h.logger.Warn("UpdateReview: User ID not found in token context")
		http.Error(w, "Unauthorized: User ID missing", http.StatusUnauthorized)
		return
	}

	var req pb.UpdateReviewRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}
	req.ReviewId = reviewID
	req.UserId = userIDFromToken

	ctx := withAuthFromHttpRequest(r.Context(), r)
	resp, err := h.client.UpdateReview(ctx, &req)
	if err != nil {
		h.logger.Error("gRPC UpdateReview call failed", zap.String("review_id", reviewID), zap.Error(err))
		handleGRPCError(w, err, "Failed to update review", h.logger)
		return
	}
	respondWithJSON(w, http.StatusOK, resp)
}

func (h *ReviewHandler) HandleDeleteReview(w http.ResponseWriter, r *http.Request) {
	reviewID := chi.URLParam(r, "reviewId")
	userIDFromToken, ok := r.Context().Value(middleware.UserIDCtxKey).(string)
	if !ok || userIDFromToken == "" {
		h.logger.Warn("DeleteReview: User ID not found in token context")
		http.Error(w, "Unauthorized: User ID missing", http.StatusUnauthorized)
		return
	}

	req := &pb.DeleteReviewRequest{
		ReviewId: reviewID,
		UserId:   userIDFromToken,
	}

	ctx := withAuthFromHttpRequest(r.Context(), r)
	_, err := h.client.DeleteReview(ctx, req)
	if err != nil {
		h.logger.Error("gRPC DeleteReview call failed", zap.String("review_id", reviewID), zap.Error(err))
		handleGRPCError(w, err, "Failed to delete review", h.logger)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *ReviewHandler) HandleListReviewsByProduct(w http.ResponseWriter, r *http.Request) {
	productID := chi.URLParam(r, "productId")
	if productID == "" {
		http.Error(w, "Missing product ID", http.StatusBadRequest)
		return
	}

	page := parseIntQueryParam(r, "page", 1)
	limit := parseIntQueryParam(r, "limit", 10)
	statusFilter := r.URL.Query().Get("status")

	req := &pb.ListReviewsByProductRequest{
		ProductId:    productID,
		Page:         page,
		Limit:        limit,
		StatusFilter: statusFilter,
	}

	resp, err := h.client.ListReviewsByProduct(context.Background(), req)
	if err != nil {
		h.logger.Error("gRPC ListReviewsByProduct call failed", zap.String("product_id", productID), zap.Error(err))
		handleGRPCError(w, err, "Failed to list reviews for product", h.logger)
		return
	}
	respondWithJSON(w, http.StatusOK, resp)
}

func (h *ReviewHandler) HandleListReviewsByUser(w http.ResponseWriter, r *http.Request) {
	userIDFromToken, ok := r.Context().Value(middleware.UserIDCtxKey).(string)
	if !ok || userIDFromToken == "" {
		h.logger.Warn("ListReviewsByUser: User ID not found in token context")
		http.Error(w, "Unauthorized: User ID missing", http.StatusUnauthorized)
		return
	}

	page := parseIntQueryParam(r, "page", 1)
	limit := parseIntQueryParam(r, "limit", 10)

	req := &pb.ListReviewsByUserRequest{
		UserId: userIDFromToken,
		Page:   page,
		Limit:  limit,
	}

	ctx := withAuthFromHttpRequest(r.Context(), r)
	resp, err := h.client.ListReviewsByUser(ctx, req)
	if err != nil {
		h.logger.Error("gRPC ListReviewsByUser call failed", zap.String("user_id", userIDFromToken), zap.Error(err))
		handleGRPCError(w, err, "Failed to list reviews for user", h.logger)
		return
	}
	respondWithJSON(w, http.StatusOK, resp)
}

func (h *ReviewHandler) HandleGetProductAverageRating(w http.ResponseWriter, r *http.Request) {
	productID := chi.URLParam(r, "productId")
	if productID == "" {
		http.Error(w, "Missing product ID", http.StatusBadRequest)
		return
	}

	req := &pb.GetProductAverageRatingRequest{ProductId: productID}
	resp, err := h.client.GetProductAverageRating(context.Background(), req)
	if err != nil {
		h.logger.Error("gRPC GetProductAverageRating call failed", zap.String("product_id", productID), zap.Error(err))
		handleGRPCError(w, err, "Failed to get product average rating", h.logger)
		return
	}
	respondWithJSON(w, http.StatusOK, resp)
}

func (h *ReviewHandler) HandleModerateReview(w http.ResponseWriter, r *http.Request) {
	reviewID := chi.URLParam(r, "reviewId")
	adminIDFromToken, ok := r.Context().Value(middleware.UserIDCtxKey).(string) // Используем UserIDCtxKey для ID админа
	if !ok || adminIDFromToken == "" {
		h.logger.Warn("ModerateReview: Admin User ID not found in token context")
		http.Error(w, "Unauthorized: Admin ID missing", http.StatusUnauthorized)
		return
	}
	// Проверка роли админа должна быть в middleware.AdminOnly

	var reqBody struct {
		NewStatus         string `json:"new_status"`
		ModerationComment string `json:"moderation_comment"`
	}
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	req := &pb.ModerateReviewRequest{
		ReviewId:          reviewID,
		AdminId:           adminIDFromToken,
		NewStatus:         reqBody.NewStatus,
		ModerationComment: reqBody.ModerationComment,
	}

	ctx := withAuthFromHttpRequest(r.Context(), r)
	resp, err := h.client.ModerateReview(ctx, req)
	if err != nil {
		h.logger.Error("gRPC ModerateReview call failed", zap.String("review_id", reviewID), zap.Error(err))
		handleGRPCError(w, err, "Failed to moderate review", h.logger)
		return
	}
	respondWithJSON(w, http.StatusOK, resp)
}
