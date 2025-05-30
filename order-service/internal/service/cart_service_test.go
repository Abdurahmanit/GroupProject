package service

import (
	"context"
	"errors"

	"testing"
	"time"

	listingpb "github.com/Abdurahmanit/GroupProject/listing-service/genproto/listing_service" // Импортируем для cartpb.CartProto
	"github.com/Abdurahmanit/GroupProject/order-service/internal/domain/entity"
	"github.com/Abdurahmanit/GroupProject/order-service/internal/platform/logger" // Импортируем для logger.Logger
	"github.com/Abdurahmanit/GroupProject/order-service/internal/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc"
)

type MockCartRepository struct {
	mock.Mock
}

func (m *MockCartRepository) GetByUserID(ctx context.Context, userID string) (*entity.Cart, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Cart), args.Error(1)
}

func (m *MockCartRepository) Save(ctx context.Context, cart *entity.Cart, ttl time.Duration) error {
	args := m.Called(ctx, cart, ttl)
	return args.Error(0)
}

func (m *MockCartRepository) DeleteByUserID(ctx context.Context, userID string) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

type MockProductDetailCache struct {
	mock.Mock
}

func (m *MockProductDetailCache) Get(ctx context.Context, productID string) (*listingpb.ListingResponse, error) {
	args := m.Called(ctx, productID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*listingpb.ListingResponse), args.Error(1)
}

func (m *MockProductDetailCache) Set(ctx context.Context, productID string, productDetails *listingpb.ListingResponse, ttl time.Duration) error {
	args := m.Called(ctx, productID, productDetails, ttl)
	return args.Error(0)
}

func (m *MockProductDetailCache) Delete(ctx context.Context, productID string) error {
	args := m.Called(ctx, productID)
	return args.Error(0)
}

type MockListingServiceClient struct {
	mock.Mock
}

func convertGRPCOptsToInterfaceSlice(opts []grpc.CallOption) []interface{} {
	s := make([]interface{}, len(opts))
	for i, v := range opts {
		s[i] = v
	}
	return s
}

func (m *MockListingServiceClient) GetListingByID(ctx context.Context, in *listingpb.GetListingRequest, opts ...grpc.CallOption) (*listingpb.ListingResponse, error) {
	allArgs := append([]interface{}{ctx, in}, convertGRPCOptsToInterfaceSlice(opts)...)
	args := m.Called(allArgs...)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*listingpb.ListingResponse), args.Error(1)
}

func (m *MockListingServiceClient) CreateListing(ctx context.Context, in *listingpb.CreateListingRequest, opts ...grpc.CallOption) (*listingpb.ListingResponse, error) {
	panic("CreateListing not implemented in mock")
}
func (m *MockListingServiceClient) UpdateListing(ctx context.Context, in *listingpb.UpdateListingRequest, opts ...grpc.CallOption) (*listingpb.ListingResponse, error) {
	panic("UpdateListing not implemented in mock")
}
func (m *MockListingServiceClient) DeleteListing(ctx context.Context, in *listingpb.DeleteListingRequest, opts ...grpc.CallOption) (*listingpb.Empty, error) {
	panic("DeleteListing not implemented in mock")
}
func (m *MockListingServiceClient) SearchListings(ctx context.Context, in *listingpb.SearchListingsRequest, opts ...grpc.CallOption) (*listingpb.SearchListingsResponse, error) {
	panic("SearchListings not implemented in mock")
}
func (m *MockListingServiceClient) UploadPhoto(ctx context.Context, in *listingpb.UploadPhotoRequest, opts ...grpc.CallOption) (*listingpb.UploadPhotoResponse, error) {
	panic("UploadPhoto not implemented in mock")
}
func (m *MockListingServiceClient) GetListingStatus(ctx context.Context, in *listingpb.GetListingRequest, opts ...grpc.CallOption) (*listingpb.ListingStatusResponse, error) {
	panic("GetListingStatus not implemented in mock")
}
func (m *MockListingServiceClient) AddFavorite(ctx context.Context, in *listingpb.AddFavoriteRequest, opts ...grpc.CallOption) (*listingpb.Empty, error) {
	panic("AddFavorite not implemented in mock")
}
func (m *MockListingServiceClient) RemoveFavorite(ctx context.Context, in *listingpb.RemoveFavoriteRequest, opts ...grpc.CallOption) (*listingpb.Empty, error) {
	panic("RemoveFavorite not implemented in mock")
}
func (m *MockListingServiceClient) GetFavorites(ctx context.Context, in *listingpb.GetFavoritesRequest, opts ...grpc.CallOption) (*listingpb.GetFavoritesResponse, error) {
	panic("GetFavorites not implemented in mock")
}
func (m *MockListingServiceClient) GetPhotoURLs(ctx context.Context, in *listingpb.GetListingRequest, opts ...grpc.CallOption) (*listingpb.PhotoURLsResponse, error) {
	panic("GetPhotoURLs not implemented in mock")
}
func (m *MockListingServiceClient) UpdateListingStatus(ctx context.Context, in *listingpb.UpdateListingStatusRequest, opts ...grpc.CallOption) (*listingpb.ListingResponse, error) {
	panic("UpdateListingStatus not implemented in mock")
}

type NoOpLogger struct{}

func (l *NoOpLogger) Init()                                        {}
func (l *NoOpLogger) Debug(args ...interface{})                    {}
func (l *NoOpLogger) Debugf(template string, args ...interface{})  {}
func (l *NoOpLogger) Info(args ...interface{})                     {}
func (l *NoOpLogger) Infof(template string, args ...interface{})   {}
func (l *NoOpLogger) Warn(args ...interface{})                     {}
func (l *NoOpLogger) Warnf(template string, args ...interface{})   {}
func (l *NoOpLogger) Error(args ...interface{})                    {}
func (l *NoOpLogger) Errorf(template string, args ...interface{})  {}
func (l *NoOpLogger) DPanic(args ...interface{})                   {}
func (l *NoOpLogger) DPanicf(template string, args ...interface{}) {}
func (l *NoOpLogger) Fatal(args ...interface{})                    {}
func (l *NoOpLogger) Fatalf(template string, args ...interface{})  {}
func (l *NoOpLogger) With(args ...interface{}) logger.Logger       { return l }

func NewNoOpLogger() logger.Logger {
	return &NoOpLogger{}
}

func TestCartService_AddItem_Success_NewItem(t *testing.T) {
	mockCartRepo := new(MockCartRepository)
	mockProductCache := new(MockProductDetailCache)
	mockListingClient := new(MockListingServiceClient)
	log := NewNoOpLogger()

	testUserID := "user1"
	testProductID := "product1"
	testQuantity := 2
	cartTTL := 24 * time.Hour
	productCacheTTL := 5 * time.Minute

	cfg := CartServiceConfig{
		CartTTL:         cartTTL,
		ProductCacheTTL: productCacheTTL,
	}
	cartSvc := NewCartService(mockCartRepo, mockProductCache, mockListingClient, log, cfg)

	emptyCart := entity.NewCart(testUserID)
	mockCartRepo.On("GetByUserID", mock.Anything, testUserID).Return(emptyCart, nil).Once()
	mockProductCache.On("Get", mock.Anything, testProductID).Return(nil, repository.ErrNotFound).Twice()
	mockListingClient.On("GetListingByID", mock.Anything, &listingpb.GetListingRequest{Id: testProductID}, mock.Anything).
		Return(&listingpb.ListingResponse{Id: testProductID, Title: "Test Product", Price: 10.0, Status: "ACTIVE"}, nil).Twice()
	mockProductCache.On("Set", mock.Anything, testProductID, mock.AnythingOfType("*listing_service.ListingResponse"), productCacheTTL).Return(nil).Twice()
	mockCartRepo.On("Save", mock.Anything, mock.MatchedBy(func(cart *entity.Cart) bool {
		return cart.UserID == testUserID && len(cart.Items) == 1 && cart.Items[0].ProductID == testProductID && cart.Items[0].Quantity == testQuantity
	}), cartTTL).Return(nil).Once()

	cartProto, err := cartSvc.AddItem(context.Background(), testUserID, testProductID, testQuantity)

	assert.NoError(t, err)
	assert.NotNil(t, cartProto)
	assert.Equal(t, testUserID, cartProto.UserId)
	assert.Len(t, cartProto.Items, 1)
	if len(cartProto.Items) == 1 {
		assert.Equal(t, testProductID, cartProto.Items[0].ProductId)
		assert.Equal(t, int32(testQuantity), cartProto.Items[0].Quantity)
		assert.Equal(t, "Test Product", cartProto.Items[0].ProductName)
		assert.Equal(t, 10.0, cartProto.Items[0].PricePerUnit)
		assert.Equal(t, 20.0, cartProto.Items[0].TotalPrice)
	}
	assert.Equal(t, 20.0, cartProto.TotalAmount)

	mockCartRepo.AssertExpectations(t)
	mockProductCache.AssertExpectations(t)
	mockListingClient.AssertExpectations(t)
}

func TestCartService_AddItem_Success_ExistingItem(t *testing.T) {
	mockCartRepo := new(MockCartRepository)
	mockProductCache := new(MockProductDetailCache)
	mockListingClient := new(MockListingServiceClient)
	log := NewNoOpLogger()

	testUserID := "user1"
	testProductID := "product1"
	initialQuantity := 1
	addQuantity := 2
	expectedTotalQuantity := initialQuantity + addQuantity
	cartTTL := 24 * time.Hour
	productCacheTTL := 5 * time.Minute

	cfg := CartServiceConfig{CartTTL: cartTTL, ProductCacheTTL: productCacheTTL}
	cartSvc := NewCartService(mockCartRepo, mockProductCache, mockListingClient, log, cfg)

	existingCart := entity.NewCart(testUserID)
	_ = existingCart.AddItem(testProductID, initialQuantity)

	mockCartRepo.On("GetByUserID", mock.Anything, testUserID).Return(existingCart, nil).Once()
	mockProductCache.On("Get", mock.Anything, testProductID).Return(nil, repository.ErrNotFound).Once()
	mockListingClient.On("GetListingByID", mock.Anything, &listingpb.GetListingRequest{Id: testProductID}, mock.Anything).
		Return(&listingpb.ListingResponse{Id: testProductID, Title: "Test Product", Price: 10.0, Status: "ACTIVE"}, nil).Once()
	mockProductCache.On("Set", mock.Anything, testProductID, mock.AnythingOfType("*listing_service.ListingResponse"), productCacheTTL).Return(nil).Once()
	mockCartRepo.On("Save", mock.Anything, mock.MatchedBy(func(cart *entity.Cart) bool {
		return cart.UserID == testUserID && len(cart.Items) == 1 && cart.Items[0].ProductID == testProductID && cart.Items[0].Quantity == expectedTotalQuantity
	}), cartTTL).Return(nil).Once()
	mockProductCache.On("Get", mock.Anything, testProductID).Return(&listingpb.ListingResponse{Id: testProductID, Title: "Test Product", Price: 10.0, Status: "ACTIVE"}, nil).Once()

	cartProto, err := cartSvc.AddItem(context.Background(), testUserID, testProductID, addQuantity)

	assert.NoError(t, err)
	assert.NotNil(t, cartProto)
	assert.Len(t, cartProto.Items, 1)
	if len(cartProto.Items) == 1 {
		assert.Equal(t, int32(expectedTotalQuantity), cartProto.Items[0].Quantity)
		assert.Equal(t, 10.0*float64(expectedTotalQuantity), cartProto.Items[0].TotalPrice)
	}
	assert.Equal(t, 10.0*float64(expectedTotalQuantity), cartProto.TotalAmount)

	mockCartRepo.AssertExpectations(t)
	mockProductCache.AssertExpectations(t)
	mockListingClient.AssertExpectations(t)
}

func TestCartService_AddItem_Fail_ListingServiceError(t *testing.T) {
	mockCartRepo := new(MockCartRepository)
	mockProductCache := new(MockProductDetailCache)
	mockListingClient := new(MockListingServiceClient)
	log := NewNoOpLogger()

	testUserID := "user1"
	testProductID := "product1"
	cartTTL := 24 * time.Hour
	productCacheTTL := 5 * time.Minute

	cfg := CartServiceConfig{CartTTL: cartTTL, ProductCacheTTL: productCacheTTL}
	cartSvc := NewCartService(mockCartRepo, mockProductCache, mockListingClient, log, cfg)

	emptyCart := entity.NewCart(testUserID)
	mockCartRepo.On("GetByUserID", mock.Anything, testUserID).Return(emptyCart, nil).Once()
	mockProductCache.On("Get", mock.Anything, testProductID).Return(nil, repository.ErrNotFound).Once()
	mockListingClient.On("GetListingByID", mock.Anything, &listingpb.GetListingRequest{Id: testProductID}, mock.Anything).
		Return(nil, errors.New("listing service unavailable")).Once()

	cartProto, err := cartSvc.AddItem(context.Background(), testUserID, testProductID, 1)

	assert.Error(t, err)
	assert.Nil(t, cartProto)
	assert.Contains(t, err.Error(), "product product1 not found or service unavailable")

	mockCartRepo.AssertExpectations(t)
	mockProductCache.AssertExpectations(t)
	mockListingClient.AssertExpectations(t)
}

func TestCartService_AddItem_Fail_ProductNotActive(t *testing.T) {
	mockCartRepo := new(MockCartRepository)
	mockProductCache := new(MockProductDetailCache)
	mockListingClient := new(MockListingServiceClient)
	log := NewNoOpLogger()

	testUserID := "user1"
	testProductID := "product1"
	cartTTL := 24 * time.Hour
	productCacheTTL := 5 * time.Minute

	cfg := CartServiceConfig{CartTTL: cartTTL, ProductCacheTTL: productCacheTTL}
	cartSvc := NewCartService(mockCartRepo, mockProductCache, mockListingClient, log, cfg)

	emptyCart := entity.NewCart(testUserID)
	mockCartRepo.On("GetByUserID", mock.Anything, testUserID).Return(emptyCart, nil).Once()
	mockProductCache.On("Get", mock.Anything, testProductID).Return(nil, repository.ErrNotFound).Once()
	mockListingClient.On("GetListingByID", mock.Anything, &listingpb.GetListingRequest{Id: testProductID}, mock.Anything).
		Return(&listingpb.ListingResponse{Id: testProductID, Title: "Inactive Product", Price: 10.0, Status: "INACTIVE"}, nil).Once()

	cartProto, err := cartSvc.AddItem(context.Background(), testUserID, testProductID, 1)

	assert.Error(t, err)
	assert.Nil(t, cartProto)
	assert.Contains(t, err.Error(), "product Inactive Product is not available for purchase")

	mockCartRepo.AssertExpectations(t)
	mockProductCache.AssertExpectations(t)
	mockListingClient.AssertExpectations(t)
}
