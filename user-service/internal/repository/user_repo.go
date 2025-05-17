package repository

import (
	"context"
	"errors"
	"time"

	"github.com/Abdurahmanit/GroupProject/user-service/internal/entity"
	"github.com/go-redis/redis/v8"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrDuplicateEmail = errors.New("email already exists")
	ErrUserNotFound   = errors.New("user not found")
)

type mongoUser struct {
	ID        primitive.ObjectID `bson:"_id,omitempty"`
	Username  string             `bson:"username"`
	Email     string             `bson:"email"`
	Password  string             `bson:"password"`
	Role      string             `bson:"role"`
	IsActive  bool               `bson:"is_active"`
	CreatedAt time.Time          `bson:"created_at"`
	UpdatedAt time.Time          `bson:"updated_at"`
}

func (m *mongoUser) toEntity() *entity.User {
	return &entity.User{
		ID:        m.ID,
		Username:  m.Username,
		Email:     m.Email,
		Password:  m.Password,
		Role:      m.Role,
		IsActive:  m.IsActive,
		CreatedAt: m.CreatedAt,
		UpdatedAt: m.UpdatedAt,
	}
}

func fromEntity(e *entity.User) *mongoUser {
	return &mongoUser{
		ID:        e.ID,
		Username:  e.Username,
		Email:     e.Email,
		Password:  e.Password,
		Role:      e.Role,
		IsActive:  e.IsActive,
		CreatedAt: e.CreatedAt,
		UpdatedAt: e.UpdatedAt,
	}
}

type UserRepository struct {
	db     *mongo.Database
	redis  *redis.Client
	logger *zap.Logger
}

// NewUserRepository now correctly typed for redis.Client from v8
func NewUserRepository(db *mongo.Database, rds *redis.Client, logger *zap.Logger) *UserRepository {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := db.Collection("users").Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "email", Value: 1}},
		Options: options.Index().SetUnique(true),
	})
	if err != nil {
		logger.Error("Failed to create unique index for email", zap.Error(err)) // Added logging
	} else {
		logger.Info("Successfully ensured unique index for email")
	}

	return &UserRepository{
		db:     db,
		redis:  rds,
		logger: logger,
	}
}

// CreateUser creates a new user in the database.
func (r *UserRepository) CreateUser(ctx context.Context, user *entity.User) (primitive.ObjectID, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return primitive.NilObjectID, err
	}

	dbUser := fromEntity(user)
	dbUser.Password = string(hashedPassword)
	if dbUser.ID.IsZero() {
		dbUser.ID = primitive.NewObjectID()
	}
	dbUser.CreatedAt = time.Now()
	dbUser.UpdatedAt = time.Now()

	_, err = r.db.Collection("users").InsertOne(ctx, dbUser)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return primitive.NilObjectID, ErrDuplicateEmail
		}
		return primitive.NilObjectID, err
	}
	return dbUser.ID, nil
}

// GetUserByEmail retrieves a user by their email.
func (r *UserRepository) GetUserByEmail(ctx context.Context, email string) (*entity.User, error) {
	var dbUser mongoUser
	err := r.db.Collection("users").FindOne(ctx, bson.M{"email": email}).Decode(&dbUser)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return dbUser.toEntity(), nil
}

// GetUserByID retrieves a user by their ID.
func (r *UserRepository) GetUserByID(ctx context.Context, userID primitive.ObjectID) (*entity.User, error) {
	var dbUser mongoUser
	err := r.db.Collection("users").FindOne(ctx, bson.M{"_id": userID}).Decode(&dbUser)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return dbUser.toEntity(), nil
}

// UpdateUser updates an existing user's details.
func (r *UserRepository) UpdateUser(ctx context.Context, user *entity.User) error {
	user.UpdatedAt = time.Now()
	dbUser := fromEntity(user)

	updateDoc := bson.M{
		"$set": bson.M{
			"username":   dbUser.Username,
			"email":      dbUser.Email,
			"role":       dbUser.Role,
			"is_active":  dbUser.IsActive,
			"updated_at": dbUser.UpdatedAt,
		},
	}

	result, err := r.db.Collection("users").UpdateOne(ctx, bson.M{"_id": dbUser.ID}, updateDoc)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return ErrDuplicateEmail
		}
		return err
	}
	if result.MatchedCount == 0 {
		return ErrUserNotFound
	}
	return nil
}

// UpdatePassword updates a user's password.
func (r *UserRepository) UpdatePassword(ctx context.Context, userID primitive.ObjectID, newPassword string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	update := bson.M{
		"$set": bson.M{
			"password":   string(hashedPassword),
			"updated_at": time.Now(),
		},
	}
	result, err := r.db.Collection("users").UpdateOne(ctx, bson.M{"_id": userID}, update)
	if err != nil {
		return err
	}
	if result.MatchedCount == 0 {
		return ErrUserNotFound
	}
	return nil
}

// HardDeleteUser permanently removes a user from the database.
func (r *UserRepository) HardDeleteUser(ctx context.Context, userID primitive.ObjectID) error {
	result, err := r.db.Collection("users").DeleteOne(ctx, bson.M{"_id": userID})
	if err != nil {
		return err
	}
	if result.DeletedCount == 0 {
		return ErrUserNotFound
	}
	// Invalidate associated tokens
	return r.InvalidateToken(ctx, userID.Hex())
}

// DeactivateUser marks a user as inactive (soft delete).
func (r *UserRepository) DeactivateUser(ctx context.Context, userID primitive.ObjectID) error {
	update := bson.M{
		"$set": bson.M{
			"is_active":  false,
			"updated_at": time.Now(),
		},
	}
	result, err := r.db.Collection("users").UpdateOne(ctx, bson.M{"_id": userID}, update)
	if err != nil {
		return err
	}
	if result.MatchedCount == 0 {
		return ErrUserNotFound
	}
	// Invalidate associated tokens
	return r.InvalidateToken(ctx, userID.Hex())
}

// ListUsers retrieves a paginated list of active users.
func (r *UserRepository) ListUsers(ctx context.Context, skip, limit int64) ([]*entity.User, error) {
	findOptions := options.Find()
	findOptions.SetSkip(skip)
	findOptions.SetLimit(limit)
	findOptions.SetSort(bson.M{"created_at": -1})

	filter := bson.M{"is_active": true}

	cursor, err := r.db.Collection("users").Find(ctx, filter, findOptions)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var dbUsers []*mongoUser
	if err = cursor.All(ctx, &dbUsers); err != nil {
		return nil, err
	}

	var users []*entity.User
	for _, dbUser := range dbUsers {
		users = append(users, dbUser.toEntity())
	}
	return users, nil
}

// SearchUsers searches for users based on a query string.
func (r *UserRepository) SearchUsers(ctx context.Context, query string, skip, limit int64) ([]*entity.User, error) {
	findOptions := options.Find()
	findOptions.SetSkip(skip)
	findOptions.SetLimit(limit)
	findOptions.SetSort(bson.M{"created_at": -1})

	filter := bson.M{
		"$or": []bson.M{
			{"username": bson.M{"$regex": query, "$options": "i"}},
			{"email": bson.M{"$regex": query, "$options": "i"}},
		},
	}

	cursor, err := r.db.Collection("users").Find(ctx, filter, findOptions)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var dbUsers []*mongoUser
	if err = cursor.All(ctx, &dbUsers); err != nil {
		return nil, err
	}
	var users []*entity.User
	for _, dbUser := range dbUsers {
		users = append(users, dbUser.toEntity())
	}
	return users, nil
}

// CacheToken stores a token in Redis.
// The keySuffix is typically the userID for JWTs.
func (r *UserRepository) CacheToken(ctx context.Context, keySuffix, token string, expiration time.Duration) error {
	return r.redis.Set(ctx, "token:"+keySuffix, token, expiration).Err()
}

// InvalidateToken removes a token from Redis.
func (r *UserRepository) InvalidateToken(ctx context.Context, keySuffix string) error {
	return r.redis.Del(ctx, "token:"+keySuffix).Err()
}

// GetToken retrieves a token from Redis.
func (r *UserRepository) GetToken(ctx context.Context, keySuffix string) (string, error) {
	token, err := r.redis.Get(ctx, "token:"+keySuffix).Result()
	if errors.Is(err, redis.Nil) {
		return "", nil // Token not found is not an application error here
	}
	return token, err
}
