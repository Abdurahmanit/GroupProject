package mongo

import (
	"context"
	"errors"
	"fmt"

	"github.com/Abdurahmanit/GroupProject/news-service/internal/entity"
	"github.com/Abdurahmanit/GroupProject/news-service/internal/port/repository"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const commentCollectionName = "comments"

type CommentMongoRepository struct {
	db *mongo.Database
}

func NewCommentMongoRepository(client *mongo.Client, dbName string) *CommentMongoRepository {
	return &CommentMongoRepository{
		db: client.Database(dbName),
	}
}

type commentDocument struct {
	ID        primitive.ObjectID `bson:"_id,omitempty"`
	NewsID    string             `bson:"news_id"`
	UserID    string             `bson:"user_id"`
	Content   string             `bson:"content"`
	CreatedAt primitive.DateTime `bson:"created_at"`
	UpdatedAt primitive.DateTime `bson:"updated_at"`
}

func toCommentDocument(c *entity.Comment) (*commentDocument, error) {
	doc := &commentDocument{
		NewsID:    c.NewsID,
		UserID:    c.UserID,
		Content:   c.Content,
		CreatedAt: primitive.NewDateTimeFromTime(c.CreatedAt),
		UpdatedAt: primitive.NewDateTimeFromTime(c.UpdatedAt),
	}
	if c.ID != "" {
		objID, err := primitive.ObjectIDFromHex(c.ID)
		if err != nil {
			return nil, fmt.Errorf("invalid comment ID format: %w", err)
		}
		doc.ID = objID
	}
	return doc, nil
}

func toCommentEntity(doc *commentDocument) *entity.Comment {
	return &entity.Comment{
		ID:        doc.ID.Hex(),
		NewsID:    doc.NewsID,
		UserID:    doc.UserID,
		Content:   doc.Content,
		CreatedAt: doc.CreatedAt.Time(),
		UpdatedAt: doc.UpdatedAt.Time(),
	}
}

func (r *CommentMongoRepository) Create(ctx context.Context, comment *entity.Comment) (string, error) {
	doc, err := toCommentDocument(comment)
	if err != nil {
		return "", err
	}

	res, err := r.db.Collection(commentCollectionName).InsertOne(ctx, doc)
	if err != nil {
		return "", fmt.Errorf("failed to create comment in mongo: %w", err)
	}

	insertedID, ok := res.InsertedID.(primitive.ObjectID)
	if !ok {
		return "", fmt.Errorf("failed to convert inserted_id to ObjectID for comment")
	}
	return insertedID.Hex(), nil
}

func (r *CommentMongoRepository) GetByID(ctx context.Context, id string) (*entity.Comment, error) {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, repository.ErrNotFound
	}

	var doc commentDocument
	err = r.db.Collection(commentCollectionName).FindOne(ctx, bson.M{"_id": objID}).Decode(&doc)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, repository.ErrNotFound
		}
		return nil, fmt.Errorf("failed to get comment by id from mongo: %w", err)
	}
	return toCommentEntity(&doc), nil
}

func (r *CommentMongoRepository) GetByNewsID(ctx context.Context, newsID string, page, pageSize int) ([]*entity.Comment, int, error) {
	skip := int64((page - 1) * pageSize)
	limit := int64(pageSize)

	findOptions := options.Find()
	findOptions.SetSkip(skip)
	findOptions.SetLimit(limit)
	findOptions.SetSort(bson.D{{"created_at", 1}})

	mongoFilter := bson.M{"news_id": newsID}

	cursor, err := r.db.Collection(commentCollectionName).Find(ctx, mongoFilter, findOptions)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list comments by news_id from mongo: %w", err)
	}
	defer cursor.Close(ctx)

	var commentDocs []commentDocument
	if err = cursor.All(ctx, &commentDocs); err != nil {
		return nil, 0, fmt.Errorf("failed to decode comment list from mongo: %w", err)
	}

	commentEntities := make([]*entity.Comment, len(commentDocs))
	for i, doc := range commentDocs {
		commentEntities[i] = toCommentEntity(&doc)
	}

	totalCount, err := r.db.Collection(commentCollectionName).CountDocuments(ctx, mongoFilter)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count comments in mongo: %w", err)
	}

	return commentEntities, int(totalCount), nil
}

func (r *CommentMongoRepository) Update(ctx context.Context, comment *entity.Comment) error {
	doc, err := toCommentDocument(comment)
	if err != nil {
		return err
	}
	if doc.ID.IsZero() {
		return fmt.Errorf("comment ID is required for update")
	}

	updateFields := bson.M{
		"$set": bson.M{
			"content":    doc.Content,
			"updated_at": doc.UpdatedAt,
		},
	}

	res, err := r.db.Collection(commentCollectionName).UpdateOne(ctx, bson.M{"_id": doc.ID}, updateFields)
	if err != nil {
		return fmt.Errorf("failed to update comment in mongo: %w", err)
	}
	if res.MatchedCount == 0 {
		return repository.ErrNotFound
	}
	return nil
}

func (r *CommentMongoRepository) Delete(ctx context.Context, id string) error {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return repository.ErrNotFound
	}

	res, err := r.db.Collection(commentCollectionName).DeleteOne(ctx, bson.M{"_id": objID})
	if err != nil {
		return fmt.Errorf("failed to delete comment from mongo: %w", err)
	}
	if res.DeletedCount == 0 {
		return repository.ErrNotFound
	}
	return nil
}
