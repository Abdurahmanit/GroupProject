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

const newsCollectionName = "news"

type NewsMongoRepository struct {
	db *mongo.Database
}

func NewNewsMongoRepository(client *mongo.Client, dbName string) *NewsMongoRepository {
	return &NewsMongoRepository{
		db: client.Database(dbName),
	}
}

type newsDocument struct {
	ID        primitive.ObjectID `bson:"_id,omitempty"`
	Title     string             `bson:"title"`
	Content   string             `bson:"content"`
	AuthorID  string             `bson:"author_id"`
	ImageURL  string             `bson:"image_url,omitempty"` // Новое поле
	CreatedAt primitive.DateTime `bson:"created_at"`
	UpdatedAt primitive.DateTime `bson:"updated_at"`
}

func toNewsDocument(n *entity.News) (*newsDocument, error) {
	doc := &newsDocument{
		Title:     n.Title,
		Content:   n.Content,
		AuthorID:  n.AuthorID,
		ImageURL:  n.ImageURL, // Добавлено
		CreatedAt: primitive.NewDateTimeFromTime(n.CreatedAt),
		UpdatedAt: primitive.NewDateTimeFromTime(n.UpdatedAt),
	}
	if n.ID != "" {
		objID, err := primitive.ObjectIDFromHex(n.ID)
		if err != nil {
			return nil, fmt.Errorf("invalid news ID format: %w", err)
		}
		doc.ID = objID
	}
	return doc, nil
}

func toNewsEntity(doc *newsDocument) *entity.News {
	return &entity.News{
		ID:        doc.ID.Hex(),
		Title:     doc.Title,
		Content:   doc.Content,
		AuthorID:  doc.AuthorID,
		ImageURL:  doc.ImageURL,
		CreatedAt: doc.CreatedAt.Time(),
		UpdatedAt: doc.UpdatedAt.Time(),
	}
}

func (r *NewsMongoRepository) Create(ctx context.Context, news *entity.News) (string, error) {
	doc, err := toNewsDocument(news)
	if err != nil {
		return "", err
	}

	res, err := r.db.Collection(newsCollectionName).InsertOne(ctx, doc)
	if err != nil {
		return "", fmt.Errorf("failed to create news in mongo: %w", err)
	}

	insertedID, ok := res.InsertedID.(primitive.ObjectID)
	if !ok {
		return "", fmt.Errorf("failed to convert inserted_id to ObjectID")
	}
	return insertedID.Hex(), nil
}

func (r *NewsMongoRepository) GetByID(ctx context.Context, id string) (*entity.News, error) {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, repository.ErrNotFound
	}

	var doc newsDocument
	err = r.db.Collection(newsCollectionName).FindOne(ctx, bson.M{"_id": objID}).Decode(&doc)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, repository.ErrNotFound
		}
		return nil, fmt.Errorf("failed to get news by id from mongo: %w", err)
	}
	return toNewsEntity(&doc), nil
}

func (r *NewsMongoRepository) Update(ctx context.Context, news *entity.News) error {
	doc, err := toNewsDocument(news)
	if err != nil {
		return err
	}
	if doc.ID.IsZero() {
		return fmt.Errorf("news ID is required for update")
	}

	updateFields := bson.M{
		"$set": bson.M{
			"title":      doc.Title,
			"content":    doc.Content,
			"author_id":  doc.AuthorID,
			"image_url":  doc.ImageURL, // Добавлено
			"updated_at": doc.UpdatedAt,
		},
	}

	res, err := r.db.Collection(newsCollectionName).UpdateOne(ctx, bson.M{"_id": doc.ID}, updateFields)
	if err != nil {
		return fmt.Errorf("failed to update news in mongo: %w", err)
	}
	if res.MatchedCount == 0 {
		return repository.ErrNotFound
	}
	return nil
}

func (r *NewsMongoRepository) Delete(ctx context.Context, id string) error {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return repository.ErrNotFound
	}

	res, err := r.db.Collection(newsCollectionName).DeleteOne(ctx, bson.M{"_id": objID})
	if err != nil {
		return fmt.Errorf("failed to delete news from mongo: %w", err)
	}
	if res.DeletedCount == 0 {
		return repository.ErrNotFound
	}
	return nil
}

func (r *NewsMongoRepository) List(ctx context.Context, page, pageSize int, filter map[string]interface{}) ([]*entity.News, int, error) {
	skip := int64((page - 1) * pageSize)
	limit := int64(pageSize)

	findOptions := options.Find()
	findOptions.SetSkip(skip)
	findOptions.SetLimit(limit)
	findOptions.SetSort(bson.D{{"created_at", -1}})

	mongoFilter := bson.M{}
	if filter != nil {
		for k, v := range filter {
			if k == "_id" || k == "id" {
				if strVal, ok := v.(string); ok {
					objID, err := primitive.ObjectIDFromHex(strVal)
					if err == nil {
						mongoFilter["_id"] = objID
						continue
					}
				}
			}
			mongoFilter[k] = v
		}
	}

	cursor, err := r.db.Collection(newsCollectionName).Find(ctx, mongoFilter, findOptions)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list news from mongo: %w", err)
	}
	defer cursor.Close(ctx)

	var newsDocs []newsDocument
	if err = cursor.All(ctx, &newsDocs); err != nil {
		return nil, 0, fmt.Errorf("failed to decode news list from mongo: %w", err)
	}

	newsEntities := make([]*entity.News, len(newsDocs))
	for i, doc := range newsDocs {
		newsEntities[i] = toNewsEntity(&doc)
	}

	totalCount, err := r.db.Collection(newsCollectionName).CountDocuments(ctx, mongoFilter)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count news in mongo: %w", err)
	}

	return newsEntities, int(totalCount), nil
}
