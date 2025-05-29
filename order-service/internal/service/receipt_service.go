package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/Abdurahmanit/GroupProject/order-service/internal/platform/logger"
	"github.com/Abdurahmanit/GroupProject/order-service/internal/repository"
)

type ReceiptService interface {
	GenerateOrderReceiptPDF(ctx context.Context, orderID, userID string) ([]byte, string, error)
}

type receiptService struct {
	orderRepo repository.OrderRepository
	log       logger.Logger
}

func NewReceiptService(
	orderRepo repository.OrderRepository,
	log logger.Logger,
) ReceiptService {
	return &receiptService{
		orderRepo: orderRepo,
		log:       log,
	}
}

func (s *receiptService) GenerateOrderReceiptPDF(ctx context.Context, orderID, userID string) ([]byte, string, error) {
	s.log.Infof("Generating PDF receipt for order ID: %s, requested by User ID: %s", orderID, userID)

	orderEntity, err := s.orderRepo.GetByID(ctx, orderID)
	if err != nil {
		s.log.Errorf("Failed to get order by ID %s for PDF generation: %v", orderID, err)
		if errors.Is(err, repository.ErrNotFound) {
			return nil, "", fmt.Errorf("order with ID %s not found", orderID)
		}
		return nil, "", fmt.Errorf("failed to retrieve order: %w", err)
	}

	if orderEntity.UserID != userID {
		s.log.Warnf("User %s attempted to generate receipt for order %s belonging to user %s", userID, orderID, orderEntity.UserID)
		return nil, "", fmt.Errorf("access denied to generate receipt for order %s", orderID)
	}

	receiptContent := fmt.Sprintf(
		"Order ID: %s\nUser ID: %s\nTotal Amount: %.2f\nStatus: %s\n\nItems:\n",
		orderEntity.ID,
		orderEntity.UserID,
		orderEntity.TotalAmount,
		orderEntity.Status,
	)
	for _, item := range orderEntity.Items {
		receiptContent += fmt.Sprintf("- %s (x%d) @ %.2f = %.2f\n",
			item.ProductName,
			item.Quantity,
			item.PricePerUnit,
			item.TotalPrice,
		)
	}
	fileName := fmt.Sprintf("receipt_%s.txt", orderID)

	s.log.Infof("Generated temporary text receipt for order ID %s", orderID)
	return []byte(receiptContent), fileName, nil
}
