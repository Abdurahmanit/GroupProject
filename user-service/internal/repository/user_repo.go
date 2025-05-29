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
	ID                             primitive.ObjectID `bson:"_id,omitempty"`
	Username                       string             `bson:"username"`
	Email                          string             `bson:"email"`
	Password                       string             `bson:"password"`
	PhoneNumber                    string             `bson:"phone_number,omitempty"`
	Role                           string             `bson:"role"`
	IsActive                       bool               `bson:"is_active"`
	CreatedAt                      time.Time          `bson:"created_at"`
	UpdatedAt                      time.Time          `bson:"updated_at"`
	IsEmailVerified                bool               `bson:"is_email_verified,omitempty"`
	EmailVerifiedAt                *time.Time         `bson:"email_verified_at,omitempty"`
	EmailVerificationCode          string             `bson:"email_verification_code,omitempty"`
	EmailVerificationCodeExpiresAt *time.Time         `bson:"email_verification_code_expires_at,omitempty"`
}

func (m *mongoUser) toEntity() *entity.User {
	return &entity.User{
		ID:                             m.ID,
		Username:                       m.Username,
		Email:                          m.Email,
		Password:                       m.Password,
		PhoneNumber:                    m.PhoneNumber,
		Role:                           m.Role,
		IsActive:                       m.IsActive,
		CreatedAt:                      m.CreatedAt,
		UpdatedAt:                      m.UpdatedAt,
		IsEmailVerified:                m.IsEmailVerified,
		EmailVerifiedAt:                m.EmailVerifiedAt,
		EmailVerificationCode:          m.EmailVerificationCode,
		EmailVerificationCodeExpiresAt: m.EmailVerificationCodeExpiresAt,
	}
}

func fromEntity(e *entity.User) *mongoUser {
	return &mongoUser{
		ID:                             e.ID,
		Username:                       e.Username,
		Email:                          e.Email,
		Password:                       e.Password,
		PhoneNumber:                    e.PhoneNumber,
		Role:                           e.Role,
		IsActive:                       e.IsActive,
		CreatedAt:                      e.CreatedAt,
		UpdatedAt:                      e.UpdatedAt,
		IsEmailVerified:                e.IsEmailVerified,
		EmailVerifiedAt:                e.EmailVerifiedAt,
		EmailVerificationCode:          e.EmailVerificationCode,
		EmailVerificationCodeExpiresAt: e.EmailVerificationCodeExpiresAt,
	}
}

type UserRepository struct {
	db     *mongo.Database
	redis  *redis.Client
	logger *zap.Logger
}

func NewUserRepository(db *mongo.Database, rds *redis.Client, logger *zap.Logger) *UserRepository {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Ensure indexes (idempotent operation)
	userCollection := db.Collection("users")
	indexes := []mongo.IndexModel{
		{Keys: bson.D{{Key: "email", Value: 1}}, Options: options.Index().SetUnique(true)},
		{Keys: bson.D{{Key: "phone_number", Value: 1}}, Options: options.Index().SetUnique(true).SetSparse(true)},
	}
	_, err := userCollection.Indexes().CreateMany(ctx, indexes)
	if err != nil {
		logger.Warn("Failed to create indexes for users collection (may already exist)", zap.Error(err))
	} else {
		logger.Info("Successfully ensured indexes for users collection")
	}

	return &UserRepository{
		db:     db,
		redis:  rds,
		logger: logger.Named("UserRepository"),
	}
}

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
	now := time.Now()
	dbUser.CreatedAt = now
	dbUser.UpdatedAt = now
	dbUser.IsEmailVerified = false // Default to not verified
	dbUser.EmailVerifiedAt = nil
	dbUser.EmailVerificationCode = ""
	dbUser.EmailVerificationCodeExpiresAt = nil

	_, err = r.db.Collection("users").InsertOne(ctx, dbUser)
	if err != nil {
		var writeException mongo.WriteException
		if errors.As(err, &writeException) {
			for _, writeError := range writeException.WriteErrors {
				if writeError.Code == 11000 {
					if strings.Contains(writeError.Message, "email_1") {
						r.logger.Warn("Duplicate email during user creation", zap.String("email", user.Email), zap.Error(writeError))
						return primitive.NilObjectID, ErrDuplicateEmail
					}
					if strings.Contains(writeError.Message, "phone_number_1") {
						r.logger.Warn("Duplicate phone number during user creation", zap.String("phoneNumber", user.PhoneNumber), zap.Error(writeError))
						return primitive.NilObjectID, ErrDuplicatePhoneNumber
					}
				}
			}
		}
		r.logger.Error("Database error during user creation", zap.String("email", user.Email), zap.Error(err))
		return primitive.NilObjectID, err
	}
	r.logger.Info("User created successfully in repository", zap.String("userID", dbUser.ID.Hex()))
	return dbUser.ID, nil
}

func (r *UserRepository) GetUserByEmail(ctx context.Context, email string) (*entity.User, error) {
	r.logger.Debug("Attempting to get user by email from repository", zap.String("email", email))
	var dbUser mongoUser
	err := r.db.Collection("users").FindOne(ctx, bson.M{"email": email}).Decode(&dbUser)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			r.logger.Debug("User not found by email in repository", zap.String("email", email))
			return nil, ErrUserNotFound
		}
		r.logger.Error("Database error fetching user by email", zap.String("email", email), zap.Error(err))
		return nil, err
	}
	r.logger.Debug("User found by email in repository", zap.String("userID", dbUser.ID.Hex()))
	return dbUser.toEntity(), nil
}

func (r *UserRepository) GetUserByID(ctx context.Context, userID primitive.ObjectID) (*entity.User, error) {
	r.logger.Debug("Attempting to get user by ID from repository", zap.String("userID", userID.Hex()))
	var dbUser mongoUser
	err := r.db.Collection("users").FindOne(ctx, bson.M{"_id": userID}).Decode(&dbUser)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			r.logger.Debug("User not found by ID in repository", zap.String("userID", userID.Hex()))
			return nil, ErrUserNotFound
		}
		r.logger.Error("Database error fetching user by ID", zap.String("userID", userID.Hex()), zap.Error(err))
		return nil, err
	}
	r.logger.Debug("User found by ID in repository", zap.String("userID", dbUser.ID.Hex()))
	return dbUser.toEntity(), nil
}

func (r *UserRepository) GetUserByPhoneNumber(ctx context.Context, phoneNumber string) (*entity.User, error) {
	r.logger.Debug("Attempting to get user by phone number from repository", zap.String("phoneNumber", phoneNumber))
	var dbUser mongoUser
	err := r.db.Collection("users").FindOne(ctx, bson.M{"phone_number": phoneNumber}).Decode(&dbUser)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			r.logger.Debug("User not found by phone number in repository", zap.String("phoneNumber", phoneNumber))
			return nil, ErrUserNotFound
		}
		r.logger.Error("Database error fetching user by phone number", zap.String("phoneNumber", phoneNumber), zap.Error(err))
		return nil, err
	}
	r.logger.Debug("User found by phone number in repository", zap.String("userID", dbUser.ID.Hex()))
	return dbUser.toEntity(), nil
}

func (r *UserRepository) UpdateUser(ctx context.Context, user *entity.User) error {
	r.logger.Info("Attempting to update user in repository", zap.String("userID", user.ID.Hex()))
	user.UpdatedAt = time.Now()
	dbUser := fromEntity(user)

	// Construct update document carefully, only $set fields that are meant to be updated by this generic method.
	// Specific updates (like password, verification status) should have their own methods or more specific update logic.
	updateDoc := bson.M{
		"$set": bson.M{
			"username":     dbUser.Username,
			"email":        dbUser.Email, // Be cautious if email changes require re-verification
			"phone_number": dbUser.PhoneNumber,
			"role":         dbUser.Role,
			"is_active":    dbUser.IsActive,
			"updated_at":   dbUser.UpdatedAt,
			// Note: is_email_verified, email_verified_at, verification codes are handled by specific methods.
		},
	}

	result, err := r.db.Collection("users").UpdateOne(ctx, bson.M{"_id": dbUser.ID}, updateDoc)
	if err != nil {
		var writeException mongo.WriteException
		if errors.As(err, &writeException) {
			for _, writeError := range writeException.WriteErrors {
				if writeError.Code == 11000 {
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

func (r *UserRepository) UpdatePassword(ctx context.Context, userID primitive.ObjectID, newPassword string) error {
	r.logger.Info("Updating password", zap.String("userID", userID.Hex()))
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		r.logger.Error("Failed to hash new password", zap.String("userID", userID.Hex()), zap.Error(err))
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
		r.logger.Error("DB error updating password", zap.String("userID", userID.Hex()), zap.Error(err))
		return err
	}
	if result.MatchedCount == 0 {
		r.logger.Warn("User not found for password update", zap.String("userID", userID.Hex()))
		return ErrUserNotFound
	}
	r.logger.Info("Password updated successfully", zap.String("userID", userID.Hex()))
	return nil
}

func (r *UserRepository) HardDeleteUser(ctx context.Context, userID primitive.ObjectID) error {
	r.logger.Info("Hard deleting user", zap.String("userID", userID.Hex()))
	result, err := r.db.Collection("users").DeleteOne(ctx, bson.M{"_id": userID})
	if err != nil {
		r.logger.Error("DB error hard deleting user", zap.String("userID", userID.Hex()), zap.Error(err))
		return err
	}
	if result.DeletedCount == 0 {
		r.logger.Warn("User not found for hard delete", zap.String("userID", userID.Hex()))
		return ErrUserNotFound
	}
	if err := r.InvalidateToken(ctx, userID.Hex()); err != nil {
		r.logger.Warn("Failed to invalidate token during hard delete, proceeding", zap.String("userID", userID.Hex()), zap.Error(err))
	}
	r.logger.Info("User hard deleted successfully", zap.String("userID", userID.Hex()))
	return nil
}

func (r *UserRepository) DeactivateUser(ctx context.Context, userID primitive.ObjectID) error {
	r.logger.Info("Deactivating user", zap.String("userID", userID.Hex()))
	update := bson.M{
		"$set": bson.M{
			"is_active":  false,
			"updated_at": time.Now(),
		},
	}
	result, err := r.db.Collection("users").UpdateOne(ctx, bson.M{"_id": userID}, update)
	if err != nil {
		r.logger.Error("DB error deactivating user", zap.String("userID", userID.Hex()), zap.Error(err))
		return err
	}
	if result.MatchedCount == 0 {
		r.logger.Warn("User not found for deactivation", zap.String("userID", userID.Hex()))
		return ErrUserNotFound
	}
	if err := r.InvalidateToken(ctx, userID.Hex()); err != nil {
		r.logger.Warn("Failed to invalidate token during deactivation, proceeding", zap.String("userID", userID.Hex()), zap.Error(err))
	}
	r.logger.Info("User deactivated successfully", zap.String("userID", userID.Hex()))
	return nil
}

func (r *UserRepository) ListUsers(ctx context.Context, skip, limit int64) ([]*entity.User, error) {
	r.logger.Debug("Listing users", zap.Int64("skip", skip), zap.Int64("limit", limit))
	findOptions := options.Find()
	findOptions.SetSkip(skip)
	findOptions.SetLimit(limit)
	findOptions.SetSort(bson.M{"created_at": -1})

	// No filter for "is_active" here by default, admin might want to see all. Filter in usecase if needed.
	cursor, err := r.db.Collection("users").Find(ctx, bson.M{}, findOptions)
	if err != nil {
		r.logger.Error("DB error listing users", zap.Error(err))
		return nil, err
	}
	defer cursor.Close(ctx)

	var dbUsers []*mongoUser
	if err = cursor.All(ctx, &dbUsers); err != nil {
		r.logger.Error("Error decoding listed users", zap.Error(err))
		return nil, err
	}

	var users []*entity.User
	for _, dbUser := range dbUsers {
		users = append(users, dbUser.toEntity())
	}
	r.logger.Debug("Users listed successfully", zap.Int("count", len(users)))
	return users, nil
}

func (r *UserRepository) SearchUsers(ctx context.Context, query string, skip, limit int64) ([]*entity.User, error) {
	r.logger.Info("Searching users in repository", zap.String("query", query), zap.Int64("skip", skip), zap.Int64("limit", limit))
	findOptions := options.Find()
	findOptions.SetSkip(skip)
	findOptions.SetLimit(limit)
	findOptions.SetSort(bson.M{"created_at": -1})

	filter := bson.M{
		"$or": []bson.M{
			{"username": bson.M{"$regex": query, "$options": "i"}},
			{"email": bson.M{"$regex": query, "$options": "i"}},
			{"phone_number": bson.M{"$regex": query, "$options": "i"}},
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

func (r *UserRepository) SaveEmailVerificationDetails(ctx context.Context, userID primitive.ObjectID, code string, expiresAt time.Time) error {
	r.logger.Info("Saving email verification details", zap.String("userID", userID.Hex()))
	update := bson.M{
		"$set": bson.M{
			"email_verification_code":            code,
			"email_verification_code_expires_at": expiresAt,
			"updated_at":                         time.Now(),
		},
	}
	result, err := r.db.Collection("users").UpdateOne(ctx, bson.M{"_id": userID}, update)
	if err != nil {
		r.logger.Error("DB error saving email verification details", zap.String("userID", userID.Hex()), zap.Error(err))
		return err
	}
	if result.MatchedCount == 0 {
		r.logger.Warn("User not found for saving email verification details", zap.String("userID", userID.Hex()))
		return ErrUserNotFound
	}
	r.logger.Info("Email verification details saved", zap.String("userID", userID.Hex()))
	return nil
}

func (r *UserRepository) MarkEmailAsVerified(ctx context.Context, userID primitive.ObjectID) error {
	r.logger.Info("Marking email as verified", zap.String("userID", userID.Hex()))
	now := time.Now()
	update := bson.M{
		"$set": bson.M{
			"is_email_verified": true,
			"email_verified_at": now,
			"updated_at":        now,
		},
		"$unset": bson.M{ // Remove code and expiry once verified
			"email_verification_code":            "",
			"email_verification_code_expires_at": "",
		},
	}
	result, err := r.db.Collection("users").UpdateOne(ctx, bson.M{"_id": userID}, update)
	if err != nil {
		r.logger.Error("DB error marking email as verified", zap.String("userID", userID.Hex()), zap.Error(err))
		return err
	}
	if result.MatchedCount == 0 {
		r.logger.Warn("User not found for marking email as verified", zap.String("userID", userID.Hex()))
		return ErrUserNotFound
	}
	r.logger.Info("Email marked as verified", zap.String("userID", userID.Hex()))
	return nil
}

// CacheToken stores a token in Redis.
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
