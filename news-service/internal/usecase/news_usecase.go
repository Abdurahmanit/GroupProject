package usecase

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/Abdurahmanit/GroupProject/news-service/internal/entity"
	"github.com/Abdurahmanit/GroupProject/news-service/internal/port/cache"
	"github.com/Abdurahmanit/GroupProject/news-service/internal/port/repository"
	"go.uber.org/zap"
)

type NATSPublisherInterface interface {
	PublishNewsCreated(ctx context.Context, news *entity.News) error
	PublishNewsUpdated(ctx context.Context, news *entity.News) error
	PublishNewsDeleted(ctx context.Context, newsID string) error
}

type NewsUseCase struct {
	newsRepo      repository.NewsRepository
	natsPublisher NATSPublisherInterface
	cacheRepo     cache.CacheRepository
	logger        *zap.Logger
}

func NewNewsUseCase(
	nr repository.NewsRepository,
	np NATSPublisherInterface,
	cr cache.CacheRepository,
	log *zap.Logger,
) *NewsUseCase {
	return &NewsUseCase{
		newsRepo:      nr,
		natsPublisher: np,
		cacheRepo:     cr,
		logger:        log,
	}
}

func newsCacheKey(newsID string) string {
	return fmt.Sprintf("news:%s", newsID)
}

const newsCacheTTL = 5 * time.Minute

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
		uc.logger.Error("Failed to create news in repository", zap.Error(err), zap.Any("input", input))
		return nil, fmt.Errorf("NewsUseCase.CreateNews: failed to create news in repo: %w", err)
	}
	news.ID = createdID

	if uc.cacheRepo != nil {
		newsBytes, marshalErrLocal := json.Marshal(news)
		if marshalErrLocal != nil {
			uc.logger.Warn("Failed to marshal news for caching after create",
				zap.Error(marshalErrLocal),
				zap.String("news_id", news.ID),
			)
		} else {
			key := newsCacheKey(news.ID)
			if setErr := uc.cacheRepo.Set(ctx, key, newsBytes, newsCacheTTL); setErr != nil {
				uc.logger.Warn("Failed to set news in cache after create",
					zap.Error(setErr),
					zap.String("key", key),
				)
			}
		}
	}

	if uc.natsPublisher != nil {
		if errPub := uc.natsPublisher.PublishNewsCreated(ctx, news); errPub != nil {
			uc.logger.Warn("Failed to publish NATS event for news created",
				zap.Error(errPub),
				zap.String("news_id", news.ID),
			)
		}
	}

	return news, nil
}

func (uc *NewsUseCase) GetNewsByID(ctx context.Context, id string) (*entity.News, error) {
	if uc.cacheRepo != nil {
		key := newsCacheKey(id)
		cachedBytes, err := uc.cacheRepo.Get(ctx, key)

		if err == nil {
			var newsFromCache entity.News
			var unmarshalErrLocal error

			unmarshalErrLocal = json.Unmarshal(cachedBytes, &newsFromCache)
			if unmarshalErrLocal == nil {
				uc.logger.Debug("News fetched from cache", zap.String("key", key))
				return &newsFromCache, nil
			}
			uc.logger.Error("Failed to unmarshal news from cache", zap.Error(unmarshalErrLocal), zap.String("key", key))
			if delErr := uc.cacheRepo.Delete(ctx, key); delErr != nil {
				uc.logger.Warn("Failed to delete corrupted data from cache", zap.String("key", key), zap.Error(delErr))
			}
		} else if !errors.Is(err, cache.ErrNotFound) {
			uc.logger.Warn("Failed to get news from cache (not a cache miss)", zap.Error(err), zap.String("key", key))
		}
	}

	uc.logger.Debug("News not found in cache or cache error, fetching from repository", zap.String("news_id", id))
	news, err := uc.newsRepo.GetByID(ctx, id)
	if err != nil {
		if !errors.Is(err, repository.ErrNotFound) {
			uc.logger.Error("Failed to get news by ID from repository", zap.Error(err), zap.String("news_id", id))
		}
		return nil, fmt.Errorf("NewsUseCase.GetNewsByID: failed to get news from repo: %w", err)
	}

	if uc.cacheRepo != nil && news != nil {
		newsBytes, marshalErrLocal := json.Marshal(news)
		if marshalErrLocal != nil {
			uc.logger.Warn("Failed to marshal news for caching after fetching from repo",
				zap.Error(marshalErrLocal),
				zap.String("news_id", news.ID),
			)
		} else {
			key := newsCacheKey(news.ID)
			if setErr := uc.cacheRepo.Set(ctx, key, newsBytes, newsCacheTTL); setErr != nil {
				uc.logger.Warn("Failed to set news in cache after fetching from repo",
					zap.Error(setErr),
					zap.String("key", key),
				)
			} else {
				uc.logger.Debug("News set to cache after fetching from repository", zap.String("key", key))
			}
		}
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
		if !errors.Is(err, repository.ErrNotFound) {
			uc.logger.Error("Failed to get news for update from repository", zap.Error(err), zap.String("news_id", input.ID))
		}
		return nil, fmt.Errorf("NewsUseCase.UpdateNews: failed to get news for update: %w", err)
	}

	updated := false
	if input.Title != nil && news.Title != *input.Title {
		news.Title = *input.Title
		updated = true
	}
	if input.Content != nil && news.Content != *input.Content {
		news.Content = *input.Content
		updated = true
	}

	if !updated {
		uc.logger.Info("No actual changes detected for news update", zap.String("news_id", input.ID))
		return news, nil
	}

	news.UpdatedAt = time.Now()

	err = uc.newsRepo.Update(ctx, news)
	if err != nil {
		uc.logger.Error("Failed to update news in repository", zap.Error(err), zap.String("news_id", news.ID))
		return nil, fmt.Errorf("NewsUseCase.UpdateNews: failed to update news in repo: %w", err)
	}

	if uc.cacheRepo != nil {
		key := newsCacheKey(news.ID)
		if delErr := uc.cacheRepo.Delete(ctx, key); delErr != nil {
			uc.logger.Warn("Failed to delete news from cache after update",
				zap.Error(delErr),
				zap.String("key", key),
			)
		} else {
			uc.logger.Debug("News deleted from cache after update", zap.String("key", key))
		}
	}

	if uc.natsPublisher != nil {
		if errPub := uc.natsPublisher.PublishNewsUpdated(ctx, news); errPub != nil {
			uc.logger.Warn("Failed to publish NATS event for news updated",
				zap.Error(errPub),
				zap.String("news_id", news.ID),
			)
		}
	}

	return news, nil
}

func (uc *NewsUseCase) DeleteNews(ctx context.Context, id string) error {
	_, err := uc.GetNewsByID(ctx, id)
	if err != nil {
		return fmt.Errorf("NewsUseCase.DeleteNews: news to delete not found or error getting it: %w", err)
	}

	err = uc.newsRepo.Delete(ctx, id)
	if err != nil {
		if !errors.Is(err, repository.ErrNotFound) {
			uc.logger.Error("Failed to delete news from repository", zap.Error(err), zap.String("news_id", id))
		}
		return fmt.Errorf("NewsUseCase.DeleteNews: failed to delete news from repo: %w", err)
	}

	if uc.cacheRepo != nil {
		key := newsCacheKey(id)
		if delErr := uc.cacheRepo.Delete(ctx, key); delErr != nil {
			uc.logger.Warn("Failed to delete news from cache after delete operation",
				zap.Error(delErr),
				zap.String("key", key),
			)
		} else {
			uc.logger.Debug("News deleted from cache after delete operation", zap.String("key", key))
		}
	}

	if uc.natsPublisher != nil {
		if errPub := uc.natsPublisher.PublishNewsDeleted(ctx, id); errPub != nil {
			uc.logger.Warn("Failed to publish NATS event for news deleted",
				zap.Error(errPub),
				zap.String("news_id", id),
			)
		}
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
		uc.logger.Error("Failed to list news from repository", zap.Error(err), zap.Any("input", input))
		return nil, fmt.Errorf("NewsUseCase.ListNews: failed to list news from repo: %w", err)
	}

	return &ListNewsOutput{News: newsList, TotalCount: total}, nil
}
