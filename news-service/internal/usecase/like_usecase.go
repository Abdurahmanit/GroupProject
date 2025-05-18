package usecase

import (
	"github.com/Abdurahmanit/GroupProject/news-service/internal/port/repository"
)

type LikeUseCase struct {
	likeRepo repository.LikeRepository
	newsRepo repository.NewsRepository
}

func NewLikeUseCase(lr repository.LikeRepository, nr repository.NewsRepository) *LikeUseCase {
	return &LikeUseCase{
		likeRepo: lr,
		newsRepo: nr,
	}
}
