package mongo

import (
	"context"
	"fmt"
	"time"

	"github.com/Abdurahmanit/GroupProject/news-service/internal/config"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

func NewMongoDBConnection(cfg *config.MongoConfig) (*mongo.Client, error) {
	ctx, cancel := context.WithTimeout(context.Background(), cfg.ConnectTimeout)
	defer cancel()

	clientOptions := options.Client().ApplyURI(cfg.URI)

	if cfg.Username != "" && cfg.Password != "" {
		clientOptions.SetAuth(options.Credential{
			Username: cfg.Username,
			Password: cfg.Password,
		})
	}

	if cfg.MinPoolSize > 0 {
		clientOptions.SetMinPoolSize(cfg.MinPoolSize)
	}
	if cfg.MaxPoolSize > 0 {
		clientOptions.SetMaxPoolSize(cfg.MaxPoolSize)
	}

	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to mongo: %w", err)
	}

	pingCtx, pingCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer pingCancel()

	err = client.Ping(pingCtx, readpref.Primary())
	if err != nil {
		client.Disconnect(ctx)
		return nil, fmt.Errorf("failed to ping mongo: %w", err)
	}

	db := client.Database(cfg.Database)

	if err := setupMongoIndexes(ctx, db); err != nil {
		client.Disconnect(ctx)
		return nil, fmt.Errorf("failed to setup mongo indexes: %w", err)
	}

	return client, nil
}

func setupMongoIndexes(ctx context.Context, db *mongo.Database) error {
	newsCollection := db.Collection("news")
	newsIndexes := []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "category", Value: 1}},
			Options: options.Index().SetName("category_idx"),
		},
		{
			Keys:    bson.D{{Key: "created_at", Value: -1}},
			Options: options.Index().SetName("created_at_desc_idx"),
		},
		{
			Keys:    bson.D{{Key: "author_id", Value: 1}},
			Options: options.Index().SetName("author_id_idx"),
		},
	}
	_, err := newsCollection.Indexes().CreateMany(ctx, newsIndexes)
	if err != nil {
		return fmt.Errorf("failed to create indexes for news collection: %w", err)
	}

	commentsCollection := db.Collection("comments")
	commentsIndexes := []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "news_id", Value: 1}},
			Options: options.Index().SetName("comments_news_id_idx"),
		},
		{
			Keys:    bson.D{{Key: "created_at", Value: 1}},
			Options: options.Index().SetName("comments_created_at_asc_idx"),
		},
	}
	_, err = commentsCollection.Indexes().CreateMany(ctx, commentsIndexes)
	if err != nil {
		return fmt.Errorf("failed to create indexes for comments collection: %w", err)
	}

	likesCollection := db.Collection("likes")
	likesIndexes := []mongo.IndexModel{
		{
			Keys: bson.D{
				{Key: "content_type", Value: 1},
				{Key: "content_id", Value: 1},
			},
			Options: options.Index().SetName("likes_content_type_id_idx"),
		},
		{
			Keys: bson.D{
				{Key: "content_type", Value: 1},
				{Key: "content_id", Value: 1},
				{Key: "user_id", Value: 1},
			},
			Options: options.Index().SetName("likes_content_user_unique_idx").SetUnique(true),
		},
	}
	_, err = likesCollection.Indexes().CreateMany(ctx, likesIndexes)
	if err != nil {
		return fmt.Errorf("failed to create indexes for likes collection: %w", err)
	}

	return nil
}
