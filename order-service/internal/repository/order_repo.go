package repository

import (
	"context"

	"github.com/Abdurahmanit/GroupProject/order-service/internal/domain/entity"
)

type CreateOrderParams struct {
	UserID          string
	Items           []entity.OrderItem
	TotalAmount     float64
	Status          entity.OrderStatus
	ShippingAddress entity.Address
	BillingAddress  entity.Address
	PaymentDetails  entity.PaymentDetails
}

type UpdateOrderPaymentDetailsParams struct {
	OrderID        string
	PaymentDetails entity.PaymentDetails
	Status         entity.OrderStatus // Новый статус заказа после обновления платежа
	Version        int
}

type UpdateOrderStatusParams struct {
	OrderID string
	Status  entity.OrderStatus
	Version int
}

type ListOrdersParams struct {
	UserID    string
	Status    string
	Page      int
	PageSize  int
	SortBy    string
	SortOrder string
}

type ListOrdersResult struct {
	Orders      []entity.Order
	TotalCount  int64
	CurrentPage int
	PageSize    int
	TotalPages  int
}

type OrderRepository interface {
	Create(ctx context.Context, params CreateOrderParams) (string, error)
	GetByID(ctx context.Context, orderID string) (*entity.Order, error)
	UpdateStatus(ctx context.Context, params UpdateOrderStatusParams) error
	UpdatePaymentDetails(ctx context.Context, params UpdateOrderPaymentDetailsParams) error
	List(ctx context.Context, params ListOrdersParams) (*ListOrdersResult, error)
}
