package mongodb

import (
    "context"
    "fmt"

    "github.com/Abdurahmanit/GroupProject/listing-service/internal/platform/logger"
    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/bson/primitive"
    "go.mongodb.org/mongo-driver/mongo"
)

type UserRepository struct {
    collection *mongo.Collection
    logger     *logger.Logger
}

func NewUserRepository(db *mongo.Database, log *logger.Logger) *UserRepository {
    return &UserRepository{
        collection: db.Collection("users"),
        logger:     log,
    }
}

// GetEmailByID получает email пользователя по его ID (hex string)
func (r *UserRepository) GetEmailByID(ctx context.Context, userID string) (string, error) {
    objID, err := primitive.ObjectIDFromHex(userID)
    if err != nil {
        r.logger.Error("GetEmailByID: invalid userID", "userID", userID, "error", err)
        return "", fmt.Errorf("invalid user ID format: %w", err)
    }

    var userDoc struct {
        Email string `bson:"email"`
    }

    err = r.collection.FindOne(ctx, bson.M{"_id": objID}).Decode(&userDoc)
    if err != nil {
        if err == mongo.ErrNoDocuments {
            r.logger.Info("GetEmailByID: user not found", "userID", userID)
            return "", fmt.Errorf("user not found")
        }
        r.logger.Error("GetEmailByID: failed to find user", "userID", userID, "error", err)
        return "", err
    }

    return userDoc.Email, nil
}
