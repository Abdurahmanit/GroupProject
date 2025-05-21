package usecase

import (
	"context"
	"errors"
	"fmt"

	"github.com/Abdurahmanit/GroupProject/news-service/internal/port/repository"
)

const (
	ContentTypeNews    = "news"
	ContentTypeComment = "comment"
)

type LikeUseCase struct {
	likeRepo    repository.LikeRepository
	newsRepo    repository.NewsRepository
	commentRepo repository.CommentRepository
}

func NewLikeUseCase(lr repository.LikeRepository, nr repository.NewsRepository, cr repository.CommentRepository) *LikeUseCase {
	return &LikeUseCase{
		likeRepo:    lr,
		newsRepo:    nr,
		commentRepo: cr,
	}
}

func (uc *LikeUseCase) validateContentExists(ctx context.Context, contentType string, contentID string) error {
	var err error
	switch contentType {
	case ContentTypeNews:
		_, err = uc.newsRepo.GetByID(ctx, contentID)
	case ContentTypeComment:
		_, err = uc.commentRepo.GetByID(ctx, contentID)
	default:
		return fmt.Errorf("unknown content type: %s", contentType)
	}

	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return fmt.Errorf("%s with id %s not found: %w", contentType, contentID, err)
		}
		return fmt.Errorf("failed to check %s existence: %w", contentType, err)
	}
	return nil
}

type AddLikeInput struct {
	ContentType string
	ContentID   string
	UserID      string
}

func (uc *LikeUseCase) AddLike(ctx context.Context, input AddLikeInput) error {
	if err := uc.validateContentExists(ctx, input.ContentType, input.ContentID); err != nil {
		return err
	}

	err := uc.likeRepo.AddLike(ctx, input.ContentType, input.ContentID, input.UserID)
	if err != nil {
		return fmt.Errorf("failed to add like: %w", err)
	}
	return nil
}

type RemoveLikeInput struct {
	ContentType string
	ContentID   string
	UserID      string
}

func (uc *LikeUseCase) RemoveLike(ctx context.Context, input RemoveLikeInput) error {
	if err := uc.validateContentExists(ctx, input.ContentType, input.ContentID); err != nil {
		return err
	}

	err := uc.likeRepo.RemoveLike(ctx, input.ContentType, input.ContentID, input.UserID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil
		}
		return fmt.Errorf("failed to remove like: %w", err)
	}
	return nil
}

type GetLikesCountInput struct {
	ContentType string
	ContentID   string
}

func (uc *LikeUseCase) GetLikesCount(ctx context.Context, input GetLikesCountInput) (int64, error) {
	if err := uc.validateContentExists(ctx, input.ContentType, input.ContentID); err != nil {
		return 0, err
	}

	count, err := uc.likeRepo.GetLikesCount(ctx, input.ContentType, input.ContentID)
	if err != nil {
		return 0, fmt.Errorf("failed to get likes count: %w", err)
	}
	return count, nil
}

type HasLikedInput struct {
	ContentType string
	ContentID   string
	UserID      string
}

func (uc *LikeUseCase) HasLiked(ctx context.Context, input HasLikedInput) (bool, error) {
	if err := uc.validateContentExists(ctx, input.ContentType, input.ContentID); err != nil {
		return false, err
	}

	liked, err := uc.likeRepo.HasLiked(ctx, input.ContentType, input.ContentID, input.UserID)
	if err != nil {
		return false, fmt.Errorf("failed to check if content was liked: %w", err)
	}
	return liked, nil
}
