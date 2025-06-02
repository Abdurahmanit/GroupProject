package grpc

import (
	"context"
	"errors"

	pb "github.com/Abdurahmanit/GroupProject/review-service"
	"github.com/Abdurahmanit/GroupProject/review-service/internal/domain"
	"github.com/Abdurahmanit/GroupProject/review-service/internal/middleware"
	"github.com/Abdurahmanit/GroupProject/review-service/internal/platform/logger"
	"github.com/Abdurahmanit/GroupProject/review-service/internal/usecase"

	"go.mongodb.org/mongo-driver/bson/primitive"
	zap "go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type ReviewHandler struct {
	pb.UnimplementedReviewServiceServer
	usecase *usecase.ReviewUsecase
	logger  *logger.Logger
}

func NewReviewHandler(uc *usecase.ReviewUsecase, log *logger.Logger) *ReviewHandler {
	return &ReviewHandler{
		usecase: uc,
		logger:  log.Named("ReviewGRPCHandler"),
	}
}

func toProtoReview(review *domain.Review) *pb.Review {
	if review == nil {
		return nil
	}
	return &pb.Review{
		Id:                review.ID.Hex(),
		UserId:            review.UserID,
		ProductId:         review.ProductID,
		SellerId:          review.SellerID,
		Rating:            review.Rating,
		Comment:           review.Comment,
		Status:            string(review.Status),
		CreatedAt:         timestamppb.New(review.CreatedAt),
		UpdatedAt:         timestamppb.New(review.UpdatedAt),
		ModerationComment: review.ModerationComment,
	}
}

func (h *ReviewHandler) CreateReview(ctx context.Context, req *pb.CreateReviewRequest) (*pb.Review, error) {
	authenticatedUserID, ok := ctx.Value(middleware.UserIDKey).(string)
	if !ok || authenticatedUserID == "" {
		h.logger.Warn("CreateReview: UserID not found in context or is empty", zap.String("request_user_id", req.GetUserId()))
		return nil, status.Errorf(codes.Unauthenticated, "user authentication required")
	}

	if req.GetUserId() != "" && req.GetUserId() != authenticatedUserID {
		h.logger.Warn("CreateReview: Authenticated user attempting to create review for another user",
			zap.String("authenticated_user_id", authenticatedUserID),
			zap.String("request_author_id", req.GetUserId()))
		return nil, status.Errorf(codes.PermissionDenied, "cannot create review for another user")
	}

	authorID := authenticatedUserID

	h.logger.Info("CreateReview RPC called",
		zap.String("author_id", authorID),
		zap.String("product_id", req.GetProductId()),
		zap.Int32("rating", req.GetRating()))

	review, err := h.usecase.CreateReview(ctx, authorID, req.GetProductId(), req.GetSellerId(), req.GetComment(), req.GetRating())
	if err != nil {
		h.logger.Error("CreateReview usecase failed", zap.Error(err), zap.String("author_id", authorID))
		if errors.Is(err, domain.ErrReviewAlreadyExists) {
			return nil, status.Errorf(codes.AlreadyExists, "%s", err.Error())
		}
		if errors.Is(err, domain.ErrInvalidInput) {
			return nil, status.Errorf(codes.InvalidArgument, "%s", err.Error())
		}
		return nil, status.Errorf(codes.Internal, "failed to create review: %v", err)
	}

	h.logger.Info("Review created successfully", zap.String("review_id", review.ID.Hex()))
	return toProtoReview(review), nil
}

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

func (h *ReviewHandler) UpdateReview(ctx context.Context, req *pb.UpdateReviewRequest) (*pb.Review, error) {
	authenticatedUserID, ok := ctx.Value(middleware.UserIDKey).(string)
	if !ok || authenticatedUserID == "" {
		h.logger.Warn("UpdateReview: UserID not found in context")
		return nil, status.Errorf(codes.Unauthenticated, "user authentication required")
	}

	if req.GetUserId() != "" && req.GetUserId() != authenticatedUserID {
		h.logger.Warn("UpdateReview: Authenticated user ID does not match user_id in request",
			zap.String("authenticated_user_id", authenticatedUserID),
			zap.String("request_user_id", req.GetUserId()))
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
	if req.Rating != 0 {
		r := req.GetRating()
		ratingToUpdate = &r
	}
	var commentToUpdate *string
	if req.Comment != "" {
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
			return nil, status.Errorf(codes.InvalidArgument, "%s", err.Error())
		}
		return nil, status.Errorf(codes.Internal, "failed to update review: %v", err)
	}

	return toProtoReview(review), nil
}

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

func (h *ReviewHandler) ListReviewsByUser(ctx context.Context, req *pb.ListReviewsByUserRequest) (*pb.ListReviewsResponse, error) {
	authenticatedUserID, ok := ctx.Value(middleware.UserIDKey).(string)
	if !ok || authenticatedUserID == "" {
		h.logger.Warn("ListReviewsByUser: UserID not found in context")
		return nil, status.Errorf(codes.Unauthenticated, "user authentication required")
	}

	targetUserID := req.GetUserId()
	if targetUserID == "" {
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

func (h *ReviewHandler) GetProductAverageRating(ctx context.Context, req *pb.GetProductAverageRatingRequest) (*pb.ProductAverageRatingResponse, error) {
	h.logger.Info("GetProductAverageRating RPC called", zap.String("product_id", req.GetProductId()))
	if req.GetProductId() == "" {
		return nil, status.Errorf(codes.InvalidArgument, "product_id is required")
	}
	avg, count, err := h.usecase.GetProductAverageRating(ctx, req.GetProductId())
	if err != nil {
		h.logger.Error("GetProductAverageRating usecase failed", zap.Error(err), zap.String("product_id", req.GetProductId()))
		if errors.Is(err, domain.ErrInvalidInput) {
			return nil, status.Errorf(codes.InvalidArgument, "%s", err.Error())
		}
		return nil, status.Errorf(codes.Internal, "failed to get product average rating: %v", err)
	}
	return &pb.ProductAverageRatingResponse{
		ProductId:     req.GetProductId(),
		AverageRating: avg,
		ReviewCount:   count,
	}, nil
}

func (h *ReviewHandler) ModerateReview(ctx context.Context, req *pb.ModerateReviewRequest) (*pb.Review, error) {
	adminID, ok := ctx.Value(middleware.UserIDKey).(string)
	if !ok || adminID == "" {
		h.logger.Warn("ModerateReview: Admin UserID not found in context")
		return nil, status.Errorf(codes.Unauthenticated, "admin authentication required")
	}

	h.logger.Info("ModerateReview RPC called",
		zap.String("review_id", req.GetReviewId()),
		zap.String("admin_id", adminID),
		zap.String("new_status", req.GetNewStatus()))

	reviewID, err := primitive.ObjectIDFromHex(req.GetReviewId())
	if err != nil {
		h.logger.Warn("ModerateReview: Invalid review_id format", zap.String("review_id", req.GetReviewId()), zap.Error(err))
		return nil, status.Errorf(codes.InvalidArgument, "invalid review ID format")
	}

	newStatus := domain.ReviewStatus(req.GetNewStatus())
	if !newStatus.IsValid() {
		return nil, status.Errorf(codes.InvalidArgument, "invalid new_status value: %s", req.GetNewStatus())
	}

	review, err := h.usecase.ModerateReview(ctx, reviewID, adminID, newStatus, req.GetModerationComment())
	if err != nil {
		h.logger.Error("ModerateReview usecase failed", zap.Error(err), zap.String("review_id", req.GetReviewId()))
		if errors.Is(err, domain.ErrNotFound) {
			return nil, status.Errorf(codes.NotFound, "review not found")
		}
		if errors.Is(err, domain.ErrForbidden) {
			return nil, status.Errorf(codes.PermissionDenied, "admin privileges required")
		}
		if errors.Is(err, domain.ErrInvalidInput) {
			return nil, status.Errorf(codes.InvalidArgument, "%s", err.Error())
		}
		return nil, status.Errorf(codes.Internal, "failed to moderate review: %v", err)
	}

	return toProtoReview(review), nil
}
