package service

import (
	"context"
	"errors"
	"fmt"

	listingpb "github.com/Abdurahmanit/GroupProject/listing-service/genproto/listing_service"
	"github.com/Abdurahmanit/GroupProject/order-service/internal/adapter/nats"
	"github.com/Abdurahmanit/GroupProject/order-service/internal/domain/entity"
	"github.com/Abdurahmanit/GroupProject/order-service/internal/platform/logger"
	"github.com/Abdurahmanit/GroupProject/order-service/internal/repository"
	commonpb "github.com/Abdurahmanit/GroupProject/order-service/proto/common"
	orderpb "github.com/Abdurahmanit/GroupProject/order-service/proto/order"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const (
	natsSubjectOrderCreated       = "order.created"
	natsSubjectOrderStatusUpdated = "order.status.updated"
)

type OrderService interface {
	PlaceOrder(ctx context.Context, userID string, shippingAddr *commonpb.AddressProto, billingAddr *commonpb.AddressProto) (*orderpb.OrderProto, error)
	GetOrderByID(ctx context.Context, orderID, userID string, isAdmin bool) (*orderpb.OrderProto, error)
	ListUserOrders(ctx context.Context, userID string, pagination *commonpb.PaginationRequest) ([]*orderpb.OrderProto, int64, error)
	CancelUserOrder(ctx context.Context, orderID, userID string) (*orderpb.OrderProto, error)
	UpdateOrderStatusByAdmin(ctx context.Context, orderID string, newStatus orderpb.OrderStatusProto, adminID string) (*orderpb.OrderProto, error)
	ListAllOrdersAdmin(ctx context.Context, adminID string, pagination *commonpb.PaginationRequest, filters map[string]string) ([]*orderpb.OrderProto, int64, error)
}

type orderService struct {
	orderRepo     repository.OrderRepository
	cartService   CartService
	listingClient listingpb.ListingServiceClient
	msgPublisher  nats.MessagePublisher
	log           logger.Logger
}

func NewOrderService(
	orderRepo repository.OrderRepository,
	cartService CartService,
	listingClient listingpb.ListingServiceClient,
	msgPublisher nats.MessagePublisher,
	log logger.Logger,
) OrderService {
	return &orderService{
		orderRepo:     orderRepo,
		cartService:   cartService,
		listingClient: listingClient,
		msgPublisher:  msgPublisher,
		log:           log,
	}
}

func mapEntityAddressToProto(addr entity.Address) *commonpb.AddressProto {
	return &commonpb.AddressProto{
		Street:     addr.Street,
		City:       addr.City,
		PostalCode: addr.PostalCode,
		Country:    addr.Country,
	}
}

func mapProtoAddressToEntity(addrProto *commonpb.AddressProto) entity.Address {
	if addrProto == nil {
		return entity.Address{}
	}
	return entity.Address{
		Street:     addrProto.Street,
		City:       addrProto.City,
		PostalCode: addrProto.PostalCode,
		Country:    addrProto.Country,
	}
}

func mapEntityOrderToProto(orderEntity *entity.Order) *orderpb.OrderProto {
	if orderEntity == nil {
		return nil
	}
	itemsProto := make([]*orderpb.OrderItemProto, len(orderEntity.Items))
	for i, item := range orderEntity.Items {
		itemsProto[i] = &orderpb.OrderItemProto{
			ProductId:    item.ProductID,
			ProductName:  item.ProductName,
			Quantity:     int32(item.Quantity),
			PricePerUnit: item.PricePerUnit,
			TotalPrice:   item.TotalPrice,
		}
	}

	var paymentDetailsProto *orderpb.PaymentDetailsProto
	if orderEntity.PaymentDetails.TransactionID != "" || orderEntity.PaymentDetails.PaymentMethodID != "" || orderEntity.PaymentDetails.PaymentStatus != "" {
		paymentDetailsProto = &orderpb.PaymentDetailsProto{
			PaymentMethodId: orderEntity.PaymentDetails.PaymentMethodID,
			TransactionId:   orderEntity.PaymentDetails.TransactionID,
			PaymentStatus:   orderEntity.PaymentDetails.PaymentStatus,
		}
	}

	var statusProto orderpb.OrderStatusProto
	statusValue, ok := orderpb.OrderStatusProto_value[string(orderEntity.Status)]
	if ok {
		statusProto = orderpb.OrderStatusProto(statusValue)
	} else {
		statusProto = orderpb.OrderStatusProto_ORDER_STATUS_PROTO_UNSPECIFIED
	}

	return &orderpb.OrderProto{
		Id:              orderEntity.ID,
		UserId:          orderEntity.UserID,
		Items:           itemsProto,
		TotalAmount:     orderEntity.TotalAmount,
		Status:          statusProto,
		ShippingAddress: mapEntityAddressToProto(orderEntity.ShippingAddress),
		BillingAddress:  mapEntityAddressToProto(orderEntity.BillingAddress),
		PaymentDetails:  paymentDetailsProto,
		CreatedAt:       timestamppb.New(orderEntity.CreatedAt),
		UpdatedAt:       timestamppb.New(orderEntity.UpdatedAt),
	}
}

func (s *orderService) PlaceOrder(ctx context.Context, userID string, shippingAddrProto *commonpb.AddressProto, billingAddrProto *commonpb.AddressProto) (*orderpb.OrderProto, error) {
	s.log.Infof("Placing order for user ID: %s", userID)

	cartPbProto, err := s.cartService.GetCart(ctx, userID)
	if err != nil {
		s.log.Errorf("Failed to get cart for user ID %s: %v", userID, err)
		return nil, fmt.Errorf("failed to retrieve cart for placing order: %w", err)
	}

	if len(cartPbProto.Items) == 0 {
		s.log.Warnf("User ID %s attempted to place an order with an empty cart", userID)
		return nil, fmt.Errorf("cannot place order with an empty cart")
	}

	orderItems := make([]entity.OrderItem, len(cartPbProto.Items))
	for i, itemProto := range cartPbProto.Items {
		newOrderItem, itemErr := entity.NewOrderItem(
			itemProto.ProductId,
			itemProto.ProductName,
			int(itemProto.Quantity),
			itemProto.PricePerUnit,
		)
		if itemErr != nil {
			s.log.Errorf("Failed to create order item for product ID %s: %v", itemProto.ProductId, itemErr)
			return nil, fmt.Errorf("invalid item in cart (product ID %s): %w", itemProto.ProductId, itemErr)
		}
		orderItems[i] = *newOrderItem
	}

	shippingAddr := mapProtoAddressToEntity(shippingAddrProto)
	billingAddr := mapProtoAddressToEntity(billingAddrProto)

	orderEntity, err := entity.NewOrder(userID, orderItems, shippingAddr, billingAddr)
	if err != nil {
		s.log.Errorf("Failed to create new order entity for user ID %s: %v", userID, err)
		return nil, fmt.Errorf("failed to prepare order: %w", err)
	}
	orderEntity.TotalAmount = cartPbProto.TotalAmount

	orderID, err := s.orderRepo.Create(ctx, repository.CreateOrderParams{
		UserID:          orderEntity.UserID,
		Items:           orderEntity.Items,
		TotalAmount:     orderEntity.TotalAmount,
		Status:          orderEntity.Status,
		ShippingAddress: orderEntity.ShippingAddress,
		BillingAddress:  orderEntity.BillingAddress,
	})
	if err != nil {
		s.log.Errorf("Failed to save order for user ID %s to repository: %v", userID, err)
		return nil, fmt.Errorf("failed to save order: %w", err)
	}
	orderEntity.ID = orderID

	if err := s.cartService.ClearCart(ctx, userID); err != nil {
		s.log.Warnf("Failed to clear cart for user ID %s after placing order %s: %v", userID, orderID, err)
	}

	if err := s.msgPublisher.Publish(ctx, natsSubjectOrderCreated, mapEntityOrderToProto(orderEntity)); err != nil {
		s.log.Warnf("Failed to publish order created event for order ID %s: %v", orderID, err)
	}

	s.log.Infof("Order %s placed successfully for user ID %s", orderID, userID)
	return mapEntityOrderToProto(orderEntity), nil
}

func (s *orderService) GetOrderByID(ctx context.Context, orderID, userID string, isAdmin bool) (*orderpb.OrderProto, error) {
	s.log.Infof("Getting order by ID: %s, UserID: %s, IsAdmin: %t", orderID, userID, isAdmin)
	orderEntity, err := s.orderRepo.GetByID(ctx, orderID)
	if err != nil {
		s.log.Errorf("Failed to get order by ID %s from repository: %v", orderID, err)
		if errors.Is(err, repository.ErrNotFound) {
			return nil, fmt.Errorf("order with ID %s not found", orderID)
		}
		return nil, fmt.Errorf("failed to retrieve order: %w", err)
	}

	if !isAdmin && orderEntity.UserID != userID {
		s.log.Warnf("User %s attempted to access order %s belonging to user %s", userID, orderID, orderEntity.UserID)
		return nil, fmt.Errorf("access denied to order %s", orderID)
	}

	s.log.Infof("Order %s retrieved successfully", orderID)
	return mapEntityOrderToProto(orderEntity), nil
}

func (s *orderService) ListUserOrders(ctx context.Context, userID string, paginationProto *commonpb.PaginationRequest) ([]*orderpb.OrderProto, int64, error) {
	s.log.Infof("Listing orders for user ID: %s", userID)
	listParams := repository.ListOrdersParams{
		UserID:   userID,
		Page:     int(paginationProto.GetPage()),
		PageSize: int(paginationProto.GetPageSize()),
	}

	result, err := s.orderRepo.List(ctx, listParams)
	if err != nil {
		s.log.Errorf("Failed to list orders for user ID %s from repository: %v", userID, err)
		return nil, 0, fmt.Errorf("failed to retrieve user orders: %w", err)
	}

	ordersProto := make([]*orderpb.OrderProto, len(result.Orders))
	for i, orderEntity := range result.Orders {
		ordersProto[i] = mapEntityOrderToProto(&orderEntity)
	}

	s.log.Infof("Listed %d orders for user ID %s", len(ordersProto), userID)
	return ordersProto, result.TotalCount, nil
}

func (s *orderService) CancelUserOrder(ctx context.Context, orderID, userID string) (*orderpb.OrderProto, error) {
	s.log.Infof("User %s attempting to cancel order %s", userID, orderID)
	orderEntity, err := s.orderRepo.GetByID(ctx, orderID)
	if err != nil {
		s.log.Errorf("Failed to get order %s for cancellation: %v", orderID, err)
		return nil, fmt.Errorf("order %s not found: %w", orderID, err)
	}

	if orderEntity.UserID != userID {
		s.log.Warnf("User %s attempted to cancel order %s not belonging to them", userID, orderID)
		return nil, fmt.Errorf("access denied: cannot cancel order %s", orderID)
	}

	if !orderEntity.CanBeCancelled() {
		s.log.Warnf("Order %s cannot be cancelled due to its current status: %s", orderID, orderEntity.Status)
		return nil, fmt.Errorf("order %s cannot be cancelled at its current status '%s'", orderID, orderEntity.Status)
	}

	currentVersion := orderEntity.Version
	err = orderEntity.UpdateStatus(entity.StatusCancelled)
	if err != nil {
		s.log.Errorf("Failed to update order entity status to cancelled for order %s: %v", orderID, err)
		return nil, fmt.Errorf("failed to set order status to cancelled: %w", err)
	}

	updateParams := repository.UpdateOrderStatusParams{
		OrderID: orderEntity.ID,
		Status:  orderEntity.Status,
		Version: currentVersion,
	}
	err = s.orderRepo.UpdateStatus(ctx, updateParams)
	if err != nil {
		s.log.Errorf("Failed to save cancelled status for order %s to repository: %v", orderID, err)
		return nil, fmt.Errorf("failed to update order status in repository: %w", err)
	}
	orderEntity.Version = currentVersion + 1

	if errPub := s.msgPublisher.Publish(ctx, natsSubjectOrderStatusUpdated, mapEntityOrderToProto(orderEntity)); errPub != nil {
		s.log.Warnf("Failed to publish order status updated event for order ID %s: %v", orderID, errPub)
	}

	s.log.Infof("Order %s cancelled successfully by user %s", orderID, userID)
	return mapEntityOrderToProto(orderEntity), nil
}

func (s *orderService) UpdateOrderStatusByAdmin(ctx context.Context, orderID string, newStatusProto orderpb.OrderStatusProto, adminID string) (*orderpb.OrderProto, error) {
	s.log.Infof("Admin %s updating status of order %s to %s", adminID, orderID, newStatusProto.String())
	orderEntity, err := s.orderRepo.GetByID(ctx, orderID)
	if err != nil {
		s.log.Errorf("Failed to get order %s for status update by admin %s: %v", orderID, adminID, err)
		return nil, fmt.Errorf("order %s not found: %w", orderID, err)
	}

	newStatusString, ok := orderpb.OrderStatusProto_name[int32(newStatusProto)]
	if !ok || newStatusProto == orderpb.OrderStatusProto_ORDER_STATUS_PROTO_UNSPECIFIED {
		s.log.Errorf("Invalid new status provided by admin %s for order %s: %s", adminID, orderID, newStatusProto.String())
		return nil, fmt.Errorf("invalid new status: %s", newStatusProto.String())
	}
	newStatusEntity := entity.OrderStatus(newStatusString)

	currentVersion := orderEntity.Version
	err = orderEntity.UpdateStatus(newStatusEntity)
	if err != nil {
		s.log.Errorf("Failed to update order entity status for order %s by admin %s: %v. Current status: %s, attempted: %s", orderID, adminID, err, orderEntity.Status, newStatusEntity)
		return nil, fmt.Errorf("failed to set order status: %w", err)
	}

	updateParams := repository.UpdateOrderStatusParams{
		OrderID: orderEntity.ID,
		Status:  orderEntity.Status,
		Version: currentVersion,
	}
	err = s.orderRepo.UpdateStatus(ctx, updateParams)
	if err != nil {
		s.log.Errorf("Failed to save updated status for order %s to repository by admin %s: %v", orderID, adminID, err)
		return nil, fmt.Errorf("failed to update order status in repository: %w", err)
	}
	orderEntity.Version = currentVersion + 1

	if errPub := s.msgPublisher.Publish(ctx, natsSubjectOrderStatusUpdated, mapEntityOrderToProto(orderEntity)); errPub != nil {
		s.log.Warnf("Failed to publish order status updated event for order ID %s: %v", orderID, errPub)
	}

	s.log.Infof("Order %s status updated to %s successfully by admin %s", orderID, newStatusEntity, adminID)
	return mapEntityOrderToProto(orderEntity), nil
}

func (s *orderService) ListAllOrdersAdmin(ctx context.Context, adminID string, paginationProto *commonpb.PaginationRequest, filters map[string]string) ([]*orderpb.OrderProto, int64, error) {
	s.log.Infof("Admin %s listing all orders with pagination and filters: %+v", adminID, filters)

	listParams := repository.ListOrdersParams{
		Page:     int(paginationProto.GetPage()),
		PageSize: int(paginationProto.GetPageSize()),
	}
	if status, ok := filters["status"]; ok {
		listParams.Status = status
	}
	if userID, ok := filters["user_id"]; ok {
		listParams.UserID = userID
	}
	if sortBy, ok := filters["sort_by"]; ok {
		listParams.SortBy = sortBy
	}
	if sortOrder, ok := filters["sort_order"]; ok {
		listParams.SortOrder = sortOrder
	}

	result, err := s.orderRepo.List(ctx, listParams)
	if err != nil {
		s.log.Errorf("Failed to list all orders for admin %s from repository: %v", adminID, err)
		return nil, 0, fmt.Errorf("failed to retrieve all orders: %w", err)
	}

	ordersProto := make([]*orderpb.OrderProto, len(result.Orders))
	for i, orderEntity := range result.Orders {
		ordersProto[i] = mapEntityOrderToProto(&orderEntity)
	}

	s.log.Infof("Listed %d total orders for admin %s", result.TotalCount, adminID)
	return ordersProto, result.TotalCount, nil
}
