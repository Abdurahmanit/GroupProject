package usecase

import (
	"context"
	"errors"

	"github.com/Abdurahmanit/GroupProject/user-service/internal/entity"
	"github.com/Abdurahmanit/GroupProject/user-service/internal/jwt"
	"github.com/gofrs/uuid"
	"golang.org/x/crypto/bcrypt"
)

type UserRepository interface {
	CreateUser(ctx context.Context, user *entity.User) error
	GetUserByEmail(ctx context.Context, email string) (*entity.User, error)
	GetUserByID(ctx context.Context, userID string) (*entity.User, error)
	CacheToken(ctx context.Context, userID, token string) error
	InvalidateToken(ctx context.Context, userID string) error
	GetToken(ctx context.Context, userID string) (string, error)
}

type UserUsecase struct {
	repo      UserRepository
	jwtSecret string
}

func NewUserUsecase(repo UserRepository, jwtSecret string) *UserUsecase {
	return &UserUsecase{
		repo:      repo,
		jwtSecret: jwtSecret,
	}
}

func (u *UserUsecase) Register(ctx context.Context, username, email, password string) (string, error) {
	// Check if user exists
	if _, err := u.repo.GetUserByEmail(ctx, email); err == nil {
		return "", errors.New("user already exists")
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}

	// Generate user ID
	userID, err := uuid.NewV4()
	if err != nil {
		return "", err
	}

	user := &entity.User{
		ID:       userID.String(),
		Username: username,
		Email:    email,
		Password: string(hashedPassword),
	}

	if err := u.repo.CreateUser(ctx, user); err != nil {
		return "", err
	}

	return user.ID, nil
}

func (u *UserUsecase) Login(ctx context.Context, email, password string) (string, error) {
	user, err := u.repo.GetUserByEmail(ctx, email)
	if err != nil {
		return "", errors.New("user not found")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return "", errors.New("invalid password")
	}

	token, err := jwt.GenerateToken(user.ID, u.jwtSecret)
	if err != nil {
		return "", err
	}

	if err := u.repo.CacheToken(ctx, user.ID, token); err != nil {
		return "", err
	}

	return token, nil
}

func (u *UserUsecase) Logout(ctx context.Context, userID string) error {
	return u.repo.InvalidateToken(ctx, userID)
}

func (u *UserUsecase) GetProfile(ctx context.Context, userID string) (*entity.User, error) {
	// Verify token exists in cache
	token, err := u.repo.GetToken(ctx, userID)
	if err != nil || token == "" {
		return nil, errors.New("user not logged in")
	}

	user, err := u.repo.GetUserByID(ctx, userID)
	if err != nil {
		return nil, errors.New("user not found")
	}

	return user, nil
}
