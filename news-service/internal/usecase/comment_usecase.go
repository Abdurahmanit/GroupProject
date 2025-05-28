package usecase

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Abdurahmanit/GroupProject/news-service/internal/entity"
	"github.com/Abdurahmanit/GroupProject/news-service/internal/port/repository"
)

type CommentUseCase struct {
	commentRepo repository.CommentRepository
	newsRepo    repository.NewsRepository
}

func NewCommentUseCase(cr repository.CommentRepository, nr repository.NewsRepository) *CommentUseCase {
	return &CommentUseCase{
		commentRepo: cr,
		newsRepo:    nr,
	}
}

type CreateCommentInput struct {
	NewsID  string
	UserID  string
	Content string
}

func (uc *CommentUseCase) CreateComment(ctx context.Context, input CreateCommentInput) (*entity.Comment, error) {
	_, err := uc.newsRepo.GetByID(ctx, input.NewsID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, fmt.Errorf("news with id %s not found: %w", input.NewsID, err)
		}
		return nil, fmt.Errorf("failed to check news existence: %w", err)
	}

	now := time.Now()
	comment := &entity.Comment{
		NewsID:    input.NewsID,
		UserID:    input.UserID,
		Content:   input.Content,
		CreatedAt: now,
		UpdatedAt: now,
	}

	createdID, err := uc.commentRepo.Create(ctx, comment)
	if err != nil {
		return nil, fmt.Errorf("failed to create comment: %w", err)
	}
	comment.ID = createdID

	return comment, nil
}

type ListCommentsInput struct {
	NewsID   string
	Page     int
	PageSize int
}

type ListCommentsOutput struct {
	Comments   []*entity.Comment
	TotalCount int
}

func (uc *CommentUseCase) GetCommentsByNewsID(ctx context.Context, input ListCommentsInput) (*ListCommentsOutput, error) {
	if input.Page <= 0 {
		input.Page = 1
	}
	if input.PageSize <= 0 {
		input.PageSize = 10
	}

	comments, total, err := uc.commentRepo.GetByNewsID(ctx, input.NewsID, input.Page, input.PageSize)
	if err != nil {
		return nil, fmt.Errorf("failed to get comments by news id: %w", err)
	}

	return &ListCommentsOutput{Comments: comments, TotalCount: total}, nil
}

type DeleteCommentInput struct {
	CommentID string
	UserID    string
}

func (uc *CommentUseCase) DeleteComment(ctx context.Context, input DeleteCommentInput) error {
	_, err := uc.commentRepo.GetByID(ctx, input.CommentID) // Используем _ для comment
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return repository.ErrNotFound
		}
		return fmt.Errorf("failed to get comment for deletion check: %w", err)
	}

	err = uc.commentRepo.Delete(ctx, input.CommentID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return repository.ErrNotFound
		}
		return fmt.Errorf("failed to delete comment: %w", err)
	}
	return nil
}
