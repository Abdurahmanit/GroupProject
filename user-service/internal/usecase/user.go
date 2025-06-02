package usecase

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"regexp"
	"time"

	"github.com/Abdurahmanit/GroupProject/user-service/internal/entity"
	"github.com/Abdurahmanit/GroupProject/user-service/internal/jwt"
	"github.com/Abdurahmanit/GroupProject/user-service/internal/mailer"
	"github.com/Abdurahmanit/GroupProject/user-service/internal/repository"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidCredentials      = errors.New("invalid email or password")
	ErrUnauthorized            = errors.New("unauthorized")
	ErrUserInactive            = errors.New("user account is inactive")
	ErrInvalidPhoneNumber      = errors.New("invalid phone number format")
	ErrPhoneNumberRequired     = errors.New("phone number is required")
	ErrDuplicatePhoneNumber    = errors.New("phone number already exists")
	ErrDuplicateEmail          = errors.New("email already exists")
	ErrEmailAlreadyVerified    = errors.New("email is already verified")
	ErrInvalidVerificationCode = errors.New("invalid or expired verification code")
	ErrMailerFailed            = errors.New("failed to send verification email")
	ErrUserNotFound            = errors.New("user not found")
)

var phoneRegex = regexp.MustCompile(`^\+?[1-9]\d{1,14}$`)

const verificationCodeLength = 6
const verificationCodeExpiryMinutes = 15

type UserUsecase struct {
	repo      *repository.UserRepository
	mailer    mailer.Mailer
	jwtSecret string
	logger    *zap.Logger
}

func NewUserUsecase(repo *repository.UserRepository, mailer mailer.Mailer, jwtSecret string, logger *zap.Logger) *UserUsecase {
	return &UserUsecase{
		repo:      repo,
		mailer:    mailer,
		jwtSecret: jwtSecret,
		logger:    logger.Named("UserUsecase"),
	}
}

func generateVerificationCode(length int) (string, error) {
	const charset = "0123456789"
	code := make([]byte, length)
	for i := range code {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			return "", fmt.Errorf("failed to generate random number for code: %w", err)
		}
		code[i] = charset[num.Int64()]
	}
	return string(code), nil
}

func (u *UserUsecase) internalSendVerificationEmail(ctx context.Context, user *entity.User) error {
	u.logger.Info("internalSendVerificationEmail: Attempting to send verification email", zap.String("userID", user.ID.Hex()), zap.String("email", user.Email))

	code, err := generateVerificationCode(verificationCodeLength)
	if err != nil {
		u.logger.Error("internalSendVerificationEmail: Failed to generate verification code", zap.String("userID", user.ID.Hex()), zap.Error(err))
		return fmt.Errorf("could not generate verification code: %w", err)
	}
	expiresAt := time.Now().Add(verificationCodeExpiryMinutes * time.Minute)

	err = u.repo.SaveEmailVerificationDetails(ctx, user.ID, code, expiresAt)
	if err != nil {
		u.logger.Error("internalSendVerificationEmail: Failed to save verification code to repository", zap.String("userID", user.ID.Hex()), zap.Error(err))
		return err
	}

	err = u.mailer.SendEmailVerification(user.Email, user.Username, code)
	if err != nil {
		u.logger.Error("internalSendVerificationEmail: Failed to send verification email via mailer", zap.String("userID", user.ID.Hex()), zap.String("email", user.Email), zap.Error(err))
		return ErrMailerFailed
	}

	u.logger.Info("internalSendVerificationEmail: Verification email sent successfully", zap.String("userID", user.ID.Hex()), zap.String("email", user.Email))
	return nil
}

func (u *UserUsecase) Register(ctx context.Context, username, email, password, phoneNumber string) (string, error) {
	u.logger.Info("Register: Attempting to register user", zap.String("email", email), zap.String("username", username), zap.String("phoneNumber", phoneNumber))

	if phoneNumber == "" {
		return "", ErrPhoneNumberRequired
	}
	if !phoneRegex.MatchString(phoneNumber) {
		return "", ErrInvalidPhoneNumber
	}

	_, err := u.repo.GetUserByEmail(ctx, email)
	if err == nil {
		return "", ErrDuplicateEmail
	} else if !errors.Is(err, repository.ErrUserNotFound) {
		return "", err
	}

	_, err = u.repo.GetUserByPhoneNumber(ctx, phoneNumber)
	if err == nil {
		return "", ErrDuplicatePhoneNumber
	} else if !errors.Is(err, repository.ErrUserNotFound) {
		return "", err
	}

	userEntity := &entity.User{
		Username:        username,
		Email:           email,
		Password:        password,
		PhoneNumber:     phoneNumber,
		Role:            "customer",
		IsActive:        true,
		IsEmailVerified: false,
		EmailVerifiedAt: nil,
	}

	objectID, err := u.repo.CreateUser(ctx, userEntity)
	if err != nil {
		u.logger.Error("Register: Failed to create user in repository", zap.Error(err))
		return "", err
	}
	u.logger.Info("Register: User created successfully in repository", zap.String("userID", objectID.Hex()))
	createdUser, err := u.repo.GetUserByID(ctx, objectID)
	if err != nil {
		u.logger.Error("Register: Failed to retrieve newly created user for sending verification email", zap.String("userID", objectID.Hex()), zap.Error(err))
	} else {
		err = u.internalSendVerificationEmail(ctx, createdUser)
		if err != nil {
			u.logger.Error("Register: Failed to send verification email automatically after registration", zap.String("userID", objectID.Hex()), zap.Error(err))
		}
	}

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

func (u *UserUsecase) RequestEmailVerification(ctx context.Context, userIDHex string) error {
	u.logger.Info("RequestEmailVerification: User requested verification email", zap.String("userID", userIDHex))
	objectID, err := primitive.ObjectIDFromHex(userIDHex)
	if err != nil {
		return errors.New("invalid user ID format")
	}

	user, err := u.repo.GetUserByID(ctx, objectID)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return ErrUserNotFound
		}
		return err
	}

	if user.IsEmailVerified {
		u.logger.Info("RequestEmailVerification: Email already verified for user", zap.String("userID", userIDHex))
		return ErrEmailAlreadyVerified
	}

	return u.internalSendVerificationEmail(ctx, user)
}

func (u *UserUsecase) VerifyEmail(ctx context.Context, userIDHex string, code string) error {
	u.logger.Info("Attempting to verify email", zap.String("userID", userIDHex))
	objectID, err := primitive.ObjectIDFromHex(userIDHex)
	if err != nil {
		return errors.New("invalid user ID format")
	}

	user, err := u.repo.GetUserByID(ctx, objectID)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return ErrUserNotFound
		}
		return err
	}

	if user.IsEmailVerified {
		u.logger.Info("Email already verified for user during verification attempt", zap.String("userID", userIDHex))
		return ErrEmailAlreadyVerified
	}

	if user.EmailVerificationCode == "" || user.EmailVerificationCodeExpiresAt == nil {
		u.logger.Warn("No verification code found or expiry not set for user", zap.String("userID", userIDHex))
		return ErrInvalidVerificationCode
	}

	if user.EmailVerificationCode != code {
		u.logger.Warn("Invalid verification code provided", zap.String("userID", userIDHex))
		return ErrInvalidVerificationCode
	}

	if time.Now().After(*user.EmailVerificationCodeExpiresAt) {
		u.logger.Warn("Verification code expired", zap.String("userID", userIDHex))
		return ErrInvalidVerificationCode
	}

	err = u.repo.MarkEmailAsVerified(ctx, user.ID)
	if err != nil {
		u.logger.Error("Failed to mark email as verified in repository", zap.String("userID", userIDHex), zap.Error(err))
		return err
	}

	u.logger.Info("Email verified successfully", zap.String("userID", userIDHex))
	return nil
}

func (u *UserUsecase) CheckEmailVerificationStatus(ctx context.Context, userIDHex string) (bool, error) {
	u.logger.Debug("Checking email verification status", zap.String("userID", userIDHex))
	objectID, err := primitive.ObjectIDFromHex(userIDHex)
	if err != nil {
		return false, errors.New("invalid user ID format")
	}

	user, err := u.repo.GetUserByID(ctx, objectID)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return false, ErrUserNotFound
		}
		return false, err
	}
	return user.IsEmailVerified, nil
}

func (u *UserUsecase) Logout(ctx context.Context, userIDHex string) error {
	u.logger.Info("Logout attempt", zap.String("userID", userIDHex))
	err := u.repo.InvalidateToken(ctx, userIDHex)
	if err != nil {
		u.logger.Error("Failed to invalidate token during logout", zap.String("userID", userIDHex), zap.Error(err))
		return err
	}
	u.logger.Info("User logged out successfully (token invalidated if applicable)", zap.String("userID", userIDHex))
	return nil
}

func (u *UserUsecase) GetProfile(ctx context.Context, userIDHex string) (*entity.User, error) {
	u.logger.Info("Attempting to get profile in usecase", zap.String("userID", userIDHex))
	objectID, err := primitive.ObjectIDFromHex(userIDHex)
	if err != nil {
		return nil, errors.New("invalid user ID format")
	}
	user, err := u.repo.GetUserByID(ctx, objectID)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return nil, ErrUserNotFound
		}
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
		return errors.New("invalid user ID format")
	}

	currentUser, err := u.repo.GetUserByID(ctx, objectID)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return ErrUserNotFound
		}
		return err
	}
	if !currentUser.IsActive {
		return ErrUserInactive
	}

	updateUser := *currentUser
	changedEmail := false

	originalIsEmailVerified := currentUser.IsEmailVerified
	originalEmailVerifiedAt := currentUser.EmailVerifiedAt

	if username != "" {
		updateUser.Username = username
	}

	if email != "" && email != currentUser.Email {
		u.logger.Info("Email change detected in UpdateProfile",
			zap.String("userID", userIDHex),
			zap.String("oldEmail", currentUser.Email),
			zap.String("newEmail", email))

		existingUserWithEmail, emailErr := u.repo.GetUserByEmail(ctx, email)
		if emailErr == nil && existingUserWithEmail.ID != objectID {
			return ErrDuplicateEmail
		} else if emailErr != nil && !errors.Is(emailErr, repository.ErrUserNotFound) {
			return emailErr
		}
		updateUser.Email = email
		updateUser.IsEmailVerified = false
		updateUser.EmailVerifiedAt = nil
		changedEmail = true
		u.logger.Info("Email verification status explicitly reset due to email change",
			zap.Bool("isEmailVerified_set_to", updateUser.IsEmailVerified),
			zap.Bool("emailVerifiedAt_is_nil_set_to", updateUser.EmailVerifiedAt == nil))
	} else {
		updateUser.IsEmailVerified = originalIsEmailVerified
		updateUser.EmailVerifiedAt = originalEmailVerifiedAt
	}

	if phoneNumber != "" && phoneNumber != currentUser.PhoneNumber {
		if !phoneRegex.MatchString(phoneNumber) {
			return ErrInvalidPhoneNumber
		}
		existingUserWithPhone, phoneErr := u.repo.GetUserByPhoneNumber(ctx, phoneNumber)
		if phoneErr == nil && existingUserWithPhone.ID != objectID {
			return ErrDuplicatePhoneNumber
		} else if phoneErr != nil && !errors.Is(phoneErr, repository.ErrUserNotFound) {
			return phoneErr
		}
		updateUser.PhoneNumber = phoneNumber
	}

	u.logger.Info("User entity state before calling repo.UpdateUser",
		zap.String("userID", updateUser.ID.Hex()),
		zap.Bool("isEmailVerified", updateUser.IsEmailVerified),
		zap.Any("emailVerifiedAt", updateUser.EmailVerifiedAt))

	err = u.repo.UpdateUser(ctx, &updateUser)
	if err != nil {
		if errors.Is(err, repository.ErrDuplicateEmail) {
			return ErrDuplicateEmail
		}
		if errors.Is(err, repository.ErrDuplicatePhoneNumber) {
			return ErrDuplicatePhoneNumber
		}
		if errors.Is(err, repository.ErrUserNotFound) {
			return ErrUserNotFound
		}
		return err
	}

	if changedEmail {
		u.logger.Info("Email changed, now clearing old verification code details from repository.", zap.String("userID", userIDHex))
		err = u.repo.SaveEmailVerificationDetails(ctx, updateUser.ID, "", time.Time{})
		if err != nil {
			u.logger.Error("Failed to clear old verification code details after email change", zap.String("userID", userIDHex), zap.Error(err))
		}
		u.logger.Info("Attempting to send verification email to new address after profile update", zap.String("newEmail", updateUser.Email))
		if errMail := u.internalSendVerificationEmail(ctx, &updateUser); errMail != nil {
			u.logger.Warn("Failed to automatically send verification email to new address after profile update", zap.Error(errMail))
		}
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
		if errors.Is(err, repository.ErrUserNotFound) {
			return ErrUserNotFound
		}
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
		if errors.Is(err, repository.ErrUserNotFound) {
			return ErrUserNotFound
		}
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
		if errors.Is(err, repository.ErrUserNotFound) {
			return ErrUserNotFound
		}
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
		if errors.Is(err, repository.ErrUserNotFound) {
			return ErrUserNotFound
		}
		return err
	}
	if !user.IsActive {
		u.logger.Info("User already inactive, no action taken for DeactivateUser", zap.String("userID", userIDHex))
		return nil
	}
	err = u.repo.DeactivateUser(ctx, objectID)
	if err != nil {
		u.logger.Error("Failed to deactivate user", zap.String("userID", userIDHex), zap.Error(err))
		if errors.Is(err, repository.ErrUserNotFound) {
			return ErrUserNotFound
		}
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
		if errors.Is(err, repository.ErrUserNotFound) {
			return nil, ErrUserNotFound
		}
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
		if errors.Is(err, repository.ErrUserNotFound) {
			return ErrUserNotFound
		}
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
		if errors.Is(err, repository.ErrUserNotFound) {
			return ErrUserNotFound
		}
		return err
	}

	oldRole := userToUpdate.Role
	userToUpdate.Role = role
	err = u.repo.UpdateUser(ctx, userToUpdate) // This will use the updated UpdateUser in repository
	if err != nil {
		u.logger.Error("Admin failed to update user role in repository", zap.String("adminID", admin.ID.Hex()), zap.String("targetUserID", userIDHex), zap.String("newRole", role), zap.Error(err))
		if errors.Is(err, repository.ErrUserNotFound) {
			return ErrUserNotFound
		}
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
			return ErrUserNotFound
		}
		u.logger.Error("Error fetching target user for AdminSetUserActiveStatus", zap.String("targetUserID", userIDHex), zap.Error(err))
		return err
	}

	if targetUser.IsActive == isActive {
		u.logger.Info("AdminSetUserActiveStatus: No change needed for user", zap.String("targetUserID", userIDHex), zap.Bool("isActive", isActive))
		return nil
	}
	targetUser.IsActive = isActive

	if err := u.repo.UpdateUser(ctx, targetUser); err != nil { // This will use the updated UpdateUser in repository
		u.logger.Error("Failed to update user active status in repo by admin", zap.String("adminID", admin.ID.Hex()), zap.String("targetUserID", targetUser.ID.Hex()), zap.Error(err))
		if errors.Is(err, repository.ErrUserNotFound) {
			return ErrUserNotFound
		}
		return errors.New("failed to update user active status")
	}
	u.logger.Info("Admin successfully set user active status", zap.String("adminID", admin.ID.Hex()), zap.String("targetUserID", targetUser.ID.Hex()), zap.Bool("newStatus", isActive))

	if !isActive {
		if err := u.repo.InvalidateToken(ctx, userIDHex); err != nil {
			u.logger.Warn("Failed to invalidate token during admin deactivation", zap.String("targetUserID", userIDHex), zap.Error(err))
		} else {
			u.logger.Info("Token invalidated for admin-deactivated user", zap.String("targetUserID", userIDHex))
		}
	}
	return nil
}
