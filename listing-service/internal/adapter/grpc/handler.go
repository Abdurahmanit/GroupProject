package grpc

import (
	"context"
	"fmt" // Для fmt.Errorf

	"github.com/Abdurahmanit/GroupProject/listing-service/internal/adapter/grpc/middleware" // Для middleware.UserIDKey
	"github.com/Abdurahmanit/GroupProject/listing-service/internal/adapter/messaging/nats"
	"github.com/Abdurahmanit/GroupProject/listing-service/internal/adapter/repository/cache"
	"github.com/Abdurahmanit/GroupProject/listing-service/internal/listing/domain"
	"github.com/Abdurahmanit/GroupProject/listing-service/internal/listing/usecase"
	pb "github.com/Abdurahmanit/GroupProject/listing-service/genproto/listing_service"
	"github.com/Abdurahmanit/GroupProject/listing-service/internal/platform/logger" // Твой логгер
	"github.com/redis/go-redis/v9"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	oteltrace "go.opentelemetry.io/otel/trace"
)

var tracer = otel.Tracer("listing-service/grpc-handler")

type Handler struct {
	pb.UnimplementedListingServiceServer
	listingUsecase  *usecase.ListingUsecase
	photoUsecase    *usecase.PhotoUsecase
	favoriteUsecase *usecase.FavoriteUsecase
	natsPublisher   *nats.Publisher
	cache           *cache.ListingCache
	logger          *logger.Logger
}

func NewHandler(
	listingRepo domain.ListingRepository,
	favoriteRepo domain.FavoriteRepository,
	storage domain.Storage,
	natsPublisher *nats.Publisher,
	cache *cache.ListingCache,
	log *logger.Logger,
) *Handler {
	listingUc := usecase.NewListingUsecase(listingRepo, log) // Передаем логгер в usecase
	photoUc := usecase.NewPhotoUsecase(storage, listingRepo, log)
	favoriteUc := usecase.NewFavoriteUsecase(favoriteRepo, log)

	return &Handler{
		listingUsecase:  listingUc,
		photoUsecase:    photoUc,
		favoriteUsecase: favoriteUc,
		natsPublisher:   natsPublisher,
		cache:           cache,
		logger:          log,
	}
}

func toProtoListingResponse(listing *domain.Listing) *pb.ListingResponse {
	if listing == nil {
		return nil
	}
	return &pb.ListingResponse{
		Id:          listing.ID,
		UserId:      listing.UserID,
		CategoryId:  listing.CategoryID,
		Title:       listing.Title,
		Description: listing.Description,
		Price:       listing.Price,
		Status:      string(listing.Status),
		Photos:      listing.Photos,
		CreatedAt:   timestamppb.New(listing.CreatedAt),
		UpdatedAt:   timestamppb.New(listing.UpdatedAt),
	}
}

// getUserIDFromContext извлекает UserID, установленный AuthInterceptor'ом.
func getUserIDFromContext(ctx context.Context, logger *logger.Logger, methodNameForLog string) (string, error) {
	authenticatedUserID, ok := ctx.Value(middleware.UserIDKey).(string)
	if !ok || authenticatedUserID == "" {
		logger.Error(methodNameForLog+": UserID not found in context or is empty",
			"detail", "This usually means the AuthInterceptor did not run or failed for a protected route, or the route was incorrectly marked public.")
		return "", status.Errorf(codes.Unauthenticated, "user authentication required and UserID missing from context")
	}
	logger.Debug(methodNameForLog+": Successfully extracted authenticated UserID from context", "auth_user_id", authenticatedUserID)
	return authenticatedUserID, nil
}

// ---- Listing Management Methods ----

func (h *Handler) CreateListing(ctx context.Context, req *pb.CreateListingRequest) (*pb.ListingResponse, error) {
	authenticatedUserID, err := getUserIDFromContext(ctx, h.logger, "CreateListing")
	if err != nil {
		return nil, err
	}

	// Важно: Убеждаемся, что пользователь создает объявление от своего имени.
	// Поле UserId в запросе должно совпадать с ID из токена.
	if req.GetUserId() == "" { // Если API Gateway не заполнил req.UserId
	    // req.UserId = authenticatedUserID // Можно установить его здесь для usecase, если он этого ожидает
	    h.logger.Info("CreateListing: req.UserId was empty, using authenticatedUserID from token for usecase call.", "auth_user_id", authenticatedUserID)
	} else if req.GetUserId() != authenticatedUserID {
		h.logger.Warn("CreateListing: UserID in request body does not match authenticated UserID from token.",
			"req_user_id", req.GetUserId(), "auth_user_id", authenticatedUserID)
		return nil, status.Errorf(codes.PermissionDenied, "cannot create listing for another user (user_id mismatch)")
	}
	// Далее в usecase передаем authenticatedUserID как источник правды.

	ctx, span := tracer.Start(ctx, "Handler.CreateListing", oteltrace.WithAttributes(
		attribute.String("authenticated_user_id", authenticatedUserID),
		attribute.String("req_user_id", req.GetUserId()), // Логируем для отладки
		attribute.String("title", req.GetTitle()),
		attribute.String("category_id", req.GetCategoryId()),
	))
	defer span.End()

	listing, err := h.listingUsecase.CreateListing(ctx, authenticatedUserID, req.GetCategoryId(), req.GetTitle(), req.GetDescription(), req.GetPrice())
	if err != nil {
		h.logger.Error("CreateListing: usecase failed", "user_id", authenticatedUserID, "title", req.GetTitle(), "error", err.Error())
		span.RecordError(err)
		return nil, status.Errorf(codes.Internal, "failed to create listing: %v", err)
	}
	span.SetAttributes(attribute.String("created_listing_id", listing.ID))

	if errCache := h.cache.SetListing(ctx, listing); errCache != nil {
		h.logger.Warn("CreateListing: SetListing to cache failed", "listing_id", listing.ID, "error", errCache.Error())
	} else {
		h.logger.Info("CreateListing: SetListing to cache successful", "listing_id", listing.ID)
	}

	_, natsSpan := tracer.Start(ctx, "NATS.Publish.listing.created")
	h.natsPublisher.Publish(ctx, "listing.created", map[string]string{"id": listing.ID, "user_id": listing.UserID, "category_id": listing.CategoryID})
	natsSpan.End()

	h.logger.Info("CreateListing: successful", "listing_id", listing.ID, "user_id", listing.UserID)
	return toProtoListingResponse(listing), nil
}

func (h *Handler) UpdateListing(ctx context.Context, req *pb.UpdateListingRequest) (*pb.ListingResponse, error) {
	authenticatedUserID, err := getUserIDFromContext(ctx, h.logger, "UpdateListing")
	if err != nil {
		return nil, err
	}
	if req.GetUserId() == "" {
	    h.logger.Info("UpdateListing: req.UserId was empty, usecase will rely on authenticatedUserID for authorization checks.", "auth_user_id", authenticatedUserID)
	} else if req.GetUserId() != authenticatedUserID {
		h.logger.Warn("UpdateListing: UserID in request body does not match authenticated UserID from token.",
			"req_user_id", req.GetUserId(), "auth_user_id", authenticatedUserID, "listing_id_to_update", req.GetId())
		return nil, status.Errorf(codes.PermissionDenied, "cannot update listing for another user (user_id mismatch)")
	}

	ctx, span := tracer.Start(ctx, "Handler.UpdateListing", oteltrace.WithAttributes(
		attribute.String("listing_id", req.GetId()),
		attribute.String("authenticated_user_id", authenticatedUserID),
		attribute.String("req_user_id", req.GetUserId()),
	))
	defer span.End()

	// Usecase должен проверить, что authenticatedUserID является владельцем объявления req.GetId()
	listing, err := h.listingUsecase.UpdateListing(ctx, req.GetId(), authenticatedUserID, req.GetCategoryId(), req.GetTitle(), req.GetDescription(), req.GetPrice(), domain.ListingStatus(req.GetStatus()))
	if err != nil {
		h.logger.Error("UpdateListing: usecase failed", "listing_id", req.GetId(), "user_id", authenticatedUserID, "error", err.Error())
		span.RecordError(err)
		// Здесь можно добавить проверку на domain.ErrForbidden, если usecase ее возвращает
		// if errors.Is(err, domain.ErrForbidden) { return nil, status.Errorf(codes.PermissionDenied, "user not authorized to update this listing")}
		return nil, status.Errorf(codes.Internal, "failed to update listing: %v", err)
	}

	if errCache := h.cache.SetListing(ctx, listing); errCache != nil {
		h.logger.Warn("UpdateListing: SetListing to cache failed", "listing_id", listing.ID, "error", errCache.Error())
	} else {
		h.logger.Info("UpdateListing: SetListing to cache successful", "listing_id", listing.ID)
	}

	_, natsSpan := tracer.Start(ctx, "NATS.Publish.listing.updated")
	h.natsPublisher.Publish(ctx, "listing.updated", map[string]string{"id": listing.ID, "user_id": listing.UserID})
	natsSpan.End()

	h.logger.Info("UpdateListing: successful", "listing_id", listing.ID, "user_id", listing.UserID)
	return toProtoListingResponse(listing), nil
}

func (h *Handler) DeleteListing(ctx context.Context, req *pb.DeleteListingRequest) (*pb.Empty, error) {
	authenticatedUserID, err := getUserIDFromContext(ctx, h.logger, "DeleteListing")
	if err != nil {
		return nil, err
	}
	if req.GetUserId() == "" {
	     h.logger.Info("DeleteListing: req.UserId was empty, usecase will rely on authenticatedUserID for authorization checks.", "auth_user_id", authenticatedUserID)
	} else if req.GetUserId() != authenticatedUserID {
		h.logger.Warn("DeleteListing: UserID in request body does not match authenticated UserID from token.",
			"req_user_id", req.GetUserId(), "auth_user_id", authenticatedUserID, "listing_id_to_delete", req.GetId())
		return nil, status.Errorf(codes.PermissionDenied, "cannot delete listing for another user (user_id mismatch)")
	}

	ctx, span := tracer.Start(ctx, "Handler.DeleteListing", oteltrace.WithAttributes(
		attribute.String("listing_id", req.GetId()),
		attribute.String("authenticated_user_id", authenticatedUserID),
		attribute.String("req_user_id", req.GetUserId()),
	))
	defer span.End()

	// Usecase должен проверить, что authenticatedUserID является владельцем объявления req.GetId()
	err = h.listingUsecase.DeleteListing(ctx, req.GetId(), authenticatedUserID)
	if err != nil {
		h.logger.Error("DeleteListing: usecase failed", "listing_id", req.GetId(), "user_id", authenticatedUserID, "error", err.Error())
		span.RecordError(err)
		return nil, status.Errorf(codes.Internal, "failed to delete listing: %v", err)
	}

	if errCache := h.cache.DeleteListing(ctx, req.GetId()); errCache != nil {
		h.logger.Warn("DeleteListing: DeleteListing from cache failed", "listing_id", req.GetId(), "error", errCache.Error())
	} else {
		h.logger.Info("DeleteListing: DeleteListing from cache successful", "listing_id", req.GetId)
	}

	_, natsSpan := tracer.Start(ctx, "NATS.Publish.listing.deleted")
	h.natsPublisher.Publish(ctx, "listing.deleted", map[string]string{"id": req.GetId(), "user_id": authenticatedUserID}) // Используем authenticatedUserID для NATS
	natsSpan.End()

	h.logger.Info("DeleteListing: successful", "listing_id", req.GetId(), "user_id", authenticatedUserID)
	return &pb.Empty{}, nil
}

func (h *Handler) UpdateListingStatus(ctx context.Context, req *pb.UpdateListingStatusRequest) (*pb.ListingResponse, error) {
	authenticatedUserID, err := getUserIDFromContext(ctx, h.logger, "UpdateListingStatus")
	if err != nil {
		return nil, err
	}
    if req.GetUserId() == "" {
	     h.logger.Info("UpdateListingStatus: req.UserId was empty, usecase will rely on authenticatedUserID for authorization checks.", "auth_user_id", authenticatedUserID)
	} else if req.GetUserId() != authenticatedUserID {
		h.logger.Warn("UpdateListingStatus: UserID in request body does not match authenticated UserID from token.",
			"req_user_id", req.GetUserId(), "auth_user_id", authenticatedUserID, "listing_id_to_update_status", req.GetId())
		return nil, status.Errorf(codes.PermissionDenied, "cannot update listing status for another user (user_id mismatch)")
	}

	ctx, span := tracer.Start(ctx, "Handler.UpdateListingStatus", oteltrace.WithAttributes(
		attribute.String("listing_id", req.GetId()),
		attribute.String("authenticated_user_id", authenticatedUserID),
		attribute.String("req_user_id", req.GetUserId()),
		attribute.String("new_status", req.GetStatus()),
	))
	defer span.End()

	// Usecase должен проверить, что authenticatedUserID является владельцем объявления req.GetId()
	listing, err := h.listingUsecase.UpdateListingStatus(ctx, req.GetId(), authenticatedUserID, domain.ListingStatus(req.GetStatus()))
	if err != nil {
		h.logger.Error("UpdateListingStatus: usecase failed", "listing_id", req.GetId(), "user_id", authenticatedUserID, "status", req.GetStatus(), "error", err.Error())
		span.RecordError(err)
		return nil, status.Errorf(codes.Internal, "failed to update listing status: %v", err)
	}

	if errCache := h.cache.SetListing(ctx, listing); errCache != nil {
		h.logger.Warn("UpdateListingStatus: SetListing to cache failed", "listing_id", listing.ID, "error", errCache.Error())
	} else {
		h.logger.Info("UpdateListingStatus: SetListing to cache successful", "listing_id", listing.ID)
	}

	_, natsSpan := tracer.Start(ctx, "NATS.Publish.listing.status.updated")
	h.natsPublisher.Publish(ctx, "listing.status.updated", map[string]string{"id": listing.ID, "status": string(listing.Status), "user_id": listing.UserID})
	natsSpan.End()

	h.logger.Info("UpdateListingStatus: successful", "listing_id", listing.ID, "new_status", string(listing.Status))
	return toProtoListingResponse(listing), nil
}

// ---- Photo Management Methods ----

func (h *Handler) UploadPhoto(ctx context.Context, req *pb.UploadPhotoRequest) (*pb.UploadPhotoResponse, error) {
	authenticatedUserID, err := getUserIDFromContext(ctx, h.logger, "UploadPhoto")
	if err != nil {
		return nil, err
	}
    if req.GetUserId() == "" {
	     h.logger.Info("UploadPhoto: req.UserId was empty, usecase will rely on authenticatedUserID for authorization checks.", "auth_user_id", authenticatedUserID)
	} else if req.GetUserId() != authenticatedUserID {
		h.logger.Warn("UploadPhoto: UserID in request body does not match authenticated UserID from token.",
			"req_user_id", req.GetUserId(), "auth_user_id", authenticatedUserID, "listing_id_for_photo", req.GetListingId())
		return nil, status.Errorf(codes.PermissionDenied, "cannot upload photo for another user's listing (user_id mismatch)")
	}

	ctx, span := tracer.Start(ctx, "Handler.UploadPhoto", oteltrace.WithAttributes(
		attribute.String("listing_id", req.GetListingId()),
		attribute.String("authenticated_user_id", authenticatedUserID),
		attribute.String("req_user_id", req.GetUserId()),
		attribute.String("file_name", req.GetFileName()),
	))
	defer span.End()

	// photoUsecase должен проверить, что authenticatedUserID является владельцем объявления req.GetListingId()
	url, err := h.photoUsecase.UploadPhoto(ctx, req.GetListingId(), authenticatedUserID, req.GetFileName(), req.GetData())
	if err != nil {
		h.logger.Error("UploadPhoto: usecase failed", "listing_id", req.GetListingId(), "user_id", authenticatedUserID, "error", err.Error())
		span.RecordError(err)
		return nil, status.Errorf(codes.Internal, "failed to upload photo: %v", err)
	}
	span.SetAttributes(attribute.String("uploaded_photo_url", url))

	if errCache := h.cache.DeleteListing(ctx, req.GetListingId()); errCache != nil { // Инвалидация кэша
		h.logger.Warn("UploadPhoto: DeleteListing from cache failed after photo upload", "listing_id", req.GetListingId(), "error", errCache.Error())
	} else {
		h.logger.Info("UploadPhoto: DeleteListing from cache successful after photo upload", "listing_id", req.GetListingId())
	}

	_, natsSpan := tracer.Start(ctx, "NATS.Publish.listing.photo.uploaded")
	h.natsPublisher.Publish(ctx, "listing.photo.uploaded", map[string]string{"id": req.GetListingId(), "photo_url": url, "user_id": authenticatedUserID})
	natsSpan.End()

	h.logger.Info("UploadPhoto: successful", "listing_id", req.GetListingId(), "url", url)
	return &pb.UploadPhotoResponse{PhotoUrl: url}, nil
}

// ---- Public Read Methods ----

func (h *Handler) GetListingByID(ctx context.Context, req *pb.GetListingRequest) (*pb.ListingResponse, error) {
	// Этот метод предполагается публичным, AuthInterceptor его пропускает.
	// UserID из контекста для авторизации здесь не извлекается.
	ctx, span := tracer.Start(ctx, "Handler.GetListingByID", oteltrace.WithAttributes(
		attribute.String("listing_id", req.GetId()),
	))
	defer span.End()

	cachedListing, errCache := h.cache.GetListing(ctx, req.GetId())
	if errCache == nil && cachedListing != nil {
		h.logger.Info("GetListingByID: Cache HIT", "listing_id", req.GetId())
		span.SetAttributes(attribute.Bool("cache_hit", true))
		return toProtoListingResponse(cachedListing), nil
	}

	span.SetAttributes(attribute.Bool("cache_hit", false))
	if errCache != nil && errCache != redis.Nil {
		h.logger.Warn("GetListingByID: GetListing from cache failed", "listing_id", req.GetId(), "error", errCache.Error())
		span.RecordError(errCache)
	} else if errCache == redis.Nil {
		h.logger.Info("GetListingByID: Cache MISS", "listing_id", req.GetId())
	}

	listing, err := h.listingUsecase.GetListingByID(ctx, req.GetId())
	if err != nil {
		h.logger.Warn("GetListingByID: usecase failed", "listing_id", req.GetId(), "error", err.Error()) // Warn, т.к. NotFound ожидаемо
		span.RecordError(err)
		return nil, status.Errorf(codes.NotFound, "listing not found: %v", err)
	}
	if listing == nil {
		h.logger.Warn("GetListingByID: usecase returned nil without error", "listing_id", req.GetId())
		span.SetAttributes(attribute.Bool("usecase_found", false))
		return nil, status.Errorf(codes.NotFound, "listing not found: %s", req.GetId())
	}
	span.SetAttributes(attribute.Bool("usecase_found", true))

	if errSetCache := h.cache.SetListing(ctx, listing); errSetCache != nil {
		h.logger.Warn("GetListingByID: SetListing to cache after fetch failed", "listing_id", listing.ID, "error", errSetCache.Error())
	} else {
		h.logger.Info("GetListingByID: SetListing to cache after fetch successful", "listing_id", listing.ID)
	}

	h.logger.Info("GetListingByID: Fetched from usecase", "listing_id", listing.ID)
	return toProtoListingResponse(listing), nil
}

func (h *Handler) SearchListings(ctx context.Context, req *pb.SearchListingsRequest) (*pb.SearchListingsResponse, error) {
	// Этот метод публичный. req.GetUserId() здесь используется как фильтр, а не для аутентификации.
	ctx, span := tracer.Start(ctx, "Handler.SearchListings", oteltrace.WithAttributes(
		attribute.String("query", req.GetQuery()),
		attribute.Float64("min_price", req.GetMinPrice()),
		attribute.Float64("max_price", req.GetMaxPrice()),
		attribute.String("status", req.GetStatus()),
		attribute.String("category_id", req.GetCategoryId()),
		attribute.String("filter_user_id", req.GetUserId()), // Для фильтрации по пользователю
		attribute.Int64("page", int64(req.GetPage())),
		attribute.Int64("limit", int64(req.GetLimit())),
		attribute.String("sort_by", req.GetSortBy()),
		attribute.String("sort_order", req.GetSortOrder()),
	))
	defer span.End()

	filter := domain.Filter{
		Query:      req.GetQuery(),
		MinPrice:   req.GetMinPrice(),
		MaxPrice:   req.GetMaxPrice(),
		Status:     domain.ListingStatus(req.GetStatus()),
		CategoryID: req.GetCategoryId(),
		UserID:     req.GetUserId(), // Передаем UserID из запроса как фильтр
		Page:       req.GetPage(),
		Limit:      req.GetLimit(),
		SortBy:     req.GetSortBy(),
		SortOrder:  req.GetSortOrder(),
	}

	listings, total, err := h.listingUsecase.SearchListings(ctx, filter)
	if err != nil {
		h.logger.Error("SearchListings: usecase failed", "filter", fmt.Sprintf("%+v", filter), "error", err.Error()) // %+v для полной структуры фильтра
		span.RecordError(err)
		return nil, status.Errorf(codes.Internal, "failed to search listings: %v", err)
	}
	span.SetAttributes(attribute.Int("search_results_count", len(listings)), attribute.Int64("search_total_count", total))

	var responses []*pb.ListingResponse
	for _, l := range listings {
		responses = append(responses, toProtoListingResponse(l))
	}

	h.logger.Info("SearchListings: successful", "count", len(responses), "total", total)
	return &pb.SearchListingsResponse{
		Listings: responses,
		Total:    total,
		Page:     req.GetPage(),
		Limit:    req.GetLimit(),
	}, nil
}

func (h *Handler) GetListingStatus(ctx context.Context, req *pb.GetListingRequest) (*pb.ListingStatusResponse, error) {
	// Этот метод публичный, если GetListingByID публичный.
	ctx, span := tracer.Start(ctx, "Handler.GetListingStatus", oteltrace.WithAttributes(
		attribute.String("listing_id", req.GetId()),
	))
	defer span.End()

	listingResp, err := h.GetListingByID(ctx, req) // Используем уже кэширующий и публичный GetListingByID
	if err != nil {
		// GetListingByID уже логирует и возвращает ошибку
		return nil, err
	}
	if listingResp == nil {
		h.logger.Warn("GetListingStatus: GetListingByID returned nil response", "listing_id", req.GetId())
		// GetListingByID должен был вернуть NotFound, но на всякий случай
		return nil, status.Errorf(codes.NotFound, "listing not found for status check: %s", req.GetId())
	}

	h.logger.Info("GetListingStatus: successful", "listing_id", req.GetId(), "status", listingResp.Status)
	return &pb.ListingStatusResponse{
		ListingId: listingResp.Id, // Добавляем listing_id в ответ, как в proto
		Status:    listingResp.Status,
	}, nil
}

func (h *Handler) GetPhotoURLs(ctx context.Context, req *pb.GetListingRequest) (*pb.PhotoURLsResponse, error) {
	// Этот метод публичный, если GetListingByID публичный.
	ctx, span := tracer.Start(ctx, "Handler.GetPhotoURLs", oteltrace.WithAttributes(
		attribute.String("listing_id", req.GetId()),
	))
	defer span.End()

	listingResp, err := h.GetListingByID(ctx, req)
	if err != nil {
		return nil, err
	}
	if listingResp == nil {
		h.logger.Warn("GetPhotoURLs: GetListingByID returned nil response", "listing_id", req.GetId())
		return nil, status.Errorf(codes.NotFound, "listing not found for photo URLs: %s", req.GetId())
	}

	h.logger.Info("GetPhotoURLs: successful", "listing_id", req.GetId(), "photo_count", len(listingResp.Photos))
	return &pb.PhotoURLsResponse{
		ListingId: listingResp.Id, // Добавляем listing_id в ответ, как в proto
		Urls:      listingResp.Photos,
	}, nil
}


// ---- Favorite Management Methods ----
// Эти методы требуют аутентификации и проверки, что пользователь оперирует своим списком избранного.

func (h *Handler) AddFavorite(ctx context.Context, req *pb.AddFavoriteRequest) (*pb.Empty, error) {
	authenticatedUserID, err := getUserIDFromContext(ctx, h.logger, "AddFavorite")
	if err != nil {
		return nil, err
	}
	// Важно: Проверяем, что пользователь (из токена) совпадает с тем, для кого добавляется избранное (из запроса).
	if req.GetUserId() != authenticatedUserID {
		h.logger.Warn("AddFavorite: Attempt to add favorite for another user or UserID mismatch.",
			"req_user_id", req.GetUserId(), "auth_user_id", authenticatedUserID, "listing_id", req.GetListingId())
		return nil, status.Errorf(codes.PermissionDenied, "cannot add/manage favorites for another user")
	}

	ctx, span := tracer.Start(ctx, "Handler.AddFavorite", oteltrace.WithAttributes(
		attribute.String("user_id", authenticatedUserID), // Используем проверенный ID
		attribute.String("listing_id", req.GetListingId()),
	))
	defer span.End()

	err = h.favoriteUsecase.AddFavorite(ctx, authenticatedUserID, req.GetListingId()) // Передаем authenticatedUserID
	if err != nil {
		h.logger.Error("AddFavorite: usecase failed", "user_id", authenticatedUserID, "listing_id", req.GetListingId(), "error", err.Error())
		span.RecordError(err)
		return nil, status.Errorf(codes.Internal, "failed to add favorite: %v", err)
	}

	h.logger.Info("AddFavorite: successful", "user_id", authenticatedUserID, "listing_id", req.GetListingId())
	return &pb.Empty{}, nil
}

func (h *Handler) RemoveFavorite(ctx context.Context, req *pb.RemoveFavoriteRequest) (*pb.Empty, error) {
	authenticatedUserID, err := getUserIDFromContext(ctx, h.logger, "RemoveFavorite")
	if err != nil {
		return nil, err
	}
	if req.GetUserId() != authenticatedUserID {
		h.logger.Warn("RemoveFavorite: Attempt to remove favorite for another user or UserID mismatch.",
			"req_user_id", req.GetUserId(), "auth_user_id", authenticatedUserID, "listing_id", req.GetListingId())
		return nil, status.Errorf(codes.PermissionDenied, "cannot add/manage favorites for another user")
	}

	ctx, span := tracer.Start(ctx, "Handler.RemoveFavorite", oteltrace.WithAttributes(
		attribute.String("user_id", authenticatedUserID),
		attribute.String("listing_id", req.GetListingId()),
	))
	defer span.End()

	err = h.favoriteUsecase.RemoveFavorite(ctx, authenticatedUserID, req.GetListingId())
	if err != nil {
		h.logger.Error("RemoveFavorite: usecase failed", "user_id", authenticatedUserID, "listing_id", req.GetListingId(), "error", err.Error())
		span.RecordError(err)
		return nil, status.Errorf(codes.Internal, "failed to remove favorite: %v", err)
	}

	h.logger.Info("RemoveFavorite: successful", "user_id", authenticatedUserID, "listing_id", req.GetListingId())
	return &pb.Empty{}, nil
}

func (h *Handler) GetFavorites(ctx context.Context, req *pb.GetFavoritesRequest) (*pb.GetFavoritesResponse, error) {
	authenticatedUserID, err := getUserIDFromContext(ctx, h.logger, "GetFavorites")
	if err != nil {
		return nil, err
	}
	if req.GetUserId() != authenticatedUserID {
		h.logger.Warn("GetFavorites: Attempt to get favorites for another user or UserID mismatch.",
			"req_user_id", req.GetUserId(), "auth_user_id", authenticatedUserID)
		return nil, status.Errorf(codes.PermissionDenied, "cannot get favorites for another user")
	}

	ctx, span := tracer.Start(ctx, "Handler.GetFavorites", oteltrace.WithAttributes(
		attribute.String("user_id", authenticatedUserID),
	))
	defer span.End()

	favorites, err := h.favoriteUsecase.GetFavorites(ctx, authenticatedUserID) // domain.Favorite
	if err != nil {
		h.logger.Error("GetFavorites: usecase failed", "user_id", authenticatedUserID, "error", err.Error())
		span.RecordError(err)
		return nil, status.Errorf(codes.Internal, "failed to get favorites: %v", err)
	}

	var listingIDs []string
	if favorites != nil {
		for _, f := range favorites {
			listingIDs = append(listingIDs, f.ListingID)
		}
	}
	span.SetAttributes(attribute.Int("favorite_count", len(listingIDs)))

	h.logger.Info("GetFavorites: successful", "user_id", authenticatedUserID, "count", len(listingIDs))
	return &pb.GetFavoritesResponse{ListingIds: listingIDs}, nil
}