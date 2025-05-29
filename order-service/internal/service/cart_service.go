package service

import (
	"context"
	"fmt"
	"time"

	listingpb "github.com/Abdurahmanit/GroupProject/listing-service/genproto/listing_service"
	"github.com/Abdurahmanit/GroupProject/order-service/internal/domain/entity"
	"github.com/Abdurahmanit/GroupProject/order-service/internal/platform/logger"
	"github.com/Abdurahmanit/GroupProject/order-service/internal/repository"
	cartpb "github.com/Abdurahmanit/GroupProject/order-service/proto/cart"
)

const (
	defaultCartTTL = 24 * time.Hour
)

type CartService interface {
	AddItem(ctx context.Context, userID, productID string, quantity int) (*cartpb.CartProto, error)
	UpdateItemQuantity(ctx context.Context, userID, productID string, newQuantity int) (*cartpb.CartProto, error)
	RemoveItem(ctx context.Context, userID, productID string) (*cartpb.CartProto, error)
	GetCart(ctx context.Context, userID string) (*cartpb.CartProto, error)
	ClearCart(ctx context.Context, userID string) error
}

type cartService struct {
	cartRepo      repository.CartRepository
	listingClient listingpb.ListingServiceClient
	log           logger.Logger
	cartTTL       time.Duration
}

type CartServiceConfig struct {
	CartTTL time.Duration
}

func NewCartService(
	cartRepo repository.CartRepository,
	listingClient listingpb.ListingServiceClient,
	log logger.Logger,
	cfg CartServiceConfig,
) CartService {
	ttl := cfg.CartTTL
	if ttl <= 0 {
		ttl = defaultCartTTL
	}
	return &cartService{
		cartRepo:      cartRepo,
		listingClient: listingClient,
		log:           log,
		cartTTL:       ttl,
	}
}

func (s *cartService) enrichAndConvertCart(ctx context.Context, cartEntity *entity.Cart) (*cartpb.CartProto, error) {
	if cartEntity == nil {
		return &cartpb.CartProto{UserId: "", Items: []*cartpb.CartItemProto{}, TotalAmount: 0}, nil
	}

	cartProto := &cartpb.CartProto{
		UserId: cartEntity.UserID,
		Items:  make([]*cartpb.CartItemProto, 0, len(cartEntity.Items)),
	}
	var totalAmount float64

	for _, itemEntity := range cartEntity.Items {
		listingResp, err := s.listingClient.GetListingByID(ctx, &listingpb.GetListingRequest{Id: itemEntity.ProductID})
		if err != nil {
			s.log.Errorf("enrichAndConvertCart: Failed to get listing details for productID %s: %v", itemEntity.ProductID, err)
			continue
		}

		if listingResp.Status != "ACTIVE" {
			s.log.Warnf("enrichAndConvertCart: Product %s (ID: %s) is not active, status: %s. Skipping item.", listingResp.Title, itemEntity.ProductID, listingResp.Status)
			continue
		}

		itemPrice := listingResp.Price
		itemTotalPrice := itemPrice * float64(itemEntity.Quantity)
		totalAmount += itemTotalPrice

		cartProto.Items = append(cartProto.Items, &cartpb.CartItemProto{
			ProductId:    itemEntity.ProductID,
			Quantity:     int32(itemEntity.Quantity),
			ProductName:  listingResp.Title,
			PricePerUnit: itemPrice,
			TotalPrice:   itemTotalPrice,
		})
	}
	cartProto.TotalAmount = totalAmount
	return cartProto, nil
}

func (s *cartService) AddItem(ctx context.Context, userID, productID string, quantity int) (*cartpb.CartProto, error) {
	s.log.Infof("Adding item to cart: UserID=%s, ProductID=%s, Quantity=%d", userID, productID, quantity)
	cartEntity, err := s.cartRepo.GetByUserID(ctx, userID)
	if err != nil {
		s.log.Errorf("Error getting cart for user %s: %v", userID, err)
		return nil, fmt.Errorf("could not retrieve cart: %w", err)
	}

	listingResp, err := s.listingClient.GetListingByID(ctx, &listingpb.GetListingRequest{Id: productID})
	if err != nil {
		s.log.Errorf("Failed to get listing details for productID %s: %v", productID, err)
		return nil, fmt.Errorf("product %s not found or service unavailable: %w", productID, err)
	}
	if listingResp.Status != "ACTIVE" {
		s.log.Warnf("Attempted to add inactive product %s (ID: %s) to cart", listingResp.Title, productID)
		return nil, fmt.Errorf("product %s is not available for purchase", listingResp.Title)
	}

	if err := cartEntity.AddItem(productID, quantity); err != nil {
		s.log.Errorf("Error adding item to cart entity for user %s: %v", productID, userID, err)
		return nil, fmt.Errorf("could not add item to cart: %w", err)
	}
	if err := s.cartRepo.Save(ctx, cartEntity, s.cartTTL); err != nil {
		s.log.Errorf("Error saving cart for user %s: %v", userID, err)
		return nil, fmt.Errorf("could not save cart: %w", err)
	}
	s.log.Infof("Item added to cart successfully for user %s", userID)
	return s.enrichAndConvertCart(ctx, cartEntity)
}

func (s *cartService) UpdateItemQuantity(ctx context.Context, userID, productID string, newQuantity int) (*cartpb.CartProto, error) {
	s.log.Infof("Updating item quantity: UserID=%s, ProductID=%s, NewQuantity=%d", userID, productID, newQuantity)
	cartEntity, err := s.cartRepo.GetByUserID(ctx, userID)
	if err != nil {
		s.log.Errorf("Error getting cart for user %s: %v", userID, err)
		return nil, fmt.Errorf("could not retrieve cart: %w", err)
	}
	if err := cartEntity.UpdateItemQuantity(productID, newQuantity); err != nil {
		s.log.Errorf("Error updating item quantity in cart entity for user %s: %v", productID, userID, err)
		return nil, fmt.Errorf("could not update item quantity: %w", err)
	}
	if err := s.cartRepo.Save(ctx, cartEntity, s.cartTTL); err != nil {
		s.log.Errorf("Error saving cart for user %s: %v", userID, err)
		return nil, fmt.Errorf("could not save cart: %w", err)
	}
	s.log.Infof("Item quantity updated successfully for user %s", userID)
	return s.enrichAndConvertCart(ctx, cartEntity)
}

func (s *cartService) RemoveItem(ctx context.Context, userID, productID string) (*cartpb.CartProto, error) {
	s.log.Infof("Removing item from cart: UserID=%s, ProductID=%s", userID, productID)
	cartEntity, err := s.cartRepo.GetByUserID(ctx, userID)
	if err != nil {
		s.log.Errorf("Error getting cart for user %s: %v", userID, err)
		return nil, fmt.Errorf("could not retrieve cart: %w", err)
	}
	if err := cartEntity.RemoveItem(productID); err != nil {
		s.log.Errorf("Error removing item from cart entity for user %s: %v", productID, userID, err)
		return nil, fmt.Errorf("could not remove item from cart: %w", err)
	}
	if err := s.cartRepo.Save(ctx, cartEntity, s.cartTTL); err != nil {
		s.log.Errorf("Error saving cart for user %s: %v", userID, err)
		return nil, fmt.Errorf("could not save cart: %w", err)
	}
	s.log.Infof("Item removed from cart successfully for user %s", userID)
	return s.enrichAndConvertCart(ctx, cartEntity)
}

func (s *cartService) GetCart(ctx context.Context, userID string) (*cartpb.CartProto, error) {
	s.log.Infof("Getting cart for user: UserID=%s", userID)
	cartEntity, err := s.cartRepo.GetByUserID(ctx, userID)
	if err != nil {
		s.log.Errorf("Error getting cart for user %s: %v", userID, err)
		return nil, fmt.Errorf("could not retrieve cart: %w", err)
	}
	return s.enrichAndConvertCart(ctx, cartEntity)
}

func (s *cartService) ClearCart(ctx context.Context, userID string) error {
	s.log.Infof("Clearing cart for user: UserID=%s", userID)
	err := s.cartRepo.DeleteByUserID(ctx, userID)
	if err != nil {
		s.log.Errorf("Error deleting cart for user %s: %v", userID, err)
		return fmt.Errorf("could not clear cart: %w", err)
	}
	s.log.Infof("Cart cleared successfully for user %s", userID)
	return nil
}
