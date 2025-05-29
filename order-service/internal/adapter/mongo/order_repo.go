package mongo

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Abdurahmanit/GroupProject/order-service/internal/app/config"
	"github.com/Abdurahmanit/GroupProject/order-service/internal/domain/entity"
	"github.com/Abdurahmanit/GroupProject/order-service/internal/repository"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	orderCollectionName = "orders"
)

type orderRepository struct {
	db         *mongo.Database
	collection *mongo.Collection
}

func NewOrderRepository(db *mongo.Client, cfg config.MongoDBConfig) repository.OrderRepository {
	database := db.Database(cfg.Database)
	collection := database.Collection(orderCollectionName)
	return &orderRepository{
		db:         database,
		collection: collection,
	}
}

func (r *orderRepository) Create(ctx context.Context, params repository.CreateOrderParams) (string, error) {
	now := time.Now().UTC()
	order := entity.Order{
		UserID:          params.UserID,
		Items:           params.Items,
		TotalAmount:     params.TotalAmount,
		Status:          params.Status,
		ShippingAddress: params.ShippingAddress,
		BillingAddress:  params.BillingAddress,
		PaymentDetails:  params.PaymentDetails,
		CreatedAt:       now,
		UpdatedAt:       now,
		Version:         1,
	}

	res, err := r.collection.InsertOne(ctx, order)
	if err != nil {
		return "", fmt.Errorf("failed to create order: %w", err)
	}

	objectID, ok := res.InsertedID.(primitive.ObjectID)
	if !ok {
		return "", fmt.Errorf("failed to convert inserted ID to ObjectID")
	}

	return objectID.Hex(), nil
}

func (r *orderRepository) GetByID(ctx context.Context, orderID string) (*entity.Order, error) {
	objID, err := primitive.ObjectIDFromHex(orderID)
	if err != nil {
		return nil, fmt.Errorf("invalid order ID format: %w", repository.ErrNotFound)
	}

	var order entity.Order
	err = r.collection.FindOne(ctx, bson.M{"_id": objID}).Decode(&order)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, repository.ErrNotFound
		}
		return nil, fmt.Errorf("failed to get order by ID %s: %w", orderID, err)
	}
	return &order, nil
}

func (r *orderRepository) UpdateStatus(ctx context.Context, params repository.UpdateOrderStatusParams) error {
	objID, err := primitive.ObjectIDFromHex(params.OrderID)
	if err != nil {
		return fmt.Errorf("invalid order ID format for update status: %w", repository.ErrUpdateFailed)
	}

	filter := bson.M{
		"_id":     objID,
		"version": params.Version,
	}
	update := bson.M{
		"$set": bson.M{
			"status":     params.Status,
			"updated_at": time.Now().UTC(),
		},
		"$inc": bson.M{"version": 1},
	}

	result, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("failed to update order status for ID %s: %w", params.OrderID, err)
	}

	if result.MatchedCount == 0 {
		var existingOrder entity.Order
		errFind := r.collection.FindOne(ctx, bson.M{"_id": objID}).Decode(&existingOrder)
		if errors.Is(errFind, mongo.ErrNoDocuments) {
			return repository.ErrNotFound
		}
		if errFind == nil && existingOrder.Version != params.Version {
			return repository.ErrOptimisticLock
		}
		return repository.ErrUpdateFailed
	}
	if result.ModifiedCount == 0 {
	}

	return nil
}

func (r *orderRepository) UpdatePaymentDetails(ctx context.Context, params repository.UpdateOrderPaymentDetailsParams) error {
	objID, err := primitive.ObjectIDFromHex(params.OrderID)
	if err != nil {
		return fmt.Errorf("invalid order ID format for update payment details: %w", repository.ErrUpdateFailed)
	}

	filter := bson.M{
		"_id":     objID,
		"version": params.Version,
	}
	updateFields := bson.M{
		"payment_details": params.PaymentDetails,
		"updated_at":      time.Now().UTC(),
	}
	if params.Status != "" {
		updateFields["status"] = params.Status
	}

	update := bson.M{
		"$set": updateFields,
		"$inc": bson.M{"version": 1},
	}

	result, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("failed to update order payment details for ID %s: %w", params.OrderID, err)
	}

	if result.MatchedCount == 0 {
		var existingOrder entity.Order
		errFind := r.collection.FindOne(ctx, bson.M{"_id": objID}).Decode(&existingOrder)
		if errors.Is(errFind, mongo.ErrNoDocuments) {
			return repository.ErrNotFound
		}
		if errFind == nil && existingOrder.Version != params.Version {
			return repository.ErrOptimisticLock
		}
		return repository.ErrUpdateFailed
	}
	return nil
}

func (r *orderRepository) List(ctx context.Context, params repository.ListOrdersParams) (*repository.ListOrdersResult, error) {
	filter := bson.M{}
	if params.UserID != "" {
		filter["user_id"] = params.UserID
	}
	if params.Status != "" {
		filter["status"] = params.Status
	}

	findOptions := options.Find()
	if params.PageSize > 0 {
		if params.Page <= 0 {
			params.Page = 1
		}
		findOptions.SetSkip(int64((params.Page - 1) * params.PageSize))
		findOptions.SetLimit(int64(params.PageSize))
	}

	if params.SortBy != "" {
		sortOrder := 1
		if params.SortOrder == "desc" {
			sortOrder = -1
		}
		findOptions.SetSort(bson.D{{Key: params.SortBy, Value: sortOrder}})
	} else {
		findOptions.SetSort(bson.D{{Key: "created_at", Value: -1}})
	}

	cursor, err := r.collection.Find(ctx, filter, findOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to list orders: %w", err)
	}
	defer cursor.Close(ctx)

	var orders []entity.Order
	if err = cursor.All(ctx, &orders); err != nil {
		return nil, fmt.Errorf("failed to decode listed orders: %w", err)
	}

	totalCount, err := r.collection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to count orders: %w", err)
	}

	totalPages := 0
	if params.PageSize > 0 {
		totalPages = (int(totalCount) + params.PageSize - 1) / params.PageSize
	} else if totalCount > 0 {
		totalPages = 1
	}

	return &repository.ListOrdersResult{
		Orders:      orders,
		TotalCount:  totalCount,
		CurrentPage: params.Page,
		PageSize:    params.PageSize,
		TotalPages:  totalPages,
	}, nil
}
