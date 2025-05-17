package grpc

import (
	"context"
	"log"

	"github.com/Abdurahmanit/GroupProject/listing-service/internal/listing/domain"
	"github.com/Abdurahmanit/GroupProject/listing-service/internal/listing/usecase"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	pb "github.com/Abdurahmanit/GroupProject/listing-service/genproto/listing_service"
)

type Handler struct {
	pb.UnimplementedListingServiceServer
	listingUsecase  *usecase.ListingUsecase
	photoUsecase    *usecase.PhotoUsecase
	favoriteUsecase *usecase.FavoriteUsecase
	natsPublisher   *nats.Publisher
}

func NewHandler(listingRepo domain.ListingRepository, favoriteRepo domain.FavoriteRepository, storage usecase.Storage, natsPublisher *nats.Publisher) *Handler {
	return &Handler{
		listingUsecase:  usecase.NewListingUsecase(listingRepo),
		photoUsecase:    usecase.NewPhotoUsecase(storage, listingRepo),
		favoriteUsecase: usecase.NewFavoriteUsecase(favoriteRepo),
		natsPublisher:   natsPublisher,
	}
}

func (h *Handler) CreateListing(ctx context.Context, req *pb.CreateListingRequest) (*pb.ListingResponse, error) {
	listing, err := h.listingUsecase.CreateListing(ctx, req.Title, req.Description, req.Price)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create listing: %v", err)
	}
	h.natsPublisher.Publish(ctx, "listing.created", map[string]string{"id": listing.ID})
	return &pb.ListingResponse{
		Id:          listing.ID,
		Title:       listing.Title,
		Description: listing.Description,
		Price:       listing.Price,
		Status:      string(listing.Status),
		Photos:      listing.Photos,
		CreatedAt:   listing.CreatedAt.String(),
		UpdatedAt:   listing.UpdatedAt.String(),
	}, nil
}

func (h *Handler) UpdateListing(ctx context.Context, req *pb.UpdateListingRequest) (*pb.ListingResponse, error) {
	listing, err := h.listingUsecase.UpdateListing(ctx, req.Id, req.Title, req.Description, req.Price, domain.ListingStatus(req.Status))
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to update listing: %v", err)
	}
	h.natsPublisher.Publish(ctx, "listing.updated", map[string]string{"id": listing.ID})
	return &pb.ListingResponse{
		Id:          listing.ID,
		Title:       listing.Title,
		Description: listing.Description,
		Price:       listing.Price,
		Status:      string(listing.Status),
		Photos:      listing.Photos,
		CreatedAt:   listing.CreatedAt.String(),
		UpdatedAt:   listing.UpdatedAt.String(),
	}, nil
}

func (h *Handler) DeleteListing(ctx context.Context, req *pb.DeleteListingRequest) (*pb.Empty, error) {
	err := h.listingUsecase.DeleteListing(ctx, req.Id)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to delete listing: %v", err)
	}
	h.natsPublisher.Publish(ctx, "listing.deleted", map[string]string{"id": req.Id})
	return &pb.Empty{}, nil
}

func (h *Handler) GetListingByID(ctx context.Context, req *pb.GetListingRequest) (*pb.ListingResponse, error) {
	listing, err := h.listingUsecase.GetListingByID(ctx, req.Id)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "listing not found: %v", err)
	}
	return &pb.ListingResponse{
		Id:          listing.ID,
		Title:       listing.Title,
		Description: listing.Description,
		Price:       listing.Price,
		Status:      string(listing.Status),
		Photos:      listing.Photos,
		CreatedAt:   listing.CreatedAt.String(),
		UpdatedAt:   listing.UpdatedAt.String(),
	}, nil
}

func (h *Handler) SearchListings(ctx context.Context, req *pb.SearchListingsRequest) (*pb.SearchListingsResponse, error) {
	listings, err := h.listingUsecase.SearchListings(ctx, domain.Filter{
		Query:    req.Query,
		MinPrice: req.MinPrice,
		MaxPrice: req.MaxPrice,
		Status:   domain.ListingStatus(req.Status),
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to search listings: %v", err)
	}
	var responses []*pb.ListingResponse
	for _, l := range listings {
		responses = append(responses, &pb.ListingResponse{
			Id:          l.ID,
			Title:       l.Title,
			Description: l.Description,
			Price:       l.Price,
			Status:      string(l.Status),
			Photos:      l.Photos,
			CreatedAt:   l.CreatedAt.String(),
			UpdatedAt:   l.UpdatedAt.String(),
		})
	}
	return &pb.SearchListingsResponse{Listings: responses}, nil
}

func (h *Handler) UploadPhoto(ctx context.Context, req *pb.UploadPhotoRequest) (*pb.UploadPhotoResponse, error) {
	url, err := h.photoUsecase.UploadPhoto(ctx, req.ListingId, req.FileName, req.Data)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to upload photo: %v", err)
	}
	return &pb.UploadPhotoResponse{Url: url}, nil
}

func (h *Handler) GetListingStatus(ctx context.Context, req *pb.GetListingRequest) (*pb.ListingStatusResponse, error) {
	listing, err := h.listingUsecase.GetListingByID(ctx, req.Id)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "listing not found: %v", err)
	}
	return &pb.ListingStatusResponse{Status: string(listing.Status)}, nil
}

func (h *Handler) AddFavorite(ctx context.Context, req *pb.AddFavoriteRequest) (*pb.Empty, error) {
	err := h.favoriteUsecase.AddFavorite(ctx, req.UserId, req.ListingId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to add favorite: %v", err)
	}
	return &pb.Empty{}, nil
}

func (h *Handler) RemoveFavorite(ctx context.Context, req *pb.RemoveFavoriteRequest) (*pb.Empty, error) {
	err := h.favoriteUsecase.RemoveFavorite(ctx, req.UserId, req.ListingId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to remove favorite: %v", err)
	}
	return &pb.Empty{}, nil
}

func (h *Handler) GetFavorites(ctx context.Context, req *pb.GetFavoritesRequest) (*pb.GetFavoritesResponse, error) {
	favorites, err := h.favoriteUsecase.GetFavorites(ctx, req.UserId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get favorites: %v", err)
	}
	var listingIDs []string
	for _, f := range favorites {
		listingIDs = append(listingIDs, f.ListingID)
	}
	return &pb.GetFavoritesResponse{ListingIds: listingIDs}, nil
}

func (h *Handler) GetPhotoURLs(ctx context.Context, req *pb.GetListingRequest) (*pb.PhotoURLsResponse, error) {
	listing, err := h.listingUsecase.GetListingByID(ctx, req.Id)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "listing not found: %v", err)
	}
	return &pb.PhotoURLsResponse{Urls: listing.Photos}, nil
}

func (h *Handler) UpdateListingStatus(ctx context.Context, req *pb.UpdateListingStatusRequest) (*pb.ListingResponse, error) {
	listing, err := h.listingUsecase.UpdateListing(ctx, req.Id, "", "", 0, domain.ListingStatus(req.Status))
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to update status: %v", err)
	}
	h.natsPublisher.Publish(ctx, "listing.status.updated", map[string]string{"id": listing.ID, "status": string(listing.Status)})
	return &pb.ListingResponse{
		Id:          listing.ID,
		Title:       listing.Title,
		Description: listing.Description,
		Price:       listing.Price,
		Status:      string(listing.Status),
		Photos:      listing.Photos,
		CreatedAt:   listing.CreatedAt.String(),
		UpdatedAt:   listing.UpdatedAt.String(),
	}, nil
}