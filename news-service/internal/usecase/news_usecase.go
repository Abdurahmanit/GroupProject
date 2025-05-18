package usecase

import (
	"context"
	"time"

	"github.com/Abdurahmanit/GroupProject/news-service/internal/entity"
	"github.com/Abdurahmanit/GroupProject/news-service/internal/port/repository"
)

type NewsUseCase struct {
	newsRepo repository.NewsRepository
}

func NewNewsUseCase(nr repository.NewsRepository) *NewsUseCase {
	return &NewsUseCase{
		newsRepo: nr,
	}
}

type CreateNewsInput struct {
	Title    string
	Content  string
	AuthorID string
}

func (uc *NewsUseCase) CreateNews(ctx context.Context, input CreateNewsInput) (*entity.News, error) {
	now := time.Now()
	news := &entity.News{
		Title:     input.Title,
		Content:   input.Content,
		AuthorID:  input.AuthorID,
		CreatedAt: now,
		UpdatedAt: now,
	}

	createdID, err := uc.newsRepo.Create(ctx, news)
	if err != nil {
		return nil, err
	}
	news.ID = createdID

	return news, nil
}

func (uc *NewsUseCase) GetNewsByID(ctx context.Context, id string) (*entity.News, error) {
	news, err := uc.newsRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return news, nil
}

type UpdateNewsInput struct {
	ID      string
	Title   *string
	Content *string
}

func (uc *NewsUseCase) UpdateNews(ctx context.Context, input UpdateNewsInput) (*entity.News, error) {
	news, err := uc.newsRepo.GetByID(ctx, input.ID)
	if err != nil {
		return nil, err
	}

	if input.Title != nil {
		news.Title = *input.Title
	}
	if input.Content != nil {
		news.Content = *input.Content
	}
	news.UpdatedAt = time.Now()

	err = uc.newsRepo.Update(ctx, news)
	if err != nil {
		return nil, err
	}

	return news, nil
}

func (uc *NewsUseCase) DeleteNews(ctx context.Context, id string) error {
	err := uc.newsRepo.Delete(ctx, id)
	if err != nil {
		return err
	}
	return nil
}

type ListNewsInput struct {
	Page     int
	PageSize int
	Filter   map[string]interface{}
}

type ListNewsOutput struct {
	News       []*entity.News
	TotalCount int
}

func (uc *NewsUseCase) ListNews(ctx context.Context, input ListNewsInput) (*ListNewsOutput, error) {
	if input.Page <= 0 {
		input.Page = 1
	}
	if input.PageSize <= 0 {
		input.PageSize = 10
	}

	newsList, total, err := uc.newsRepo.List(ctx, input.Page, input.PageSize, input.Filter)
	if err != nil {
		return nil, err
	}

	return &ListNewsOutput{News: newsList, TotalCount: total}, nil
}
