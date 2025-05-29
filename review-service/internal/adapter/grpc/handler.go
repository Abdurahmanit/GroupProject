package grpc

import (
	"context"
	"errors"

	pb "github.com/Abdurahmanit/GroupProject/review-service/genproto/review_service" // Path to your generated protobuf code
	"github.com/Abdurahmanit/GroupProject/review-service/internal/middleware"        // Assuming middleware is in a shared or local internal path
	"github.com/Abdurahmanit/GroupProject/review-service/internal/platform/logger"
	"github.com/Abdurahmanit/GroupProject/review-service/internal/review/domain"
	"github.com/Abdurahmanit/GroupProject/review-service/internal/review/usecase"

	"go.mongodb.org/mongo-driver/bson/primitive"
	zap "go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// ReviewHandler implements the gRPC service for reviews.
type ReviewHandler struct {
	pb.UnimplementedReviewServiceServer
	usecase *usecase.ReviewUsecase
	logger  *logger.Logger
}

// NewReviewHandler creates a new gRPC handler for the review service.
func NewReviewHandler(uc *usecase.ReviewUsecase, log *logger.Logger) *ReviewHandler {
	return &ReviewHandler{
		usecase: uc,
		logger:  log.Named("ReviewGRPCHandler"),
	}
}

// toProtoReview converts a domain.Review to its protobuf representation.
func toProtoReview(review *domain.Review) *pb.Review {
	if review == nil {
		return nil
	}
	return &pb.Review{
		Id:        review.ID.Hex(),
		UserId:    review.UserID,
		ProductId: review.ProductID,
		SellerId:  review.SellerID,
		Rating:    review.Rating,
		Comment:   review.Comment,
		Status:    string(review.Status),
		CreatedAt: timestamppb.New(review.CreatedAt),
		UpdatedAt: timestamppb.New(review.UpdatedAt),
		// ModerationComment: review.ModerationComment, // Add if in proto
	}
}

// CreateReview handles the creation of a new review.
func (h *ReviewHandler) CreateReview(ctx context.Context, req *pb.CreateReviewRequest) (*pb.Review, error) {
	// UserID should be extracted from the JWT token by an auth interceptor and put into context.
	// The API Gateway should pass the JWT, and the gRPC auth interceptor in review-service should validate it.
	authenticatedUserID, ok := ctx.Value(middleware.UserIDKey).(string)
	if !ok || authenticatedUserID == "" {
		h.logger.Warn("CreateReview: UserID not found in context or is empty", zap.String("request_user_id", req.GetUserId()))
		return nil, status.Errorf(codes.Unauthenticated, "user authentication required")
	}

	// Ensure the user is creating a review for themselves, or if req.UserId is meant to be the review author.
	// Typically, the authenticatedUserID IS the author.
	if req.GetUserId() != "" && req.GetUserId() != authenticatedUserID {
		h.logger.Warn("CreateReview: Authenticated user attempting to create review for another user",
			zap.String("authenticated_user_id", authenticatedUserID),
			zap.String("request_author_id", req.GetUserId()))
		return nil, status.Errorf(codes.PermissionDenied, "cannot create review for another user")
	}

	// Use authenticatedUserID as the author.
	authorID := authenticatedUserID

	h.logger.Info("CreateReview RPC called",
		zap.String("author_id", authorID),
		zap.String("product_id", req.GetProductId()),
		zap.Int32("rating", req.GetRating()))

	review, err := h.usecase.CreateReview(ctx, authorID, req.GetProductId(), req.GetSellerId(), req.GetComment(), req.GetRating())
	if err != nil {
		h.logger.Error("CreateReview usecase failed", zap.Error(err), zap.String("author_id", authorID))
		if errors.Is(err, domain.ErrReviewAlreadyExists) { // Example custom error
			return nil, status.Errorf(codes.AlreadyExists, err.Error())
		}
		if errors.Is(err, domain.ErrInvalidInput) {
			return nil, status.Errorf(codes.InvalidArgument, err.Error())
		}
		return nil, status.Errorf(codes.Internal, "failed to create review: %v", err)
	}

	h.logger.Info("Review created successfully", zap.String("review_id", review.ID.Hex()))
	return toProtoReview(review), nil
}

// GetReview retrieves a review by its ID.
func (h *ReviewHandler) GetReview(ctx context.Context, req *pb.GetReviewRequest) (*pb.Review, error) {
	h.logger.Info("GetReview RPC called", zap.String("review_id", req.GetReviewId()))

	reviewID, err := primitive.ObjectIDFromHex(req.GetReviewId())
	if err != nil {
		h.logger.Warn("GetReview: Invalid review_id format", zap.String("review_id", req.GetReviewId()), zap.Error(err))
		return nil, status.Errorf(codes.InvalidArgument, "invalid review ID format")
	}

	review, err := h.usecase.GetReview(ctx, reviewID)
	if err != nil {
		h.logger.Error("GetReview usecase failed", zap.Error(err), zap.String("review_id", req.GetReviewId()))
		if errors.Is(err, domain.ErrNotFound) {
			return nil, status.Errorf(codes.NotFound, "review not found")
		}
		return nil, status.Errorf(codes.Internal, "failed to get review: %v", err)
	}

	return toProtoReview(review), nil
}

// UpdateReview allows a user to update their own review.
func (h *ReviewHandler) UpdateReview(ctx context.Context, req *pb.UpdateReviewRequest) (*pb.Review, error) {
	authenticatedUserID, ok := ctx.Value(middleware.UserIDKey).(string)
	if !ok || authenticatedUserID == "" {
		h.logger.Warn("UpdateReview: UserID not found in context")
		return nil, status.Errorf(codes.Unauthenticated, "user authentication required")
	}

	// Ensure the user_id in the request matches the authenticated user, if provided for an extra check.
	// The primary authorization check should be in the usecase against the review's actual author.
	if req.GetUserId() != "" && req.GetUserId() != authenticatedUserID {
		h.logger.Warn("UpdateReview: Authenticated user ID does not match user_id in request",
			zap.String("authenticated_user_id", authenticatedUserID),
			zap.String("request_user_id", req.GetUserId()))
		// Depending on policy, this could be an error or just a log.
		// For now, we rely on the usecase to check ownership based on reviewID and authenticatedUserID.
	}

	h.logger.Info("UpdateReview RPC called",
		zap.String("review_id", req.GetReviewId()),
		zap.String("user_id_performing_update", authenticatedUserID))

	reviewID, err := primitive.ObjectIDFromHex(req.GetReviewId())
	if err != nil {
		h.logger.Warn("UpdateReview: Invalid review_id format", zap.String("review_id", req.GetReviewId()), zap.Error(err))
		return nil, status.Errorf(codes.InvalidArgument, "invalid review ID format")
	}

	var ratingToUpdate *int32
	if req.Rating != 0 { // Assuming 0 means not provided for update, adjust if 0 is a valid rating
		r := req.GetRating()
		ratingToUpdate = &r
	}
	var commentToUpdate *string
	if req.Comment != "" { // Assuming empty string means not provided for update
		c := req.GetComment()
		commentToUpdate = &c
	}

	review, err := h.usecase.UpdateReview(ctx, reviewID, authenticatedUserID, ratingToUpdate, commentToUpdate)
	if err != nil {
		h.logger.Error("UpdateReview usecase failed", zap.Error(err), zap.String("review_id", req.GetReviewId()))
		if errors.Is(err, domain.ErrNotFound) {
			return nil, status.Errorf(codes.NotFound, "review not found")
		}
		if errors.Is(err, domain.ErrForbidden) {
			return nil, status.Errorf(codes.PermissionDenied, "user not authorized to update this review")
		}
		if errors.Is(err, domain.ErrInvalidInput) {
			return nil, status.Errorf(codes.InvalidArgument, err.Error())
		}
		return nil, status.Errorf(codes.Internal, "failed to update review: %v", err)
	}

	return toProtoReview(review), nil
}

// DeleteReview allows a user to delete their own review.
func (h *ReviewHandler) DeleteReview(ctx context.Context, req *pb.DeleteReviewRequest) (*emptypb.Empty, error) {
	authenticatedUserID, ok := ctx.Value(middleware.UserIDKey).(string)
	if !ok || authenticatedUserID == "" {
		h.logger.Warn("DeleteReview: UserID not found in context")
		return nil, status.Errorf(codes.Unauthenticated, "user authentication required")
	}

	if req.GetUserId() != "" && req.GetUserId() != authenticatedUserID {
		h.logger.Warn("DeleteReview: Authenticated user ID does not match user_id in request",
			zap.String("authenticated_user_id", authenticatedUserID),
			zap.String("request_user_id", req.GetUserId()))
	}

	h.logger.Info("DeleteReview RPC called",
		zap.String("review_id", req.GetReviewId()),
		zap.String("user_id_performing_delete", authenticatedUserID))

	reviewID, err := primitive.ObjectIDFromHex(req.GetReviewId())
	if err != nil {
		h.logger.Warn("DeleteReview: Invalid review_id format", zap.String("review_id", req.GetReviewId()), zap.Error(err))
		return nil, status.Errorf(codes.InvalidArgument, "invalid review ID format")
	}

	err = h.usecase.DeleteReview(ctx, reviewID, authenticatedUserID)
	if err != nil {
		h.logger.Error("DeleteReview usecase failed", zap.Error(err), zap.String("review_id", req.GetReviewId()))
		if errors.Is(err, domain.ErrNotFound) {
			return nil, status.Errorf(codes.NotFound, "review not found")
		}
		if errors.Is(err, domain.ErrForbidden) {
			return nil, status.Errorf(codes.PermissionDenied, "user not authorized to delete this review")
		}
		return nil, status.Errorf(codes.Internal, "failed to delete review: %v", err)
	}

	return &emptypb.Empty{}, nil
}

// ListReviewsByProduct lists reviews for a given product. This is likely a public endpoint.
func (h *ReviewHandler) ListReviewsByProduct(ctx context.Context, req *pb.ListReviewsByProductRequest) (*pb.ListReviewsResponse, error) {
	h.logger.Info("ListReviewsByProduct RPC called", zap.String("product_id", req.GetProductId()))

	if req.GetProductId() == "" {
		return nil, status.Errorf(codes.InvalidArgument, "product_id is required")
	}

	var statusFilter *string
	if req.GetStatusFilter() != "" {
		sf := req.GetStatusFilter()
		statusFilter = &sf
	}

	reviews, total, err := h.usecase.ListReviewsByProduct(ctx, req.GetProductId(), req.GetPage(), req.GetLimit(), statusFilter)
	if err != nil {
		h.logger.Error("ListReviewsByProduct usecase failed", zap.Error(err), zap.String("product_id", req.GetProductId()))
		return nil, status.Errorf(codes.Internal, "failed to list reviews by product: %v", err)
	}

	protoReviews := make([]*pb.Review, len(reviews))
	for i, r := range reviews {
		protoReviews[i] = toProtoReview(r)
	}

	return &pb.ListReviewsResponse{
		Reviews: protoReviews,
		Total:   total,
		Page:    req.GetPage(),
		Limit:   req.GetLimit(),
	}, nil
}

// ListReviewsByUser lists reviews written by a specific user. Requires authentication.
func (h *ReviewHandler) ListReviewsByUser(ctx context.Context, req *pb.ListReviewsByUserRequest) (*pb.ListReviewsResponse, error) {
	authenticatedUserID, ok := ctx.Value(middleware.UserIDKey).(string)
	if !ok || authenticatedUserID == "" {
		h.logger.Warn("ListReviewsByUser: UserID not found in context")
		return nil, status.Errorf(codes.Unauthenticated, "user authentication required")
	}

	// User can only list their own reviews.
	targetUserID := req.GetUserId()
	if targetUserID == "" { // If not specified in request, defaults to authenticated user
		targetUserID = authenticatedUserID
	} else if targetUserID != authenticatedUserID {
		h.logger.Warn("ListReviewsByUser: Attempt to list reviews for another user",
			zap.String("authenticated_user_id", authenticatedUserID),
			zap.String("requested_user_id", targetUserID))
		return nil, status.Errorf(codes.PermissionDenied, "cannot list reviews for another user")
	}

	h.logger.Info("ListReviewsByUser RPC called", zap.String("user_id", targetUserID))

	reviews, total, err := h.usecase.ListReviewsByUser(ctx, targetUserID, req.GetPage(), req.GetLimit())
	if err != nil {
		h.logger.Error("ListReviewsByUser usecase failed", zap.Error(err), zap.String("user_id", targetUserID))
		return nil, status.Errorf(codes.Internal, "failed to list reviews by user: %v", err)
	}

	protoReviews := make([]*pb.Review, len(reviews))
	for i, r := range reviews {
		protoReviews[i] = toProtoReview(r)
	}

	return &pb.ListReviewsResponse{
		Reviews: protoReviews,
		Total:   total,
		Page:    req.GetPage(),
		Limit:   req.GetLimit(),
	}, nil
}

// ModerateReview allows an admin to moderate a review.
func (h *ReviewHandler) ModerateReview(ctx context.Context, req *pb.ModerateReviewRequest) (*pb.Review, error) {
	adminID, ok := ctx.Value(middleware.UserIDKey).(string) // Assuming admin ID is also in UserIDKey after role check
	if !ok || adminID == "" {
		h.logger.Warn("ModerateReview: Admin UserID not found in context")
		return nil, status.Errorf(codes.Unauthenticated, "admin authentication required")
	}
	// Further role check (e.g., `ctx.Value(middleware.UserRoleKey) == "admin"`) should happen in AuthInterceptor or usecase.
	// For simplicity, usecase.ModerateReview should verify admin privileges.

	h.logger.Info("ModerateReview RPC called",
		zap.String("review_id", req.GetReviewId()),
		zap.String("admin_id", adminID), // Log the admin performing the action
		zap.String("new_status", req.GetNewStatus()))

	reviewID, err := primitive.ObjectIDFromHex(req.GetReviewId())
	if err != nil {
		h.logger.Warn("ModerateReview: Invalid review_id format", zap.String("review_id", req.GetReviewId()), zap.Error(err))
		return nil, status.Errorf(codes.InvalidArgument, "invalid review ID format")
	}

	newStatus := domain.ReviewStatus(req.GetNewStatus())
	// Validate newStatus if necessary (e.g., ensure it's one of the allowed enum values)
	if newStatus != domain.ReviewStatusApproved && newStatus != domain.ReviewStatusRejected && newStatus != domain.ReviewStatusHidden && newStatus != domain.ReviewStatusPending {
		return nil, status.Errorf(codes.InvalidArgument, "invalid new_status value")
	}

	review, err := h.usecase.ModerateReview(ctx, reviewID, adminID, newStatus, req.GetModerationComment())
	if err != nil {
		h.logger.Error("ModerateReview usecase failed", zap.Error(err), zap.String("review_id", req.GetReviewId()))
		if errors.Is(err, domain.ErrNotFound) {
			return nil, status.Errorf(codes.NotFound, "review not found")
		}
		if errors.Is(err, domain.ErrForbidden) { // If usecase checks admin role
			return nil, status.Errorf(codes.PermissionDenied, "admin privileges required")
		}
		if errors.Is(err, domain.ErrInvalidInput) {
			return nil, status.Errorf(codes.InvalidArgument, err.Error())
		}
		return nil, status.Errorf(codes.Internal, "failed to moderate review: %v", err)
	}

	return toProtoReview(review), nil
}
