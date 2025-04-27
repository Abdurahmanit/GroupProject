package usecase

import (
	"context"
	"errors"
	"time"

	"github.com/Abdurahmanit/GroupProject/user-service/internal/entity"
	"github.com/Abdurahmanit/GroupProject/user-service/internal/repository"
	"github.com/gofrs/uuid"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrUnauthorized       = errors.New("unauthorized")
)

type UserUsecase struct {
	repo      *repository.UserRepository
	jwtSecret string
}

func NewUserUsecase(repo *repository.UserRepository, jwtSecret string) *UserUsecase {
	return &UserUsecase{
		repo:      repo,
		jwtSecret: jwtSecret,
	}
}

func (u *UserUsecase) Register(ctx context.Context, username, email, password string) (string, error) {
	userID, err := uuid.NewV4()
	if err != nil {
		return "", err
	}

	user := &entity.User{
		ID:              userID.String(),
		Username:        username,
		Email:           email,
		Password:        password, // Will be hashed in the repository
		Role:            "customer",
		IsEmailVerified: false,
		IsActive:        true,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	err = u.repo.CreateUser(ctx, user)
	if err != nil {
		return "", err
	}

	return user.ID, nil
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
		return "", ErrUnauthorized
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return "", ErrInvalidCredentials
	}

	token, err := u.repo.GetToken(ctx, user.ID)
	if err != nil {
		return "", err
	}
	if token != "" {
		return token, nil
	}

	// Generate a new token (simplified for this example; in production, use JWT)
	token = uuid.Must(uuid.NewV4()).String()
	err = u.repo.CacheToken(ctx, user.ID, token)
	if err != nil {
		return "", err
	}

	return token, nil
}

func (u *UserUsecase) Logout(ctx context.Context, userID string) error {
	return u.repo.InvalidateToken(ctx, userID)
}

func (u *UserUsecase) GetProfile(ctx context.Context, userID string) (*entity.User, error) {
	user, err := u.repo.GetUserByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if !user.IsActive {
		return nil, ErrUnauthorized
	}
	return user, nil
}

func (u *UserUsecase) UpdateProfile(ctx context.Context, userID, username, email string) error {
	user, err := u.repo.GetUserByID(ctx, userID)
	if err != nil {
		return err
	}
	if !user.IsActive {
		return ErrUnauthorized
	}

	user.Username = username
	user.Email = email
	user.UpdatedAt = time.Now()

	return u.repo.UpdateUser(ctx, user)
}

func (u *UserUsecase) ChangePassword(ctx context.Context, userID, oldPassword, newPassword string) error {
	user, err := u.repo.GetUserByID(ctx, userID)
	if err != nil {
		return err
	}
	if !user.IsActive {
		return ErrUnauthorized
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(oldPassword))
	if err != nil {
		return ErrInvalidCredentials
	}

	return u.repo.UpdatePassword(ctx, userID, newPassword)
}

func (u *UserUsecase) VerifyEmail(ctx context.Context, userID string) error {
	user, err := u.repo.GetUserByID(ctx, userID)
	if err != nil {
		return err
	}
	if !user.IsActive {
		return ErrUnauthorized
	}

	user.IsEmailVerified = true
	user.UpdatedAt = time.Now()

	return u.repo.UpdateUser(ctx, user)
}

func (u *UserUsecase) DeleteUser(ctx context.Context, userID string) error {
	user, err := u.repo.GetUserByID(ctx, userID)
	if err != nil {
		return err
	}
	if !user.IsActive {
		return ErrUnauthorized
	}

	return u.repo.DeleteUser(ctx, userID)
}

func (u *UserUsecase) AdminDeleteUser(ctx context.Context, adminID, userID string) error {
	admin, err := u.repo.GetUserByID(ctx, adminID)
	if err != nil {
		return err
	}
	if admin.Role != "admin" {
		return ErrUnauthorized
	}

	return u.repo.HardDeleteUser(ctx, userID)
}

func (u *UserUsecase) AdminListUsers(ctx context.Context, adminID string, skip, limit int64) ([]*entity.User, error) {
	admin, err := u.repo.GetUserByID(ctx, adminID)
	if err != nil {
		return nil, err
	}
	if admin.Role != "admin" {
		return nil, ErrUnauthorized
	}

	return u.repo.ListUsers(ctx, skip, limit)
}

func (u *UserUsecase) AdminSearchUsers(ctx context.Context, adminID, query string, skip, limit int64) ([]*entity.User, error) {
	admin, err := u.repo.GetUserByID(ctx, adminID)
	if err != nil {
		return nil, err
	}
	if admin.Role != "admin" {
		return nil, ErrUnauthorized
	}

	return u.repo.SearchUsers(ctx, query, skip, limit)
}

func (u *UserUsecase) AdminUpdateUserRole(ctx context.Context, adminID, userID, role string) error {
	admin, err := u.repo.GetUserByID(ctx, adminID)
	if err != nil {
		return err
	}
	if admin.Role != "admin" {
		return ErrUnauthorized
	}

	user, err := u.repo.GetUserByID(ctx, userID)
	if err != nil {
		return err
	}
	if !user.IsActive {
		return ErrUnauthorized
	}

	user.Role = role
	user.UpdatedAt = time.Now()

	return u.repo.UpdateUser(ctx, user)
}
