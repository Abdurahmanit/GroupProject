package repository

import (
	"context"
	"errors"
	"strings"
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
	ErrDuplicateEmail       = errors.New("email already exists")
	ErrDuplicatePhoneNumber = errors.New("phone number already exists") // New error
	ErrUserNotFound         = errors.New("user not found")
)

type mongoUser struct {
	ID          primitive.ObjectID `bson:"_id,omitempty"`
	Username    string             `bson:"username"`
	Email       string             `bson:"email"`
	Password    string             `bson:"password"`
	PhoneNumber string             `bson:"phone_number,omitempty"` // New field
	Role        string             `bson:"role"`
	IsActive    bool               `bson:"is_active"`
	CreatedAt   time.Time          `bson:"created_at"`
	UpdatedAt   time.Time          `bson:"updated_at"`
}

func (m *mongoUser) toEntity() *entity.User {
	return &entity.User{
		ID:          m.ID,
		Username:    m.Username,
		Email:       m.Email,
		Password:    m.Password,
		PhoneNumber: m.PhoneNumber, // New field
		Role:        m.Role,
		IsActive:    m.IsActive,
		CreatedAt:   m.CreatedAt,
		UpdatedAt:   m.UpdatedAt,
	}
}

func fromEntity(e *entity.User) *mongoUser {
	return &mongoUser{
		ID:          e.ID,
		Username:    e.Username,
		Email:       e.Email,
		Password:    e.Password,
		PhoneNumber: e.PhoneNumber, // New field
		Role:        e.Role,
		IsActive:    e.IsActive,
		CreatedAt:   e.CreatedAt,
		UpdatedAt:   e.UpdatedAt,
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

	_, err = db.Collection("users").Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "phone_number", Value: 1}},
		Options: options.Index().SetUnique(true).SetSparse(true),
	})
	if err != nil {
		logger.Error("Failed to create unique index for phone_number in UserRepository", zap.Error(err))
	} else {
		logger.Info("Successfully ensured unique index for phone_number in UserRepository")
	}

	return &UserRepository{
		db:     db,
		redis:  rds,
		logger: logger,
	}
}

// CreateUser creates a new user in the database.
func (r *UserRepository) CreateUser(ctx context.Context, user *entity.User) (primitive.ObjectID, error) {
	r.logger.Info("Attempting to create user in repository", zap.String("email", user.Email), zap.String("phoneNumber", user.PhoneNumber))
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		r.logger.Error("Failed to hash password during user creation", zap.String("email", user.Email), zap.Error(err))
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
		// Check for MongoDB duplicate key error (code 11000)
		var writeException mongo.WriteException
		if errors.As(err, &writeException) {
			for _, writeError := range writeException.WriteErrors {
				if writeError.Code == 11000 {
					if strings.Contains(writeError.Message, "email_1") { // Check for index name or key pattern
						r.logger.Warn("Duplicate email during user creation", zap.String("email", user.Email), zap.Error(writeError))
						return primitive.NilObjectID, ErrDuplicateEmail
					}
					if strings.Contains(writeError.Message, "phone_number_1") { // Check for index name or key pattern
						r.logger.Warn("Duplicate phone number during user creation", zap.String("phoneNumber", user.PhoneNumber), zap.Error(writeError))
						return primitive.NilObjectID, ErrDuplicatePhoneNumber
					}
				}
			}
		}
		// Fallback for other InsertOne errors or if specific duplicate not identified
		r.logger.Error("Database error during user creation", zap.String("email", user.Email), zap.Error(err))
		return primitive.NilObjectID, err
	}
	r.logger.Info("User created successfully in repository", zap.String("userID", dbUser.ID.Hex()))
	return dbUser.ID, nil
}

// GetUserByEmail retrieves a user by their email.
func (r *UserRepository) GetUserByEmail(ctx context.Context, email string) (*entity.User, error) {
	r.logger.Debug("Attempting to get user by email from repository", zap.String("email", email))
	var dbUser mongoUser
	err := r.db.Collection("users").FindOne(ctx, bson.M{"email": email}).Decode(&dbUser)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			r.logger.Warn("User not found by email in repository", zap.String("email", email))
			return nil, ErrUserNotFound
		}
		r.logger.Error("Database error fetching user by email", zap.String("email", email), zap.Error(err))
		return nil, err
	}
	r.logger.Debug("User found by email in repository", zap.String("userID", dbUser.ID.Hex()))
	return dbUser.toEntity(), nil
}

// GetUserByID retrieves a user by their ID.
func (r *UserRepository) GetUserByID(ctx context.Context, userID primitive.ObjectID) (*entity.User, error) {
	r.logger.Debug("Attempting to get user by ID from repository", zap.String("userID", userID.Hex()))
	var dbUser mongoUser
	err := r.db.Collection("users").FindOne(ctx, bson.M{"_id": userID}).Decode(&dbUser)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			r.logger.Warn("User not found by ID in repository", zap.String("userID", userID.Hex()))
			return nil, ErrUserNotFound
		}
		r.logger.Error("Database error fetching user by ID", zap.String("userID", userID.Hex()), zap.Error(err))
		return nil, err
	}
	r.logger.Debug("User found by ID in repository", zap.String("userID", dbUser.ID.Hex()))
	return dbUser.toEntity(), nil
}

// GetUserByPhoneNumber retrieves a user by their phone number.
func (r *UserRepository) GetUserByPhoneNumber(ctx context.Context, phoneNumber string) (*entity.User, error) {
	r.logger.Debug("Attempting to get user by phone number from repository", zap.String("phoneNumber", phoneNumber))
	var dbUser mongoUser
	err := r.db.Collection("users").FindOne(ctx, bson.M{"phone_number": phoneNumber}).Decode(&dbUser)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			r.logger.Warn("User not found by phone number in repository", zap.String("phoneNumber", phoneNumber))
			return nil, ErrUserNotFound // Or a more specific "ErrPhoneNumberNotFound" if needed
		}
		r.logger.Error("Database error fetching user by phone number", zap.String("phoneNumber", phoneNumber), zap.Error(err))
		return nil, err
	}
	r.logger.Debug("User found by phone number in repository", zap.String("userID", dbUser.ID.Hex()))
	return dbUser.toEntity(), nil
}

// UpdateUser updates an existing user's details.
func (r *UserRepository) UpdateUser(ctx context.Context, user *entity.User) error {
	r.logger.Info("Attempting to update user in repository", zap.String("userID", user.ID.Hex()))
	user.UpdatedAt = time.Now()
	dbUser := fromEntity(user)

	updateDoc := bson.M{
		"$set": bson.M{
			"username":     dbUser.Username,
			"email":        dbUser.Email,
			"phone_number": dbUser.PhoneNumber, // Updated field
			"role":         dbUser.Role,
			"is_active":    dbUser.IsActive,
			"updated_at":   dbUser.UpdatedAt,
		},
	}

	result, err := r.db.Collection("users").UpdateOne(ctx, bson.M{"_id": dbUser.ID}, updateDoc)
	if err != nil {
		var writeException mongo.WriteException
		if errors.As(err, &writeException) {
			for _, writeError := range writeException.WriteErrors {
				if writeError.Code == 11000 { // Duplicate key error code
					if strings.Contains(writeError.Message, "email_1") {
						r.logger.Warn("Duplicate email during user update", zap.String("userID", user.ID.Hex()), zap.String("email", user.Email), zap.Error(writeError))
						return ErrDuplicateEmail
					}
					if strings.Contains(writeError.Message, "phone_number_1") {
						r.logger.Warn("Duplicate phone number during user update", zap.String("userID", user.ID.Hex()), zap.String("phoneNumber", user.PhoneNumber), zap.Error(writeError))
						return ErrDuplicatePhoneNumber
					}
				}
			}
		}
		r.logger.Error("Database error during user update", zap.String("userID", user.ID.Hex()), zap.Error(err))
		return err
	}
	if result.MatchedCount == 0 {
		r.logger.Warn("User not found during update attempt", zap.String("userID", user.ID.Hex()))
		return ErrUserNotFound
	}
	r.logger.Info("User updated successfully in repository", zap.String("userID", user.ID.Hex()))
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

// SearchUsers searches for users based on a query string (username, email, or phone number).
func (r *UserRepository) SearchUsers(ctx context.Context, query string, skip, limit int64) ([]*entity.User, error) {
	r.logger.Info("Searching users in repository", zap.String("query", query), zap.Int64("skip", skip), zap.Int64("limit", limit))
	findOptions := options.Find()
	findOptions.SetSkip(skip)
	findOptions.SetLimit(limit)
	findOptions.SetSort(bson.M{"created_at": -1})

	// Case-insensitive regex search for username, email, or phone number
	filter := bson.M{
		"is_active": true, // Continue to filter by active status
		"$or": []bson.M{
			{"username": bson.M{"$regex": query, "$options": "i"}},
			{"email": bson.M{"$regex": query, "$options": "i"}},
			{"phone_number": bson.M{"$regex": query, "$options": "i"}}, // Added phone number search
		},
	}

	cursor, err := r.db.Collection("users").Find(ctx, filter, findOptions)
	if err != nil {
		r.logger.Error("Database error during user search", zap.String("query", query), zap.Error(err))
		return nil, err
	}
	defer cursor.Close(ctx)

	var dbUsers []*mongoUser
	if err = cursor.All(ctx, &dbUsers); err != nil {
		r.logger.Error("Error decoding search results", zap.String("query", query), zap.Error(err))
		return nil, err
	}
	var users []*entity.User
	for _, dbUser := range dbUsers {
		users = append(users, dbUser.toEntity())
	}
	r.logger.Info("User search completed", zap.String("query", query), zap.Int("count", len(users)))
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
