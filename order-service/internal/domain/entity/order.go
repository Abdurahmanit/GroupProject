package entity

import (
	"errors"
	"fmt"
	"time"
)

type OrderStatus string

const (
	StatusPendingPayment OrderStatus = "PENDING_PAYMENT"
	StatusPaid           OrderStatus = "PAID"
	StatusProcessing     OrderStatus = "PROCESSING"
	StatusShipped        OrderStatus = "SHIPPED"
	StatusDelivered      OrderStatus = "DELIVERED"
	StatusCancelled      OrderStatus = "CANCELLED"
	StatusFailed         OrderStatus = "FAILED"
)

type Address struct {
	Street     string `bson:"street,omitempty"`
	City       string `bson:"city,omitempty"`
	PostalCode string `bson:"postal_code,omitempty"`
	Country    string `bson:"country,omitempty"`
}

type OrderItem struct {
	ProductID    string  `bson:"product_id"`
	ProductName  string  `bson:"product_name"`
	Quantity     int     `bson:"quantity"`
	PricePerUnit float64 `bson:"price_per_unit"`
	TotalPrice   float64 `bson:"total_price"`
}

func NewOrderItem(productID, productName string, quantity int, pricePerUnit float64) (*OrderItem, error) {
	if productID == "" {
		return nil, errors.New("product ID cannot be empty")
	}
	if productName == "" {
		return nil, errors.New("product name cannot be empty")
	}
	if quantity <= 0 {
		return nil, errors.New("quantity must be positive")
	}
	if pricePerUnit < 0 {
		return nil, errors.New("price per unit cannot be negative")
	}
	return &OrderItem{
		ProductID:    productID,
		ProductName:  productName,
		Quantity:     quantity,
		PricePerUnit: pricePerUnit,
		TotalPrice:   float64(quantity) * pricePerUnit,
	}, nil
}

type PaymentDetails struct {
	PaymentMethodID string `bson:"payment_method_id,omitempty"`
	TransactionID   string `bson:"transaction_id,omitempty"`
	PaymentStatus   string `bson:"payment_status,omitempty"`
}

type Order struct {
	ID              string         `bson:"_id,omitempty"`
	UserID          string         `bson:"user_id"`
	Items           []OrderItem    `bson:"items"`
	TotalAmount     float64        `bson:"total_amount"`
	Status          OrderStatus    `bson:"status"`
	ShippingAddress Address        `bson:"shipping_address,omitempty"`
	BillingAddress  Address        `bson:"billing_address,omitempty"`
	PaymentDetails  PaymentDetails `bson:"payment_details,omitempty"`
	CreatedAt       time.Time      `bson:"created_at"`
	UpdatedAt       time.Time      `bson:"updated_at"`
	Version         int            `bson:"version"`
}

func NewOrder(userID string, items []OrderItem, shippingAddr, billingAddr Address) (*Order, error) {
	if userID == "" {
		return nil, errors.New("user ID cannot be empty")
	}
	if len(items) == 0 {
		return nil, errors.New("order must contain at least one item")
	}

	order := &Order{
		UserID:          userID,
		Items:           items,
		Status:          StatusPendingPayment,
		ShippingAddress: shippingAddr,
		BillingAddress:  billingAddr,
		CreatedAt:       time.Now().UTC(),
		UpdatedAt:       time.Now().UTC(),
		Version:         1,
	}
	order.CalculateTotalAmount()
	return order, nil
}

func (o *Order) CalculateTotalAmount() {
	var total float64
	for _, item := range o.Items {
		total += item.TotalPrice
	}
	o.TotalAmount = total
}

func (o *Order) CanBeCancelled() bool {
	switch o.Status {
	case StatusPendingPayment, StatusPaid, StatusProcessing:
		return true
	default:
		return false
	}
}

func (o *Order) UpdateStatus(newStatus OrderStatus) error {
	if o.Status == newStatus {
		return nil
	}
	validTransitions := map[OrderStatus][]OrderStatus{
		StatusPendingPayment: {StatusPaid, StatusCancelled, StatusFailed},
		StatusPaid:           {StatusProcessing, StatusCancelled},
		StatusProcessing:     {StatusShipped, StatusCancelled},
		StatusShipped:        {StatusDelivered, StatusCancelled},
		StatusDelivered:      {},
		StatusCancelled:      {},
		StatusFailed:         {StatusPendingPayment},
	}
	allowed, ok := validTransitions[o.Status]
	if !ok {
		return fmt.Errorf("cannot transition from unknown status %s", o.Status)
	}
	canTransition := false
	for _, s := range allowed {
		if s == newStatus {
			canTransition = true
			break
		}
	}
	if !canTransition && newStatus != StatusFailed {
		return fmt.Errorf("invalid status transition from %s to %s", o.Status, newStatus)
	}
	o.Status = newStatus
	o.UpdatedAt = time.Now().UTC()
	o.Version++
	return nil
}

func (o *Order) AddPaymentDetails(details PaymentDetails) {
	o.PaymentDetails = details
	o.UpdatedAt = time.Now().UTC()
}
