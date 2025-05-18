package usecase

import (
	"context"
	"errors"
	"regexp" // For basic phone number validation

	"github.com/Abdurahmanit/GroupProject/user-service/internal/entity"
	"github.com/Abdurahmanit/GroupProject/user-service/internal/jwt"
	"github.com/Abdurahmanit/GroupProject/user-service/internal/repository"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidCredentials   = errors.New("invalid email or password")
	ErrUnauthorized         = errors.New("unauthorized")
	ErrUserInactive         = errors.New("user account is inactive")
	ErrInvalidPhoneNumber   = errors.New("invalid phone number format")
	ErrPhoneNumberRequired  = errors.New("phone number is required")
	ErrDuplicatePhoneNumber = errors.New("phone number already exists") // New errorы to match repository
	ErrDuplicateEmail       = errors.New("email already exists")        // Usecase level error for consistency
)

var phoneRegex = regexp.MustCompile(`^\+?[1-9]\d{1,14}$`)

type UserUsecase struct {
	repo      *repository.UserRepository
	jwtSecret string
	logger    *zap.Logger // Logger instance
}

func NewUserUsecase(repo *repository.UserRepository, jwtSecret string, logger *zap.Logger) *UserUsecase {
	return &UserUsecase{
		repo:      repo,
		jwtSecret: jwtSecret,
		logger:    logger,
	}
}

func (u *UserUsecase) Register(ctx context.Context, username, email, password, phoneNumber string) (string, error) {
	u.logger.Info("Attempting to register user in usecase", zap.String("email", email), zap.String("username", username), zap.String("phoneNumber", phoneNumber))

	if phoneNumber == "" {
		u.logger.Warn("Registration attempt with empty phone number", zap.String("email", email))
		return "", ErrPhoneNumberRequired
	}
	if !phoneRegex.MatchString(phoneNumber) {
		u.logger.Warn("Registration attempt with invalid phone number format", zap.String("email", email), zap.String("phoneNumber", phoneNumber))
		return "", ErrInvalidPhoneNumber
	}

	// Check if phone number already exists
	_, err := u.repo.GetUserByPhoneNumber(ctx, phoneNumber)
	if err == nil { // User found with this phone number, means it's a duplicate
		u.logger.Warn("Registration attempt with already existing phone number", zap.String("phoneNumber", phoneNumber))
		return "", ErrDuplicatePhoneNumber
	} else if !errors.Is(err, repository.ErrUserNotFound) { // An unexpected error occurred during the check
		u.logger.Error("Error checking for existing phone number during registration", zap.String("phoneNumber", phoneNumber), zap.Error(err))
		return "", err
	}
	// If err is ErrUserNotFound, then the phone number is not taken, which is good.

	// Check if email already exists (optional here if DB index is primary guard, but good for early feedback)
	_, err = u.repo.GetUserByEmail(ctx, email)
	if err == nil {
		u.logger.Warn("Registration attempt with already existing email", zap.String("email", email))
		return "", ErrDuplicateEmail
	} else if !errors.Is(err, repository.ErrUserNotFound) {
		u.logger.Error("Error checking for existing email during registration", zap.String("email", email), zap.Error(err))
		return "", err
	}

	user := &entity.User{
		Username:    username,
		Email:       email,
		Password:    password,
		PhoneNumber: phoneNumber,
		Role:        "customer",
		IsActive:    true,
	}

	objectID, err := u.repo.CreateUser(ctx, user)
	if err != nil {
		u.logger.Error("Failed to register user (repository CreateUser error)", zap.String("email", email), zap.Error(err))
		// Map repository-specific errors to usecase errors if necessary, or propagate directly
		if errors.Is(err, repository.ErrDuplicateEmail) {
			return "", ErrDuplicateEmail
		}
		if errors.Is(err, repository.ErrDuplicatePhoneNumber) {
			return "", ErrDuplicatePhoneNumber
		}
		return "", err // Generic error
	}
	u.logger.Info("User registered successfully in usecase", zap.String("userID", objectID.Hex()), zap.String("email", email))
	return objectID.Hex(), nil
}

func (u *UserUsecase) Login(ctx context.Context, email, password string) (string, error) {
	u.logger.Info("Login attempt", zap.String("email", email))
	user, err := u.repo.GetUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			u.logger.Warn("Login attempt for non-existent user", zap.String("email", email))
			return "", ErrInvalidCredentials
		}
		u.logger.Error("Error fetching user by email during login", zap.String("email", email), zap.Error(err))
		return "", err
	}

	if !user.IsActive {
		u.logger.Warn("Login attempt for inactive user", zap.String("email", email), zap.String("userID", user.ID.Hex()))
		return "", ErrUserInactive
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		u.logger.Warn("Invalid password attempt", zap.String("email", email), zap.String("userID", user.ID.Hex()))
		return "", ErrInvalidCredentials
	}

	tokenString, err := jwt.GenerateToken(user.ID.Hex(), u.jwtSecret)
	if err != nil {
		u.logger.Error("Failed to generate JWT", zap.String("userID", user.ID.Hex()), zap.Error(err))
		return "", errors.New("failed to generate token")
	}
	u.logger.Info("User logged in successfully", zap.String("userID", user.ID.Hex()), zap.String("email", email))
	return tokenString, nil
}

func (u *UserUsecase) Logout(ctx context.Context, userIDHex string) error {
	u.logger.Info("Logout attempt", zap.String("userID", userIDHex))
	err := u.repo.InvalidateToken(ctx, "jwt:"+userIDHex)
	if err != nil {
		u.logger.Error("Failed to invalidate token during logout", zap.String("userID", userIDHex), zap.Error(err))
		return err
	}
	u.logger.Info("User logged out successfully", zap.String("userID", userIDHex))
	return nil
}

func (u *UserUsecase) GetProfile(ctx context.Context, userIDHex string) (*entity.User, error) {
	u.logger.Info("Attempting to get profile in usecase", zap.String("userID", userIDHex))
	objectID, err := primitive.ObjectIDFromHex(userIDHex)
	if err != nil {
		u.logger.Error("Invalid user ID format for GetProfile in usecase", zap.String("userIDHex", userIDHex), zap.Error(err))
		return nil, errors.New("invalid user ID format")
	}
	user, err := u.repo.GetUserByID(ctx, objectID)
	if err != nil {
		u.logger.Error("Failed to get user profile from repository", zap.String("userID", userIDHex), zap.Error(err))
		return nil, err
	}
	u.logger.Info("User profile retrieved successfully in usecase", zap.String("userID", userIDHex))
	return user, nil
}

func (u *UserUsecase) UpdateProfile(ctx context.Context, userIDHex, username, email, phoneNumber string) error {
	u.logger.Info("Attempting to update profile in usecase",
		zap.String("userID", userIDHex),
		zap.String("newUsername", username),
		zap.String("newEmail", email),
		zap.String("newPhoneNumber", phoneNumber))

	objectID, err := primitive.ObjectIDFromHex(userIDHex)
	if err != nil {
		u.logger.Error("Invalid user ID format for UpdateProfile in usecase", zap.String("userIDHex", userIDHex), zap.Error(err))
		return errors.New("invalid user ID format")
	}

	currentUser, err := u.repo.GetUserByID(ctx, objectID)
	if err != nil {
		u.logger.Error("Failed to get user for UpdateProfile from repository", zap.String("userID", userIDHex), zap.Error(err))
		return err
	}
	if !currentUser.IsActive {
		u.logger.Warn("Attempt to update profile of inactive user in usecase", zap.String("userID", userIDHex))
		return ErrUserInactive
	}

	updateUser := *currentUser // Create a mutable copy

	if username != "" { // Allow partial updates
		updateUser.Username = username
	}
	if email != "" && email != currentUser.Email {
		// Optional: Check for email duplication if changing
		_, err := u.repo.GetUserByEmail(ctx, email)
		if err == nil { // Email taken by someone else
			u.logger.Warn("UpdateProfile attempt with email already in use by another user", zap.String("userID", userIDHex), zap.String("email", email))
			return ErrDuplicateEmail
		} else if !errors.Is(err, repository.ErrUserNotFound) {
			u.logger.Error("Error checking for existing email during profile update", zap.String("email", email), zap.Error(err))
			return err
		}
		updateUser.Email = email
	}

	if phoneNumber != "" && phoneNumber != currentUser.PhoneNumber {
		if !phoneRegex.MatchString(phoneNumber) {
			u.logger.Warn("UpdateProfile attempt with invalid phone number format", zap.String("userID", userIDHex), zap.String("phoneNumber", phoneNumber))
			return ErrInvalidPhoneNumber
		}
		// Check if new phone number is already taken by another user
		existingUserWithPhone, err := u.repo.GetUserByPhoneNumber(ctx, phoneNumber)
		if err == nil && existingUserWithPhone.ID != objectID { // Phone taken by someone else
			u.logger.Warn("UpdateProfile attempt with phone number already in use by another user", zap.String("userID", userIDHex), zap.String("phoneNumber", phoneNumber))
			return ErrDuplicatePhoneNumber
		} else if err != nil && !errors.Is(err, repository.ErrUserNotFound) { // Unexpected error
			u.logger.Error("Error checking for existing phone number during profile update", zap.String("phoneNumber", phoneNumber), zap.Error(err))
			return err
		}
		updateUser.PhoneNumber = phoneNumber
	}

	err = u.repo.UpdateUser(ctx, &updateUser)
	if err != nil {
		u.logger.Error("Failed to update user profile in repository via usecase", zap.String("userID", userIDHex), zap.Error(err))
		if errors.Is(err, repository.ErrDuplicateEmail) {
			return ErrDuplicateEmail
		}
		if errors.Is(err, repository.ErrDuplicatePhoneNumber) {
			return ErrDuplicatePhoneNumber
		}
		return err
	}
	u.logger.Info("User profile updated successfully in usecase", zap.String("userID", userIDHex))
	return nil
}

func (u *UserUsecase) ChangePassword(ctx context.Context, userIDHex, oldPassword, newPassword string) error {
	u.logger.Info("Attempting to change password", zap.String("userID", userIDHex))
	objectID, err := primitive.ObjectIDFromHex(userIDHex)
	if err != nil {
		u.logger.Error("Invalid user ID format for ChangePassword", zap.String("userIDHex", userIDHex), zap.Error(err))
		return errors.New("invalid user ID format")
	}
	user, err := u.repo.GetUserByID(ctx, objectID)
	if err != nil {
		u.logger.Error("Failed to get user for ChangePassword", zap.String("userID", userIDHex), zap.Error(err))
		return err
	}
	if !user.IsActive {
		u.logger.Warn("Attempt to change password for inactive user", zap.String("userID", userIDHex))
		return ErrUserInactive
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(oldPassword))
	if err != nil {
		u.logger.Warn("Invalid old password provided for ChangePassword", zap.String("userID", userIDHex), zap.Error(err))
		return ErrInvalidCredentials
	}

	err = u.repo.UpdatePassword(ctx, objectID, newPassword)
	if err != nil {
		u.logger.Error("Failed to update password in repository", zap.String("userID", userIDHex), zap.Error(err))
		return err
	}
	u.logger.Info("Password changed successfully", zap.String("userID", userIDHex))
	return nil
}

func (u *UserUsecase) DeleteUser(ctx context.Context, userIDHex string) error {
	u.logger.Info("Attempting to hard delete user (user initiated)", zap.String("userID", userIDHex))
	objectID, err := primitive.ObjectIDFromHex(userIDHex)
	if err != nil {
		u.logger.Error("Invalid user ID format for DeleteUser", zap.String("userIDHex", userIDHex), zap.Error(err))
		return errors.New("invalid user ID format")
	}
	err = u.repo.HardDeleteUser(ctx, objectID)
	if err != nil {
		u.logger.Error("Failed to hard delete user", zap.String("userID", userIDHex), zap.Error(err))
		return err
	}
	u.logger.Info("User hard deleted successfully", zap.String("userID", userIDHex))
	return nil
}

func (u *UserUsecase) DeactivateUser(ctx context.Context, userIDHex string) error {
	u.logger.Info("Attempting to deactivate user (user initiated)", zap.String("userID", userIDHex))
	objectID, err := primitive.ObjectIDFromHex(userIDHex)
	if err != nil {
		u.logger.Error("Invalid user ID format for DeactivateUser", zap.String("userIDHex", userIDHex), zap.Error(err))
		return errors.New("invalid user ID format")
	}
	user, err := u.repo.GetUserByID(ctx, objectID)
	if err != nil {
		u.logger.Error("Failed to get user for DeactivateUser", zap.String("userID", userIDHex), zap.Error(err))
		return err
	}
	if !user.IsActive {
		u.logger.Info("User already inactive, no action taken for DeactivateUser", zap.String("userID", userIDHex))
		return nil
	}
	err = u.repo.DeactivateUser(ctx, objectID)
	if err != nil {
		u.logger.Error("Failed to deactivate user", zap.String("userID", userIDHex), zap.Error(err))
		return err
	}
	u.logger.Info("User deactivated successfully", zap.String("userID", userIDHex))
	return nil
}

// --- Admin Functions ---

func (u *UserUsecase) AdminCheck(ctx context.Context, adminIDHex string) (*entity.User, error) {
	u.logger.Debug("Performing admin check", zap.String("adminID", adminIDHex))
	adminObjectID, err := primitive.ObjectIDFromHex(adminIDHex)
	if err != nil {
		u.logger.Error("Invalid admin ID format for AdminCheck", zap.String("adminIDHex", adminIDHex), zap.Error(err))
		return nil, errors.New("invalid admin ID format")
	}
	admin, err := u.repo.GetUserByID(ctx, adminObjectID)
	if err != nil {
		u.logger.Error("Failed to get admin user for AdminCheck", zap.String("adminID", adminIDHex), zap.Error(err))
		return nil, err
	}
	if admin.Role != "admin" || !admin.IsActive {
		u.logger.Warn("Admin authorization failed for AdminCheck", zap.String("adminID", adminIDHex), zap.String("role", admin.Role), zap.Bool("isActive", admin.IsActive))
		return nil, ErrUnauthorized
	}
	u.logger.Debug("Admin check successful", zap.String("adminID", adminIDHex))
	return admin, nil
}

func (u *UserUsecase) AdminDeleteUser(ctx context.Context, adminIDHex, userIDHex string) error {
	u.logger.Info("Admin attempting to hard delete user", zap.String("adminID", adminIDHex), zap.String("targetUserID", userIDHex))
	admin, err := u.AdminCheck(ctx, adminIDHex)
	if err != nil {
		// AdminCheck already logs, no need to log again here unless for context
		return err
	}
	userObjectID, err := primitive.ObjectIDFromHex(userIDHex)
	if err != nil {
		u.logger.Error("Invalid target user ID format for AdminDeleteUser", zap.String("userIDHex", userIDHex), zap.Error(err))
		return errors.New("invalid user ID format for deletion")
	}
	err = u.repo.HardDeleteUser(ctx, userObjectID)
	if err != nil {
		u.logger.Error("Admin failed to hard delete user", zap.String("adminID", admin.ID.Hex()), zap.String("targetUserID", userIDHex), zap.Error(err))
		return err
	}
	u.logger.Info("Admin successfully hard deleted user", zap.String("adminID", admin.ID.Hex()), zap.String("targetUserID", userIDHex))
	return nil
}

func (u *UserUsecase) AdminListUsers(ctx context.Context, adminIDHex string, skip, limit int64) ([]*entity.User, error) {
	u.logger.Info("Admin attempting to list users", zap.String("adminID", adminIDHex), zap.Int64("skip", skip), zap.Int64("limit", limit))
	admin, err := u.AdminCheck(ctx, adminIDHex)
	if err != nil {
		return nil, err
	}
	users, err := u.repo.ListUsers(ctx, skip, limit)
	if err != nil {
		u.logger.Error("Admin failed to list users", zap.String("adminID", admin.ID.Hex()), zap.Error(err))
		return nil, err
	}
	u.logger.Info("Admin successfully listed users", zap.String("adminID", admin.ID.Hex()), zap.Int("count", len(users)))
	return users, nil
}

func (u *UserUsecase) AdminSearchUsers(ctx context.Context, adminIDHex, query string, skip, limit int64) ([]*entity.User, error) {
	u.logger.Info("Admin attempting to search users (usecase)", zap.String("adminID", adminIDHex), zap.String("query", query), zap.Int64("skip", skip), zap.Int64("limit", limit))
	admin, err := u.AdminCheck(ctx, adminIDHex)
	if err != nil {
		return nil, err
	}
	users, err := u.repo.SearchUsers(ctx, query, skip, limit)
	if err != nil {
		u.logger.Error("Admin failed to search users (repository error)", zap.String("adminID", admin.ID.Hex()), zap.String("query", query), zap.Error(err))
		return nil, err
	}
	u.logger.Info("Admin successfully searched users (usecase)", zap.String("adminID", admin.ID.Hex()), zap.String("query", query), zap.Int("count", len(users)))
	return users, nil
}

func (u *UserUsecase) AdminUpdateUserRole(ctx context.Context, adminIDHex, userIDHex, role string) error {
	u.logger.Info("Admin attempting to update user role", zap.String("adminID", adminIDHex), zap.String("targetUserID", userIDHex), zap.String("newRole", role))
	admin, err := u.AdminCheck(ctx, adminIDHex)
	if err != nil {
		return err
	}
	userObjectID, err := primitive.ObjectIDFromHex(userIDHex)
	if err != nil {
		u.logger.Error("Invalid target user ID format for AdminUpdateUserRole", zap.String("userIDHex", userIDHex), zap.Error(err))
		return errors.New("invalid user ID format for role update")
	}
	userToUpdate, err := u.repo.GetUserByID(ctx, userObjectID)
	if err != nil {
		u.logger.Error("Failed to get user for AdminUpdateUserRole", zap.String("targetUserID", userIDHex), zap.Error(err))
		return err
	}

	oldRole := userToUpdate.Role
	userToUpdate.Role = role
	err = u.repo.UpdateUser(ctx, userToUpdate)
	if err != nil {
		u.logger.Error("Admin failed to update user role", zap.String("adminID", admin.ID.Hex()), zap.String("targetUserID", userIDHex), zap.String("newRole", role), zap.Error(err))
		return err
	}
	u.logger.Info("Admin successfully updated user role", zap.String("adminID", admin.ID.Hex()), zap.String("targetUserID", userIDHex), zap.String("oldRole", oldRole), zap.String("newRole", role))
	return nil
}

func (u *UserUsecase) AdminSetUserActiveStatus(ctx context.Context, adminIDHex, userIDHex string, isActive bool) error {
	u.logger.Info("Admin attempting to set user active status", zap.String("adminID", adminIDHex), zap.String("targetUserID", userIDHex), zap.Bool("isActive", isActive))
	admin, err := u.AdminCheck(ctx, adminIDHex)
	if err != nil {
		u.logger.Warn("Admin check failed for AdminSetUserActiveStatus", zap.String("attemptedAdminID", adminIDHex), zap.Error(err))
		return err
	}

	userObjectID, err := primitive.ObjectIDFromHex(userIDHex)
	if err != nil {
		u.logger.Error("Invalid target user ID format for AdminSetUserActiveStatus", zap.String("userIDHex", userIDHex), zap.Error(err))
		return errors.New("invalid target user ID format")
	}
	targetUser, err := u.repo.GetUserByID(ctx, userObjectID)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			u.logger.Warn("Target user not found for AdminSetUserActiveStatus", zap.String("targetUserID", userIDHex), zap.Error(err))
			return repository.ErrUserNotFound
		}
		u.logger.Error("Error fetching target user for AdminSetUserActiveStatus", zap.String("targetUserID", userIDHex), zap.Error(err))
		return err
	}

	if targetUser.IsActive == isActive {
		u.logger.Info("AdminSetUserActiveStatus: No change needed for user", zap.String("targetUserID", userIDHex), zap.Bool("isActive", isActive))
		return nil
	}
	oldStatus := targetUser.IsActive
	targetUser.IsActive = isActive

	if err := u.repo.UpdateUser(ctx, targetUser); err != nil {
		u.logger.Error("Failed to update user active status in repo by admin", zap.String("adminID", admin.ID.Hex()), zap.String("targetUserID", targetUser.ID.Hex()), zap.Error(err))
		return errors.New("failed to update user active status")
	}
	u.logger.Info("Admin successfully set user active status", zap.String("adminID", admin.ID.Hex()), zap.String("targetUserID", targetUser.ID.Hex()), zap.Bool("oldStatus", oldStatus), zap.Bool("newStatus", isActive))

	if !isActive {
		if err := u.repo.InvalidateToken(ctx, "jwt:"+userIDHex); err != nil {
			u.logger.Warn("Failed to invalidate token during admin deactivation", zap.String("targetUserID", userIDHex), zap.Error(err))
		} else {
			u.logger.Info("Token invalidated for admin-deactivated user", zap.String("targetUserID", userIDHex))
		}
	}
	return nil
}
