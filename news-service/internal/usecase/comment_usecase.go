package usecase

import (
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
