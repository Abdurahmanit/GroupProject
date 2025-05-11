// File: user-service/internal/usecase/user.go
package usecase

import (
	"context"
	// "crypto/rand" // No longer needed as email verification is removed
	// "encoding/hex" // No longer needed
	"errors"

	"github.com/Abdurahmanit/GroupProject/user-service/internal/entity"
	"github.com/Abdurahmanit/GroupProject/user-service/internal/jwt"
	"github.com/Abdurahmanit/GroupProject/user-service/internal/repository"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/crypto/bcrypt"
	// "go.uber.org/zap" // Logger can be added if needed
)

var (
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrUnauthorized       = errors.New("unauthorized")
	ErrUserInactive       = errors.New("user account is inactive")
	// Email verification specific errors removed
)

// Email verification constants removed

type UserUsecase struct {
	repo      *repository.UserRepository // This field is unexported (lowercase 'r')
	jwtSecret string
	// logger    *zap.Logger // Uncomment and initialize if logging is desired here
}

func NewUserUsecase(repo *repository.UserRepository, jwtSecret string /*, logger *zap.Logger*/) *UserUsecase {
	return &UserUsecase{
		repo:      repo,
		jwtSecret: jwtSecret,
		// logger:    logger,
	}
}

func (u *UserUsecase) Register(ctx context.Context, username, email, password string) (string, error) {
	user := &entity.User{
		Username: username,
		Email:    email,
		Password: password, // Will be hashed in the repository
		Role:     "customer",
		IsActive: true,
		// ID, CreatedAt, UpdatedAt will be handled by the repository
	}

	objectID, err := u.repo.CreateUser(ctx, user)
	if err != nil {
		return "", err
	}
	// Call to InitiateEmailVerification removed
	return objectID.Hex(), nil
}

func (u *UserUsecase) Login(ctx context.Context, email, password string) (string, error) {
	user, err := u.repo.GetUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return "", ErrInvalidCredentials
		}
		return "", err
	}

	if !user.IsActive {
		return "", ErrUserInactive
	}
	// Check for IsEmailVerified removed

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return "", ErrInvalidCredentials
	}

	tokenString, err := jwt.GenerateToken(user.ID.Hex(), u.jwtSecret)
	if err != nil {
		// u.logger.Error("Failed to generate JWT", zap.Error(err)) // Example logging
		return "", errors.New("failed to generate token")
	}
	return tokenString, nil
}

func (u *UserUsecase) Logout(ctx context.Context, userIDHex string) error {
	// For JWTs, logout is often client-side. Server-side invalidation is for blacklisting.
	return u.repo.InvalidateToken(ctx, "jwt:"+userIDHex) // Assuming a prefix for JWTs in Redis
}

func (u *UserUsecase) GetProfile(ctx context.Context, userIDHex string) (*entity.User, error) {
	objectID, err := primitive.ObjectIDFromHex(userIDHex)
	if err != nil {
		return nil, errors.New("invalid user ID format")
	}
	user, err := u.repo.GetUserByID(ctx, objectID)
	if err != nil {
		return nil, err // errors like ErrUserNotFound will be propagated
	}
	// No need to check IsActive here if JWT implies an active session was established.
	// If this method can be called in a context without prior JWT validation for activity,
	// then an IsActive check might be warranted here.
	return user, nil
}

func (u *UserUsecase) UpdateProfile(ctx context.Context, userIDHex, username, email string) error {
	objectID, err := primitive.ObjectIDFromHex(userIDHex)
	if err != nil {
		return errors.New("invalid user ID format")
	}
	user, err := u.repo.GetUserByID(ctx, objectID)
	if err != nil {
		return err
	}
	if !user.IsActive {
		return ErrUserInactive
	}

	user.Username = username
	user.Email = email
	// user.UpdatedAt will be set by repository's UpdateUser method
	// Logic for setting IsEmailVerified to false on email change removed

	return u.repo.UpdateUser(ctx, user)
}

func (u *UserUsecase) ChangePassword(ctx context.Context, userIDHex, oldPassword, newPassword string) error {
	objectID, err := primitive.ObjectIDFromHex(userIDHex)
	if err != nil {
		return errors.New("invalid user ID format")
	}
	user, err := u.repo.GetUserByID(ctx, objectID)
	if err != nil {
		return err
	}
	if !user.IsActive {
		return ErrUserInactive
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(oldPassword))
	if err != nil {
		return ErrInvalidCredentials
	}

	return u.repo.UpdatePassword(ctx, objectID, newPassword)
}

// DeleteUser (Hard Delete) - User initiated
func (u *UserUsecase) DeleteUser(ctx context.Context, userIDHex string) error {
	objectID, err := primitive.ObjectIDFromHex(userIDHex)
	if err != nil {
		return errors.New("invalid user ID format")
	}
	// Further checks like ensuring the user is deleting their own account
	// are typically handled by ensuring userIDHex matches the ID from the JWT.
	return u.repo.HardDeleteUser(ctx, objectID)
}

// DeactivateUser (Soft Delete) - User initiated
func (u *UserUsecase) DeactivateUser(ctx context.Context, userIDHex string) error {
	objectID, err := primitive.ObjectIDFromHex(userIDHex)
	if err != nil {
		return errors.New("invalid user ID format")
	}
	user, err := u.repo.GetUserByID(ctx, objectID) // Fetch to check current status
	if err != nil {
		return err
	}
	if !user.IsActive { // Already inactive
		return nil
	}
	return u.repo.DeactivateUser(ctx, objectID)
}

// --- Admin Functions ---

// AdminCheck verifies if the provided adminIDHex corresponds to an active admin.
func (u *UserUsecase) AdminCheck(ctx context.Context, adminIDHex string) (*entity.User, error) {
	adminObjectID, err := primitive.ObjectIDFromHex(adminIDHex)
	if err != nil {
		return nil, errors.New("invalid admin ID format")
	}
	admin, err := u.repo.GetUserByID(ctx, adminObjectID)
	if err != nil {
		return nil, err // Could be ErrUserNotFound or other DB error
	}
	if admin.Role != "admin" || !admin.IsActive {
		return nil, ErrUnauthorized
	}
	return admin, nil
}

// AdminDeleteUser (Hard Delete) - Admin initiated
func (u *UserUsecase) AdminDeleteUser(ctx context.Context, adminIDHex, userIDHex string) error {
	if _, err := u.AdminCheck(ctx, adminIDHex); err != nil {
		return err
	}
	userObjectID, err := primitive.ObjectIDFromHex(userIDHex)
	if err != nil {
		return errors.New("invalid user ID format for deletion")
	}
	return u.repo.HardDeleteUser(ctx, userObjectID)
}

func (u *UserUsecase) AdminListUsers(ctx context.Context, adminIDHex string, skip, limit int64) ([]*entity.User, error) {
	if _, err := u.AdminCheck(ctx, adminIDHex); err != nil {
		return nil, err
	}
	return u.repo.ListUsers(ctx, skip, limit)
}

func (u *UserUsecase) AdminSearchUsers(ctx context.Context, adminIDHex, query string, skip, limit int64) ([]*entity.User, error) {
	if _, err := u.AdminCheck(ctx, adminIDHex); err != nil {
		return nil, err
	}
	return u.repo.SearchUsers(ctx, query, skip, limit)
}

func (u *UserUsecase) AdminUpdateUserRole(ctx context.Context, adminIDHex, userIDHex, role string) error {
	if _, err := u.AdminCheck(ctx, adminIDHex); err != nil {
		return err
	}
	userObjectID, err := primitive.ObjectIDFromHex(userIDHex)
	if err != nil {
		return errors.New("invalid user ID format for role update")
	}
	userToUpdate, err := u.repo.GetUserByID(ctx, userObjectID)
	if err != nil {
		return err
	}
	// Add role validation if necessary (e.g., ensure role is one of predefined values)
	userToUpdate.Role = role
	// userToUpdate.UpdatedAt will be set by repository's UpdateUser method
	return u.repo.UpdateUser(ctx, userToUpdate)
}

// AdminSetUserActiveStatus sets the active status of a user by an admin.
// This is the NEW PUBLIC METHOD that the handler will call.
func (u *UserUsecase) AdminSetUserActiveStatus(ctx context.Context, adminIDHex, userIDHex string, isActive bool) error {
	// 1. Authenticate Admin
	if _, err := u.AdminCheck(ctx, adminIDHex); err != nil {
		return err // ErrUnauthorized or other error from AdminCheck
	}

	// 2. Get Target User
	userObjectID, err := primitive.ObjectIDFromHex(userIDHex)
	if err != nil {
		return errors.New("invalid target user ID format")
	}
	targetUser, err := u.repo.GetUserByID(ctx, userObjectID) // Accessing its own 'repo' field
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return repository.ErrUserNotFound // Propagate specific error
		}
		return err // Other repository error
	}

	// 3. Update Status if changed
	if targetUser.IsActive == isActive {
		return nil // No change needed
	}
	targetUser.IsActive = isActive
	// targetUser.UpdatedAt will be set by u.repo.UpdateUser

	if err := u.repo.UpdateUser(ctx, targetUser); err != nil { // Accessing its own 'repo' field
		// u.logger.Error("Failed to update user active status in repo by admin", zap.Error(err))
		return errors.New("failed to update user active status")
	}

	// 4. If deactivating, optionally invalidate token
	if !isActive {
		if err := u.repo.InvalidateToken(ctx, "jwt:"+userIDHex); err != nil { // Accessing its own 'repo' field
			// u.logger.Warn("Failed to invalidate token during admin deactivation", zap.Error(err))
		}
	}
	return nil
}

// Email verification related methods (generateSecureToken, InitiateEmailVerification, ConfirmEmailVerification, ResendVerificationEmail) removed.
